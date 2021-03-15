package main

import (
	"net/http"
	"os"
)

type MyAppComponent struct {
	Handler http.Handler
}

func NewMyAppComponent(auth_url string) *MyAppComponent {

	os.Setenv("AUTH_URL", auth_url)

	return &MyAppComponent{
		Handler: NewRouter(),
	}
}
