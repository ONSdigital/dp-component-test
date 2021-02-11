package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"testing"

	featuretest "github.com/armakuni/dp-go-featuretest"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

type FeatureTestSuite struct {
	Mongo *featuretest.MongoFeature
}

var componentFlag = flag.Bool("component", false, "perform component tests")
var allFlag = flag.Bool("all", false, "perform all tests")

func (m *MyAppFeature) initialiser(h http.Handler) featuretest.ServiceInitialiser {
	return func() (http.Handler, error) {
		m.Handler = h
		return h, nil
	}
}

func (t *FeatureTestSuite) InitializeScenario(ctx *godog.ScenarioContext) {
	server := NewServer()

	feature := NewMyAppFeature(server.Handler, t.Mongo.Server.URI())
	apiFeature := featuretest.NewAPIFeature(feature.initialiser(server.Handler))

	ctx.BeforeScenario(func(*godog.Scenario) {
		t.Mongo.Reset()
		apiFeature.Reset()
	})

	ctx.AfterScenario(func(*godog.Scenario, error) {})

	apiFeature.RegisterSteps(ctx)
	t.Mongo.RegisterSteps(ctx)

}

func (t *FeatureTestSuite) InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
		mongoOptions := featuretest.MongoOptions{
			Port:         27017,
			MongoVersion: "4.0.5",
			DatabaseName: "testing",
		}
		t.Mongo = featuretest.NewMongoFeature(mongoOptions)
	})

	ctx.AfterSuite(func() {
		t.Mongo.Close()
	})
}

func TestMain(m *testing.M) {
	flag.Parse()
	status := 0
	if *componentFlag || *allFlag {
		var opts = godog.Options{
			Output: colors.Colored(os.Stdout),
			Format: "pretty",
		}

		ts := &FeatureTestSuite{}

		status = godog.TestSuite{
			Name:                 "feature_tests",
			ScenarioInitializer:  ts.InitializeScenario,
			TestSuiteInitializer: ts.InitializeTestSuite,
			Options:              &opts,
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
