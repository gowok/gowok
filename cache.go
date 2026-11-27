package gowok

import "github.com/gowok/gowok/cache"

func Cache(name ...string) cache.Cache {
	return cache.Client(name...)
}
