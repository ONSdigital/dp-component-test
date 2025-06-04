package main

import (
	"flag"
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

	component := NewMyAppComponent(server.Handler, t.Redis.Server.Addr())
	apiFeature := componenttest.NewAPIFeature(component.initialiser(server.Handler))

	ctx.BeforeScenario(func(*godog.Scenario) {
		t.Redis.Reset()
		apiFeature.Reset()
	})

	ctx.AfterScenario(func(*godog.Scenario, error) {
		t.Redis.Reset()
		apiFeature.Reset()
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
