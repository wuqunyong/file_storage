package mongodb

import (
	"context"

	"github.com/wuqunyong/file_storage/pkg/actor"
	"github.com/wuqunyong/file_storage/pkg/common/concepts"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoComponent struct {
	*actor.Actor
	ctx    context.Context
	config *Config
	db     *mongo.Database
}

func NewMongoComponent(engine concepts.IEngine, ctx context.Context, config *Config) *MongoComponent {
	return &MongoComponent{
		Actor:  actor.NewActor("mongodb.1", engine),
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

func (component *MongoComponent) OnInit() error {
	db, err := NewMongoDB(component.ctx, component.config)
	if err != nil {
		return err
	}
	component.db = db
	return nil
}

func (component *MongoComponent) OnCleanup() {

}

func (component *MongoComponent) GetDatabase() *mongo.Database {
	return component.db
}
