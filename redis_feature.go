package componenttest

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/cucumber/godog"
	testRedis "github.com/testcontainers/testcontainers-go/modules/redis"
)

// RedisFeature is a struct containing a testcontainer redis database
type RedisFeature struct {
	Server *testRedis.RedisContainer
	Client *redis.Client
}

type RedisOptions struct {
	RedisVersion string
}

// NewRedisFeature creates a new testcontainer redis database using the supplied options
func NewRedisFeature(opts RedisOptions) *RedisFeature {
	ctx := context.Background()

	s, err := testRedis.Run(ctx, fmt.Sprintf("redis:%s", opts.RedisVersion))
	if err != nil {
		panic(err)
	}

	err = s.Start(ctx)
	if err != nil {
		panic(err)
	}

	connectionString, err := s.ConnectionString(ctx)
	if err != nil {
		panic(err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: strings.ReplaceAll(connectionString, "redis://", ""),
	})

	return &RedisFeature{
		Server: s,
		Client: client,
	}
}

// Reset drops all keys from the testcontainer redis
func (r *RedisFeature) Reset() error {
	return r.Client.FlushAll(context.Background()).Err()
}

// Close stops the testcontainer redis
func (r *RedisFeature) Close() error {
	return r.Server.Terminate(context.Background())
}

func (r *RedisFeature) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^the key "([^"]*)" is already set to a value of "([^"]*)" in the Redis store$`, r.theKeyIsAlreadySetToAValueOfInTheRedisStore)
	ctx.Step(`^the key "([^"]*)" has a value of "([^"]*)" in the Redis store$`, r.theKeyHasAValueOfInTheRedisStore)
	ctx.Step(`^the key "([^"]*)" has a value of a set of the following values in the Redis store$`, r.theKeyHasAValueOfASetOfValuesInTheRedisStore)
	ctx.Step(`^the key "([^"]*)" is already set to a value of a set of the following values in the Redis store$`, r.theKeyIsAlreadySetToAValueOfASetOfValuesInTheRedisStore)
	ctx.Step(`^redis contains no value for key "([^"]*)"$`, r.redisContainsNoValueFor)
	ctx.Step(`^redis is healthy$`, r.redisIsHealthy)
	ctx.Step(`^redis stops running$`, r.redisStopsRunning)
}

func (r *RedisFeature) theKeyHasAValueOfASetOfValuesInTheRedisStore(key string, table *godog.Table) error {
	values := make([]interface{}, 0, len(table.Rows)-1)

	for i := range table.Rows {
		if i > 0 {
			values = append(values, table.Rows[i].Cells[0].Value)
		}
	}

	actual := r.Client.SMembers(context.Background(), key).Val()

	if len(actual) != len(values) {
		return fmt.Errorf("unexpected number of members for key %q: got %d, want %d", key, len(actual), len(values))
	}
	for _, v := range values {
		found := false
		for _, a := range actual {
			if a == v {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("expected value %q for key %q not found in Redis set", v, key)
		}
	}
	return nil
}

func (r *RedisFeature) theKeyIsAlreadySetToAValueOfASetOfValuesInTheRedisStore(key string, table *godog.Table) error {
	values := make([]interface{}, 0, len(table.Rows)-1)

	for i := range table.Rows {
		if i > 0 {
			values = append(values, table.Rows[i].Cells[0].Value)
		}
	}

	return r.Client.SAdd(context.Background(), key, values...).Err()
}

func (r *RedisFeature) theKeyIsAlreadySetToAValueOfInTheRedisStore(key, value string) error {
	return r.Client.Set(context.Background(), key, value, 0).Err()
}

func (r *RedisFeature) theKeyHasAValueOfInTheRedisStore(key, expected string) error {
	actual, err := r.Client.Get(context.Background(), key).Result()
	if err != nil {
		return fmt.Errorf("failed to get key %q from Redis: %w", key, err)
	}

	if actual != expected {
		return fmt.Errorf("unexpected value for key %q: got %q, want %q", key, actual, expected)
	}

	return nil
}

func (r *RedisFeature) redisContainsNoValueFor(key string) error {
	ctx := context.Background()
	exists, err := r.Client.Exists(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("error checking existence of key %q", key)
	}

	if exists == 1 {
		val, err := r.Client.Get(ctx, key).Result()
		if err != nil {
			return fmt.Errorf("error getting value of key %q", key)
		}
		return fmt.Errorf("expected no value for key %q, but found %q", key, val)
	}
	return nil
}

func (r *RedisFeature) redisIsHealthy() error {
	return r.Client.Ping(context.Background()).Err()
}

func (r *RedisFeature) redisStopsRunning() error {
	if r.Server != nil {
		timeout := time.Second * 1
		return r.Server.Stop(context.Background(), &timeout)
	}
	return nil
}
