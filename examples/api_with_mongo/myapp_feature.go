package main

import (
	"net/http"
	"os"
)

type MyAppFeature struct {
	Handler http.Handler
}

func NewMyAppFeature(handler http.Handler, mongoUrl string) *MyAppFeature {

	os.Setenv("MONGO_URL", mongoUrl)
	os.Setenv("DATABASE_NAME", "testing")

	return &MyAppFeature{
		Handler: handler,
	}
}
