package componenttest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gopkg.in/mgo.v2/bson"

	mim "github.com/ONSdigital/dp-mongodb-in-memory"
	"github.com/cucumber/godog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoFeature is a struct containing an in-memory mongo database
type MongoFeature struct {
	Server   *mim.Server
	Client   mongo.Client
	Database *mongo.Database
}

// MongoOptions contains a set of options required to create a new MongoFeature
type MongoOptions struct {
	MongoVersion string
	DatabaseName string
}

// NewMongoFeature creates a new in-memory mongo database using the supplied options
func NewMongoFeature(mongoOptions MongoOptions) *MongoFeature {

	mongoServer, err := mim.Start(mongoOptions.MongoVersion)
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
	ctx.Step(`^the document with "([^"]*)" set to "([^"]*)" does not exist in the "([^"]*)" collection$`, m.theDocumentWithSetToDoesNotExistInTheCollection)
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

func (m *MongoFeature) theDocumentWithSetToDoesNotExistInTheCollection(key, value, collectionName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := m.Database.Collection(collectionName)
	var documentJSON interface{}

	err := collection.FindOne(ctx, bson.M{key: value}).Decode(&documentJSON)

	if err == mongo.ErrNoDocuments {
		return nil
	}

	if err != nil {
		return err
	}

	return errors.New(fmt.Sprintf("Document with property %s: %s was found in the collection", key, value))
}
