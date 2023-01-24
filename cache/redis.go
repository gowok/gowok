package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	goredis "github.com/go-redis/redis/v8"
	"github.com/gowok/gowok/driver"
)

type redis struct {
	client *goredis.Client
}

func NewRedis(config goredis.Options) (driver.Cache, error) {
	rdb := goredis.NewClient(&config)

	return &redis{
		client: rdb,
	}, nil
}

func (c redis) IsAvailable(ctx context.Context) bool {
	if c.client == nil {
		return true
	}

	err := c.client.Ping(ctx).Err()
	return err == nil
}

func (c redis) SetContext(ctx context.Context, key string, val any) error {
	if !c.IsAvailable(ctx) {
		return errors.New("cache is not available")
	}

	expireAt := time.Duration(0)

	ttlFromCtx, ok := ctx.Value("ttl").(time.Duration)
	if ok {
		expireAt = ttlFromCtx
	}

	jsonMarshal, err := json.Marshal(val)
	if err == nil {
		return c.client.Set(ctx, key, string(jsonMarshal), expireAt).Err()
	}

	return c.client.Set(ctx, key, val, expireAt).Err()
}

func (c redis) GetContext(ctx context.Context, key string) (any, error) {
	if !c.IsAvailable(ctx) {
		return nil, errors.New("cache is not available")
	}

	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var jsonUnmarshal any
	err = json.Unmarshal([]byte(val), &jsonUnmarshal)
	if err == nil {
		return jsonUnmarshal, nil
	}

	return val, nil
}
