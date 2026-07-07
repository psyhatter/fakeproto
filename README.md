# fakeproto

[![Go Reference](https://pkg.go.dev/badge/github.com/psyhatter/fakeproto.svg)](https://pkg.go.dev/github.com/psyhatter/fakeproto)

Fill [protobuf](https://protobuf.dev/) messages with random fake data using
[gofakeit](https://github.com/brianvoe/gofakeit) as the generation engine.

## `gofakeit.Struct` vs `fakeproto.Message`

`gofakeit.Struct` works great for plain Go structs annotated with `fake` tags,
but falls short on protobuf-generated structs.

<table>
<tr>
<th><code>gofakeit.Struct</code></th>
<th><code>fakeproto.Message</code></th>
</tr>
<tr>
<td>

```go
user := &pb.User{}
err := gofakeit.Struct(user)
if err != nil {
    // TODO: Handle error.
}

fmt.Println(protojson.Format(user))
```
```json
{
 "id": "CrbhJHy",
❌// oneof field is skipped...
 "name": "dDheSag",
❌"status": 12374249, // out of enum
❌"createdAt": "283539-01-05T10:32:30Z",
 "age": 1308958553,
 "tags": [
  "lmVwo",
  "UFBN",
  "CIiUYBUi"
 ]
}

// email/phone are absent: gofakeit.Struct
// sets Go struct fields directly,
// bypassing proto reflection - oneof
// fields are invisible to it.
```

</td>
<td>

```go
user := &pb.User{}
err := fakeproto.Message(user)
if err != nil {
    // TODO: Handle error.
}

fmt.Println(protojson.Format(user))
```
```json
{
 "id": "z3f7cb4zc8kn4ukbk56x",
✅"email": "clifford@clark.com",
 "name": "Aleen Schroeder",
✅"status": "STATUS_ACTIVE",
✅"createdAt": "2006-07-23T05:03:41Z",
 "age": 29,
 "tags": [
  "not",
  "which",
  "example"
 ]
}

// email is set, phone is absent: login
// is a oneof - exactly one variant is
// chosen at random.
//
```

</td>
</tr>
</table>

`fakeproto` walks the message descriptor via `protoreflect` - the same API
used by the protobuf runtime itself - so it correctly handles `protobuf` features (see
[supported features](#supported-features)), using
`gofakeit` as the random-data engine under the hood.

## Why a separate module?

[gofakeit](https://github.com/brianvoe/gofakeit) explicitly advertises **zero dependencies** - it does not pull in
`protobuf`, reflection extensions,
or generated code. Adding `protobuf` support to `gofakeit` itself would break that guarantee for every user who doesn't
need it.

`fakeproto` is an opt-in companion module:

Users of `gofakeit` who don't need `protobuf` pay nothing. Users who do need it get a drop-in solution that delegates
all data generation back to `gofakeit`.

> *Protobuf support lives in `fakeproto`, not in `gofakeit`, to preserve `gofakeit`'s zero-dependency guarantee.*

## Installation

```
go get github.com/psyhatter/fakeproto@latest
```

Requires Go 1.23+.

## Quick start

```go
import (
    "github.com/brianvoe/gofakeit/v7"
    "github.com/psyhatter/fakeproto"
)

// Fill a message (non-deterministic).
user := &pb.User{}
err := fakeproto.Message(user)
if err != nil {
	// TODO: Handle error.
}

// Fill a message with a seeded faker for deterministic output (tests, snapshots).
user2 := &pb.User{}
err = fakeproto.Message(user2, fakeproto.WithFaker(gofakeit.New(42)))
if err != nil {
	// TODO: Handle error.
}

// Fill a standalone enum variable with a random declared value.
var s pb.Status
fakeproto.Enum(&s)
```

## Supported features

- [x] All 15 proto3 scalar types
- [x] Nested messages (recursive, depth-limited)
- [x] `repeated` fields and `map` fields
- [x] `enum` - value chosen from the descriptor's declared set (including negative values); also available as standalone
  `Enum[E](&e)`
- [x] `oneof` - one variant selected at random
- [x] Proto3 `optional` scalars (explicit presence)
- [x] Well-known types: `Timestamp`, `Duration`, `StringValue`/`Int32Value`/… wrappers, `Struct`, `ListValue`,
  `FieldMask`
- [x] `google.protobuf.Any` - filled with a random `structpb.Value`
- [x] Field name heuristics: `*email*`, `*name*`, `*city*`, `*zip*`, `*phone*`, `*url*`, `*uuid*`, `*currency*`, …

## Behavior

### Field name heuristics

For string and numeric fields, fakeproto looks up the field's text name in
gofakeit's generator registry before falling back to a random value. Fields
named `email`, `city`, `zip`, `latitude`, `year`, and hundreds of others
automatically receive semantically correct values without any configuration.

### `oneof`

One variant is chosen at random per oneof. All other variants are left unset.
Use `WithSkipOneof` to leave every oneof completely unset instead.

### `enum`

Reserved numbers and names are never produced - only values declared in the `.proto` file can appear.

### `repeated` fields and `maps`

The number of elements is chosen uniformly at random in `[1, maxRepeated]`
(default max is 5). Map keys are guaranteed to be unique because they are
generated independently and inserted into the map one by one - later entries
silently overwrite earlier ones on collision, so the actual count may be
slightly less than the chosen length for types with small key spaces (e.g.
`bool` keys).

### proto3 `optional` scalars

A field declared `optional` in proto3 has explicit presence. fakeproto leaves
it unset with probability `optionalProbability` (default `10%`). Set
`WithOptionalProbability(0)` to always fill every optional field.

### Depth limiting

`fakeproto` counts nesting depth starting at 1 for the root message. When a
message-type field would be filled at depth `> maxDepth`, it is left nil
instead. This prevents infinite recursion in self-referential schemas.

### Determinism

Pass a seeded `*gofakeit.Faker` via `WithFaker(gofakeit.New(seed))` to get
fully reproducible output. The default `gofakeit.GlobalFaker` is shared and
non-deterministic.

## Options

| Option                               | Default | Description                                                                                                |
|--------------------------------------|---------|------------------------------------------------------------------------------------------------------------|
| `WithMaxDepth(n int)`                | `10`    | Stop recursing into message fields at depth n.                                                             |
| `WithMaxRepeated(n int)`             | `5`     | Generate at most n elements in repeated fields and maps.                                                   |
| `WithOptionalProbability(p float64)` | `0.10`  | Probability that an explicit proto3 `optional` scalar is left unset. `0` = always set, `1` = always unset. |
| `WithSkipOneof()`                    | -       | Leave all oneof fields unset.                                                                              |

## Limitations

### No per-field generation control

`gofakeit.Struct` lets you override the generator for any field via a struct tag:

```go
type User struct {
    Name  string `fake:"{firstname} {lastname}"`
    Phone string `fake:"(###) ###-####"`
}
```

fakeproto has no equivalent. Generation is driven by field names: fakeproto
checks whether the field name matches any generator registered in gofakeit
(`email`, `name`, `city`, `phone`, `url`, and hundreds of others) and calls it
if found. When no match exists, a random value is used as a fallback. There is
no way to attach a specific template to a field whose name does not match any
registered generator.

This limitation may be addressed in a future version. One approach under
consideration is a custom proto field option (`fake.proto` extension) that
mirrors the struct-tag syntax. See [docs/fake-field-option.md](docs/fake-field-option.md)
for the design sketch and the open problems that need to be solved first.

## Acknowledgements

`fakeproto` would not exist without [gofakeit](https://github.com/brianvoe/gofakeit)
by [@brianvoe](https://github.com/brianvoe). `gofakeit` is one of the most thoughtfully
designed fake-data libraries in the Go ecosystem - with zero dependencies, a clean API, and an impressive breadth of
generators. `fakeproto` uses it as its sole data-generation engine. Thank you, **Brian**, and everyone who has
contributed to `gofakeit`.

## Relationship to `gofakeit`

fakeproto builds directly on top of `gofakeit`'s data generators and is designed to feel like a natural extension of it.
Once the API and feature coverage here stabilize, we intend to propose upstreaming this protobuf support to
[github.com/brianvoe/gofakeit](https://github.com/brianvoe/gofakeit) - either as an optional submodule or as a
referenced companion library - while preserving `gofakeit`'s zero-dependency guarantee for users who don't need
`protobuf`.
