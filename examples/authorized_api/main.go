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

func ExampleHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(201)
	fmt.Fprintf(w, "accepted")
}

func NewRouter() http.Handler {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/example1", ExampleHandler1).Methods(http.MethodGet)
	router.HandleFunc("/example3", MustAuthorize(ExampleHandler)).Methods(http.MethodPost)
	router.HandleFunc("/example4", MustBeIdentified(ExampleHandler)).Methods(http.MethodPost)

	return router
}

func main() {
	log.Fatal(http.ListenAndServe(":10000", NewRouter()))
}
