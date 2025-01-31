package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func ExampleHandler1(w http.ResponseWriter, r *http.Request) {
	data := struct {
		ExampleType int `json:"example_type"`
	}{ExampleType: 1}

	w.Header().Set("Content-Type", "application/json")

	resp, _ := json.Marshal(data)

	fmt.Fprintf(w, string(resp))
}

func ExampleHandler2(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(403)
	fmt.Fprintf(w, "403 - Forbidden")
}

func newRouter() http.Handler {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/example1", ExampleHandler1).Methods("GET")
	router.HandleFunc("/example2", ExampleHandler2).Methods("POST")

	return router
}

func NewServer() *http.Server {
	return &http.Server{
		Handler: newRouter(),
	}
}

func main() {
	server := NewServer()
	log.Fatal(http.ListenAndServe(":10000", server.Handler))
}
