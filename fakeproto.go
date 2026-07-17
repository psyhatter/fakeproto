package fakeproto

import (
	"github.com/brianvoe/gofakeit/v7"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Option configures the behavior of [Message] and [Enum].
type Option func(*config)

// config holds resolved settings for a single fill call.
type config struct {
	faker               *gofakeit.Faker
	maxDepth            int
	maxRepeated         int
	optionalProbability float64 // probability of leaving an optional field unset
	skipOneof           bool
}

func defaultConfig() config {
	return config{
		maxDepth:            5,
		maxRepeated:         5,
		optionalProbability: 0.10,
	}
}

// WithFaker sets the [gofakeit.Faker] instance used for random data generation.
// A seeded faker ([gofakeit.New]) produces deterministic output, which is
// useful for snapshot tests and godoc examples. Defaults to [gofakeit.GlobalFaker].
func WithFaker(f *gofakeit.Faker) Option {
	return func(c *config) { c.faker = f }
}

// WithMaxDepth sets the maximum JSON-container nesting depth of the generated
// output. Each message, array, and map object counts as one container level.
// Containers that would exceed d are left empty or nil. Default is 5.
func WithMaxDepth(d int) Option {
	return func(c *config) { c.maxDepth = d }
}

// WithMaxRepeated sets the maximum length of generated repeated fields and map
// entries. The actual length is chosen randomly in [1, n]. Default is 5.
func WithMaxRepeated(n int) Option {
	return func(c *config) { c.maxRepeated = n }
}

// WithOptionalProbability sets the probability p in [0, 1] that an explicit
// proto3 optional scalar field is left unset. p=0 means always set; p=1 means
// always unset. Default is 0.10.
func WithOptionalProbability(p float64) Option {
	return func(c *config) { c.optionalProbability = p }
}

// WithSkipOneof instructs the generator to leave all oneof fields unset.
func WithSkipOneof() Option {
	return func(c *config) { c.skipOneof = true }
}

// Message fills msg with random fake data.
//
// Use [WithFaker] to supply a seeded [gofakeit.Faker] for deterministic output:
//
//	f := gofakeit.New(42)
//	err := fakeproto.Message(msg, fakeproto.WithFaker(f))
func Message(msg proto.Message, opts ...Option) error {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}
	f := cfg.faker
	if f == nil {
		f = gofakeit.GlobalFaker
	}
	fi := &filler{f: f, cfg: cfg}
	return fi.message(msg.ProtoReflect(), 1)
}

// ProtoEnum is satisfied by every protobuf-generated enum type. All such types
// have int32 as their underlying type.
type ProtoEnum interface {
	protoreflect.Enum
	~int32
}

// Enum sets *e to a random value chosen from the enum's declared set.
// Reserved numbers are never produced.
//
//	var p pb.Priority
//	fakeproto.Enum(&p)
func Enum[E ProtoEnum](e *E, opts ...Option) {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}
	f := cfg.faker
	if f == nil {
		f = gofakeit.GlobalFaker
	}
	var zero E
	vals := zero.Descriptor().Values()
	*e = E(vals.Get(f.IntN(vals.Len())).Number())
}
