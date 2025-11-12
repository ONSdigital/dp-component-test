package main

import (
	"context"
	"flag"
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

func (t *componenttestSuite) InitializeScenario(godogCtx *godog.ScenarioContext) {
	// create client created the test mongo URI
	uri, err := t.Mongo.URI()
	if err != nil {
		panic(err)
	}

	client, err := NewMongoClient(uri)
	if err != nil {
		panic(err)
	}

	// create server and component once per scenario
	server := NewServer(client, "testing", ":0") // :0 so it doesn't bind when using in-process tests

	component := NewMyAppComponentWithClient(server.Handler, client, "testing")
	apiFeature := componenttest.NewAPIFeature(component.initialiser(server.Handler))

	// Reset DB and API feature before each scenario
	godogCtx.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		if err := t.Mongo.Reset(); err != nil {
			return nil, err
		}
		apiFeature.Reset()
		return ctx, nil
	})

	godogCtx.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		if err := t.Mongo.Reset(); err != nil {
			return ctx, err
		}
		apiFeature.Reset()
		_ = client.Disconnect(ctx)
		return ctx, nil
	})

	apiFeature.RegisterSteps(godogCtx)
	t.Mongo.RegisterSteps(godogCtx)
}

func (t *componenttestSuite) InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
		mongoOptions := componenttest.MongoOptions{
			MongoVersion: "4.4.8",
			DatabaseName: "testing",
		}
		t.Mongo = componenttest.NewMongoFeature(mongoOptions)
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
