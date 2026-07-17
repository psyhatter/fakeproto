[![Go Reference](https://pkg.go.dev/badge/github.com/psyhatter/fakeproto.svg)](https://pkg.go.dev/github.com/psyhatter/fakeproto)

# Changelog

All notable changes to this project will be documented in this file.

## [0.0.1] - 2026-07-17

### Added

- `Message` fills proto messages with random fake data via gofakeit.
- `Enum[E]` fills a standalone enum variable with a random declared value.
- Options: `WithFaker`, `WithMaxDepth`, `WithMaxRepeated`, `WithOptionalProbability`, `WithSkipOneof`.
- Field name heuristics: field names registered in gofakeit (`email`, `city`, `year`, …) produce semantically correct
  values automatically.
- Well-known type support: `Timestamp`, `Duration`, `FieldMask`, `Any`, `Struct`, `ListValue`, `Value`, and all wrapper
  types.
- proto2 `GroupKind` support.
- proto3 `optional` scalar support (explicit presence via synthetic oneof).
- Deterministic output via seeded `*gofakeit.Faker`.
