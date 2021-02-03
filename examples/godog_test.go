package main

import (
	featuretest "github.com/armakuni/dp-go-featuretest"
	"github.com/cucumber/godog"
)

func InitializeScenario(ctx *godog.ScenarioContext) {
	myAppFeature := NewMyAppFeature()
	apiFeature := featuretest.NewAPIFeature(myAppFeature.HTTPServer)

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
