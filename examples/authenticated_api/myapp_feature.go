package main

import (
	"net/http"
	"os"
)

type MyAppFeature struct {
	Handler http.Handler
}

func NewMyAppFeature(auth_url string) *MyAppFeature {

	os.Setenv("AUTH_URL", auth_url)

	return &MyAppFeature{
		Handler: NewRouter(),
	}
}
