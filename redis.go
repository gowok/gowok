package gowok

import (
	"log/slog"

	"github.com/go-redis/redis/v8"
	"github.com/gowok/gowok/config"
	"github.com/gowok/gowok/optional"
)

type Redis map[string]*redis.Client

func NewRedis(config map[string]config.Cache) (Redis, error) {
	redises := Redis{}

	for name, dbC := range config {
		if dbC.Driver == "redis" {
			clientOpt := Must(redis.ParseURL(dbC.DSN))
			client := redis.NewClient(clientOpt)

			redises[name] = client
		}
	}

	return redises, nil
}

func (d Redis) Get(name ...string) optional.Optional[*redis.Client] {
	n := ""
	if len(name) > 0 {
		n = name[0]
		if db, ok := d[n]; ok {
			return optional.New(&db)
		}
	}

	if n != "" {
		slog.Info("using default connection", "not_found", n)
	}

	if db, ok := d["default"]; ok {
		return optional.New(&db)
	}

	var db *redis.Client
	return optional.New(&db)
}
