package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	RedisURL string
}

var cfg *Config

// Get returns the default config with any modifications through environment
// variables
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{
		RedisURL: "localhost:6379",
	}

	return cfg, envconfig.Process("", cfg)
}

type Data struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func NewConfig() *Config {
	return &Config{
		RedisURL: os.Getenv("REDIS_URL"),
	}
}

func ExampleHandler(w http.ResponseWriter, r *http.Request) {
	post := mux.Vars(r)
	key := post["id"]
	config := NewConfig()
	client := NewRedisClient(config.RedisURL)

	result, err := client.Get(context.TODO(), key).Result()
	if err != nil {
		w.WriteHeader(404)
		fmt.Println(err.Error())
		return
	}

	resultBody := Data{
		Key:   key,
		Value: result,
	}

	responseBody, err := json.Marshal(resultBody)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	if r.Header.Get("Accept") != "text/html" {
		w.Header().Add("Content-Type", "application/json")
		if _, err := w.Write(responseBody); err != nil {
			log.Printf("failed to write JSON response: %v", err)
		}
	} else {
		w.Header().Add("Content-Type", "text/html")
		response := fmt.Sprintf(`<value id="key">%s</value><value id="value">%s</value>`,
			html.EscapeString(resultBody.Key), html.EscapeString(resultBody.Value))
		if _, err := w.Write([]byte(response)); err != nil {
			log.Printf("failed to write HTML response: %v", err)
		}
	}
}

func ExampleDeleteHandler(w http.ResponseWriter, r *http.Request) {
	post := mux.Vars(r)
	key := post["id"]
	config := NewConfig()
	client := NewRedisClient(config.RedisURL)

	err := client.Del(context.TODO(), key).Err()
	if err != nil {
		w.WriteHeader(404)
		fmt.Println(err.Error())
		return
	}

	w.WriteHeader(204)
}

func ExampleHealthHandler(w http.ResponseWriter, _ *http.Request) {
	config := NewConfig()
	client := NewRedisClient(config.RedisURL)
	ctx := context.Background()
	err := client.Ping(ctx).Err()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func newRouter() http.Handler {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/desserts/{id}", ExampleHandler).Methods("GET")
	router.HandleFunc("/desserts/{id}", ExampleDeleteHandler).Methods("DELETE")
	router.HandleFunc("/health", ExampleHealthHandler).Methods("GET")
	return router
}

func NewServer() *http.Server {
	return &http.Server{
		Handler:     newRouter(),
		ReadTimeout: 10 * time.Second,
	}
}

func NewRedisClient(redisURL string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: "",
		DB:       0,
	})
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
