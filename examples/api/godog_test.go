package main

import (
	featuretest "github.com/armakuni/dp-go-featuretest"
	"github.com/cucumber/godog"
	"net/http"
)


func (m* MyAppFeature) initialiser(h http.Handler) featuretest.ServiceInitialiser {
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
