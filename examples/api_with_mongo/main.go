package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
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

// NewMongoClient creates and returns a connected MongoDB client
func NewMongoClient(mongoURL string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		return nil, err
	}

	// Ping to ensure the connection is established
	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, err
	}

	return client, nil
}

// ExampleHandler returns an http.HandlerFunc that uses the provided mongo client.
func ExampleHandler(client *mongo.Client, dbName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		post := mux.Vars(r)
		collection := client.Database(dbName).Collection("datasets")
		var result Data

		err := collection.FindOne(context.Background(), bson.M{"id": post["id"]}).Decode(&result)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(os.Stderr, err.Error())
			return
		}
		resultBody, err := json.Marshal(result)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if r.Header.Get("Accept") != "text/html" {
			w.Header().Set("Content-Type", "application/json")
			if _, err := w.Write(resultBody); err != nil {
				fmt.Fprintf(os.Stderr, "failed to write JSON response: %v\n", err)
			}
		} else {
			w.Header().Set("Content-Type", "text/html")
			response := fmt.Sprintf(
				`<value id="_id">%s</value><value id="id">%s</value><value id="example_data">%s</value>`,
				result.MongoID, result.ID, result.ExampleData)

			if _, err := w.Write([]byte(response)); err != nil {
				fmt.Fprintf(os.Stderr, "failed to write HTML response: %v\n", err)
			}
		}
	}
}

func ExampleDeleteHandler(client *mongo.Client, dbName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		post := mux.Vars(r)
		collection := client.Database(dbName).Collection("datasets")

		_, err := collection.DeleteOne(context.Background(), bson.M{"id": post["id"]})
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func ExamplePutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

// NewServerHandler returns an http.Handler configured with the given client and database name.
func NewServerHandler(client *mongo.Client, dbName string) http.Handler {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/datasets/{id}", ExampleHandler(client, dbName)).Methods("GET")
	router.HandleFunc("/datasets/{id}", ExampleDeleteHandler(client, dbName)).Methods("DELETE")
	router.HandleFunc("/datasets/{id}", ExamplePutHandler()).Methods("PUT")
	return router
}

// NewServer creates an *http.Server configured with the given handler.
func NewServer(client *mongo.Client, dbName string, addr string) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           NewServerHandler(client, dbName),
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func main() {
	cfg := NewConfig()
	if cfg.MongoURL == "" || cfg.DatabaseName == "" {
		log.Fatal("MONGO_URL and DATABASE_NAME must be set")
	}

	client, err := NewMongoClient(cfg.MongoURL)
	if err != nil {
		log.Fatalf("failed to connect to mongo: %v", err)
	}
	// Ensure client is disconnected on shutdown
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = client.Disconnect(ctx)
	}()

	server := NewServer(client, cfg.DatabaseName, ":10000")

	// Graceful shutdown handling
	idleConnsClosed := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
		close(idleConnsClosed)
	}()

	log.Printf("starting server on %s\n", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("ListenAndServe error: %v", err)
	}

	<-idleConnsClosed
}
