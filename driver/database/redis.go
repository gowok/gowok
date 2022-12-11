package database

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type Redis struct {
	*redis.Client
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
