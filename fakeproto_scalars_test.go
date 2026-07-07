package fakeproto_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"

	"github.com/psyhatter/fakeproto"
	testpb "github.com/psyhatter/fakeproto/internal/test/gen"
)

func TestScalars(t *testing.T) {
	t.Parallel()

	// A single random call could produce zero for any numeric field (probability
	// ~1/2^32), making single-shot non-zero assertions flaky. Instead, run over
	// a range of seeds and verify that each field is non-zero at least once.
	var (
		gotDouble, gotFloat      bool
		gotInt32, gotInt64       bool
		gotUint32, gotUint64     bool
		gotSint32, gotSint64     bool
		gotFixed32, gotFixed64   bool
		gotSfixed32, gotSfixed64 bool
		gotString, gotBytes      bool
		gotBool                  bool
	)
	for seed := range uint64(100) {
		msg := &testpb.AllScalars{}
		if err := fakeproto.Message(msg, fakeproto.WithFaker(gofakeit.New(seed))); err != nil {
			t.Fatal(err)
		}
		gotDouble = gotDouble || msg.GetDoubleField() != 0
		gotFloat = gotFloat || msg.GetFloatField() != 0
		gotInt32 = gotInt32 || msg.GetInt32Field() != 0
		gotInt64 = gotInt64 || msg.GetInt64Field() != 0
		gotUint32 = gotUint32 || msg.GetUint32Field() != 0
		gotUint64 = gotUint64 || msg.GetUint64Field() != 0
		gotSint32 = gotSint32 || msg.GetSint32Field() != 0
		gotSint64 = gotSint64 || msg.GetSint64Field() != 0
		gotFixed32 = gotFixed32 || msg.GetFixed32Field() != 0
		gotFixed64 = gotFixed64 || msg.GetFixed64Field() != 0
		gotSfixed32 = gotSfixed32 || msg.GetSfixed32Field() != 0
		gotSfixed64 = gotSfixed64 || msg.GetSfixed64Field() != 0
		gotString = gotString || msg.GetStringField() != ""
		gotBytes = gotBytes || len(msg.GetBytesField()) != 0
		gotBool = gotBool || msg.GetBoolField()
	}
	if !gotDouble {
		t.Error("DoubleField was zero across all seeds")
	}
	if !gotFloat {
		t.Error("FloatField was zero across all seeds")
	}
	if !gotInt32 {
		t.Error("Int32Field was zero across all seeds")
	}
	if !gotInt64 {
		t.Error("Int64Field was zero across all seeds")
	}
	if !gotUint32 {
		t.Error("Uint32Field was zero across all seeds")
	}
	if !gotUint64 {
		t.Error("Uint64Field was zero across all seeds")
	}
	if !gotSint32 {
		t.Error("Sint32Field was zero across all seeds")
	}
	if !gotSint64 {
		t.Error("Sint64Field was zero across all seeds")
	}
	if !gotFixed32 {
		t.Error("Fixed32Field was zero across all seeds")
	}
	if !gotFixed64 {
		t.Error("Fixed64Field was zero across all seeds")
	}
	if !gotSfixed32 {
		t.Error("Sfixed32Field was zero across all seeds")
	}
	if !gotSfixed64 {
		t.Error("Sfixed64Field was zero across all seeds")
	}
	if !gotString {
		t.Error("StringField was empty across all seeds")
	}
	if !gotBytes {
		t.Error("BytesField was empty across all seeds")
	}
	if !gotBool {
		t.Error("BoolField was false across all seeds")
	}
}

// TestSemanticScalars verifies that fields whose names are registered in
// gofakeit's function registry produce semantically correct values instead of
// arbitrary random numbers.
func TestSemanticScalars(t *testing.T) {
	t.Parallel()

	msg := &testpb.SemanticScalars{}
	if err := fakeproto.Message(msg, fakeproto.WithFaker(gofakeit.New(1))); err != nil {
		t.Fatal(err)
	}

	if msg.GetAge() < 0 || msg.GetAge() > 120 {
		t.Errorf("Age=%d out of [0, 120]", msg.GetAge())
	}
	if msg.GetYear() < 1900 || msg.GetYear() > 2100 {
		t.Errorf("Year=%d out of [1900, 2100]", msg.GetYear())
	}
	if msg.GetMonth() < 1 || msg.GetMonth() > 12 {
		t.Errorf("Month=%d out of [1, 12]", msg.GetMonth())
	}
	if msg.GetDay() < 1 || msg.GetDay() > 31 {
		t.Errorf("Day=%d out of [1, 31]", msg.GetDay())
	}
	if msg.GetHour() < 0 || msg.GetHour() > 23 {
		t.Errorf("Hour=%d out of [0, 23]", msg.GetHour())
	}
	if msg.GetMinute() < 0 || msg.GetMinute() > 59 {
		t.Errorf("Minute=%d out of [0, 59]", msg.GetMinute())
	}
	if msg.GetSecond() < 0 || msg.GetSecond() > 59 {
		t.Errorf("Second=%d out of [0, 59]", msg.GetSecond())
	}
	if msg.GetLatitude() < -90 || msg.GetLatitude() > 90 {
		t.Errorf("Latitude=%f out of [-90, 90]", msg.GetLatitude())
	}
	if msg.GetLongitude() < -180 || msg.GetLongitude() > 180 {
		t.Errorf("Longitude=%f out of [-180, 180]", msg.GetLongitude())
	}
	if msg.GetPrice() <= 0 {
		t.Errorf("Price=%f must be positive", msg.GetPrice())
	}
	// "email" is a string generator in gofakeit; on an int32 field num() must
	// fall back to a random value, not return zero.
	if msg.GetEmail() == 0 {
		t.Error("Email (int32) is zero - num() fallback was not called for a string-typed registry entry")
	}
	// "galaxy" is not registered in gofakeit at all; num() must fall back to a
	// random value, not return zero.
	if msg.GetGalaxy() == 0 {
		t.Error("Galaxy (int32) is zero - num() fallback was not called for an unknown name")
	}
	// "numerify" is registered in gofakeit but requires a mandatory "str" param
	// with no default; Generate returns an error, so str() must fall back to a
	// random word/phrase rather than returning an empty string.
	if msg.GetNumerify() == "" {
		t.Error("Numerify (string) is empty - str() fallback was not called when Generate errored")
	}
}

func FuzzSemanticScalars(f *testing.F) {
	f.Fuzz(func(t *testing.T, seed uint64) {
		t.Parallel()
		msg := &testpb.SemanticScalars{}
		if err := fakeproto.Message(msg, fakeproto.WithFaker(gofakeit.New(seed))); err != nil {
			t.Fatal(err)
		}
		if msg.GetAge() < 0 || msg.GetAge() > 120 {
			t.Errorf("seed=%d: Age=%d out of [0, 120]", seed, msg.GetAge())
		}
		if msg.GetYear() < 1900 || msg.GetYear() > 2100 {
			t.Errorf("seed=%d: Year=%d out of [1900, 2100]", seed, msg.GetYear())
		}
		if msg.GetMonth() < 1 || msg.GetMonth() > 12 {
			t.Errorf("seed=%d: Month=%d out of [1, 12]", seed, msg.GetMonth())
		}
		if msg.GetDay() < 1 || msg.GetDay() > 31 {
			t.Errorf("seed=%d: Day=%d out of [1, 31]", seed, msg.GetDay())
		}
		if msg.GetHour() < 0 || msg.GetHour() > 23 {
			t.Errorf("seed=%d: Hour=%d out of [0, 23]", seed, msg.GetHour())
		}
		if msg.GetMinute() < 0 || msg.GetMinute() > 59 {
			t.Errorf("seed=%d: Minute=%d out of [0, 59]", seed, msg.GetMinute())
		}
		if msg.GetSecond() < 0 || msg.GetSecond() > 59 {
			t.Errorf("seed=%d: Second=%d out of [0, 59]", seed, msg.GetSecond())
		}
		if msg.GetLatitude() < -90 || msg.GetLatitude() > 90 {
			t.Errorf("seed=%d: Latitude=%f out of [-90, 90]", seed, msg.GetLatitude())
		}
		if msg.GetLongitude() < -180 || msg.GetLongitude() > 180 {
			t.Errorf("seed=%d: Longitude=%f out of [-180, 180]", seed, msg.GetLongitude())
		}
		if msg.GetPrice() <= 0 {
			t.Errorf("seed=%d: Price=%f must be positive", seed, msg.GetPrice())
		}
		if msg.GetEmail() == 0 {
			t.Errorf("seed=%d: Email (int32) is zero - num() fallback was not called", seed)
		}
		if msg.GetGalaxy() == 0 {
			t.Errorf("seed=%d: Galaxy (int32) is zero - num() fallback was not called", seed)
		}
	})
}

func BenchmarkSemanticScalars(b *testing.B) {
	opt := fakeproto.WithFaker(gofakeit.New(0))
	b.ResetTimer()
	for range b.N {
		msg := &testpb.SemanticScalars{}
		_ = fakeproto.Message(msg, opt)
	}
}

func BenchmarkScalars(b *testing.B) {
	opt := fakeproto.WithFaker(gofakeit.New(0))
	b.ResetTimer()
	for range b.N {
		msg := &testpb.AllScalars{}
		_ = fakeproto.Message(msg, opt)
	}
}

func FuzzScalars(f *testing.F) {
	f.Fuzz(func(t *testing.T, seed uint64) {
		t.Parallel()
		msg := &testpb.AllScalars{}
		faker := gofakeit.New(seed)
		if err := fakeproto.Message(msg, fakeproto.WithFaker(faker)); err != nil {
			t.Fatal(err)
		}
	})
}
