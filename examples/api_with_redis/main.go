package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"github.com/redis/go-redis/v9"
	"html"
	"log"
	"net/http"
	"os"
)

type Config struct {
	RedisUrl string
}

var cfg *Config

// Get returns the default config with any modifications through environment
// variables
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{
		RedisUrl: "localhost:6379",
	}

	return cfg, envconfig.Process("", cfg)
}

type Data struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func NewConfig() *Config {
	return &Config{
		RedisUrl: os.Getenv("REDIS_URL"),
	}
}

func ExampleHandler(w http.ResponseWriter, r *http.Request) {
	post := mux.Vars(r)
	key := post["id"]
	config := NewConfig()
	client := NewRedisClient(config.RedisUrl)

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
		w.Write(responseBody)

	} else {
		w.Header().Add("Content-Type", "text/html")
		response := fmt.Sprintf(`<value id="key">%s</value><value id="value">%s</value>`, html.EscapeString(resultBody.Key), html.EscapeString(resultBody.Value))
		w.Write([]byte(response))
	}
}

func ExampleDeleteHandler(w http.ResponseWriter, r *http.Request) {
	post := mux.Vars(r)
	key := post["id"]
	config := NewConfig()
	client := NewRedisClient(config.RedisUrl)

	err := client.Del(context.TODO(), key).Err()
	if err != nil {
		w.WriteHeader(404)
		fmt.Println(err.Error())
		return
	}

	w.WriteHeader(204)
}

func ExampleHealthHandler(w http.ResponseWriter, r *http.Request) {
	config := NewConfig()
	client := NewRedisClient(config.RedisUrl)
	ctx := context.Background()
	err := client.Ping(ctx).Err()

	if err != nil {
		w.WriteHeader(500)
		fmt.Println(err.Error())
		return
	} else {
		w.WriteHeader(200)
		return
	}
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
		Handler: newRouter(),
	}
}

func NewRedisClient(redisUrl string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     redisUrl,
		Password: "",
		DB:       0,
	})
}

func main() {
	server := NewServer()
	log.Fatal(http.ListenAndServe(":10000", server.Handler))
}
