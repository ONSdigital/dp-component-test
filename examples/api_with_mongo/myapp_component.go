package main

import (
	"net/http"
	"os"
)

type MyAppComponent struct {
	Handler http.Handler
}

func NewMyAppComponent(handler http.Handler, mongoUrl string) *MyAppComponent {

	os.Setenv("MONGO_URL", mongoUrl)
	os.Setenv("DATABASE_NAME", "testing")

	return &MyAppComponent{
		Handler: handler,
	}
}
