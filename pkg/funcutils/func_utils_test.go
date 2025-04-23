package funcutils

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/wuqunyong/file_storage/pkg/encoders"
	"github.com/wuqunyong/file_storage/pkg/errs"
	testdata "github.com/wuqunyong/file_storage/proto"
	"google.golang.org/protobuf/proto"
)

type Reply1 struct {
	Value int32
}

type Handler struct {
}

func (h *Handler) Func1(ctx context.Context, arg *int32, reply *Reply1) errs.CodeError {
	reply.Value += *arg
	fmt.Printf("inside value:%v\n", reply)
	return errs.NewCodeError(errors.New("invalid"))
}

func (h *Handler) Func2(ctx context.Context, arg *int32, reply *testdata.Person) errs.CodeError {
	reply.Age += *arg
	fmt.Printf("inside value:%v\n", reply)
	return nil
}

func Test1(t *testing.T) {
	handler := Handler{}
	ptrMethod := GetRPCReflectFunc(handler.Func1, true)
	if ptrMethod == nil {
		return
	}

	reply := &Reply1{}
	var args *int32 = new(int32)
	*args = 10
	CallPRCReflectRequestFunc(ptrMethod, context.Background(), args, reply)
	fmt.Printf("Reply1 value:%v\n", reply)
}

func Test2(t *testing.T) {
	handler := Handler{}
	ptrMethod := GetRPCReflectFunc(handler.Func2, true)
	if ptrMethod == nil {
		return
	}

	reply := &testdata.Person{Name: "小明", Age: 18}
	data, err := proto.Marshal(reply)
	if err != nil {
		t.Fatal("Couldn't serialize object", err)
	}

	argValue := reflect.New(ptrMethod.ReplyType.Elem()).Interface()
	decoder := encoders.NewProtobufEncoder()
	decoder.Decode(data, argValue)

	var args *int32 = new(int32)
	*args = 10
	CallPRCReflectRequestFunc(ptrMethod, context.Background(), args, argValue)
	fmt.Printf("Reply2 value:%v\n", reply)
}
