package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/gorilla/mux"
)

func ExampleHandler1(w http.ResponseWriter, r *http.Request) {
	data := struct {
		ExampleType int `json:"example_type"`
	}{ExampleType: 1}

	w.Header().Set("Content-Type", "application/json")

	resp, _ := json.Marshal(data)

	fmt.Fprintf(w, "%s", string(resp))
}

func ExampleHandler2(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(403)
	fmt.Fprintf(w, "403 - Forbidden")
}

func ExampleHealthHandler(w http.ResponseWriter, r *http.Request) {
	var (
		checkTime       = time.Now()
		gitCommit       = "6584b786caac36b6214ffe04bf62f058d4021538"
		language        = "go"
		languageVersion = "go1.24.2"
		msgHealthy      = "redis is healthy"
		name            = "Redis"
		statusCode      = 200
		statusOK        = "OK"
		version         = "v1.2.3"
	)

	healthVersion := healthcheck.VersionInfo{
		BuildTime:       checkTime,
		GitCommit:       gitCommit,
		Version:         version,
		Language:        language,
		LanguageVersion: languageVersion,
	}
	healthCheck := componenttest.Check{
		Name:        name,
		Status:      statusOK,
		StatusCode:  statusCode,
		Message:     msgHealthy,
		LastChecked: &checkTime,
		LastSuccess: &checkTime,
	}
	responseBody := componenttest.HealthCheckTest{
		Status:  statusOK,
		Version: healthVersion,
		Uptime:  time.Duration(4),
		Checks:  []*componenttest.Check{&healthCheck},
	}

	healthResponse, err := json.Marshal(responseBody)
	if err != nil {
		w.WriteHeader(500)
		fmt.Println(err.Error())
		return
	} else {
		w.WriteHeader(200)
		w.Write(healthResponse)
		return
	}

}

func newRouter() http.Handler {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/example1", ExampleHandler1).Methods("GET")
	router.HandleFunc("/example2", ExampleHandler2).Methods("POST")
	router.HandleFunc("/health", ExampleHealthHandler).Methods("GET")

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
