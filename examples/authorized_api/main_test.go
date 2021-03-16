package main

import (
	"flag"
	"fmt"
	"os"
	"testing"

	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var componentFlag = flag.Bool("component", false, "perform component tests")
var allFlag = flag.Bool("all", false, "perform all tests")

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

func TestMain(m *testing.M) {
	flag.Parse()
	status := 0
	if *componentFlag || *allFlag {
		var opts = godog.Options{
			Output: colors.Colored(os.Stdout),
			Format: "pretty",
		}

		godog.TestSuite{
			Name:                "component_tests",
			ScenarioInitializer: InitializeScenario,
			Options:             &opts,
		}.Run()
	}

	if !*componentFlag || *allFlag {
		if st := m.Run(); st > status {
			status = st
		}
	}

	if *componentFlag {
		fmt.Printf("coverage: %.1f%s\n", testing.Coverage()*100, "% of all statements")
	}

	os.Exit(status)
}
