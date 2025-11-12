package main

import (
	"net/http"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
)

type MyAppComponent struct {
	Handler  http.Handler
	DBClient *mongo.Client
	Database string
}

// NewMyAppComponentWithClient creates the component using an injected *mongo.Client.
// Prefer this in tests (or production wiring) — no global env side-effects.
func NewMyAppComponentWithClient(handler http.Handler, client *mongo.Client, database string) *MyAppComponent {
	return &MyAppComponent{
		Handler:  handler,
		DBClient: client,
		Database: database,
	}
}

// NewMyAppComponent is kept for backwards compatibility with code that expects a mongo URL
// and for callers that rely on environment variables. Prefer NewMyAppComponentWithClient instead.
func NewMyAppComponent(handler http.Handler, mongoURL string) *MyAppComponent {
	// keep original behaviour so existing tests still work
	// caller is responsible for setting DATABASE_NAME if they want a different value
	_ = os.Setenv("MONGO_URL", mongoURL)
	_ = os.Setenv("DATABASE_NAME", "testing")

	return &MyAppComponent{
		Handler: handler,
	}
}

// initialiser returns a function compatible with component-test style initialisers.
// It captures the handler and stores it on the component.
func (m *MyAppComponent) initialiser(h http.Handler) func() (http.Handler, error) {
	return func() (http.Handler, error) {
		m.Handler = h
		return h, nil
	}
}
