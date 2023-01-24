package cache

import (
	"context"
	"errors"

	"github.com/gowok/gowok/driver"
)

type def struct {
}

func NewDefault() (driver.Cache, error) {
	return &def{}, nil
}

func (c def) IsAvailable(ctx context.Context) bool {
	return false
}

func (c def) SetContext(ctx context.Context, key string, val any) error {
	return errors.New("cache is not available")
}

func (c def) GetContext(ctx context.Context, key string) (any, error) {
	return nil, errors.New("cache is not available")
}
