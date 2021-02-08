package main

import (
	"net/http"
)

type MyAppFeature struct {
	Handler http.Handler
}

func NewMyAppFeature(handler http.Handler) *MyAppFeature {

	return &MyAppFeature{
		Handler: handler,
	}
}
