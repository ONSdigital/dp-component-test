package main

import (
	"net/http"
	"os"
)

type MyAppComponent struct {
	Handler http.Handler
}

func NewMyAppComponent(handler http.Handler, mongoURL string) *MyAppComponent {
	os.Setenv("MONGO_URL", mongoURL)
	os.Setenv("DATABASE_NAME", "testing")

	return &MyAppComponent{
		Handler: handler,
	}
}
