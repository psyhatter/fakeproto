package fakeproto_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"

	"github.com/psyhatter/fakeproto"
	testpb "github.com/psyhatter/fakeproto/internal/test/gen"
)

func TestGroupKind(t *testing.T) {
	t.Parallel()

	msg := &testpb.WithGroup{}
	if err := fakeproto.Message(msg, fakeproto.WithOptionalProbability(0)); err != nil {
		t.Fatal(err)
	}

	inner := msg.GetInner()
	if inner == nil {
		t.Fatal("Inner group is nil")
	}
	if inner.GetName() == "" {
		t.Error("Inner.Name is empty")
	}
	if inner.GetValue() == 0 {
		t.Error("Inner.Value is zero")
	}
}

func BenchmarkGroupKind(b *testing.B) {
	opt := fakeproto.WithFaker(gofakeit.New(0))
	b.ResetTimer()
	for range b.N {
		msg := &testpb.WithGroup{}
		_ = fakeproto.Message(msg, opt)
	}
}

func FuzzGroupKind(f *testing.F) {
	f.Fuzz(func(t *testing.T, seed uint64) {
		t.Parallel()
		msg := &testpb.WithGroup{}
		if err := fakeproto.Message(msg, fakeproto.WithFaker(gofakeit.New(seed))); err != nil {
			t.Fatal(err)
		}
	})
}
