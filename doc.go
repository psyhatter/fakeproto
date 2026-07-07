// Package fakeproto fills protobuf messages with random fake data.
//
// It uses [github.com/brianvoe/gofakeit/v7] as the data-generation engine and
// walks the protobuf message descriptor via protoreflect, which means it
// correctly handles oneof, map, repeated, proto3 optional presence, enum values
// from the declared set, and well-known types.
//
// # Why a separate module?
//
// Protobuf support lives here, not in gofakeit, to preserve gofakeit's
// zero-dependency guarantee. Users who don't need protobuf don't pay for it.
//
// # Basic usage
//
//	var msg mypb.MyMessage
//	if err := fakeproto.Message(&msg); err != nil {
//	    log.Fatal(err)
//	}
//
// Use [WithFaker] to supply a seeded [gofakeit.Faker] for deterministic output:
//
//	f := gofakeit.New(42) // fixed seed
//	var msg mypb.MyMessage
//	if err := fakeproto.Message(&msg, fakeproto.WithFaker(f)); err != nil {
//	    log.Fatal(err)
//	}
//
// Use [Enum] to fill a single enum variable with a random declared value:
//
//	var s mypb.Status
//	fakeproto.Enum(&s)
package fakeproto
