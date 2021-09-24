package main

import (
	"flag"
	"os"
	"testing"

	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var componentFlag = flag.Bool("component", false, "perform component tests")

func InitializeScenario(ctx *godog.ScenarioContext) {
	authorizationFeature := componenttest.NewAuthorizationFeature()
	myAppFeature := NewMyAppComponent(authorizationFeature.FakeAuthService.ResolveURL(""))
	apiFeature := componenttest.NewAPIFeatureWithHandler(myAppFeature.Handler)

	ctx.BeforeScenario(func(*godog.Scenario) {
		apiFeature.Reset()
		authorizationFeature.Reset()
	})

	ctx.AfterScenario(func(*godog.Scenario, error) {
		authorizationFeature.Close()
	})

	apiFeature.RegisterSteps(ctx)
	authorizationFeature.RegisterSteps(ctx)
}

func TestComponent(t *testing.T) {
	if *componentFlag {
		var opts = godog.Options{
			Output: colors.Colored(os.Stdout),
			Paths:  flag.Args(),
			Format: "pretty",
		}

		status := godog.TestSuite{
			Name:                "component_tests",
			ScenarioInitializer: InitializeScenario,
			Options:             &opts,
		}.Run()

		if status > 0 {
			t.Fail()
		}

	} else {
		t.Skip()
	}
}
