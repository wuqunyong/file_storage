package actor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"reflect"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/wuqunyong/file_storage/pkg/actor"
	"github.com/wuqunyong/file_storage/pkg/common/concepts"
	"github.com/wuqunyong/file_storage/pkg/component/mongodb"
	"github.com/wuqunyong/file_storage/pkg/errs"
	"github.com/wuqunyong/file_storage/pkg/logger"
	"github.com/wuqunyong/file_storage/pkg/msg"
	testdata "github.com/wuqunyong/file_storage/protobuf"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/protobuf/proto"
)

type ActorObjA struct {
	*actor.Actor
	inited atomic.Bool
}

func (actor *ActorObjA) Init() error {
	if actor.inited.Load() {
		return errors.New("duplicate init")
	}
	actor.Register("Func1", actor.Func1)
	actor.Register("Func2", actor.Func2)
	actor.inited.Store(true)
	return nil
}

func (actor *ActorObjA) Func1(ctx context.Context, arg *testdata.Person, reply *testdata.Person) errs.CodeError {
	reply.Age += arg.Age
	reply.Name = "Func1"
	reply.Address = actor.ActorId().ID
	fmt.Printf("inside value:%v\n", reply)

	for i := 0; i < 10; i++ {
		func1 := func() {
			if i == 5 {
				panic("in func1 5 =======")
			}
			fmt.Println("task func1", i)
		}
		actor.PostTask(func1)
	}

	return nil
}

func (actor *ActorObjA) Func2(ctx context.Context, arg *testdata.Person, reply *testdata.Person) errs.CodeError {
	reply.Age += arg.Age
	reply.Name = "Func1"
	reply.Address = actor.ActorId().ID
	fmt.Printf("inside value:%v\n", reply)

	return errs.NewCodeError(errors.New("invalid"), 123)
}

func (actor *ActorObjA) Func3() {
	fmt.Println("func3")
}

func Must[T proto.Message](arg []byte, object T) T {
	err := proto.Unmarshal(arg, object)
	if err != nil {
		panic(err)
	}
	return object
}

var emptyMsgType = reflect.TypeOf(&msg.MsgReq{})

func SwitchFunc(obj any) {
	switch msg := obj.(type) {
	case func():
		msg()
	default:
		fmt.Printf("未知型")
	}
}

var asciiSpace = [256]uint8{'\t': 1, '\n': 1, '\v': 1, '\f': 1, '\r': 1, ' ': 1}

type Black interface {
	Create(ctx context.Context, blacks []*BlackObj) (err error)
}

func NewBlackMongo(db *mongo.Database) (Black, error) {
	coll := db.Collection("black")
	_, err := coll.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.D{
			{Key: "owner_user_id", Value: 1},
			{Key: "block_user_id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return nil, err
	}
	return &BlackMgo{coll: coll}, nil
}

type BlackMgo struct {
	coll *mongo.Collection
}

type BlackObj struct {
	OwnerUserID    string    `bson:"owner_user_id"`
	BlockUserID    string    `bson:"block_user_id"`
	CreateTime     time.Time `bson:"create_time"`
	AddSource      int32     `bson:"add_source"`
	OperatorUserID string    `bson:"operator_user_id"`
	Ex             string    `bson:"ex"`
}

func (b *BlackMgo) Create(ctx context.Context, blacks []*BlackObj) (err error) {
	return mongodb.InsertMany(ctx, b.coll, blacks)
}

func TestClient(t *testing.T) {
	logger.CreateLogger("log.txt")

	fmt.Printf("%v, %T, %s\n", asciiSpace, asciiSpace, reflect.TypeOf(asciiSpace).Name())

	engine := actor.NewEngine("test", "1.2.3", true, "nats://127.0.0.1:4222")

	var mongoConfig mongodb.Config
	mongoConfig.Uri = "mongodb://admin:123456@127.0.0.1:27018"
	mongoConfig.Database = "vcity"
	component := mongodb.NewMongoComponent(engine, context.Background(), &mongoConfig)
	engine.AddComponent(component)

	err := engine.Init()
	if err != nil {
		t.Fatal("init err", err)
	}
	actorObj1 := &ActorObjA{
		Actor: actor.NewActor("1", engine),
	}
	actorObj2 := &ActorObjA{
		Actor: actor.NewActor("2", engine),
	}
	engine.SpawnActor(actorObj1)
	engine.SpawnActor(actorObj2)
	engine.Start()

	time.Sleep(time.Duration(3) * time.Second)

	mongoComponent := engine.GetComponent(mongodb.ComponentName).(*mongodb.MongoComponent)
	blackTable, err := NewBlackMongo(mongoComponent.GetDatabase())
	if err != nil {
		t.Fatal("init err", err)
	}
	var blacks []*BlackObj
	blacks = append(blacks, &BlackObj{
		OwnerUserID: "1",
		CreateTime:  time.Now(),
	}, &BlackObj{
		OwnerUserID: "2",
		CreateTime:  time.Now(),
	})
	blackTable.Create(context.Background(), blacks)

	person := &testdata.Person{Name: "小明", Age: 18}
	request := actorObj1.Request(concepts.NewActorId("engine.test.server.1.2.345", "1"), "Func1", person)
	if reflect.TypeOf(request) == emptyMsgType {
		slog.Info("test", "type", "same")
	}

	fmt.Printf("request:%T, %v\n", request, request)
	obj, err := msg.GetResult[testdata.Person](request)
	if err != nil {
		//t.Fatal("DecodeResponse1", err)
	}
	fmt.Printf("obj:%T, %v\n", obj, obj)

	request = actorObj1.Request(actorObj2.ActorId(), "Func1", person)
	if reflect.TypeOf(request) == emptyMsgType {
		fmt.Printf("Same\n")
	}
	fmt.Printf("request:%T, %v\n", request, request)
	obj, err = msg.GetResult[testdata.Person](request)
	if err != nil {
		t.Fatal("DecodeResponse2", err)
	}
	fmt.Printf("obj2:%T, %v\n", obj, obj)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}

func TestServer(t *testing.T) {

	engine := actor.NewEngine("test", "1.2.345", true, "nats://127.0.0.1:4222")
	engine.Init()
	actorObj1 := &ActorObjA{
		Actor: actor.NewActor("1", engine),
	}
	engine.SpawnActor(actorObj1)
	engine.Start()

	// time.Sleep(time.Duration(6) * time.Second)

	// person := &testdata.Person{Name: "小明", Age: 18}
	// request := actorObj1.Request(concepts.NewActorId("identify.server.1.2.3", "12"), "Func1", person, 600*time.Second)
	// fmt.Printf("request:%T, %v\n", request, request)
	// obj, err := msg.GetResult[testdata.Person](request)
	// if err != nil {
	// 	t.Fatal("DecodeResponse", err)
	// }
	// fmt.Printf("obj:%T, %v\n", obj, obj)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}
