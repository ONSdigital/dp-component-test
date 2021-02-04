package main

import (
	"net/http"
)

type MyAppFeature struct {
	Handler http.Handler
}

func NewMyAppFeature() *MyAppFeature {

	return &MyAppFeature{
		Handler: NewRouter(),
	}
}
