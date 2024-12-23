package mongodb

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	Uri         string
	Address     []string
	Database    string
	Username    string
	Password    string
	MaxPoolSize int
	MaxRetry    int
}

// CheckMongo tests the MongoDB connection without retries.
func Check(ctx context.Context, config *Config) error {
	if err := config.ValidateAndSetDefaults(); err != nil {
		return err
	}

	clientOpts := options.Client().ApplyURI(config.Uri)
	mongoClient, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return fmt.Errorf("%s %s %s %s %s %s %s %s", err, "MongoDB connect failed", "URI", config.Uri, "Database", config.Database, "MaxPoolSize", config.MaxPoolSize)
	}

	defer func() {
		if err := mongoClient.Disconnect(ctx); err != nil {
			_ = mongoClient.Disconnect(ctx)
		}
	}()

	if err = mongoClient.Ping(ctx, nil); err != nil {
		return fmt.Errorf("%s %s %s %s %s %s %s %s", err, "MongoDB ping failed", "URI", config.Uri, "Database", config.Database, "MaxPoolSize", config.MaxPoolSize)
	}

	return nil
}

// ValidateAndSetDefaults validates the configuration and sets default values.
func (c *Config) ValidateAndSetDefaults() error {
	if c.Uri == "" && len(c.Address) == 0 {
		return errors.New("either Uri or Address must be provided")
	}
	if c.Database == "" {
		return errors.New("database is required")
	}
	if c.MaxPoolSize <= 0 {
		c.MaxPoolSize = defaultMaxPoolSize
	}
	if c.MaxRetry <= 0 {
		c.MaxRetry = defaultMaxRetry
	}
	if c.Uri == "" {
		c.Uri = buildMongoURI(c)
	}
	return nil
}
