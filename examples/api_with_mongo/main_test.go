package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"testing"

	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

type componenttestSuite struct {
	Mongo *componenttest.MongoFeature
}

var componentFlag = flag.Bool("component", false, "perform component tests")

func (m *MyAppComponent) initialiser(h http.Handler) componenttest.ServiceInitialiser {
	return func() (http.Handler, error) {
		m.Handler = h
		return h, nil
	}
}

func (t *componenttestSuite) InitializeScenario(godogCtx *godog.ScenarioContext) {
	server := NewServer()

	component := NewMyAppComponent(server.Handler)
	apiFeature := componenttest.NewAPIFeature(component.initialiser(server.Handler))

	godogCtx.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		t.Mongo.Reset()
		apiFeature.Reset()
		return ctx, nil
	})

	godogCtx.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		t.Mongo.Reset()
		apiFeature.Reset()
		return ctx, nil
	})

	apiFeature.RegisterSteps(godogCtx)
	t.Mongo.RegisterSteps(godogCtx)
}

func (t *componenttestSuite) InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
		t.Mongo = componenttest.NewMongoFeature(componenttest.MongoOptions{
			ClusterEndpoint: "mongodb:27017",
			DatabaseName:    "test",
		})
	})

	ctx.AfterSuite(func() {
		t.Mongo.Close()
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
