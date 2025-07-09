package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

type Config struct {
	authorizationServiceURL string
}

func NewConfig() *Config {
	return &Config{
		authorizationServiceURL: os.Getenv("AUTH_URL"),
	}
}

func ExampleHandler2(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(500)
}

func ExampleHandler1(w http.ResponseWriter, _ *http.Request) {
	data := struct {
		ExampleType int `json:"example_type"`
	}{ExampleType: 1}

	w.Header().Set("Content-Type", "application/json")

	resp, _ := json.Marshal(data)

	fmt.Fprintf(w, "%s", string(resp))
}

func ExampleHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(201)
	fmt.Fprintf(w, "accepted")
}

func NewRouter() http.Handler {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/example1", ExampleHandler1).Methods(http.MethodGet)
	router.HandleFunc("/example3", MustAuthorize(ExampleHandler)).Methods(http.MethodPost)
	router.HandleFunc("/example4", MustBeIdentified(ExampleHandler)).Methods(http.MethodPost)
	router.HandleFunc("/example5", ZebedeeMustAuthorize(ExampleHandler2)).Methods(http.MethodPost)
	router.HandleFunc("/example6", ZebedeeMustPermitUser(ExampleHandler2)).Methods(http.MethodPost)

	return router
}

func main() {
	server := &http.Server{
		Addr:              ":10000",
		Handler:           NewRouter(),
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Fatal(server.ListenAndServe())
}
