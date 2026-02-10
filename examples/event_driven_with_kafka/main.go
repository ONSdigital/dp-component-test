package main

import (
	"context"
	"encoding/json"
	"fmt"

	kafka "github.com/ONSdigital/dp-kafka/v4"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/google/uuid"
)

const (
	kafkaGroup   = "kafka-example"
	kafkaVersion = "3.8.0"
)

// Input represents an input example event
type Input struct {
	Input string `json:"input"`
}

// Output represents an output example event
type Output struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

// Service represents a service that consumes and produces example kafka events
type Service struct {
	InputTopic     string
	OutputTopic    string
	inputConsumer  kafka.IConsumerGroup
	outputProducer kafka.IProducer
	KafkaBrokers   []string
}

func (s *Service) Start(ctx context.Context) {
	// Init consumer
	s.inputConsumer = getConsumer(ctx, s.KafkaBrokers, s.InputTopic)

	// Init Producer
	s.outputProducer = getProducer(ctx, s.KafkaBrokers, s.OutputTopic)

	// Register the handler
	handler := &Handler{
		OutputProducer: s.outputProducer,
	}
	if err := s.inputConsumer.RegisterHandler(ctx, handler.Handle); err != nil {
		panic(err)
	}

	// Start consuming
	err := s.inputConsumer.Start()
	if err != nil {
		panic(err)
	}
}

func (s *Service) Close(ctx context.Context) {
	if s.inputConsumer != nil {
		err := s.inputConsumer.StopAndWait()
		if err != nil {
			panic(err)
		}
	}

	if s.outputProducer != nil {
		err := s.outputProducer.Close(ctx)
		if err != nil {
			panic(err)
		}
	}
}

func getConsumer(ctx context.Context, brokers []string, topic string) *kafka.ConsumerGroup {
	kafkaOffset := kafka.OffsetOldest
	version := kafkaVersion
	minHealthy := 1
	cgConfig := &kafka.ConsumerGroupConfig{
		BrokerAddrs:       brokers,
		Topic:             topic,
		GroupName:         kafkaGroup,
		MinBrokersHealthy: &minHealthy,
		KafkaVersion:      &version,
		Offset:            &kafkaOffset,
	}
	consumer, err := kafka.NewConsumerGroup(ctx, cgConfig)
	if err != nil {
		panic(err)
	}
	return consumer
}

func getProducer(ctx context.Context, brokers []string, topic string) *kafka.Producer {
	minHealthy := 1
	version := kafkaVersion
	pConfig := &kafka.ProducerConfig{
		BrokerAddrs:       brokers,
		Topic:             topic,
		MinBrokersHealthy: &minHealthy,
		KafkaVersion:      &version,
	}
	producer, err := kafka.NewProducer(ctx, pConfig)
	if err != nil {
		panic(err)
	}
	return producer
}

// Handler represents a kafka handler for the example app
type Handler struct {
	OutputProducer kafka.IProducer
}

// Handle consumes an input event and produces an output event based on it unless the input id is `no-output`
func (h *Handler) Handle(ctx context.Context, _ int, msg kafka.Message) error {
	inputEvent := Input{}
	if err := json.Unmarshal(msg.GetData(), &inputEvent); err != nil {
		return err
	}

	id := inputEvent.Input
	log.Info(ctx, "received input event from kafka consumer", log.Data{"id": id})

	// special case to suppress an output event being produced
	if id == "no-output" {
		log.Info(ctx, "no output was produced")
		return nil
	}

	outputEvent := Output{
		Input:  inputEvent.Input,
		Output: "World!",
	}
	err := h.OutputProducer.SendJSON(ctx, outputEvent)
	if err != nil {
		return err
	}

	return nil
}

// Run the example service against a real kakfa
func main() {
	service := &Service{
		InputTopic:   fmt.Sprintf("input-%s", uuid.NewString()),
		OutputTopic:  fmt.Sprintf("output-%s", uuid.NewString()),
		KafkaBrokers: []string{"localhost:9092", "localhost:9093", "localhost:9094"},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	service.Start(ctx)
	defer service.Close(ctx)

	fireExampleEvent(ctx, service)
}

// Example event, not part of service
func fireExampleEvent(ctx context.Context, s *Service) {
	inputProducer := getProducer(ctx, s.KafkaBrokers, s.InputTopic)
	err := inputProducer.Initialise(ctx)
	if err != nil {
		panic(err)
	}
	defer inputProducer.Close(ctx)

	id := uuid.NewString()
	msg := Input{
		Input: id,
	}
	log.Info(ctx, "sending example event", log.Data{"id": id})
	err = inputProducer.SendJSON(ctx, msg)
	if err != nil {
		panic(err)
	}

	// wait for produced output
	done := make(chan bool)
	outputConsumer := getConsumer(ctx, s.KafkaBrokers, s.OutputTopic)
	handler := func(ctx context.Context, _ int, msg kafka.Message) error {
		output := &Output{}
		err := json.Unmarshal(msg.GetData(), output)
		if err != nil {
			return err
		}
		if output.Input == id {
			log.Info(ctx, "example output event consumed", log.Data{"id": output.Input})
			done <- true
		}
		return nil
	}
	if err := outputConsumer.RegisterHandler(ctx, handler); err != nil {
		panic(err)
	}
	defer outputConsumer.Close(ctx)

	// Start consuming
	err = outputConsumer.Start()
	if err != nil {
		panic(err)
	}

	// Wait for done
	<-done
}
