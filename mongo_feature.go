package componenttest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	mim "github.com/ONSdigital/dp-mongodb-in-memory"
	"github.com/cucumber/godog"
	"go.mongodb.org/mongo-driver/bson"
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

// MongoDeletedDocs contains a list of counts for all deleted documents
// against a given collection of a mongo database
type MongoDeletedDocs struct {
	Database    string
	Count       int64
	Collections []MongoCollectionDeletedDocs
}

// MongoCollectionDeletedDocs contains the number of document deleted from collection
type MongoCollectionDeletedDocs struct {
	Name  string
	Count int64
}

// NewMongoFeature creates a new in-memory mongo database using the supplied options
func NewMongoFeature(mongoOptions MongoOptions) *MongoFeature {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoServer, err := mim.Start(ctx, mongoOptions.MongoVersion)
	if err != nil {
		panic(err)
	}

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

// ResetDatabase removes all data in all collections within database
func (m *MongoFeature) ResetDatabase(ctx context.Context, databaseName string) (*MongoDeletedDocs, error) {
	dbCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	collectionNames, err := m.Client.Database(databaseName).ListCollectionNames(dbCtx, bson.D{})
	if err != nil {
		return nil, err
	}

	deletedDocs, err := m.ResetCollections(ctx, databaseName, collectionNames)
	if err != nil {
		return nil, err
	}

	return deletedDocs, nil
}

// ResetCollections removes all data in all collections specified within database
func (m *MongoFeature) ResetCollections(ctx context.Context, databaseName string, collectionNames []string) (*MongoDeletedDocs, error) {
	collCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	deletedDocs := &MongoDeletedDocs{
		Database: databaseName,
	}

	for _, collectionName := range collectionNames {
		collection := m.Client.Database(databaseName).Collection(collectionName)

		deleteOp, err := collection.DeleteMany(collCtx, bson.D{})
		if err != nil {
			return deletedDocs, err
		}

		count := deleteOp.DeletedCount

		deletedDocs.Collections = append(deletedDocs.Collections, MongoCollectionDeletedDocs{
			Name:  collectionName,
			Count: count,
		})

		deletedDocs.Count = deletedDocs.Count + count
	}

	return deletedDocs, nil
}

// Close stops the in-memory mongo database
func (m *MongoFeature) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	m.Server.Stop(ctx)
	return nil
}

func (m *MongoFeature) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^remove all documents from database: "([^"]*)"`, m.RemoveAllDataFromDatabase)
	ctx.Step(`^remove all documents in the "([^"]*)" collection from the "([^"]*)" database`, m.RemoveAllDataFromCollection)
	ctx.Step(`^the following document exists in the "([^"]*)" collection:$`, m.TheFollowingDocumentExistsInTheCollection)
	ctx.Step(`^the document with "([^"]*)" set to "([^"]*)" does not exist in the "([^"]*)" collection$`, m.theDocumentWithSetToDoesNotExistInTheCollection)
}

func (m *MongoFeature) RemoveAllDataFromDatabase(databaseName string) error {
	deletedDocs, err := m.ResetDatabase(context.Background(), databaseName)
	if err != nil {
		return err
	}

	if deletedDocs.Count == 0 {
		return fmt.Errorf("no documents were deleted in database: %s", databaseName)
	}

	return nil
}

func (m *MongoFeature) RemoveAllDataFromCollection(collectionName, databaseName string) error {
	deletedDocs, err := m.ResetCollections(context.Background(), databaseName, []string{collectionName})
	if err != nil {
		return err
	}

	if deletedDocs.Count == 0 {
		return fmt.Errorf("no documents were deleted in database: %s", databaseName)
	}

	for i := range deletedDocs.Collections {
		if deletedDocs.Collections[i].Count == 0 {
			return fmt.Errorf("no documents were deleted for collection: %s in database: %s", collectionName, databaseName)
		}
	}

	return nil
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
