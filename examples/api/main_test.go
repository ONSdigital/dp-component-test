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

var componentFlag = flag.Bool("component", false, "perform component tests")

func (m *MyAppComponent) initialiser(h http.Handler) componenttest.ServiceInitialiser {
	return func() (http.Handler, error) {
		m.Handler = h
		return h, nil
	}
}

func InitializeScenario(godogCtx *godog.ScenarioContext) {
	server := NewServer()
	component := NewMyAppComponent(server.Handler)

	apiFeature := componenttest.NewAPIFeature(component.initialiser(server.Handler))

	godogCtx.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		apiFeature.Reset()
		return ctx, nil
	})

	apiFeature.RegisterSteps(godogCtx)
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
