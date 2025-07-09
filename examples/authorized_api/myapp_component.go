package main

import (
	"net/http"
	"os"
)

type MyAppComponent struct {
	Handler http.Handler
}

func NewMyAppComponent(authURL string) *MyAppComponent {
	os.Setenv("AUTH_URL", authURL)

	return &MyAppComponent{
		Handler: NewRouter(),
	}
}
