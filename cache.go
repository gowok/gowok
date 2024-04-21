package gowok

import (
	"log/slog"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	store_redis "github.com/eko/gocache/store/redis/v4"
	"github.com/gowok/gowok/config"
	"github.com/gowok/gowok/optional"
	"github.com/redis/go-redis/v9"
)

type Cache map[string]store.StoreInterface

func NewCache(config map[string]config.Cache) (Cache, error) {
	redises := Cache{}

	for name, dbC := range config {
		if dbC.Driver == "redis" {
			clientOpt := Must(redis.ParseURL(dbC.DSN))
			client := store_redis.NewRedis(redis.NewClient(clientOpt))
			redises[name] = client
		}
	}

	return redises, nil
}

func (d Cache) Get(name ...string) optional.Optional[*cache.Cache[any]] {
	n := ""
	if len(name) > 0 {
		n = name[0]
		if db, ok := d[n]; ok {
			c := cache.New[any](db)
			return optional.New(&c)
		}
	}

	if n != "" {
		slog.Info("using default connection", "not_found", n)
	}

	if db, ok := d["default"]; ok {
		c := cache.New[any](db)
		return optional.New(&c)
	}

	return optional.Empty[*cache.Cache[any]]()
}
