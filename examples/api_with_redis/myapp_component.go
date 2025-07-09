package main

import (
	"net/http"
	"os"

	componentTest "github.com/ONSdigital/dp-component-test"
)

type MyAppComponent struct {
	Handler      http.Handler
	redisFeature *componentTest.RedisFeature
	Config       *Config
}

func NewMyAppComponent(handler http.Handler, redisURL string, redisFeat *componentTest.RedisFeature) (*MyAppComponent, error) {
	os.Setenv("REDIS_URL", redisURL)

	c := &MyAppComponent{
		Handler:      handler,
		redisFeature: redisFeat,
	}

	var err error

	c.Config, err = Get()
	if err != nil {
		return nil, err
	}

	c.redisFeature = redisFeat
	c.Config.RedisURL = c.redisFeature.Server.Addr()

	return c, nil
}
