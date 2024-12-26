package mongodb

import (
	"context"

	"github.com/wuqunyong/file_storage/pkg/common/concepts"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoComponent struct {
	ctx    context.Context
	engine concepts.IEngine
	config *Config
	db     *mongo.Database
}

func NewMongoComponent(ctx context.Context, config *Config) *MongoComponent {
	return &MongoComponent{
		ctx:    ctx,
		config: config,
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
	db, err := NewMongoDB(component.ctx, component.config)
	if err != nil {
		return err
	}
	component.db = db
	return nil
}

func (component *MongoComponent) OnStart() {

}

func (component *MongoComponent) OnCleanup() {

}

func (component *MongoComponent) GetDatabase() *mongo.Database {
	return component.db
}
