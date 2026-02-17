package main

import (
	"context"
	"flag"
	"os"
	"testing"

	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

type componentTestSuite struct {
	Kafka *componenttest.KafkaFeature
}

var componentFlag = flag.Bool("component", false, "perform component tests")

func (t *componentTestSuite) InitializeScenario(godogCtx *godog.ScenarioContext) {
	component := NewMyAppComponent(t.Kafka)

	godogCtx.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		component.Initialize(ctx)
		ctx = component.ScenarioContext(ctx)
		return ctx, nil
	})

	godogCtx.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		component.Close(ctx)
		return ctx, nil
	})
	t.Kafka.RegisterSteps(godogCtx)
}

func (t *componentTestSuite) InitializeTestSuite(godogCtx *godog.TestSuiteContext) {
	godogCtx.BeforeSuite(func() {
		t.Kafka = componenttest.NewKafkaFeature(&componenttest.KafkaOptions{KafkaVersion: kafkaVersion})
	})

	godogCtx.AfterSuite(func() {
		t.Kafka.Close()
	})
}

func TestComponent(t *testing.T) {
	if *componentFlag {
		var opts = godog.Options{
			Output:   colors.Colored(os.Stdout),
			Paths:    flag.Args(),
			Format:   "pretty",
			TestingT: t,
		}

		ts := &componentTestSuite{}

		status := godog.TestSuite{
			Name:                 "component_tests",
			ScenarioInitializer:  ts.InitializeScenario,
			TestSuiteInitializer: ts.InitializeTestSuite,
			Options:              &opts,
		}.Run()

		if status > 0 {
			t.Fail()
		}
	} else {
		t.Skip()
	}
}
