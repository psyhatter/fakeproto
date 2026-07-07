# Fake field option - design sketch

## Problem

`fakeproto` currently picks a generator for a field based on its name: a field
called `email` gets `gofakeit`'s email generator, a field called `city` gets
its city generator, and so on. Fields whose names don't match any generator
fall back to a random word or phrase.

This is convenient out of the box but gives users no way to override the
choice. `gofakeit.Struct` solves an equivalent problem for plain Go structs
with struct tags:

```go
type User struct {
    Name  string `fake:"{firstname} {lastname}"`
    Phone string `fake:"(###) ###-####"`
    Age   int    `fake:"{number:18,80}"`
}
```

There is no equivalent mechanism for protobuf fields today.

## Idea: proto field option

Protobuf supports custom field options via `extend google.protobuf.FieldOptions`.
The idea is to ship a `fake.proto` extension alongside `fakeproto`:

```proto
syntax = "proto3";

package fakeproto.v1;

import "google/protobuf/descriptor.proto";

option go_package = "github.com/psyhatter/fakeproto/fakepb";

extend google.protobuf.FieldOptions {
  // fake specifies a gofakeit template for this field. The value is passed to
  // gofakeit.Faker.Generate and supports the full template syntax:
  // named generators ({name}, {email}), format patterns (###-###-####), and
  // combined templates ({firstname} {lastname}).
  string fake = 50000;
}
```

Users would annotate their own proto files:

```proto
syntax = "proto3";
import "fake.proto";

message User {
  string name  = 1 [(fakeproto.v1.fake) = "{firstname} {lastname}"];
  string phone = 2 [(fakeproto.v1.fake) = "(###) ###-####"];
  int32  age   = 3 [(fakeproto.v1.fake) = "{number:18,80}"];
}
```

At runtime fakeproto would read the option via `proto.GetExtension` and pass
its value to `gofakeit.Faker.Generate`. Fields without the option would
continue to use the current name-based lookup and random fallback.

## Open problems

### 1. Fields you cannot annotate

The option only helps when you own and can edit the proto source. There is no
way to attach it to:

- **Well-known types** (`google.protobuf.Timestamp`, `Duration`, `Struct`, …) - fakeproto handles these with hardcoded logic today, which would remain the only option.
- **Third-party protos** - fields from external dependencies cannot be annotated without forking the source.

An alternative approach would be some kind of field-path-to-template mapping
passed as a Go `Option` (e.g. `WithFieldTemplate("user.phone", "(###) ###-####")`),
which would not require touching the proto files at all. This is worth
exploring independently.

### 2. Distribution and installation

For the option to work, `fake.proto` must be on the protoc include path when
users compile their own protos. This is a non-trivial distribution problem.

**With buf (Buf Schema Registry):**
The cleanest solution. Publish `fake.proto` to `buf.build/psyhatter/fakeproto`
and users add a single line to their `buf.yaml`:

```yaml
deps:
  - buf.build/psyhatter/fakeproto
```

After `buf dep update`, the file is available automatically - no explicit `-I`
flags needed. This is the same mechanism used by most modern protobuf libraries.

**With raw protoc:**
There is no `go install` equivalent for proto files. Since fakeproto is already
a Go dependency, the proto file sits in the module cache after `go get`. Users
must locate it manually:

```bash
protoc \
  -I "$(go list -m -f '{{.Dir}}' github.com/psyhatter/fakeproto)/proto" \
  -I . \
  --go_out=. \
  your/service.proto
```

The `go list` invocation is fragile in environments where the module cache path
is non-standard, and the command is noticeably more complex than what users
expect from a library. Projects that use raw protoc and cannot migrate to buf
would face a meaningfully worse experience.
