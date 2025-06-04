package componenttest

import (
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
	ctx.Step(`^the key "([^"]*)" has a value of "([^"]*)" in the Redis store$`, r.theKeyHasAValueOfInTheRedisStore)
}

func (r *RedisFeature) theKeyHasAValueOfInTheRedisStore(key, value string) error {
	return r.Server.Set(key, value)
}
