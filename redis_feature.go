package componenttest

import (
	"context"
	"fmt"

	disRedis "github.com/ONSdigital/dis-redis"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/alicebob/miniredis/v2"
	"github.com/cucumber/godog"
)

// RedisFeature is a struct containing an in-memory redis database
type RedisFeature struct {
	Server *miniredis.Miniredis
}

// NewRedisFeature creates a new in-memory redis database using the supplied options
func NewRedisFeature() *RedisFeature {
	s := miniredis.NewMiniRedis()

	err := s.StartAddr("localhost:6379")
	if err != nil {
		panic(err)
	}

	return &RedisFeature{
		Server: s,
	}
}

// Reset drops all keys from the in memory server
func (r *RedisFeature) Reset() error {
	r.Server.FlushAll()
	return nil
}

// Close stops the in-memory redis database
func (r *RedisFeature) Close() error {
	r.Server.Close()
	return nil
}

func (r *RedisFeature) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^the key "([^"]*)" is already set to a value of "([^"]*)" in the Redis store$`, r.theKeyIsAlreadySetToAValueOfInTheRedisStore)
	ctx.Step(`^the key "([^"]*)" has a value of "([^"]*)" in the Redis store$`, r.theKeyHasAValueOfInTheRedisStore)
	ctx.Step(`^redis contains no value for key "([^"]*)"$`, r.redisContainsNoValueFor)
	ctx.Step(`^redis is healthy$`, r.redisIsHealthy)
	ctx.Step(`^redis stops running$`, r.redisStopsRunning)
}

func (r *RedisFeature) theKeyIsAlreadySetToAValueOfInTheRedisStore(key, value string) error {
	return r.Server.Set(key, value)
}

func (r *RedisFeature) theKeyHasAValueOfInTheRedisStore(key, expected string) error {
	actual, err := r.Server.Get(key)
	if err != nil {
		return fmt.Errorf("failed to get key %q from Redis: %w", key, err)
	}

	if actual != expected {
		return fmt.Errorf("unexpected value for key %q: got %q, want %q", key, actual, expected)
	}

	return nil
}

func (r *RedisFeature) redisContainsNoValueFor(key string) error {
	if r.Server.Exists(key) {
		val, _ := r.Server.Get(key)
		return fmt.Errorf("expected no value for key %q, but found %q", key, val)
	}
	return nil
}

func (r *RedisFeature) redisIsHealthy() error {
	ctx := context.Background()
	clientConfig := &disRedis.ClientConfig{}
	redisClient, err := disRedis.NewClient(ctx, clientConfig)

	if err != nil {
		log.Error(ctx, "Failed to create dis-redis client", err)
		return err
	}

	_, err = redisClient.Ping(ctx)

	return err
}

func (r *RedisFeature) redisStopsRunning() error {
	if r.Server != nil {
		r.Server.Close()
	}
	return nil
}
