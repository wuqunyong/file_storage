package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/wuqunyong/file_storage/pkg/actor"
	"github.com/wuqunyong/file_storage/pkg/component/mongodb"
	"github.com/wuqunyong/file_storage/pkg/logger"
	testdata "github.com/wuqunyong/file_storage/protobuf"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type ItemTbl interface {
	Create(ctx context.Context, items []*testdata.Item) (err error)
	Find(ctx context.Context) (items []bson.D, err error)
}

func NewItemTblMongo(db *mongo.Database) (ItemTbl, error) {
	coll := db.Collection("item_test")
	_, err := coll.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.D{
			{Key: "Name", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return nil, err
	}
	return &ItemTblMgo{coll: coll}, nil
}

type ItemTblMgo struct {
	coll *mongo.Collection
}

func (b *ItemTblMgo) Create(ctx context.Context, items []*testdata.Item) (err error) {
	bsonArray := make([]bson.D, 0)
	for _, value := range items {
		jsonObj, err := marshalProtoToJSON(value)
		if err != nil {
			//todo
		}
		var doc bson.D
		err = bson.UnmarshalExtJSON([]byte(jsonObj), true, &doc)
		if err != nil {
			//todo
		}
		doc = append(doc, bson.E{Key: "_id", Value: value.Id})
		bsonArray = append(bsonArray, doc)
		fmt.Printf("doc:%T, %v\n", doc, doc)

	}
	return mongodb.InsertMany(ctx, b.coll, bsonArray)
}

func (b *ItemTblMgo) Find(ctx context.Context) (items []bson.D, err error) {
	return mongodb.Find[bson.D](ctx, b.coll, bson.D{})
}

func marshalProtoToJSON(item *testdata.Item) (string, error) {
	marshalOpts := protojson.MarshalOptions{
		Multiline:       true,
		Indent:          "  ",
		UseProtoNames:   true,
		EmitUnpopulated: true,
	}

	jsonBytes, err := marshalOpts.Marshal(item)
	if err != nil {
		return "", fmt.Errorf("failed to marshal item to JSON: %w", err)
	}

	return string(jsonBytes), nil
}

func unmarshalJSONToProto(jsonStr string) (*testdata.Item, error) {
	unmarshalOpts := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}

	var item testdata.Item
	if err := unmarshalOpts.Unmarshal([]byte(jsonStr), &item); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to item: %w", err)
	}

	return &item, nil
}

func TestMongo(t *testing.T) {
	logger.CreateLogger("log.txt")

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
	engine.Start()
	defer engine.Stop()

	time.Sleep(time.Duration(3) * time.Second)

	mongoComponent := engine.GetComponent(mongodb.ComponentName).(*mongodb.MongoComponent)

	itemTable, err := NewItemTblMongo(mongoComponent.GetDatabase("test"))
	if err != nil {
		t.Fatal("init err", err)
	}
	var items []*testdata.Item
	itemObj1 := &testdata.Item{
		Id:   12,
		Name: "obj56",
		Age:  12,
		Type: testdata.ItemType_IT_Phone,
	}
	deskObj := make(map[int32]*testdata.Desk)
	deskObj[1] = &testdata.Desk{
		Num: 5555,
	}
	deskObj[2] = &testdata.Desk{
		Num: 666,
	}
	itemObj1.Msg = &testdata.Item_PhoneInfo{
		PhoneInfo: &testdata.Phone{
			Num:   123,
			Price: 345,
			Data:  deskObj,
		},
	}
	items = append(items, itemObj1)

	jsonItem1, err := proto.Marshal(itemObj1)
	if err != nil {
		//todo
	}
	itemObj11 := &testdata.Item{}
	err = proto.Unmarshal(jsonItem1, itemObj11)
	if err != nil {
		//todo
	}
	fmt.Printf("obj:%T, %v\n", itemObj11, itemObj11)

	json1, err := marshalProtoToJSON(itemObj1)
	if err != nil {
		//todo
	}
	fmt.Printf("json1:%T, %v\n", json1, json1)

	proto1, err := unmarshalJSONToProto(json1)
	if err != nil {
		//todo
	}
	fmt.Printf("proto1:%T, %v\n", proto1, proto1)

	itemObj2 := &testdata.Item{
		Id:   123,
		Name: "obj6",
		Age:  123,
		Type: testdata.ItemType_IT_Watch,
	}
	itemObj2.Msg = &testdata.Item_WatchInfo{
		WatchInfo: &testdata.Watch{
			Name: "test",
		},
	}
	items = append(items, itemObj2)
	jsonItem2, err := proto.Marshal(itemObj2)
	if err != nil {
		//todo
	}
	jsonItem12 := &testdata.Item{}
	err = proto.Unmarshal(jsonItem2, jsonItem12)
	if err != nil {
		//todo
	}
	fmt.Printf("obj:%T, %v\n", jsonItem12, jsonItem12)

	itemTable.Create(context.Background(), items)

	findResult, err := itemTable.Find(context.Background())
	if err != nil {
		//todo
	}
	fmt.Printf("obj:%T, %v\n", findResult, findResult)
	for _, value := range findResult {
		fmt.Printf("value:%T, %v\n", value, value)

		jsonByte, err := bson.MarshalExtJSON(value, false, false)
		if err != nil {
			//todo
		}

		proto1, err := unmarshalJSONToProto(string(jsonByte))
		if err != nil {
			//todo
		}

		fmt.Printf("proto1:%T, %v\n", proto1, proto1)
	}
}
