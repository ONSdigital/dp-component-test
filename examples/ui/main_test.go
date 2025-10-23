package main

import (
	"context"
	"flag"
	"os"
	"testing"

	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var componentFlag = flag.Bool("component", false, "perform component tests")

func InitializeScenario(godogCtx *godog.ScenarioContext) {
	server := NewServer()
	component := NewMyAppComponent(server.Handler)

	uiFeature := componenttest.NewUIFeature("http://" + component.Config.SiteDomain + component.Config.BindAddr)

	godogCtx.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		uiFeature.Reset()
		return ctx, nil
	})

	godogCtx.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		if err := component.Close(); err != nil {
			log.Warn(context.Background(), "error closing component", log.FormatErrors([]error{err}))
		}
		uiFeature.Close()
		return ctx, nil
	})

	uiFeature.RegisterSteps(godogCtx)
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
