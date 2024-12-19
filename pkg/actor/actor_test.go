package actor

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/wuqunyong/file_storage/pkg/errs"
	"github.com/wuqunyong/file_storage/pkg/msg"
	testdata "github.com/wuqunyong/file_storage/protobuf"
	"google.golang.org/protobuf/proto"
)

type ActorObjA struct {
	*Actor
}

func (actor *ActorObjA) Init() error {
	actor.Register("Func1", actor.Func1)
	actor.Register("Func2", actor.Func2)
	return nil
}

func (actor *ActorObjA) Func1(arg *testdata.Person, reply *testdata.Person) errs.CodeError {
	reply.Age += arg.Age
	reply.Name = "Func1"
	reply.Address = actor.actorId.ID
	fmt.Printf("inside value:%v\n", reply)

	return nil
}

func (actor *ActorObjA) Func2(arg *testdata.Person, reply *testdata.Person) errs.CodeError {
	reply.Age += arg.Age
	reply.Name = "Func1"
	reply.Address = actor.actorId.ID
	fmt.Printf("inside value:%v\n", reply)

	return errs.NewCodeError(errors.New("invalid"), 123)
}

func Must[T proto.Message](arg []byte, object T) T {
	err := proto.Unmarshal(arg, object)
	if err != nil {
		panic(err)
	}
	return object
}

func Test(t *testing.T) {

	engine := NewEngine(LocalLookupAddr)
	actorObj1 := &ActorObjA{
		Actor: NewActor("1", engine),
	}
	actorObj2 := &ActorObjA{
		Actor: NewActor("2", engine),
	}
	engine.SpawnActor(actorObj1)
	engine.SpawnActor(actorObj2)

	for i := 0; i < 1000; i++ {
		age := int32(i)
		person := &testdata.Person{Name: "小明", Age: age}
		request := actorObj1.Request(actorObj2.ActorId(), "Func1", person, 600*time.Second)
		fmt.Printf("request:%T, %v\n", request, request)
		obj, err := msg.GetResult[testdata.Person](request)
		if err != nil {
			t.Fatal("DecodeResponse", err)
		}
		fmt.Printf("obj:%T, %v\n", obj, obj)
		fmt.Printf("i:%v\n", i)
	}

	for i := 0; i < 1000; i++ {
		age := int32(i)
		person := &testdata.Person{Name: "小张", Age: age}
		request := actorObj2.Request(actorObj1.ActorId(), "Func2", person, 600*time.Second)
		fmt.Printf("request:%T, %v\n", request, request)
		obj, err := msg.GetResult[testdata.Person](request)
		if err != nil {
			sError := fmt.Sprintf("DecodeResponse: %v\n", err)
			t.Fatal(sError)
		}
		fmt.Printf("obj:%T, %v\n", obj, obj)
		fmt.Printf("i:%v\n", i)
	}

	time.Sleep(time.Duration(6) * time.Second)

	actorObj1.Stop()
	actorObj2.Stop()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}
