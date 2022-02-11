package main

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// Config holds configuration used by MyAppComponent.
type Config struct {
	BindAddr                string
	SiteDomain              string
	GracefulShutdownTimeout time.Duration
}

// NewConfig initialises hard-coded configuration for MyAppComponent.
func NewConfig() *Config {
	return &Config{
		BindAddr:                ":26601",
		SiteDomain:              "localhost",
		GracefulShutdownTimeout: 5 * time.Second,
	}
}

// MyAppComponent holds the initialized http server and config required for running component tests.
type MyAppComponent struct {
	errorChan      chan error
	Config         *Config
	HTTPServer     *http.Server
	ServiceRunning bool
}

// NewMyAppComponent initializes server used for running component.
func NewMyAppComponent(handler http.Handler) *MyAppComponent {
	c := &MyAppComponent{
		errorChan: make(chan error, 1),
		Config:    NewConfig(),
		HTTPServer: &http.Server{
			Handler: handler,
		},
		ServiceRunning: false,
	}

	c.run()
	c.ServiceRunning = true

	return c
}

func (c *MyAppComponent) run() {
	c.HTTPServer.Addr = c.Config.BindAddr

	// Start HTTP server
	go func() {
		if err := c.HTTPServer.ListenAndServe(); err != nil {
			c.errorChan <- err
		}
	}()
}

// Close server running component.
func (c *MyAppComponent) Close() error {
	if c.ServiceRunning {
		err := c.close(context.Background())
		c.ServiceRunning = false
		return err
	}
	return nil
}

func (c *MyAppComponent) close(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.Config.GracefulShutdownTimeout)
	hasShutdownError := false

	go func() {
		defer cancel()

		// stop any incoming requests
		if err := c.HTTPServer.Shutdown(ctx); err != nil {
			hasShutdownError = true
		}
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-ctx.Done()

	// timeout expired
	if ctx.Err() == context.DeadlineExceeded {
		return ctx.Err()
	}

	// other error
	if hasShutdownError {
		err := errors.New("failed to shutdown gracefully")
		return err
	}

	return nil
}
