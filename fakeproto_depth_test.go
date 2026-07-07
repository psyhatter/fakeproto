package fakeproto_test

import (
	"slices"
	"strconv"
	"strings"
	"testing"
	_ "unsafe" // For go:linkname,

	"github.com/brianvoe/gofakeit/v7"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/psyhatter/fakeproto"
	testpb "github.com/psyhatter/fakeproto/internal/test/gen"
)

func TestDepthDefault(t *testing.T) {
	t.Parallel()

	// With the default maxDepth=5 the chain must reach exactly 5 levels.
	msg := &testpb.Node{}
	if err := fakeproto.Message(msg); err != nil {
		t.Fatal(err)
	}
	if d := checkDepth(t, nil, msg.ProtoReflect(), 5); d != 5 {
		t.Errorf("got depth %d, want exactly 5 (default maxDepth=5)", d)
	}
}

func TestDepthCustom(t *testing.T) {
	t.Parallel()

	const limit = 3
	for i := range limit {
		msg := &testpb.Node{}
		if err := fakeproto.Message(msg, fakeproto.WithMaxDepth(i)); err != nil {
			t.Fatal(err)
		}

		nestedness := checkDepth(t, nil, msg.ProtoReflect(), limit)
		if i != nestedness {
			t.Errorf("WithMaxDepth(%d): got JSON depth %d, want %d", i, nestedness, i)
		}
		if i > 0 {
			if msg.GetWellknown() == nil {
				t.Errorf("WithMaxDepth(%d): wellknown field is nil", i)
			}
		}
		if i > 1 {
			if len(msg.GetWellknownList()) == 0 {
				t.Errorf("WithMaxDepth(%d): wellknown_list is empty", i)
			}
		}
	}
}

func TestDepthOne(t *testing.T) {
	t.Parallel()

	// maxDepth=1 means the root is filled but child must be nil.
	msg := &testpb.Node{}
	if err := fakeproto.Message(msg, fakeproto.WithMaxDepth(1)); err != nil {
		t.Fatal(err)
	}
	if msg.GetScalar() == "" {
		t.Error("Node.Scalar should be filled at depth 1")
	}
	if msg.GetChild() != nil {
		t.Error("Node.Child must be nil when maxDepth=1")
	}
}

func BenchmarkDepth(b *testing.B) {
	opt := fakeproto.WithFaker(gofakeit.New(0))
	b.ResetTimer()
	for range b.N {
		msg := &testpb.Node{}
		_ = fakeproto.Message(msg, opt)
	}
}

func FuzzDepth(f *testing.F) {
	f.Fuzz(func(t *testing.T, seed uint64) {
		t.Parallel()
		msg := &testpb.Node{}
		if err := fakeproto.Message(msg, fakeproto.WithFaker(gofakeit.New(seed))); err != nil {
			t.Fatal(err)
		}
		if d := checkDepth(t, nil, msg.ProtoReflect(), 5); d != 5 {
			t.Errorf("got depth %d, want exactly 5 (default maxDepth=5)", d)
		}
	})
}

func checkDepth(t *testing.T, path []string, m protoreflect.Message, maxDepth int) (maxFound int) {
	t.Helper()

	if len(path) >= maxDepth {
		t.Fatalf("$.%s: depth exceeds maxDepth=%d", strings.Join(path, "."), maxDepth)
	}

	m.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		t.Helper()
		fieldPath := append(slices.Clip(path), string(fd.Name()))
		var inner int
		switch {
		case fd.IsMap():
			inner = checkDepthMap(t, fieldPath, fd, v.Map(), maxDepth)
		case fd.IsList():
			inner = checkDepthList(t, fieldPath, fd, v.List(), maxDepth)
		case isHeavyProtoMessage(fd):
			inner = checkDepth(t, fieldPath, v.Message(), maxDepth)
		}
		maxFound = max(maxFound, inner+1)
		return true
	})

	return maxFound
}

func checkDepthList(
	t *testing.T,
	path []string,
	fd protoreflect.FieldDescriptor,
	list protoreflect.List,
	maxDepth int,
) (maxFound int) {
	t.Helper()
	if len(path) >= maxDepth {
		t.Fatalf("$.%s: depth exceeds maxDepth=%d", strings.Join(path, "."), maxDepth)
	}
	for i := range list.Len() {
		elemPath := append(slices.Clip(path), "["+strconv.Itoa(i)+"]")
		var inner int
		if isHeavyProtoMessage(fd) {
			inner = checkDepth(t, elemPath, list.Get(i).Message(), maxDepth)
		}
		maxFound = max(maxFound, inner+1)
	}
	return maxFound
}

func checkDepthMap(
	t *testing.T,
	path []string,
	fd protoreflect.FieldDescriptor,
	mp protoreflect.Map,
	maxDepth int,
) (maxFound int) {
	t.Helper()

	if len(path) >= maxDepth {
		t.Fatalf("$.%s: depth exceeds maxDepth=%d", strings.Join(path, "."), maxDepth)
	}

	valFd := fd.MapValue()
	mp.Range(func(k protoreflect.MapKey, v protoreflect.Value) bool {
		t.Helper()
		keyPath := append(slices.Clip(path), k.String())
		var inner int
		if isHeavyProtoMessage(valFd) {
			inner = checkDepth(t, keyPath, v.Message(), maxDepth)
		}
		maxFound = max(maxFound, inner+1)
		return true
	})

	return maxFound
}

//go:linkname isHeavyProtoMessage github.com/psyhatter/fakeproto.isHeavyMessage
func isHeavyProtoMessage(fd protoreflect.FieldDescriptor) bool
