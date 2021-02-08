package main

import (
	goflag "flag"
	"net/http"
	"os"
	"testing"

	featuretest "github.com/armakuni/dp-go-featuretest"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	flag "github.com/spf13/pflag"
)

var componentFlag = false

func init() {
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	flag.BoolVar(&componentFlag, "component", false, "set this flag to run the component tests")
	flag.Parse()
}

func (m *MyAppFeature) initialiser(h http.Handler) featuretest.ServiceInitialiser {
	return func() (http.Handler, error) {
		m.Handler = h
		return h, nil
	}
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	server := NewServer()
	feature := NewMyAppFeature(server.Handler)

	apiFeature := featuretest.NewAPIFeature(feature.initialiser(server.Handler))

	ctx.BeforeScenario(func(*godog.Scenario) {
		apiFeature.Reset()
	})

	ctx.AfterScenario(func(*godog.Scenario, error) {

	})

	apiFeature.RegisterSteps(ctx)
}

func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
	})
}

func TestFeatures(t *testing.T) {
	if componentFlag == true {
		var opts = godog.Options{
			Output: colors.Colored(os.Stdout),
			Format: "pretty",
		}

		godog.TestSuite{
			Name:                 "feature_tests",
			TestSuiteInitializer: InitializeTestSuite,
			ScenarioInitializer:  InitializeScenario,
			Options:              &opts,
		}.Run()
	}
}
