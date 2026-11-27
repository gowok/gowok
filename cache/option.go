package cache

import "time"

type GetOptionFunc func(*GetOption)
type GetOption struct {
	Fallback any
}

type SetOptionFunc func(*SetOption)
type SetOption struct {
	TTL time.Duration
}

func WithFallback(fallback any) GetOptionFunc {
	return func(o *GetOption) {
		o.Fallback = fallback
	}
}

func WithTTL(ttl time.Duration) SetOptionFunc {
	return func(o *SetOption) {
		o.TTL = ttl
	}
}
