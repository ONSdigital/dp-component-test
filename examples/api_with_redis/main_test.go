package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"testing"

	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

type componenttestSuite struct {
	Redis *componenttest.RedisFeature
}

var componentFlag = flag.Bool("component", false, "perform component tests")

func (m *MyAppComponent) initialiser(h http.Handler) componenttest.ServiceInitialiser {
	return func() (http.Handler, error) {
		m.Handler = h
		return h, nil
	}
}

func (t *componenttestSuite) InitializeScenario(ctx *godog.ScenarioContext) {
	server := NewServer()

	component, err := NewMyAppComponent(server.Handler, t.Redis.Server.Addr(), t.Redis)
	if err != nil {
		fmt.Printf("failed to create redis app component - error: %v", err)
		os.Exit(1)
	}
	apiFeature := componenttest.NewAPIFeature(component.initialiser(server.Handler))

	ctx.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		t.Redis.Reset()
		apiFeature.Reset()
		return ctx, nil
	})

	ctx.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		t.Redis.Reset()
		apiFeature.Reset()
		return ctx, nil
	})

	apiFeature.RegisterSteps(ctx)
	t.Redis.RegisterSteps(ctx)
}

func (t *componenttestSuite) InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
		t.Redis = componenttest.NewRedisFeature()
	})

	ctx.AfterSuite(func() {
		t.Redis.Close()
	})
}

func TestComponent(t *testing.T) {
	if *componentFlag {
		var opts = godog.Options{
			Output: colors.Colored(os.Stdout),
			Paths:  flag.Args(),
			Format: "pretty",
		}

		ts := &componenttestSuite{}

		status := godog.TestSuite{
			Name:                 "component_tests",
			ScenarioInitializer:  ts.InitializeScenario,
			TestSuiteInitializer: ts.InitializeTestSuite,
			Options:              &opts,
		}.Run()

		if status > 0 {
			t.Fail()
		}
	} else {
		t.Skip()
	}
}
