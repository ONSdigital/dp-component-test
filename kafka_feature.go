package componenttest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	kafka "github.com/ONSdigital/dp-kafka/v4"
	"github.com/cucumber/godog"
	tckafka "github.com/testcontainers/testcontainers-go/modules/kafka"
)

// KafkaFeature represents a component test feature that tests kafka functionality via testcontainers
type KafkaFeature struct {
	kafkaContainer *tckafka.KafkaContainer
	KafkaVersion   string
}

const defaultKafkaContainerName = "confluentinc/confluent-local:7.5.0"
const defaultKafkaVersion = "3.8.0"

// KafkaOptions are optional configuration options for the kafka feature initialisation
type KafkaOptions struct {
	ContainerName string
	KafkaVersion  string
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

	return &KafkaFeature{
		kafkaContainer: kafkaContainer,
		KafkaVersion:   opts.KafkaVersion,
	}
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

// Close stops the kafka testcontainer
func (kf *KafkaFeature) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
	defer cancel()

	return kf.kafkaContainer.Terminate(ctx)
}

type topicMapKeyType string

var topicMapKey = topicMapKeyType("topicMapKey")

// ContextWithTopicMap adds a mapping from a named topic to a specific topic for the current context, this allows an app
// component to define random topics per scenario but for the feature scenarios to reference them by easy names
func (kf *KafkaFeature) ContextWithTopicMap(ctx context.Context, from, to string) context.Context {
	var topicMap map[string]string
	if ctxTopicMap := ctx.Value(topicMapKey); ctxTopicMap != nil {
		topicMap = ctxTopicMap.(map[string]string)
	} else {
		topicMap = make(map[string]string)
	}
	topicMap[from] = to
	return context.WithValue(ctx, topicMapKey, topicMap)
}

// RegisterSteps adds the kafka feature's steps to the godog ScenarioContext
func (kf *KafkaFeature) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^this "([^"]*)" JSON event is queued, to be consumed:$`, kf.thisJSONEventIsQueued)
	ctx.Step(`^this "([^"]*)" JSON event is produced:$`, kf.thisJSONEventIsProduced)
	ctx.Step(`^no "([^"]*)" JSON event is produced within (\d+) seconds$`, kf.noJSONEventIsProducedInTime)
}

func (kf *KafkaFeature) thisJSONEventIsQueued(ctx context.Context, topic string, document *godog.DocString) error {
	topic = kf.unmapTopic(ctx, topic)

	// ensure document is valid json
	if !json.Valid([]byte(document.Content)) {
		return fmt.Errorf("not a valid json document")
	}

	producer := kf.getProducer(ctx, topic)
	return producer.SendBytes(ctx, []byte(document.Content))
}

func (kf *KafkaFeature) thisJSONEventIsProduced(ctx context.Context, topic string, document *godog.DocString) error {
	topic = kf.unmapTopic(ctx, topic)

	// ensure expected document is valid json
	buffer := bytes.NewBuffer([]byte{})
	err := json.Compact(buffer, []byte(document.Content))
	if err != nil {
		return fmt.Errorf("not a valid json document: %w", err)
	}
	wantedDoc := buffer.Bytes()

	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	done := make(chan []byte)

	consumer := kf.getConsumer(ctx, topic)
	handler := func(_ context.Context, _ int, msg kafka.Message) error {
		done <- msg.GetData()
		return nil
	}
	if err := consumer.RegisterHandler(ctx, handler); err != nil {
		return err
	}
	defer consumer.Close(ctx)

	// Start consuming
	if err := consumer.Start(); err != nil {
		return err
	}

	// wait for done or timeout
	select {
	case msg := <-done:
		msgBuffer := bytes.NewBuffer([]byte{})
		err := json.Compact(msgBuffer, msg)
		if err != nil {
			return fmt.Errorf("not a valid json event: %w", err)
		}
		gotDoc := msgBuffer.Bytes()

		if !bytes.Equal(gotDoc, wantedDoc) {
			return fmt.Errorf("expected produced event to contain '%s', got '%s'", document.Content, string(msg))
		}
		return nil
	case <-ctx.Done():
		return fmt.Errorf("no event was produced in time")
	}
}

func (kf *KafkaFeature) noJSONEventIsProducedInTime(ctx context.Context, topic string, seconds int) error {
	topic = kf.unmapTopic(ctx, topic)

	ctx, cancel := context.WithTimeout(ctx, time.Duration(seconds)*time.Second)
	defer cancel()

	eventRecieved := make(chan bool)

	consumer := kf.getConsumer(ctx, topic)
	handler := func(_ context.Context, _ int, _ kafka.Message) error {
		eventRecieved <- true
		return nil
	}
	if err := consumer.RegisterHandler(ctx, handler); err != nil {
		return err
	}
	defer consumer.Close(ctx)

	// Start consuming
	if err := consumer.Start(); err != nil {
		return err
	}

	// wait for eventRecieved or timeout
	select {
	case <-eventRecieved:
		return fmt.Errorf("unexpected event produced in %d seconds", seconds)
	case <-ctx.Done():
		return nil
	}
}

// unmap topic will check the context for any mapped topics and if so will see if a mapping has been defined for the
// requested topic. If so it will use that instead.
func (kf *KafkaFeature) unmapTopic(ctx context.Context, topic string) string {
	if topicMap := ctx.Value(topicMapKey); topicMap != nil {
		if v, ok := topicMap.(map[string]string)[topic]; ok {
			topic = v
		}
	}
	return topic
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
