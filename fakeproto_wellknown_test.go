package fakeproto_test

import (
	"strings"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/psyhatter/fakeproto"
	testpb "github.com/psyhatter/fakeproto/internal/test/gen"
)

func TestWellKnown(t *testing.T) {
	t.Parallel()

	msg := &testpb.WellKnownTypes{}
	if err := fakeproto.Message(msg, fakeproto.WithFaker(gofakeit.New(1))); err != nil {
		t.Fatal(err)
	}

	t.Run("Timestamp", func(t *testing.T) {
		t.Parallel()
		ts := msg.GetCreatedAt()
		if ts == nil {
			t.Fatal("CreatedAt is nil")
		}
		// Seconds must be a plausible Unix timestamp (after year 2000).
		year2000 := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		if !ts.AsTime().After(year2000) {
			t.Errorf("CreatedAt.Seconds=%d looks too old (before year %d)", ts.GetSeconds(), year2000.Year())
		}
		if ts.GetNanos() < 0 || ts.GetNanos() >= 1e9 {
			t.Errorf("CreatedAt.Nanos=%d out of [0, 1e9)", ts.GetNanos())
		}
		if msg.GetUpdatedAt() == nil {
			t.Error("UpdatedAt is nil")
		}
	})

	t.Run("Duration", func(t *testing.T) {
		t.Parallel()
		d := msg.GetTtl()
		if d == nil {
			t.Fatal("Ttl is nil")
		}
		// Duration must be positive and at most one year.
		if d.GetSeconds() <= 0 {
			t.Errorf("Ttl.Seconds=%d must be positive", d.GetSeconds())
		}
		if d.GetSeconds() > 86400*365 {
			t.Errorf("Ttl.Seconds=%d exceeds one year", d.GetSeconds())
		}
		if msg.GetProcessingTime() == nil {
			t.Error("ProcessingTime is nil")
		}
	})

	t.Run("Wrappers", func(t *testing.T) {
		t.Parallel()
		if msg.GetOptionalTitle() == nil {
			t.Error("OptionalTitle wrapper is nil")
		} else if msg.GetOptionalTitle().GetValue() == "" {
			t.Error("OptionalTitle.Value is empty")
		}
		if msg.GetOptionalCount() == nil {
			t.Error("OptionalCount wrapper is nil")
		}
		if msg.GetOptionalFlag() == nil {
			t.Error("OptionalFlag wrapper is nil")
		}
		if msg.GetOptionalScore() == nil {
			t.Error("OptionalScore wrapper is nil")
		}
	})

	t.Run("StructAndList", func(t *testing.T) {
		t.Parallel()
		md := msg.GetMetadata()
		if md == nil {
			t.Fatal("Metadata Struct is nil")
		}
		if len(md.GetFields()) == 0 {
			t.Error("Metadata.Fields is empty")
		}
		for k, v := range md.GetFields() {
			if v == nil {
				t.Errorf("Metadata.Fields[%q] value is nil", k)
			} else if v.GetKind() == nil {
				t.Errorf("Metadata.Fields[%q].Kind is nil", k)
			}
		}

		tags := msg.GetTags()
		if tags == nil {
			t.Fatal("Tags ListValue is nil")
		}
		if len(tags.GetValues()) == 0 {
			t.Error("Tags.Values is empty")
		}
		for i, v := range tags.GetValues() {
			if v == nil {
				t.Errorf("Tags.Values[%d] is nil", i)
			} else if v.GetKind() == nil {
				t.Errorf("Tags.Values[%d].Kind is nil", i)
			}
		}
	})

	t.Run("StructValue", func(t *testing.T) {
		t.Parallel()
		sv := msg.GetSampleValue()
		if sv == nil {
			t.Fatal("SampleValue is nil")
		}
		if sv.GetKind() == nil {
			t.Fatal("SampleValue.Kind is nil")
		}
		switch sv.GetKind().(type) {
		case *structpb.Value_NullValue,
			*structpb.Value_NumberValue,
			*structpb.Value_StringValue,
			*structpb.Value_BoolValue,
			*structpb.Value_StructValue,
			*structpb.Value_ListValue:
			// ok
		default:
			t.Errorf("SampleValue.Kind has unexpected type %T", sv.GetKind())
		}
	})

	t.Run("FieldMask", func(t *testing.T) {
		t.Parallel()
		fm := msg.GetUpdateMask()
		if fm == nil {
			t.Fatal("UpdateMask FieldMask is nil")
		}
		if len(fm.GetPaths()) == 0 {
			t.Error("UpdateMask.Paths is empty")
		}
		for i, p := range fm.GetPaths() {
			if p == "" {
				t.Errorf("UpdateMask.Paths[%d] is empty", i)
			}
			if strings.ContainsAny(p, " \t\n\r") {
				t.Errorf("UpdateMask.Paths[%d]=%q contains whitespace", i, p)
			}
		}
	})

	t.Run("Any", func(t *testing.T) {
		t.Parallel()
		details := msg.GetDetails()
		if details == nil {
			t.Fatal("Details (Any) is nil - should be filled with a structpb.Value")
		}
		// Type URL must point to structpb.Value - the type used by wellKnown.
		const wantTypeURL = "type.googleapis.com/google.protobuf.Value"
		if details.GetTypeUrl() != wantTypeURL {
			t.Errorf("Details.TypeUrl=%q, want %q", details.GetTypeUrl(), wantTypeURL)
		}
		// Must unmarshal cleanly.
		var sv structpb.Value
		if err := details.UnmarshalTo(&sv); err != nil {
			t.Errorf("Details.UnmarshalTo Value: %v", err)
		}
		if sv.GetKind() == nil {
			t.Error("Details Value.Kind is nil after unmarshal")
		}
	})
}

func BenchmarkWellKnown(b *testing.B) {
	opt := fakeproto.WithFaker(gofakeit.New(0))
	b.ResetTimer()
	for range b.N {
		msg := &testpb.WellKnownTypes{}
		_ = fakeproto.Message(msg, opt)
	}
}

func FuzzWellKnown(f *testing.F) {
	f.Fuzz(func(t *testing.T, seed uint64) {
		t.Parallel()
		msg := &testpb.WellKnownTypes{}
		if err := fakeproto.Message(msg, fakeproto.WithFaker(gofakeit.New(seed))); err != nil {
			t.Fatal(err)
		}
	})
}
