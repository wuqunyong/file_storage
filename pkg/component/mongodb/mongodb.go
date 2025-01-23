package mongodb

import (
	"context"

	"github.com/wuqunyong/file_storage/pkg/concepts"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoComponent struct {
	ctx     context.Context
	engine  concepts.IEngine
	configs map[string]*Config
	dbs     map[string][]*mongo.Database
}

type DBHashFunc func([]*mongo.Database) int

func NewMongoComponent(ctx context.Context, configs map[string]*Config) *MongoComponent {
	return &MongoComponent{
		ctx:     ctx,
		configs: configs,
		dbs:     make(map[string][]*mongo.Database),
	}
}

func (component *MongoComponent) Name() string {
	return ComponentName
}

func (component *MongoComponent) Priority() int32 {
	return 1
}

func (component *MongoComponent) SetEngine(engine concepts.IEngine) {
	component.engine = engine
}

func (component *MongoComponent) GetEngine() concepts.IEngine {
	return component.engine
}

func (component *MongoComponent) OnInit() error {
	for key, config := range component.configs {
		clients := make([]*mongo.Database, 0, config.ConnectNum)
		for index := 0; index < int(config.ConnectNum); index++ {
			db, err := NewMongoDB(component.ctx, config)
			if err != nil {
				return err
			}
			clients = append(clients, db)
		}
		component.dbs[key] = clients
	}
	return nil
}

func (component *MongoComponent) OnStart() {

}

func (component *MongoComponent) OnCleanup() {

}

func (component *MongoComponent) GetDatabase(name string, funObj ...DBHashFunc) *mongo.Database {
	clients, ok := component.dbs[name]
	if !ok {
		return nil
	}
	iSize := len(clients)
	if iSize <= 0 {
		return nil
	}

	if len(funObj) == 0 {
		return clients[0]
	}

	functor := funObj[0]
	if functor == nil {
		return nil
	}

	iIndex := functor(clients)
	iSelect := iIndex % iSize
	return clients[iSelect]
}
