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
	MongoUrl     string
	DatabaseName string
}

type Data struct {
	MongoId     string `bson:"_id" json:"_id"`
	ID          string `bson:"id" json:"id"`
	ExampleData string `bson:"example_data" json:"example_data"`
}

func NewConfig() *Config {
	return &Config{
		MongoUrl:     os.Getenv("MONGO_URL"),
		DatabaseName: os.Getenv("DATABASE_NAME"),
	}
}

func ExampleHandler(w http.ResponseWriter, r *http.Request) {
	post := mux.Vars(r)
	config := NewConfig()
	client, _ := NewMongoClient(config.MongoUrl)
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
		w.Write(resultBody)

	} else {
		w.Header().Add("Content-Type", "text/html")
		response := fmt.Sprintf(`<value id="_id">%s</value><value id="id">%s</value><value id="example_data">%s</value>`, result.MongoId, result.ID, result.ExampleData)
		w.Write([]byte(response))
	}
}

func ExampleDeleteHandler(w http.ResponseWriter, r *http.Request) {
	post := mux.Vars(r)
	config := NewConfig()
	client, _ := NewMongoClient(config.MongoUrl)
	collection := client.Database(config.DatabaseName).Collection("datasets")

	_, err := collection.DeleteOne(context.Background(), bson.M{"id": post["id"]})
	if err != nil {
		w.WriteHeader(404)
		return
	}

	w.WriteHeader(204)
}

func ExamplePutHandler(w http.ResponseWriter, r *http.Request) {
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
		Handler: newRouter(),
	}
}

func NewMongoClient(mongoUrl string) (*mongo.Client, error) {
	config := NewConfig()
	client, err := mongo.NewClient(options.Client().ApplyURI(config.MongoUrl))
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		return nil, err
	}

	return client, nil
}

func main() {
	server := NewServer()
	log.Fatal(http.ListenAndServe(":10000", server.Handler))
}
