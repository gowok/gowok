package database

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/gowok/gowok/config"
)

type Redis struct {
	*redis.Client
}

var _ KVClient = Redis{}
var _ KVReader = Redis{}
var _ KVWriter = Redis{}

func NewRedis(conf config.Database) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr: conf.DSN,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return &Redis{client}, err
}

func (kv Redis) PingContext(ctx context.Context) error {
	_, err := kv.Ping(ctx).Result()
	if err != nil {
		return err
	}

	return nil
}

func (kv Redis) SetContext(ctx context.Context, key string, value any) error {
	_, err := kv.Set(ctx, key, value, 0).Result()
	if err != nil {
		return err
	}

	return nil
}

func (kv Redis) GetContext(ctx context.Context, key string) (any, error) {
	data, err := kv.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	return data, nil
}
