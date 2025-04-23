package proto

import (
	"fmt"
	"testing"

	"github.com/wuqunyong/file_storage/proto/common_msg"
	"google.golang.org/protobuf/proto"
)

func BenchmarkProtobufMarshalStruct(b *testing.B) {
	me := &common_msg.Person{Name: "derek", Age: 22, Address: "140 New Montgomery St"}
	me.Children = make(map[string]*common_msg.Person)

	me.Children["sam"] = &common_msg.Person{Name: "sam", Age: 19, Address: "140 New Montgomery St"}
	me.Children["meg"] = &common_msg.Person{Name: "meg", Age: 17, Address: "140 New Montgomery St"}

	data, err := proto.Marshal(me)
	if err != nil {
		b.Fatal("Couldn't serialize object", err)
	}

	fmt.Printf("data:%+v\n", data)

	other := &common_msg.Person{}

	err = proto.Unmarshal(data, other)
	if err != nil {
		b.Fatal("Couldn't Unmarshal object", err)
	}
	fmt.Printf("other:%+v\n", other)
}
