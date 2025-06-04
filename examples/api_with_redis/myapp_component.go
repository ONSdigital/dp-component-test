package main

import (
	"net/http"
	"os"
)

type MyAppComponent struct {
	Handler http.Handler
}

func NewMyAppComponent(handler http.Handler, redisUrl string) *MyAppComponent {

	os.Setenv("REDIS_URL", redisUrl)

	return &MyAppComponent{
		Handler: handler,
	}
}
