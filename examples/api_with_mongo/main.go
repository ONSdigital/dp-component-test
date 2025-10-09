package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	MongoURL     string
	DatabaseName string
}

type Data struct {
	MongoID     string `bson:"_id" json:"_id"`
	ID          string `bson:"id" json:"id"`
	ExampleData string `bson:"example_data" json:"example_data"`
}

func NewConfig() *Config {
	return &Config{
		MongoURL:     os.Getenv("MONGO_URL"),
		DatabaseName: os.Getenv("DATABASE_NAME"),
	}
}

func ExampleHandler(w http.ResponseWriter, r *http.Request) {
	post := mux.Vars(r)
	config := NewConfig()
	client, _ := NewMongoClient(config.MongoURL)
	collection := client.Database(config.DatabaseName).Collection("datasets")
	var result Data

	err := collection.FindOne(context.Background(), bson.M{"id": post["id"]}).Decode(&result)
	if err != nil {
		w.WriteHeader(404)
		fmt.Println(err.Error())
		return
	}
	resultBody, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	if r.Header.Get("Accept") != "text/html" {
		w.Header().Add("Content-Type", "application/json")
		if _, err := w.Write(resultBody); err != nil {
			fmt.Printf("failed to write JSON response: %v\n", err)
		}
	} else {
		w.Header().Add("Content-Type", "text/html")
		response := fmt.Sprintf(
			`<value id="_id">%s</value><value id="id">%s</value><value id="example_data">%s</value>`,
			result.MongoID, result.ID, result.ExampleData)

		if _, err := w.Write([]byte(response)); err != nil {
			fmt.Printf("failed to write HTML response: %v\n", err)
		}
	}
}

func ExampleDeleteHandler(w http.ResponseWriter, r *http.Request) {
	post := mux.Vars(r)
	config := NewConfig()
	client, _ := NewMongoClient(config.MongoURL)
	collection := client.Database(config.DatabaseName).Collection("datasets")

	_, err := collection.DeleteOne(context.Background(), bson.M{"id": post["id"]})
	if err != nil {
		w.WriteHeader(404)
		return
	}

	w.WriteHeader(204)
}

func ExamplePutHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(200)
}

func newRouter() http.Handler {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/datasets/{id}", ExampleHandler).Methods("GET")
	router.HandleFunc("/datasets/{id}", ExampleDeleteHandler).Methods("DELETE")
	router.HandleFunc("/datasets/{id}", ExamplePutHandler).Methods("PUT")
	return router
}

func NewServer() *http.Server {
	return &http.Server{
		Handler:           newRouter(),
		ReadHeaderTimeout: 5 * time.Second,
	}
}

// NewMongoClient creates and returns a connected MongoDB client
func NewMongoClient(mongoURL string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(fmt.Sprintf("mongodb://%s", mongoURL)))
	if err != nil {
		return nil, err
	}

	// Ping to ensure the connection is established
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return client, nil
}

func main() {
	server := &http.Server{
		Addr:              ":10000",
		Handler:           NewServer().Handler,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Fatal(server.ListenAndServe())
}
