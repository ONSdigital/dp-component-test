package componenttest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	kafka "github.com/ONSdigital/dp-kafka/v4"
	"github.com/ONSdigital/dp-kafka/v4/avro"
	"github.com/cucumber/godog"
	"github.com/google/uuid"
	tckafka "github.com/testcontainers/testcontainers-go/modules/kafka"
)

// KafkaFeature represents a component test feature that tests kafka functionality via testcontainers
type KafkaFeature struct {
	kafkaContainer *tckafka.KafkaContainer
	KafkaVersion   string
	EventEncoders  map[string]map[string]EventEncoder
}

const defaultKafkaContainerName = "confluentinc/confluent-local:7.5.0"
const defaultKafkaVersion = "3.8.0"

// KafkaOptions are optional configuration options for the kafka feature initialisation
// If no encoders are supplied for a topic then the default encoding of JSON is assumed for that topic
type KafkaOptions struct {
	ContainerName string
	KafkaVersion  string
	Encoders      []KafkaEncoderOption
}

// KafkaEncoderOption links an envent Encoder to a topic and encoding type
type KafkaEncoderOption struct {
	Topic    string
	Encoding string // Eg, Avro
	Encoder  EventEncoder
}

// NewKafkaFeature creates a new feature with the supplied optional configuration options
func NewKafkaFeature(opts *KafkaOptions) *KafkaFeature {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if opts == nil {
		opts = &KafkaOptions{}
	}

	if opts.KafkaVersion == "" {
		opts.KafkaVersion = defaultKafkaVersion
	}

	if opts.ContainerName == "" {
		opts.ContainerName = defaultKafkaContainerName
	}

	kafkaContainer, err := tckafka.Run(ctx, opts.ContainerName)
	if err != nil {
		panic(err)
	}

	kf := &KafkaFeature{
		kafkaContainer: kafkaContainer,
		KafkaVersion:   opts.KafkaVersion,
		EventEncoders:  make(map[string]map[string]EventEncoder),
	}
	for _, encoderOption := range opts.Encoders {
		if kf.EventEncoders[encoderOption.Topic] == nil {
			kf.EventEncoders[encoderOption.Topic] = make(map[string]EventEncoder)
		}
		kf.EventEncoders[encoderOption.Topic][encoderOption.Encoding] = encoderOption.Encoder
	}
	return kf
}

// GetBrokers returns the kafka brokers of the underlying testcontainers instance. I.e. these are the addresses of the
// brokers which can be used for the app under test's kafka client
func (kf *KafkaFeature) GetBrokers(ctx context.Context) []string {
	brokers, err := kf.kafkaContainer.Brokers(ctx)
	if err != nil {
		panic(err)
	}

	return brokers
}

// NewScenario initiates a new KafkaScenario with features scoped to the current schenario
func (kf *KafkaFeature) NewScenario() *KafkaScenario {
	return &KafkaScenario{
		KafkaFeature: kf,
	}
}

// Close stops the kafka testcontainer
func (kf *KafkaFeature) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
	defer cancel()

	return kf.kafkaContainer.Terminate(ctx)
}

// KafkaScenario represents the kafka features scoped to the currently running scenario
type KafkaScenario struct {
	mu           sync.Mutex
	KafkaFeature *KafkaFeature
	topics       map[string]*kafkaScenarioTopic
}

type kafkaScenarioTopic struct {
	mu               sync.Mutex
	topic            string
	mappedTopic      string
	producer         *kafka.Producer
	consumer         *kafka.ConsumerGroup
	ConsumedMessages [][]byte
}

// GetMappedTopic returns a topic that has been mapped in the current scenario. If this is the first time it has been
// called it will create a new random mappping. Subsequent calls return the same value.
func (ks *KafkaScenario) GetMappedTopic(topic string) string {
	return ks.getScenarioTopic(topic).mappedTopic
}

// RegisterSteps adds the kafka feature's steps to the godog ScenarioContext
func (ks *KafkaScenario) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^this "([^"]*)" event is queued, to be consumed:$`, ks.thisEventIsQueued)
	ctx.Step(`^this "([^"]*)" ([^"]*) event is queued, to be consumed:$`, ks.thisEncodedEventIsQueued)
	ctx.Step(`^this "([^"]*)" event is produced:$`, ks.thisEventIsProduced)
	ctx.Step(`^this "([^"]*)" ([^"]*) event is produced:$`, ks.thisEncodedEventIsProduced)
	ctx.Step(`^no "([^"]*)" event is produced within (\d+) seconds$`, ks.noEventIsProducedInTime)
}

// Close cleans up any consumers and producers being used by the current scenairo once finished with
func (ks *KafkaScenario) Close(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	ks.mu.Lock()
	defer ks.mu.Unlock()

	for _, topic := range ks.topics {
		if producer := topic.producer; producer != nil {
			if err := producer.Close(ctx); err != nil {
				return err
			}
		}
		if consumer := topic.consumer; consumer != nil {
			if err := consumer.Close(ctx); err != nil {
				return err
			}
		}
	}

	return nil
}

func (ks *KafkaScenario) thisEventIsQueued(ctx context.Context, topic string, document *godog.DocString) error {
	return ks.thisEncodedEventIsQueued(ctx, topic, "JSON", document)
}

func (ks *KafkaScenario) thisEncodedEventIsQueued(ctx context.Context, topic, encoding string, document *godog.DocString) error {
	encoder, ok := ks.KafkaFeature.EventEncoders[topic][encoding]
	if !ok {
		encoder = compactJSON
	}

	// encode message
	wireMsg, err := encoder([]byte(document.Content))
	if err != nil {
		return err
	}

	producer := ks.getProducer(ctx, topic)
	return producer.SendBytes(ctx, wireMsg)
}

func (ks *KafkaScenario) thisEventIsProduced(ctx context.Context, topic string, document *godog.DocString) error {
	return ks.thisEncodedEventIsProduced(ctx, topic, "JSON", document)
}

func (ks *KafkaScenario) thisEncodedEventIsProduced(ctx context.Context, topic, encoding string, document *godog.DocString) error {
	encoder, ok := ks.KafkaFeature.EventEncoders[topic][encoding]
	if !ok {
		encoder = compactJSON
	}

	err := ks.startConsuming(ctx, topic)
	if err != nil {
		return err
	}
	scenarioTopic := ks.getScenarioTopic(topic)

	// encode expected document
	wantedEvent, err := encoder([]byte(document.Content))
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	errChan := make(chan error)

	go func() {
		for {
			select {
			case <-ticker.C:
				func() {
					scenarioTopic.mu.Lock()
					defer scenarioTopic.mu.Unlock()

					for _, msg := range scenarioTopic.ConsumedMessages {
						if bytes.Equal(msg, wantedEvent) {
							errChan <- nil
							ticker.Stop()
							return
						}
					}
				}()

			case <-ctx.Done():
				ticker.Stop()
				scenarioTopic.mu.Lock()
				received := bytes.Join(scenarioTopic.ConsumedMessages, []byte(","))
				scenarioTopic.mu.Unlock()
				errChan <- fmt.Errorf("no matching event was produced in time - actual [%s]", string(received))
				return
			}
		}
	}()

	// wait for done or error
	return <-errChan
}

func (ks *KafkaScenario) noEventIsProducedInTime(ctx context.Context, topic string, seconds int) error {
	err := ks.startConsuming(ctx, topic)
	if err != nil {
		return err
	}
	scenarioTopic := ks.getScenarioTopic(topic)

	ctx, cancel := context.WithTimeout(ctx, time.Duration(seconds)*time.Second)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	errChan := make(chan error)

	go func() {
		for {
			select {
			case <-ticker.C:
				if len(scenarioTopic.ConsumedMessages) > 0 {
					scenarioTopic.mu.Lock()
					received := bytes.Join(scenarioTopic.ConsumedMessages, []byte(","))
					scenarioTopic.mu.Unlock()
					errChan <- fmt.Errorf("unexpected event(s) produced in %d seconds - actual [%s]", seconds, string(received))
					ticker.Stop()
					return
				}
			case <-ctx.Done():
				ticker.Stop()
				errChan <- nil
				return
			}
		}
	}()

	// wait for done or error
	return <-errChan
}

func (ks *KafkaScenario) getScenarioTopic(topic string) *kafkaScenarioTopic {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	if ks.topics == nil {
		ks.topics = make(map[string]*kafkaScenarioTopic)
	}
	if _, ok := ks.topics[topic]; !ok {
		ks.topics[topic] = &kafkaScenarioTopic{
			topic:       topic,
			mappedTopic: fmt.Sprintf("%s-%s", topic, uuid.NewString()),
		}
	}
	return ks.topics[topic]
}

func (ks *KafkaScenario) getProducer(ctx context.Context, topic string) *kafka.Producer {
	scenarioTopic := ks.getScenarioTopic(topic)
	scenarioTopic.mu.Lock()
	defer scenarioTopic.mu.Unlock()
	if scenarioTopic.producer != nil {
		return scenarioTopic.producer
	}
	producer := ks.KafkaFeature.getProducer(ctx, scenarioTopic.mappedTopic)
	scenarioTopic.producer = producer
	return producer
}

func (ks *KafkaScenario) startConsuming(ctx context.Context, topic string) error {
	scenarioTopic := ks.getScenarioTopic(topic)
	scenarioTopic.mu.Lock()
	defer scenarioTopic.mu.Unlock()
	if scenarioTopic.consumer != nil {
		// already started - do nothing
		return nil
	}
	consumer := ks.KafkaFeature.getConsumer(ctx, scenarioTopic.mappedTopic)
	scenarioTopic.consumer = consumer

	handler := func(_ context.Context, _ int, msg kafka.Message) error {
		scenarioTopic.mu.Lock()
		defer scenarioTopic.mu.Unlock()
		scenarioTopic.ConsumedMessages = append(scenarioTopic.ConsumedMessages, msg.GetData())
		return nil
	}
	if err := consumer.RegisterHandler(ctx, handler); err != nil {
		return err
	}

	// Start consuming
	if err := consumer.Start(); err != nil {
		return err
	}
	return nil
}

// EventEncoder represents a function that can take in a JSON representation and output an encoded message
type EventEncoder func([]byte) ([]byte, error)

func compactJSON(data []byte) ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	err := json.Compact(buffer, data)
	if err != nil {
		return nil, fmt.Errorf("not a valid json document: %w", err)
	}
	return buffer.Bytes(), nil
}

// NewAvroEncoder creates a [EventEncoder] that encodes the model supplied using the supplied avro schema
func NewAvroEncoder[T any](schema *avro.Schema) func([]byte) ([]byte, error) {
	return func(jsonData []byte) ([]byte, error) {
		var e T
		err := json.Unmarshal(jsonData, &e)
		if err != nil {
			return nil, err
		}
		avroData, err := schema.Marshal(&e)
		if err != nil {
			return nil, err
		}
		return avroData, nil
	}
}

func (kf *KafkaFeature) getProducer(ctx context.Context, topic string) *kafka.Producer {
	minHealthy := 1
	version := kf.KafkaVersion
	pConfig := &kafka.ProducerConfig{
		BrokerAddrs:       kf.GetBrokers(ctx),
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

func (kf *KafkaFeature) getConsumer(ctx context.Context, topic string) *kafka.ConsumerGroup {
	kafkaOffset := kafka.OffsetOldest
	version := kf.KafkaVersion
	minHealthy := 1
	cgConfig := &kafka.ConsumerGroupConfig{
		BrokerAddrs:       kf.GetBrokers(ctx),
		Topic:             topic,
		GroupName:         "dummy_group",
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
