package main

import (
	featuretest "github.com/armakuni/dp-go-featuretest"
	"github.com/cucumber/godog"
)

func InitializeScenario(ctx *godog.ScenarioContext) {
	myAppFeature := NewMyAppFeature()
	apiFeature := featuretest.NewAPIFeatureWithHandler(myAppFeature.Handler)

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
