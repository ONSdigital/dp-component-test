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

type FeatureTestSuite struct {
	Mongo *featuretest.MongoCapability
	T     *testing.T
}

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

var componentFlag = false

func init() {
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	flag.BoolVar(&componentFlag, "component", false, "set this flag to run the component tests")
	flag.Parse()
}

func (t *FeatureTestSuite) InitializeTestSuite(ctx *godog.TestSuiteContext) {

	mongoOptions := featuretest.MongoOptions{
		Port:         27017,
		MongoVersion: "4.0.5",
		DatabaseName: "testing",
	}
	mongoFeature, err := featuretest.NewMongoCapability(mongoOptions)
	if err != nil {
		panic(err)
	}
	t.Mongo = mongoFeature
	ctx.AfterSuite(func() {
		mongoFeature.Close()
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
