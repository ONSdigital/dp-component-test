package featuretest

import (
	"context"
	"encoding/json"
	"time"

	"github.com/benweissmann/memongo"
	"github.com/cucumber/godog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoCapability is a struct containing an in-memory mongo database
type MongoCapability struct {
	Server   *memongo.Server
	Client   mongo.Client
	Database *mongo.Database
}

// MongoOptions contains a set of options required to create a new MongoCapability
type MongoOptions struct {
	Port         int
	MongoVersion string
	// Logger       *log.Logger
	DatabaseName string
}

// NewMongoCapability creates a new in-memory mongo database using the supplied options
func NewMongoCapability(mongoOptions MongoOptions) (*MongoCapability, error) {

	opts := memongo.Options{
		Port:           mongoOptions.Port,
		MongoVersion:   mongoOptions.MongoVersion,
		StartupTimeout: time.Second * 10,
		// Logger:         mongoOptions.Logger,
	}

	mongoServer, err := memongo.StartWithOptions(&opts)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoServer.URI()))
	if err != nil {
		return nil, err
	}

	database := client.Database(mongoOptions.DatabaseName)

	return &MongoCapability{
		Server:   mongoServer,
		Client:   *client,
		Database: database,
	}, nil
}

// Reset is currently not implemented
func (m *MongoCapability) Reset() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	m.Database.Drop(ctx)
	return nil
}

// Close stops the in-memory mongo database
func (m *MongoCapability) Close() error {
	m.Server.Stop()
	return nil
}

func (m *MongoCapability) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^the following document exists in the "([^"]*)" collection:$`, m.TheFollowingDocumentExistsInTheCollection)
}

func (m *MongoCapability) TheFollowingDocumentExistsInTheCollection(collectionName string, document *godog.DocString) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := m.Database.Collection(collectionName)

	var documentJson map[string]interface{}

	if err := json.Unmarshal([]byte(document.Content), &documentJson); err != nil {
		return err
	}
	if _, err := collection.InsertOne(ctx, documentJson); err != nil {
		return err
	}
	return nil
}
