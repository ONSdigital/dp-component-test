package main

import (
	"bytes"
	"encoding/json"
	"errors"
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

func ExampleHandler4(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Status string `json:"status"`
	}{Status: "ok"}

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
	router.HandleFunc("/example3", MustAuthenticate(ExampleHandler3)).Methods(http.MethodPost)
	router.HandleFunc("/example4", MustHavePostPermission(ExampleHandler4)).Methods(http.MethodPost)

	return router
}

func main() {
	log.Fatal(http.ListenAndServe(":10000", NewRouter()))
}

func MustAuthenticate(h func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	checkauth := func(w http.ResponseWriter, r *http.Request) {
		err := validateAuth("some token")
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		h(w, r)
	}
	return checkauth
}

func validateAuth(token string) error {
	config := NewConfig()
	response, err := http.Get(config.authorizationServiceUrl + "/identity")
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		return errors.New("401 - Unauthorized")
	}

	return nil
}

func MustHavePostPermission(h func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	checkauth := func(w http.ResponseWriter, r *http.Request) {
		err := validatePermission("some token", "POST")
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		h(w, r)
	}
	return checkauth
}

func validatePermission(token, permission string) error {
	config := NewConfig()
	response, err := http.Post(
		config.authorizationServiceUrl+"/permissions",
		"application/json",
		bytes.NewReader([]byte(`{"token": "`+token+`"}`)),
	)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		return errors.New("401 - Unauthorized")
	}

	return nil
}
