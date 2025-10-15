package main

import (
	"net/http"
)

type MyAppComponent struct {
	Handler http.Handler
}

func NewMyAppComponent(handler http.Handler) *MyAppComponent {
	return &MyAppComponent{
		Handler: handler,
	}
}
