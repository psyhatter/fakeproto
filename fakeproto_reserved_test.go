package fakeproto_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"

	"github.com/psyhatter/fakeproto"
	testpb "github.com/psyhatter/fakeproto/internal/test/gen"
)

// TestReservedFields verifies that Proto fills all active fields correctly when
// reserved numbers and names are interleaved in the descriptor. Reserved entries
// never appear in Fields()/Oneofs()/Enum.Values(), so the walk must not skip or
// corrupt the surrounding active fields.
func TestReservedFields(t *testing.T) {
	t.Parallel()

	// User has reserved 3, 5 / "nickname", "phone" after the active fields
	// (id=1, name=2, age=6, login oneof email/phone).
	t.Run("message", func(t *testing.T) {
		t.Parallel()
		msg := &testpb.User{}
		if err := fakeproto.Message(msg, fakeproto.WithFaker(gofakeit.New(0))); err != nil {
			t.Fatal(err)
		}
		if msg.GetId() == "" {
			t.Error("User.Id is empty")
		}
		if msg.GetName() == "" {
			t.Error("User.Name is empty")
		}
		if msg.GetLogin() == nil {
			t.Error("User.Login (oneof) is nil")
		}
	})

	// PaymentOneof has reserved 2, 4 / "wire_transfer", "check_number" before the
	// oneof. Active oneof fields: card_number=1, bank_account=3, crypto_wallet=5, cash=6.
	t.Run("oneof", func(t *testing.T) {
		t.Parallel()
		for seed := range uint64(20) {
			msg := &testpb.PaymentOneof{}
			if err := fakeproto.Message(msg, fakeproto.WithFaker(gofakeit.New(seed))); err != nil {
				t.Fatalf("seed=%d: %v", seed, err)
			}
			if msg.GetPaymentMethod() == nil {
				t.Fatalf("seed=%d: PaymentOneof.PaymentMethod is nil - reserved entries disrupted oneof walk", seed)
			}
		}
	})

	// WithEnum uses Status which has reserved 2, 4 / "STATUS_BANNED", "STATUS_ARCHIVED".
	// Active values: UNSPECIFIED=0, DELETED=-1, ACTIVE=1, INACTIVE=3, SUSPENDED=6.
	// Reserved numbers 2 and 4 must never appear.
	t.Run("enum", func(t *testing.T) {
		t.Parallel()
		declared := map[int32]bool{
			0:  true, // STATUS_UNSPECIFIED
			-1: true, // STATUS_DELETED
			1:  true, // STATUS_ACTIVE
			3:  true, // STATUS_INACTIVE
			6:  true, // STATUS_SUSPENDED
		}
		for seed := range uint64(50) {
			msg := &testpb.WithEnum{}
			if err := fakeproto.Message(msg, fakeproto.WithFaker(gofakeit.New(seed))); err != nil {
				t.Fatalf("seed=%d: %v", seed, err)
			}
			if !declared[int32(msg.GetStatus())] {
				t.Errorf("seed=%d: Status=%d is not a declared enum value", seed, msg.GetStatus())
			}
		}
	})
}
