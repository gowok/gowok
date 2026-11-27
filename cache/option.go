package cache

import "time"

type GetOption func(*getOption)
type getOption struct {
	fallback any
}

type SetOption func(*setOption)
type setOption struct {
	ttl time.Duration
}

func WithFallback(fallback any) GetOption {
	return func(o *getOption) {
		o.fallback = fallback
	}
}

func WithTTL(ttl time.Duration) SetOption {
	return func(o *setOption) {
		o.ttl = ttl
	}
}
