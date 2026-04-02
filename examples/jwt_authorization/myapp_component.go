package main

import (
	"net/http"
)

type MyAppComponent struct {
	Handler http.Handler
	Config  *Config
}

func NewMyAppComponent(cfg *Config) *MyAppComponent {
	return &MyAppComponent{
		Config:  cfg,
		Handler: NewRouter(cfg),
	}
}
