package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

type MyAppFeature struct {
	HTTPServer *http.Server
	// errorChan  error
}

func NewMyAppFeature() *MyAppFeature {
	router := mux.NewRouter().StrictSlash(true)
	HandleRequests(router)

	f := &MyAppFeature{
		HTTPServer: &http.Server{
			Handler: router,
		},
		// errorChan: make(chan error),
	}

	return f
}
