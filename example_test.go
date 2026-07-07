package fakeproto_test

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/brianvoe/gofakeit/v7"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/psyhatter/fakeproto"
	testpb "github.com/psyhatter/fakeproto/internal/test/gen"
)

func ExampleMessage() {
	user := &testpb.User{}
	// WithFaker sets a fixed seed so the output is always identical.
	if err := fakeproto.Message(user, fakeproto.WithFaker(gofakeit.New(657))); err != nil {
		panic(err) // TODO: Handle error.
	}

	// protojson.Marshal works better with timestamppb.Timestamp than json.Marshal.
	b, err := protojson.Marshal(user)
	if err != nil {
		panic(err) // TODO: Handle error.
	}

	// To make the output prettier.
	buf := bytes.NewBuffer(nil)
	err = json.Indent(buf, b, "", " ")
	if err != nil {
		panic(err) // TODO: Handle error.
	}
	fmt.Println(buf.String())

	// Output:
	// {
	//  "id": "z3f7cb4zc8kn4ukbk56x",
	//  "email": "cliffordking@clark.com",
	//  "name": "Aleen Schroeder",
	//  "status": "STATUS_ACTIVE",
	//  "createdAt": "2006-07-23T05:03:41.200315279Z",
	//  "age": 29,
	//  "tags": [
	//   "not",
	//   "which",
	//   "example"
	//  ]
	// }
}

func ExampleEnum() {
	var p testpb.Priority
	fakeproto.Enum(&p, fakeproto.WithFaker(gofakeit.New(42)))
	fmt.Println(p)
	// Output:
	// PRIORITY_MEDIUM
}
