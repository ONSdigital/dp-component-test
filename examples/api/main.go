package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type EmbeddedTimestamp struct {
	InnerTimestamp time.Time `json:"inner_timestamp"`
}

type DynamicResponse struct {
	Timestamp     time.Time         `json:"timestamp"`
	ID            string            `json:"id"`
	EmbeddedField EmbeddedTimestamp `json:"embedded"`
	URL           string            `json:"url"`
}

func ExampleHandler1(w http.ResponseWriter, _ *http.Request) {
	data := struct {
		ExampleType int `json:"example_type"`
	}{ExampleType: 1}

	w.Header().Set("Content-Type", "application/json")

	resp, _ := json.Marshal(data)

	fmt.Fprintf(w, "%s", string(resp))
}

func ExampleHandler2(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(403)
	fmt.Fprintf(w, "403 - Forbidden")
}

func ExampleHealthHandler(w http.ResponseWriter, _ *http.Request) {
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, writeErr := w.Write(healthResponse); writeErr != nil {
		// optionally log the error or handle it
		fmt.Printf("failed to write response: %v\n", writeErr)
	}
}

func dynamicValidationHandler(w http.ResponseWriter, _ *http.Request) {
	response := DynamicResponse{
		Timestamp: time.Now(),
		ID:        uuid.New().String(),
		EmbeddedField: EmbeddedTimestamp{
			InnerTimestamp: time.Now(),
		},
		URL: fmt.Sprintf("http://localhost/endpoint/%s", uuid.New().String()),
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func newRouter() http.Handler {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/example1", ExampleHandler1).Methods("GET")
	router.HandleFunc("/example2", ExampleHandler2).Methods("POST")
	router.HandleFunc("/health", ExampleHealthHandler).Methods("GET")
	router.HandleFunc("/dynamic/validation", dynamicValidationHandler).Methods("GET")

	return router
}

func NewServer() *http.Server {
	return &http.Server{
		Handler:     newRouter(),
		ReadTimeout: 10 * time.Second,
	}
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
