package main

import (
	"os"
	"testing"

	featuretest "github.com/armakuni/dp-go-featuretest"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

type FeatureTestSuite struct {
	T *testing.T
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
	var opts = godog.Options{
		Output: colors.Colored(os.Stdout),
		Format: "pretty", // can define default values
	}

	ts := &FeatureTestSuite{T: t}

	godog.TestSuite{
		Name:                 "feature_tests",
		TestSuiteInitializer: ts.InitializeTestSuite,
		ScenarioInitializer:  ts.InitializeScenario,
		Options:              &opts,
	}.Run()
}
