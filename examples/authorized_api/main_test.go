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

var componentFlag = flag.Bool("component", false, "perform component tests")

func InitializeScenario(godogCtx *godog.ScenarioContext) {
	authorizationFeature := componenttest.NewAuthorizationFeature()
	myAppFeature := NewMyAppComponent(authorizationFeature.FakeAuthService.ResolveURL(""))
	apiFeature := componenttest.NewAPIFeatureWithHandler(myAppFeature.Handler)

	godogCtx.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		apiFeature.Reset()
		authorizationFeature.Reset()
		return ctx, nil
	})

	godogCtx.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		authorizationFeature.Close()
		return ctx, nil
	})

	apiFeature.RegisterSteps(godogCtx)
	authorizationFeature.RegisterSteps(godogCtx)
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
