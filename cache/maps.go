package cache

import (
	"context"
	"fmt"

	"github.com/gowok/gowok/driver"
)

type maps map[string]any

func NewMap() (driver.Cache, error) {
	return &maps{}, nil
}

func (c maps) IsAvailable(ctx context.Context) bool {
	return false
}

func (c *maps) SetContext(ctx context.Context, key string, val any) error {
	d := *c
	d[key] = val
	return nil
}

func (c maps) GetContext(ctx context.Context, key string) (any, error) {
	data, ok := c[key]
	if !ok {
		return nil, fmt.Errorf("no cache with key %s", key)
	}

	return data, nil
}
