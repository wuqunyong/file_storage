package test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
	"unsafe"

	"github.com/wuqunyong/file_storage/pkg/actor"
	"github.com/wuqunyong/file_storage/pkg/component/mongodb"
	"github.com/wuqunyong/file_storage/pkg/component/tcpserver"
	"github.com/wuqunyong/file_storage/pkg/concepts"
	"github.com/wuqunyong/file_storage/pkg/easytcp"
	"github.com/wuqunyong/file_storage/pkg/errs"
	logger "github.com/wuqunyong/file_storage/pkg/logger"
	"github.com/wuqunyong/file_storage/pkg/msg"
	"github.com/wuqunyong/file_storage/pkg/tick"
	testdata "github.com/wuqunyong/file_storage/proto"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/protobuf/proto"
)

type ActorObjA struct {
	*actor.Actor
	inited atomic.Bool
	id     int
}

func (actor *ActorObjA) OnInit() error {
	if actor.inited.Load() {
		return errors.New("duplicate init")
	}
	actor.Register(1, actor.Func1)
	actor.Register(2, actor.Func2)
	actor.Register(103, actor.Func3)
	actor.inited.Store(true)
	return nil
}

func (actor *ActorObjA) OnShutdown() {

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

	actor.id = 10000
	// funcCb := func(id uint64) {
	// 	fmt.Println("Id:", id)

	// 	actor.id++
	// 	mongoComponent := actor.GetEngine().GetComponent(mongodb.ComponentName).(*mongodb.MongoComponent)
	// 	blackTable, err := NewBlackMongo(mongoComponent.GetDatabase())
	// 	if err != nil {
	// 		return
	// 	}
	// 	var blacks []*BlackObj
	// 	blacks = append(blacks, &BlackObj{
	// 		OwnerUserID: strconv.Itoa(actor.id),
	// 		CreateTime:  time.Now(),
	// 	})
	// 	blackTable.Create(context.Background(), blacks)
	// }
	// item := tick.NewPersistentTimer(2, 15*time.Second, func(id uint64) {
	// 	fmt.Println("Id:", id)
	// 	funcCb(id)
	// })
	// actor.GetTimerQueue().Restore(item)

	iTimestamp := GetNextMinuteTimestamp()
	item := tick.NewPersistentTimerWithExpires(3, iTimestamp, 60*time.Second, func(id uint64) {
		fmt.Println("offset Id:", id, time.Now())
	})
	actor.GetTimerQueue().Restore(item)
	// actor.GetTimerQueue().Restore(item1)

	return nil
}

func (actor *ActorObjA) Func2(ctx context.Context, arg *testdata.Person, reply *testdata.Person) errs.CodeError {
	reply.Age += arg.Age
	reply.Name = "Func1"
	reply.Address = actor.ActorId().ID
	fmt.Printf("inside value:%v\n", reply)

	return errs.NewCodeError(errors.New("invalid"), 123)
}

func (actor *ActorObjA) Func3(ctx context.Context, arg *testdata.MSG_NOTICE_INSTANCE) {
	fmt.Printf("inside value arg:%v\n", arg)
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

const TimeOffset = 8 * 3600

func GetCurDayZeroTimestamp() int64 {
	now := time.Now()
	timeStr := now.Format("2006-01-02")
	t, _ := time.Parse("2006-01-02", timeStr)
	return t.Unix() - TimeOffset
}

func GetNextHourOffset() int64 {
	now := time.Now()
	timeStr := now.Format("2006-01-02 15")
	t, _ := time.Parse("2006-01-02 15", timeStr)
	iPass := (now.Unix() - t.Unix()) % 3600
	iValue := 3600 - iPass
	return iValue
}

func GetNextMinuteTimestamp() int64 {
	now := time.Now()
	timeStr := now.Format("2006-01-02 15:04")
	t, _ := time.Parse("2006-01-02 15:04", timeStr)
	t = t.Add(-TimeOffset * time.Second)
	t = t.Add(60 * time.Second)
	return t.UnixNano()
}

type ChildActorObjA struct {
	concepts.ChildActor
}

func (a *ChildActorObjA) OnInit() error {
	logger.Log(logger.InfoLevel, "ChildActorObjA OnInit", "actorId", a.ActorId().String(), "address", uintptr(unsafe.Pointer(a)))
	return nil
}

func (a *ChildActorObjA) OnShutdown() {
	logger.Log(logger.InfoLevel, "ChildActorObjA OnShutdown", "actorId", a.ActorId().String())
}

func TestClient(t *testing.T) {
	iZero := GetNextMinuteTimestamp()
	fmt.Println(iZero)

	engine := actor.NewEngine(0, 1, 1001, "nats://127.0.0.1:4222")

	var mongoConfig mongodb.Config
	mongoConfig.Uri = "mongodb://admin:123456@127.0.0.1:27018"
	mongoConfig.Database = "vcity"
	mongoConfig.ConnectNum = 2

	configs := map[string]*mongodb.Config{}
	configs["test"] = &mongoConfig

	component := mongodb.NewMongoComponent(context.Background(), configs)
	engine.MustAddComponent(component)

	engine.MustInit()
	actorObj1 := &ActorObjA{
		Actor: actor.NewActor("1", engine),
	}
	actorObj2 := &ActorObjA{
		Actor: actor.NewActor("2", engine),
	}
	for i := 0; i < 3; i++ {
		childObj := &ChildActorObjA{}
		actorObj2.SpawnChild(childObj, fmt.Sprintf("ppp:%d", i))
		// childObj.OnShutdown()

		logger.Log(logger.InfoLevel, "aaaa", "actorId", childObj.ActorId().String(), "childObj address", childObj.GetObjAddress())

		findObj := engine.GetRegistry().GetByID(childObj.ActorId().ID)
		logger.Log(logger.InfoLevel, "bbbb", "actorId", findObj.ActorId().String(), "childObj address", findObj.GetObjAddress())
		// findObj.OnShutdown()

		// findObj.Stop()

		// findObj.OnShutdown()
	}

	// engine.SpawnActor(actorObj1)
	engine.SpawnActor(actorObj2)
	engine.Start()

	// for i := 200; i < 300; i++ {
	// 	actorObj2.SpawnChild(&ChildActorObjA{}, fmt.Sprintf("ppp:%d", i))
	// }

	defer engine.Stop()

	time.Sleep(time.Duration(3) * time.Second)

	mongoComponent := engine.GetComponent(mongodb.ComponentName).(*mongodb.MongoComponent)
	blackTable, err := NewBlackMongo(mongoComponent.GetDatabase("test"))
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
	// request := actorObj1.Request(concepts.NewActorId("engine.test.server.1.2.345", "1"), "Func1", person)
	// obj, err := msg.GetResult[testdata.Person](request)
	obj, err := actor.SendRequest[testdata.Person](actorObj1, concepts.NewActorId("engine.0.1.1001.server", "1"), 1, person)
	if err != nil {
		//t.Fatal("DecodeResponse1", err)
	}
	fmt.Printf("obj:%T, %v\n", obj, obj)

	request := actorObj1.Request(actorObj2.ActorId(), 1, person)
	if reflect.TypeOf(request) == emptyMsgType {
		fmt.Printf("Same\n")
	}
	fmt.Printf("request:%T, %v\n", request, request)
	obj, err = msg.GetResult[testdata.Person](request)
	if err != nil {
		t.Fatal("DecodeResponse2", err)
	}
	fmt.Printf("obj2:%T, %v\n", obj, obj)

	time.Sleep(30 * time.Second)
	fmt.Println("over")
	// common.WaitForShutdown()
}

func TestServer(t *testing.T) {

	engine := actor.NewEngine(0, 1, 1001, "nats://127.0.0.1:4222")
	engine.MustInit()
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

func NewPBClientOption() *easytcp.ClientOption {
	packer := tcpserver.NewPBPacker()
	codec := &easytcp.ProtobufCodec{}
	return &easytcp.ClientOption{Packer: packer,
		Codec: codec}
}

type Options struct {
	// Addrs sets the addresses of auth
	Addrs []string
}

type Option func(o *Options)

// Addrs is the auth addresses to use.
func Addrs(addrs ...string) Option {
	return func(o *Options) {
		//o.Addrs = addrs
		o.Addrs = append(o.Addrs, addrs...)
	}
}

func TestTCPClient(t *testing.T) {

	optObj := &Options{}
	optFunc := Addrs()
	optFunc(optObj)

	if optObj.Addrs == nil {
		fmt.Printf("%v", optObj)
	}

	client := easytcp.NewClient(NewPBClientOption())
	err := client.Dial("127.0.0.1:16007")
	if err != nil {
		return
	}

	reqData := &testdata.AccountLoginRequest{}
	reqData.AccountId = 1234
	err = client.SendRequest(1001, reqData)
	if err != nil {
		return
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}
