package fakeproto_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"

	"github.com/psyhatter/fakeproto"
	testpb "github.com/psyhatter/fakeproto/internal/test/gen"
)

func TestOptionalAlwaysSet(t *testing.T) {
	t.Parallel()

	// With probability=0 every optional field must be populated.
	msg := &testpb.OptionalFields{}
	if err := fakeproto.Message(msg, fakeproto.WithOptionalProbability(0)); err != nil {
		t.Fatal(err)
	}
	if msg.MaybeCount == nil {
		t.Error("MaybeCount should be set when optionalProbability=0")
	}
	if msg.MaybeName == nil {
		t.Error("MaybeName should be set when optionalProbability=0")
	}
	if msg.MaybeFlag == nil {
		t.Error("MaybeFlag should be set when optionalProbability=0")
	}
}

func TestOptionalAlwaysUnset(t *testing.T) {
	t.Parallel()

	// With probability=1 every optional field must be left nil.
	msg := &testpb.OptionalFields{}
	if err := fakeproto.Message(msg, fakeproto.WithOptionalProbability(1)); err != nil {
		t.Fatal(err)
	}
	if msg.MaybeCount != nil {
		t.Error("MaybeCount should be nil when optionalProbability=1")
	}
	if msg.MaybeName != nil {
		t.Error("MaybeName should be nil when optionalProbability=1")
	}
	if msg.MaybeFlag != nil {
		t.Error("MaybeFlag should be nil when optionalProbability=1")
	}
}

func BenchmarkOptional(b *testing.B) {
	opt := fakeproto.WithFaker(gofakeit.New(0))
	b.ResetTimer()
	for range b.N {
		msg := &testpb.OptionalFields{}
		_ = fakeproto.Message(msg, opt)
	}
}

func FuzzOptional(f *testing.F) {
	f.Fuzz(func(t *testing.T, seed uint64) {
		t.Parallel()
		msg := &testpb.OptionalFields{}
		if err := fakeproto.Message(msg, fakeproto.WithFaker(gofakeit.New(seed))); err != nil {
			t.Fatal(err)
		}
	})
}
