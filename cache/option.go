package cache

import "time"

type GetOption func(*getOption)
type getOption struct{}

type SetOption func(*setOption)
type setOption struct {
	ttl time.Duration
}

func WithTTL(ttl time.Duration) SetOption {
	return func(o *setOption) {
		o.ttl = ttl
	}
}
