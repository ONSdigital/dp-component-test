package main

import (
	"context"

	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/cucumber/godog"
)

// MyAppComponent represents an example app under test
type MyAppComponent struct {
	svc           *Service
	kafkaScenario *componenttest.KafkaScenario
}

// NewMyAppComponent creates a new component using the supplied kafka component
func NewMyAppComponent(kafkaScenario *componenttest.KafkaScenario) *MyAppComponent {
	c := &MyAppComponent{
		kafkaScenario: kafkaScenario,
	}
	return c
}

func (c *MyAppComponent) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^the service is started with ([^"]*) configured`, c.theServiceStarts)
}

func (c *MyAppComponent) theServiceStarts(ctx context.Context, msgType string) error {
	if msgType == "Avro" {
		c.svc.UseAvro = true
	} else {
		c.svc.UseAvro = false
	}

	c.svc.Start(ctx)
	return nil
}

// Initialize sets up the component for the current scenario with random kafka topics. It starts the underlying service
// ready to consume and produce events
func (c *MyAppComponent) Initialize(ctx context.Context) error {
	c.svc = &Service{
		InputTopic:   c.kafkaScenario.GetMappedTopic("input"),
		OutputTopic:  c.kafkaScenario.GetMappedTopic("output"),
		KafkaBrokers: c.kafkaScenario.KafkaFeature.GetBrokers(ctx),
	}
	return nil
}

// Close closes the underlying service, stopping kafka consumers etc.
func (c *MyAppComponent) Close(ctx context.Context) error {
	c.svc.Close(ctx)
	return nil
}
