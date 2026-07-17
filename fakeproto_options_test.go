package fakeproto_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"google.golang.org/protobuf/proto"

	"github.com/psyhatter/fakeproto"
	testpb "github.com/psyhatter/fakeproto/internal/test/gen"
)

// TestDeterminism verifies that two calls with the same seed produce identical
// messages for every proto type in the test suite.
func TestDeterminism(t *testing.T) {
	t.Parallel()

	const seed = uint64(42)

	cases := []struct {
		name string
		new  func() proto.Message
	}{
		{"AllScalars", func() proto.Message { return &testpb.AllScalars{} }},
		{"StringList", func() proto.Message { return &testpb.StringList{} }},
		{"NestedMap", func() proto.Message { return &testpb.NestedMap{} }},
		{"Company", func() proto.Message { return &testpb.Company{} }},
		{"Node", func() proto.Message { return &testpb.Node{} }},
		{"WithEnum", func() proto.Message { return &testpb.WithEnum{} }},
		{"PaymentOneof", func() proto.Message { return &testpb.PaymentOneof{} }},
		{"WellKnownTypes", func() proto.Message { return &testpb.WellKnownTypes{} }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			a, b := tc.new(), tc.new()
			if err := fakeproto.Message(a, fakeproto.WithFaker(gofakeit.New(seed))); err != nil {
				t.Fatal(err)
			}
			if err := fakeproto.Message(b, fakeproto.WithFaker(gofakeit.New(seed))); err != nil {
				t.Fatal(err)
			}
			if !proto.Equal(a, b) {
				t.Errorf("two calls with seed=%d produced different messages", seed)
			}
		})
	}
}

// TestProtoNoError verifies that Proto (using the global faker) does not return
// an error for any of the 7 test proto types.
func TestProtoNoError(t *testing.T) {
	t.Parallel()

	msgs := []proto.Message{
		&testpb.AllScalars{},
		&testpb.StringList{},
		&testpb.NestedMap{},
		&testpb.Company{},
		&testpb.WithEnum{},
		&testpb.PaymentOneof{},
		&testpb.WellKnownTypes{},
	}
	for _, msg := range msgs {
		if err := fakeproto.Message(msg); err != nil {
			t.Errorf("%T: unexpected error: %v", msg, err)
		}
	}
}

// TestOptionalProbabilityStatistical verifies that WithOptionalProbability(0.5)
// leaves roughly 50 % of optional fields unset across many iterations.
func TestOptionalProbabilityStatistical(t *testing.T) {
	t.Parallel()

	const (
		iterations = 1000
		wantRate   = 0.5
		tolerance  = 0.10
	)

	probability := fakeproto.WithOptionalProbability(wantRate)

	unset := 0
	for i := range iterations {
		msg := &testpb.OptionalFields{}
		err := fakeproto.Message(msg, fakeproto.WithFaker(gofakeit.New(uint64(i))), probability)
		if err != nil {
			t.Fatal(err)
		}
		if msg.MaybeCount == nil {
			unset++
		}
	}

	rate := float64(unset) / iterations
	if rate < wantRate-tolerance || rate > wantRate+tolerance {
		t.Errorf("unset rate=%.3f, want %.3f±%.3f", rate, wantRate, tolerance)
	}
}

// TestWithMaxRepeatedOne verifies that WithMaxRepeated(1) produces exactly one
// element in every repeated field.
func TestWithMaxRepeatedOne(t *testing.T) {
	t.Parallel()

	msg := &testpb.StringList{}
	if err := fakeproto.Message(msg, fakeproto.WithMaxRepeated(1)); err != nil {
		t.Fatal(err)
	}
	if n := len(msg.GetItems()); n != 1 {
		t.Errorf("expected 1 item, got %d", n)
	}
}

// TestWithMaxDepthZero verifies that WithMaxDepth(0) leaves nested message
// fields nil (depth budget exhausted before entering the root's children).
func TestWithMaxDepthZero(t *testing.T) {
	t.Parallel()

	msg := &testpb.Node{}
	if err := fakeproto.Message(msg, fakeproto.WithMaxDepth(0)); err != nil {
		t.Fatal(err)
	}
	// With maxDepth=0 the root itself gets filled (depth starts at 1),
	// but child must be nil because depth 1 >= maxDepth 0.
	if msg.GetChild() != nil {
		t.Error("Node.Child must be nil when WithMaxDepth(0)")
	}
}
