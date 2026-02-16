package main

import (
	"context"
	"fmt"

	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/google/uuid"
)

// MyAppComponent represents an example app under test
type MyAppComponent struct {
	svc          *Service
	kafkaFeature *componenttest.KafkaFeature
}

// NewMyAppComponent creates a new component using the supplied kafka component
func NewMyAppComponent(kafkaFeature *componenttest.KafkaFeature) *MyAppComponent {
	c := &MyAppComponent{
		kafkaFeature: kafkaFeature,
	}
	return c
}

// Initialize sets up the component for the current scenario with random kafka topics. It starts the underlying service
// ready to consume and produce events
func (c *MyAppComponent) Initialize(ctx context.Context) error {
	c.svc = &Service{
		InputTopic:   fmt.Sprintf("input-%s", uuid.NewString()),
		OutputTopic:  fmt.Sprintf("output-%s", uuid.NewString()),
		KafkaBrokers: c.kafkaFeature.GetBrokers(ctx),
	}
	c.svc.Start(ctx)
	return nil
}

// ScenarioContext populates the supplied context with variable relating to the current scenario. Specifically it adds
// mappings for the random topics so that scenarios can refer to them using friendly ids
func (c *MyAppComponent) ScenarioContext(ctx context.Context) context.Context {
	ctx = c.kafkaFeature.ContextWithTopicMap(ctx, "input", c.svc.InputTopic)
	ctx = c.kafkaFeature.ContextWithTopicMap(ctx, "output", c.svc.OutputTopic)
	return ctx
}

// Close closes the underlying service, stopping kafka consumers etc.
func (c *MyAppComponent) Close(ctx context.Context) error {
	c.svc.Close(ctx)
	return nil
}
