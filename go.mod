module github.com/psyhatter/fakeproto

// Protobuf support lives in github.com/psyhatter/fakeproto, not in
// github.com/brianvoe/gofakeit, to preserve gofakeit's zero-dependency guarantee.

go 1.23

require (
	github.com/brianvoe/gofakeit/v7 v7.15.0
	google.golang.org/protobuf v1.36.11
)
