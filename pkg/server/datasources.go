package server

import (
	"context"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spring-financial-group/peacock/pkg/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type DataSources struct {
	cfg           *config.DataSources
	MongoDBClient *mongo.Client
}

// NewDataSources creates and initialises the data sources for the server.
func NewDataSources(cfg *config.DataSources) (*DataSources, error) {
	log.Info("Initialising data sources")
	ds := &DataSources{
		cfg: cfg,
	}
	if err := ds.initialiseMongoDB(); err != nil {
		return nil, errors.Wrap(err, "failed to initialise mongodb")
	}
	return ds, nil
}

func (ds *DataSources) initialiseMongoDB() error {
	client, err := mongo.NewClient(options.Client().ApplyURI(ds.cfg.MongoDB.ConnectionString))
	if err != nil {
		return errors.Wrap(err, "failed to create client")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to connect")
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to ping")
	}

	ds.MongoDBClient = client
	log.Info("Successfully initialised mongo")
	return nil
}
