package fakeproto

import (
	"fmt"
	"time"
	_ "unsafe" // For go:linkname,

	"github.com/brianvoe/gofakeit/v7"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// filler carries the faker and resolved config for one fill call.
type filler struct {
	f   *gofakeit.Faker
	cfg config
}

// message fills m at the given depth. The depth argument is already the
// effective depth of this message: value() increments it before calling here,
// so message() itself never needs to touch the counter.
//
// The guard fires only for the root invocation: [Message] starts at depth=1, so
// WithMaxDepth(0) produces an empty message because 1 > 0. In all other call
// paths value() pre-checks the limit and never calls message when the limit
// would be exceeded.
func (fi *filler) message(m protoreflect.Message, depth int) error {
	if depth > fi.cfg.maxDepth {
		return nil
	}

	if ok, err := fi.wellKnown(m, depth); ok || err != nil {
		return err
	}

	fields := m.Descriptor().Fields()
	for i := range fields.Len() {
		fd := fields.Get(i)
		// Members of real oneofs are handled in the oneof loop below. Proto3 optional
		// fields live in synthetic oneofs and go through the normal walk.
		if od := fd.ContainingOneof(); od != nil && !od.IsSynthetic() {
			continue
		}
		if err := fi.field(m, fd, depth); err != nil {
			return err
		}
	}

	if !fi.cfg.skipOneof {
		oneofs := m.Descriptor().Oneofs()
		for i := range oneofs.Len() {
			od := oneofs.Get(i)
			if od.IsSynthetic() {
				continue
			}
			fields := od.Fields()
			fd := fields.Get(fi.f.IntN(fields.Len()))
			if err := fi.field(m, fd, depth); err != nil {
				return err
			}
		}
	}

	return nil
}

func (fi *filler) field(m protoreflect.Message, fd protoreflect.FieldDescriptor, depth int) error {
	// Explicit presence (proto3 optional scalar): leave unset with configured
	// probability. Members of real oneofs also report presence but are not optional
	// - the caller has already committed to setting this variant.
	if od := fd.ContainingOneof(); fd.HasPresence() && !isMessage(fd) &&
		(od == nil || od.IsSynthetic()) &&
		fi.f.Float64Range(0, 1) < fi.cfg.optionalProbability {
		return nil
	}

	switch {
	case fd.IsMap():
		// A map adds one JSON container level. Skip if there is no room even for the
		// map itself.
		if fi.cfg.maxRepeated <= 0 || depth >= fi.cfg.maxDepth {
			return nil
		}

		// Skip if there is room for the map but not for a heavy (non-well-known message
		// considered as a scalar) message value inside it
		mapValue := fd.MapValue()
		if depth+1 >= fi.cfg.maxDepth && isHeavyMessage(mapValue) {
			return nil
		}

		mp := m.Mutable(fd).Map()
		for range fi.f.IntRange(1, fi.cfg.maxRepeated) {
			// Map keys are always scalars in proto3, so skip value() overhead.
			k, err := fi.scalar(fd.MapKey())
			if err != nil {
				return err
			}
			// depth+1: the map object occupies one nesting level, so values inside it are
			// one level deeper.
			v, err := fi.value(mapValue, mp.NewValue, depth+1)
			if err != nil {
				return err
			}
			mp.Set(k.MapKey(), v)
		}
	case fd.IsList():
		// A list adds one JSON container level. Same budget logic as maps.
		if fi.cfg.maxRepeated <= 0 ||
			depth >= fi.cfg.maxDepth {
			return nil
		}

		if depth+1 >= fi.cfg.maxDepth && isHeavyMessage(fd) {
			return nil
		}

		list := m.Mutable(fd).List()
		for range fi.f.IntRange(1, fi.cfg.maxRepeated) {
			// depth+1: array elements are one level deeper than the containing object.
			v, err := fi.value(fd, list.NewElement, depth+1)
			if err != nil {
				return err
			}
			list.Append(v)
		}
	default:
		v, err := fi.value(fd, func() protoreflect.Value { return m.NewField(fd) }, depth)
		if err != nil {
			return err
		}
		if v.IsValid() {
			m.Set(fd, v)
		}
	}
	return nil
}

// value produces a value for fd. newMsg allocates an empty message value when
// fd is a message kind. An invalid protoreflect.Value signals that the caller
// should leave the field unset.
//
// Depth accounting lives here: value() increments the counter for regular
// messages before the limit check, then passes the updated counter to
// message(). Well-known types rendered as JSON scalars by protojson (Timestamp,
// Duration, wrappers, ...) are exempt from the increment because they do not
// add a JSON container level.
func (fi *filler) value(
	fd protoreflect.FieldDescriptor,
	newMsg func() protoreflect.Value,
	depth int,
) (v protoreflect.Value, err error) {
	if !isMessage(fd) {
		v, err = fi.scalar(fd)
		return v, err
	}

	// Well-known types rendered as JSON scalars (Timestamp → string, wrappers →
	// number/bool/string, ...) do not add a JSON container level, so they do not
	// consume depth budget. All other messages increment the counter.
	if !isWellKnownScalarMessage(fd.Message()) {
		depth++
	}

	if depth > fi.cfg.maxDepth {
		return protoreflect.Value{}, nil
	}

	v = newMsg()
	return v, fi.message(v.Message(), depth)
}

func (fi *filler) scalar(fd protoreflect.FieldDescriptor) (protoreflect.Value, error) {
	switch fd.Kind() {
	default:
		return protoreflect.Value{}, fmt.Errorf("fakeproto: unsupported field kind %v", fd.Kind())
	case protoreflect.BoolKind:
		return protoreflect.ValueOfBool(fi.f.Bool()), nil
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return protoreflect.ValueOfInt32(num(fd, fi, fi.f.Int32)), nil
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return protoreflect.ValueOfInt64(num(fd, fi, fi.f.Int64)), nil
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return protoreflect.ValueOfUint32(num(fd, fi, fi.f.Uint32)), nil
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return protoreflect.ValueOfUint64(num(fd, fi, fi.f.Uint64)), nil
	case protoreflect.FloatKind:
		return protoreflect.ValueOfFloat32(num(fd, fi, fi.f.Float32)), nil
	case protoreflect.DoubleKind:
		return protoreflect.ValueOfFloat64(num(fd, fi, fi.f.Float64)), nil
	case protoreflect.StringKind:
		return protoreflect.ValueOfString(fi.str(fd)), nil
	case protoreflect.BytesKind:
		return protoreflect.ValueOfBytes([]byte(fi.str(fd))), nil
	case protoreflect.EnumKind:
		vals := fd.Enum().Values()
		return protoreflect.ValueOfEnum(vals.Get(fi.f.IntN(vals.Len())).Number()), nil
	}
}

// str asks gofakeit's function registry for a generator matching the field name
// (email, city, zip, …); unknown names or generators that require mandatory
// params fall back to random text.
func (fi *filler) str(fd protoreflect.FieldDescriptor) string {
	if info := gofakeit.GetFuncLookup(fd.TextName()); info != nil {
		// nil MapParams tells GetField to use each param's Default value. Generators
		// with required params that have no default (e.g. numerify, lexify) return an
		// error here; the fallback below handles that case.
		if v, err := info.Generate(fi.f, nil, info); err == nil {
			if s, ok := v.(string); ok {
				return s
			}
			return fmt.Sprint(v)
		}
	}
	if fi.f.Bool() {
		return fi.f.Word()
	}
	return fi.f.Phrase()
}

type numConstraint interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

// num asks gofakeit's function registry for a generator matching the field name
// and converts the result to T. Falls back to fallback() if the name is unknown
// or the returned value is not a numeric type (e.g. a string from "email").
func num[T numConstraint](fd protoreflect.FieldDescriptor, fi *filler, fallback func() T) T {
	info := gofakeit.GetFuncLookup(fd.TextName())
	if info == nil {
		return fallback()
	}

	// nil MapParams - see comment in filler.str().
	v, err := info.Generate(fi.f, nil, info)
	if err != nil {
		return fallback()
	}

	switch n := v.(type) {
	case int:
		return T(n)
	case int8:
		return T(n)
	case int16:
		return T(n)
	case int32:
		return T(n)
	case int64:
		return T(n)
	case uint:
		return T(n)
	case uint8:
		return T(n)
	case uint16:
		return T(n)
	case uint32:
		return T(n)
	case uint64:
		return T(n)
	case float32:
		return T(n)
	case float64:
		return T(n)
	default:
		return fallback()
	}
}

// wellKnown fills types that need semantic constraints the generic walk cannot
// provide. Returns true when handled; the caller skips the generic walk.
//
// Wrappers (StringValue, Int32Value, …) are intentionally NOT listed here: the
// generic walk fills their inner "value" field correctly.
func (fi *filler) wellKnown(m protoreflect.Message, depth int) (bool, error) {
	switch m.Descriptor().FullName() {
	default:
		return false, nil

	case "google.protobuf.Timestamp":
		// Use a date in the valid Unix timestamp range.
		proto.Merge(m.Interface(), timestamppb.New(fi.f.Date()))

	case "google.protobuf.Duration":
		// Keep duration positive and bounded to one year.
		proto.Merge(m.Interface(), durationpb.New(time.Duration(fi.f.IntRange(1, 86400*365))*time.Second))

	case "google.protobuf.FieldMask":
		// Paths must be valid proto field-name identifiers (no spaces, no dots
		// within a single component). gofakeit.Word() returns a single lowercase
		// word, which is always a valid field-path component.
		n := fi.f.IntRange(1, 3)
		paths := make([]string, n)
		for i := range paths {
			paths[i] = fi.f.Word()
		}
		proto.Merge(m.Interface(), &fieldmaskpb.FieldMask{Paths: paths})

	case "google.protobuf.Any":
		// Pack a random structpb.Value so Any always carries a typed payload. Use
		// Deterministic marshaling so that map fields (e.g. google.protobuf.Struct) are
		// encoded in a stable key order - proto.Equal compares Any.value as raw bytes,
		// so non-deterministic ordering would break equality checks.
		var payload structpb.Value
		if err := fi.message(payload.ProtoReflect(), depth); err != nil {
			return false, fmt.Errorf("%T generation: %w", &payload, err)
		}

		var a anypb.Any
		err := anypb.MarshalFrom(&a, &payload, proto.MarshalOptions{Deterministic: true})
		if err != nil {
			return false, fmt.Errorf("marshaling %T: %w", &payload, err)
		}

		proto.Merge(m.Interface(), &a)
	}
	return true, nil
}

func isMessage(fd protoreflect.FieldDescriptor) bool {
	kind := fd.Kind()
	return kind == protoreflect.MessageKind || kind == protoreflect.GroupKind
}

// isWellKnownScalarMessage reports whether fd's message type is rendered as a
// JSON scalar by protojson rather than a JSON object or array. Such types do
// not add a container nesting level and are exempt from the depth guard.
func isWellKnownScalarMessage(fd protoreflect.MessageDescriptor) bool {
	switch fd.FullName() {
	case "google.protobuf.Timestamp",
		"google.protobuf.Duration",
		"google.protobuf.FieldMask",
		"google.protobuf.BoolValue",
		"google.protobuf.StringValue",
		"google.protobuf.BytesValue",
		"google.protobuf.Int32Value",
		"google.protobuf.Int64Value",
		"google.protobuf.UInt32Value",
		"google.protobuf.UInt64Value",
		"google.protobuf.FloatValue",
		"google.protobuf.DoubleValue":
		return true
	}
	return false
}

// isHeavyMessage reports whether fd is a message type that adds a JSON
// container level. Well-known scalar types (Timestamp rendered as string, etc.)
// are "light" and do not consume depth budget.
//
//go:linkname isHeavyMessage
func isHeavyMessage(fd protoreflect.FieldDescriptor) bool {
	return isMessage(fd) && !isWellKnownScalarMessage(fd.Message())
}
