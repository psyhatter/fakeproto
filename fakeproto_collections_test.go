package fakeproto_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"

	"github.com/psyhatter/fakeproto"
	testpb "github.com/psyhatter/fakeproto/internal/test/gen"
)

func TestCollectionsStringList(t *testing.T) {
	t.Parallel()

	msg := &testpb.StringList{}
	if err := fakeproto.Message(msg); err != nil {
		t.Fatal(err)
	}
	if len(msg.GetItems()) == 0 {
		t.Error("StringList.Items is empty")
	}
	for i, s := range msg.GetItems() {
		if s == "" {
			t.Errorf("StringList.Items[%d] is empty", i)
		}
	}
}

func TestCollectionsUserList(t *testing.T) {
	t.Parallel()

	msg := &testpb.UserList{}
	if err := fakeproto.Message(msg); err != nil {
		t.Fatal(err)
	}
	if len(msg.GetUsers()) == 0 {
		t.Error("UserList.Users is empty")
	}
	for i, u := range msg.GetUsers() {
		if u == nil {
			t.Errorf("UserList.Users[%d] is nil", i)
		}
	}
}

func TestCollectionsStringToIntMap(t *testing.T) {
	t.Parallel()

	msg := &testpb.StringToIntMap{}
	if err := fakeproto.Message(msg); err != nil {
		t.Fatal(err)
	}
	if len(msg.GetCounts()) == 0 {
		t.Error("StringToIntMap.Counts is empty")
	}
}

func TestCollectionsNestedMap(t *testing.T) {
	t.Parallel()

	msg := &testpb.NestedMap{}
	if err := fakeproto.Message(msg); err != nil {
		t.Fatal(err)
	}
	if len(msg.GetUsersById()) == 0 {
		t.Error("NestedMap.UsersById is empty")
	}
	for k, v := range msg.GetUsersById() {
		if v == nil {
			t.Errorf("NestedMap.UsersById[%d] is nil", k)
		}
	}
}

func TestCollectionsBinaryFields(t *testing.T) {
	t.Parallel()

	msg := &testpb.BinaryFields{}
	if err := fakeproto.Message(msg); err != nil {
		t.Fatal(err)
	}
	if len(msg.GetRaw()) == 0 {
		t.Error("BinaryFields.Raw is empty")
	}
	if len(msg.GetChunks()) == 0 {
		t.Error("BinaryFields.Chunks is empty")
	}
}

func TestCollectionsMaxRepeated(t *testing.T) {
	t.Parallel()

	const limit = 2
	msg := &testpb.StringList{}
	if err := fakeproto.Message(msg, fakeproto.WithMaxRepeated(limit)); err != nil {
		t.Fatal(err)
	}
	if n := len(msg.GetItems()); n > limit {
		t.Errorf("StringList.Items len=%d exceeds WithMaxRepeated(%d)", n, limit)
	}
}

// TestCollectionsMaxRepeatedZero documents the behavior of WithMaxRepeated(0):
// field() short-circuits with an early return before calling IntRange, so
// repeated fields are always left empty (0 elements).
func TestCollectionsMaxRepeatedZero(t *testing.T) {
	t.Parallel()

	for seed := range uint64(50) {
		msg := &testpb.StringList{}
		err := fakeproto.Message(msg, fakeproto.WithFaker(gofakeit.New(seed)), fakeproto.WithMaxRepeated(0))
		if err != nil {
			t.Fatal(err)
		}
		if n := len(msg.GetItems()); n > 1 {
			t.Errorf("seed=%d: StringList.Items len=%d, want at most 1 (WithMaxRepeated(0) edge case)", seed, n)
		}
	}
}

func BenchmarkCollectionsStringList(b *testing.B) {
	opt := fakeproto.WithFaker(gofakeit.New(0))
	b.ResetTimer()
	for range b.N {
		msg := &testpb.StringList{}
		_ = fakeproto.Message(msg, opt)
	}
}

func BenchmarkCollectionsNestedMap(b *testing.B) {
	opt := fakeproto.WithFaker(gofakeit.New(0))
	b.ResetTimer()
	for range b.N {
		msg := &testpb.NestedMap{}
		_ = fakeproto.Message(msg, opt)
	}
}

func FuzzCollectionsStringList(f *testing.F) {
	f.Fuzz(func(t *testing.T, seed uint64) {
		t.Parallel()
		msg := &testpb.StringList{}
		if err := fakeproto.Message(msg, fakeproto.WithFaker(gofakeit.New(seed))); err != nil {
			t.Fatal(err)
		}
	})
}
