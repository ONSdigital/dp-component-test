package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"testing"

	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

type componenttestSuite struct {
	Mongo *componenttest.MongoFeature
}

var componentFlag = flag.Bool("component", false, "perform component tests")
var allFlag = flag.Bool("all", false, "perform all tests")

func (m *MyAppComponent) initialiser(h http.Handler) componenttest.ServiceInitialiser {
	return func() (http.Handler, error) {
		m.Handler = h
		return h, nil
	}
}

func (t *componenttestSuite) InitializeScenario(ctx *godog.ScenarioContext) {
	server := NewServer()

	component := NewMyAppComponent(server.Handler, t.Mongo.Server.URI())
	apiFeature := componenttest.NewAPIFeature(component.initialiser(server.Handler))

	ctx.BeforeScenario(func(*godog.Scenario) {
		t.Mongo.Reset()
		apiFeature.Reset()
	})

	ctx.AfterScenario(func(*godog.Scenario, error) {
		t.Mongo.Reset()
		apiFeature.Reset()
	})

	apiFeature.RegisterSteps(ctx)
	t.Mongo.RegisterSteps(ctx)

}

func (t *componenttestSuite) InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
		mongoOptions := componenttest.MongoOptions{
			MongoVersion: "4.0.23",
			DatabaseName: "testing",
		}
		t.Mongo = componenttest.NewMongoFeature(mongoOptions)
	})

	ctx.AfterSuite(func() {
		t.Mongo.Close()
	})
}

func TestMain(t *testing.T) {
	if *componentFlag {
		var opts = godog.Options{
			Output: colors.Colored(os.Stdout),
			Paths:  flag.Args(),
			Format: "pretty",
		}

		ts := &componenttestSuite{}

		status := godog.TestSuite{
			Name:                 "component_tests",
			ScenarioInitializer:  ts.InitializeScenario,
			TestSuiteInitializer: ts.InitializeTestSuite,
			Options:              &opts,
		}.Run()

		fmt.Printf("coverage: %.1f%s\n", testing.Coverage()*100, "% of all statements")

		if status > 0 {
			t.Fail()
		}
	} else {
		t.Skip()
	}

}
