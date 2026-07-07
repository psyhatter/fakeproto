package fakeproto_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"

	"github.com/psyhatter/fakeproto"
	testpb "github.com/psyhatter/fakeproto/internal/test/gen"
)

func TestOneofSet(t *testing.T) {
	t.Parallel()

	msg := &testpb.PaymentOneof{}
	if err := fakeproto.Message(msg); err != nil {
		t.Fatal(err)
	}
	if msg.GetPaymentMethod() == nil {
		t.Error("PaymentOneof.PaymentMethod is nil - exactly one variant must be set")
	}
}

func TestOneofAllVariantsReachable(t *testing.T) {
	t.Parallel()

	type counts struct{ card, bank, crypto, cash int }
	var c counts
	for seed := range uint64(200) {
		msg := &testpb.PaymentOneof{}
		if err := fakeproto.Message(msg, fakeproto.WithFaker(gofakeit.New(seed))); err != nil {
			t.Fatal(err)
		}
		switch msg.GetPaymentMethod().(type) {
		case *testpb.PaymentOneof_CardNumber:
			c.card++
		case *testpb.PaymentOneof_BankAccount:
			c.bank++
		case *testpb.PaymentOneof_CryptoWallet:
			c.crypto++
		case *testpb.PaymentOneof_Cash:
			c.cash++
		}
	}
	if c.card == 0 {
		t.Error("card_number variant never selected")
	}
	if c.bank == 0 {
		t.Error("bank_account variant never selected")
	}
	if c.crypto == 0 {
		t.Error("crypto_wallet variant never selected")
	}
	if c.cash == 0 {
		t.Error("cash variant never selected")
	}
}

func TestOneofSkip(t *testing.T) {
	t.Parallel()

	msg := &testpb.PaymentOneof{}
	if err := fakeproto.Message(msg, fakeproto.WithSkipOneof()); err != nil {
		t.Fatal(err)
	}
	if msg.GetPaymentMethod() != nil {
		t.Error("expected PaymentMethod to be nil with WithSkipOneof()")
	}
}

func BenchmarkOneof(b *testing.B) {
	opt := fakeproto.WithFaker(gofakeit.New(0))
	b.ResetTimer()
	for range b.N {
		msg := &testpb.PaymentOneof{}
		_ = fakeproto.Message(msg, opt)
	}
}

func FuzzOneof(f *testing.F) {
	f.Fuzz(func(t *testing.T, seed uint64) {
		t.Parallel()
		msg := &testpb.PaymentOneof{}
		if err := fakeproto.Message(msg, fakeproto.WithFaker(gofakeit.New(seed))); err != nil {
			t.Fatal(err)
		}
		if msg.GetPaymentMethod() == nil {
			t.Error("PaymentMethod is nil")
		}
	})
}
