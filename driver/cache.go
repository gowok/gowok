package driver

import "time"

type Cache interface {
	Set(key string, value any, ttl ...time.Duration)
	Get(key string) (any, error)
	IsAvailable() bool
}
