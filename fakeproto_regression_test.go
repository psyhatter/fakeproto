package fakeproto_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"

	"github.com/psyhatter/fakeproto"
	testpb "github.com/psyhatter/fakeproto/internal/test/gen"
)

// Regression test for bug #1: oneof members were subject to optionalProbability.
//
// Every member of a oneof reports HasPresence() == true, so the generic
// "leave optional fields unset with probability p" check used to apply to the
// oneof variant that had already been chosen. With the default p=0.10 roughly
// every tenth generated oneof ended up completely empty; with p=1.0 the oneof
// was always empty. Once the generator has picked a variant, it must commit to
// setting it - presence of oneof members is not optionality.
func TestRegressionOneofNotAffectedByOptionalProbability(t *testing.T) {
	t.Parallel()

	// p=1.0 makes the bug deterministic: if oneof members went through the
	// optionality check, the variant would be skipped on every single seed.
	for seed := range uint64(100) {
		msg := &testpb.PaymentOneof{}
		err := fakeproto.Message(msg, fakeproto.WithFaker(gofakeit.New(seed)), fakeproto.WithOptionalProbability(1))
		if err != nil {
			t.Fatal(err)
		}
		if msg.GetPaymentMethod() == nil {
			t.Fatalf("seed=%d: oneof variant left unset - optionalProbability leaked into oneof handling", seed)
		}
	}
}

// Regression test for bug #2: proto3 optional fields are synthetic oneofs.
//
// The protobuf compiler wraps every proto3 `optional` scalar into a
// single-member synthetic oneof to implement explicit presence. Code that
// filters "fields belonging to a oneof" by ContainingOneof() != nil alone
// accidentally captures optional fields too, routing them through the oneof
// code path: WithSkipOneof would wrongly leave them unset, and
// WithOptionalProbability would never apply. Real and synthetic oneofs must
// be told apart via OneofDescriptor.IsSynthetic().
func TestRegressionOptionalIsSyntheticOneofNotRealOneof(t *testing.T) {
	t.Parallel()

	// WithSkipOneof must skip real oneofs only. If synthetic oneofs were
	// treated as real, the optional fields below would be left unset.
	msg := &testpb.OptionalFields{}
	err := fakeproto.Message(msg,
		fakeproto.WithSkipOneof(),
		fakeproto.WithOptionalProbability(0))
	if err != nil {
		t.Fatal(err)
	}
	if msg.MaybeCount == nil || msg.MaybeName == nil || msg.MaybeFlag == nil {
		t.Error("proto3 optional fields skipped by WithSkipOneof - synthetic oneof treated as a real one")
	}

	// And the mirror check: WithOptionalProbability(1) must leave optional
	// fields unset even though they live inside (synthetic) oneofs, which are
	// normally exempt from the optionality check.
	msg2 := &testpb.OptionalFields{}
	err = fakeproto.Message(msg2,
		fakeproto.WithOptionalProbability(1))
	if err != nil {
		t.Fatal(err)
	}
	if msg2.MaybeCount != nil || msg2.MaybeName != nil || msg2.MaybeFlag != nil {
		t.Error("optionalProbability ignored for proto3 optional fields - synthetic oneof exempted like a real one")
	}
}
