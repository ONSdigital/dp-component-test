package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/google/uuid"
)

type componentTestSuite struct {
	Kafka *componenttest.KafkaFeature
}

var componentFlag = flag.Bool("component", false, "perform component tests")

func (t *componentTestSuite) InitializeScenario(godogCtx *godog.ScenarioContext) {
	var svc Service

	godogCtx.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		svc := &Service{
			InputTopic:   fmt.Sprintf("input-%s", uuid.NewString()),
			OutputTopic:  fmt.Sprintf("output-%s", uuid.NewString()),
			KafkaBrokers: t.Kafka.GetBrokers(ctx),
		}
		ctx = t.Kafka.ContextWithTopicMap(ctx, "input", svc.InputTopic)
		ctx = t.Kafka.ContextWithTopicMap(ctx, "output", svc.OutputTopic)
		svc.Start(ctx)
		return ctx, nil
	})

	godogCtx.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		svc.Close(ctx)
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
			Output: colors.Colored(os.Stdout),
			Paths:  flag.Args(),
			Format: "pretty",
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
