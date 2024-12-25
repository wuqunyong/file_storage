package testdata

import (
	"fmt"
	"testing"

	"google.golang.org/protobuf/proto"
)

func BenchmarkProtobufMarshalStruct(b *testing.B) {
	me := &Person{Name: "derek", Age: 22, Address: "140 New Montgomery St"}
	me.Children = make(map[string]*Person)

	me.Children["sam"] = &Person{Name: "sam", Age: 19, Address: "140 New Montgomery St"}
	me.Children["meg"] = &Person{Name: "meg", Age: 17, Address: "140 New Montgomery St"}

	data, err := proto.Marshal(me)
	if err != nil {
		b.Fatal("Couldn't serialize object", err)
	}

	fmt.Printf("data:%+v\n", data)

	req := &api.Req{}
	req.Data = data
	fmt.Printf("req:%+v\n", req)

	other := &Person{}

	err = proto.Unmarshal(req.Data, other)
	if err != nil {
		b.Fatal("Couldn't Unmarshal object", err)
	}
	fmt.Printf("other:%+v\n", other)
}
