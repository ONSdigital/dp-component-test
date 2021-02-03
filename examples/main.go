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

func HandleRequests(router *mux.Router) {
	router.HandleFunc("/example1", ExampleHandler1)
	router.HandleFunc("/example2", ExampleHandler2)
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	HandleRequests(router)

	log.Fatal(http.ListenAndServe(":10000", router))
}
