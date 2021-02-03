package main

import (
	"net/http"
)

type MyAppFeature struct {
	Handler http.Handler
	// errorChan  error
}

func NewMyAppFeature() *MyAppFeature {

	return &MyAppFeature{
		Handler: NewRouter(),
		// errorChan: make(chan error),
	}
}
