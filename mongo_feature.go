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

// MongoFeature is a struct containing an in-memory mongo database
type MongoFeature struct {
	Server   *memongo.Server
	Client   mongo.Client
	Database *mongo.Database
}

// MongoOptions contains a set of options required to create a new MongoFeature
type MongoOptions struct {
	Port         int
	MongoVersion string
	// Logger       *log.Logger
	DatabaseName string
}

// NewMongoFeature creates a new in-memory mongo database using the supplied options
func NewMongoFeature(mongoOptions MongoOptions) *MongoFeature {

	opts := memongo.Options{
		Port:           mongoOptions.Port,
		MongoVersion:   mongoOptions.MongoVersion,
		StartupTimeout: time.Minute,
	}

	mongoServer, err := memongo.StartWithOptions(&opts)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoServer.URI()))
	if err != nil {
		panic(err)
	}

	database := client.Database(mongoOptions.DatabaseName)

	return &MongoFeature{
		Server:   mongoServer,
		Client:   *client,
		Database: database,
	}
}

// Reset is currently not implemented
func (m *MongoFeature) Reset() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	m.Database.Drop(ctx)
	return nil
}

// Close stops the in-memory mongo database
func (m *MongoFeature) Close() error {
	m.Server.Stop()
	return nil
}

func (m *MongoFeature) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^the following document exists in the "([^"]*)" collection:$`, m.TheFollowingDocumentExistsInTheCollection)
}

func (m *MongoFeature) TheFollowingDocumentExistsInTheCollection(collectionName string, document *godog.DocString) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := m.Database.Collection(collectionName)

	var documentJSON map[string]interface{}

	if err := json.Unmarshal([]byte(document.Content), &documentJSON); err != nil {
		return err
	}
	if _, err := collection.InsertOne(ctx, documentJSON); err != nil {
		return err
	}
	return nil
}
