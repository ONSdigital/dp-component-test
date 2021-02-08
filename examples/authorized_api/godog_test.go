package main

import (
	goflag "flag"
	"os"
	"testing"

	featuretest "github.com/armakuni/dp-go-featuretest"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	flag "github.com/spf13/pflag"
)

type FeatureTestSuite struct {
	T *testing.T
}

var componentFlag = false

func init() {
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	flag.BoolVar(&componentFlag, "component", false, "set this flag to run the component tests")
	flag.Parse()
}

func (t *FeatureTestSuite) InitializeScenario(ctx *godog.ScenarioContext) {
	authorizationFeature := featuretest.NewAuthorizationFeature(t.T)
	myAppFeature := NewMyAppFeature(authorizationFeature.FakeAuthService.ResolveURL(""))
	apiFeature := featuretest.NewAPIFeatureWithHandler(myAppFeature.Handler)

	ctx.BeforeScenario(func(*godog.Scenario) {
		apiFeature.Reset()
	})

	ctx.AfterScenario(func(*godog.Scenario, error) {

	})

	apiFeature.RegisterSteps(ctx)
	authorizationFeature.RegisterSteps(ctx)
}

func (t *FeatureTestSuite) InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
	})
}

func TestFeatures(t *testing.T) {
	if componentFlag == true {
		var opts = godog.Options{
			Output: colors.Colored(os.Stdout),
			Format: "pretty",
		}

		ts := &FeatureTestSuite{T: t}

		godog.TestSuite{
			Name:                 "feature_tests",
			TestSuiteInitializer: ts.InitializeTestSuite,
			ScenarioInitializer:  ts.InitializeScenario,
			Options:              &opts,
		}.Run()
	}
}
