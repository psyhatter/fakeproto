package fakeproto_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"

	"github.com/psyhatter/fakeproto"
	testpb "github.com/psyhatter/fakeproto/internal/test/gen"
)

func TestNestedAddress(t *testing.T) {
	t.Parallel()

	msg := &testpb.Address{}
	if err := fakeproto.Message(msg); err != nil {
		t.Fatal(err)
	}
	if msg.GetStreet() == "" {
		t.Error("Street is empty")
	}
	if msg.GetCity() == "" {
		t.Error("City is empty")
	}
	if msg.GetCountry() == "" {
		t.Error("Country is empty")
	}
	if msg.GetZip() == "" {
		t.Error("Zip is empty")
	}
}

func BenchmarkNestedAddress(b *testing.B) {
	opt := fakeproto.WithFaker(gofakeit.New(0))
	b.ResetTimer()
	for range b.N {
		msg := &testpb.Address{}
		_ = fakeproto.Message(msg, opt)
	}
}

func FuzzNestedAddress(f *testing.F) {
	f.Fuzz(func(t *testing.T, seed uint64) {
		t.Parallel()
		msg := &testpb.Address{}
		if err := fakeproto.Message(msg, fakeproto.WithFaker(gofakeit.New(seed))); err != nil {
			t.Fatal(err)
		}
	})
}

func TestNestedCompany(t *testing.T) {
	t.Parallel()

	msg := &testpb.Company{}
	if err := fakeproto.Message(msg); err != nil {
		t.Fatal(err)
	}
	if msg.GetName() == "" {
		t.Error("Company.Name is empty")
	}
	if msg.GetHeadquarters() == nil {
		t.Fatal("Company.Headquarters is nil")
	}
	if msg.GetHeadquarters().GetCity() == "" {
		t.Error("Company.Headquarters.City is empty")
	}
}

func BenchmarkNestedCompany(b *testing.B) {
	opt := fakeproto.WithFaker(gofakeit.New(0))
	b.ResetTimer()
	for range b.N {
		msg := &testpb.Company{}
		_ = fakeproto.Message(msg, opt)
	}
}

func FuzzNestedCompany(f *testing.F) {
	f.Fuzz(func(t *testing.T, seed uint64) {
		t.Parallel()
		msg := &testpb.Company{}
		if err := fakeproto.Message(msg, fakeproto.WithFaker(gofakeit.New(seed))); err != nil {
			t.Fatal(err)
		}
	})
}
