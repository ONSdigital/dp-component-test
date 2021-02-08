package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

type Config struct {
	authorizationServiceUrl string
}

func NewConfig() *Config {
	return &Config{
		authorizationServiceUrl: os.Getenv("AUTH_URL"),
	}
}

func ExampleHandler1(w http.ResponseWriter, r *http.Request) {
	data := struct {
		ExampleType int `json:"example_type"`
	}{ExampleType: 1}

	w.Header().Set("Content-Type", "application/json")

	resp, _ := json.Marshal(data)

	fmt.Fprintf(w, string(resp))
}

func ExampleHandler3(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(403)
	fmt.Fprintf(w, "403 - Forbidden")
}

func NewRouter() http.Handler {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/example1", ExampleHandler1).Methods(http.MethodGet)
	router.HandleFunc("/example3", MustAuthorize(ExampleHandler3)).Methods(http.MethodPost)

	return router
}

func main() {
	log.Fatal(http.ListenAndServe(":10000", NewRouter()))
}
