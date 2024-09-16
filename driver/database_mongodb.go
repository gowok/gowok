package driver

import (
	"context"
	"log/slog"

	"github.com/gowok/gowok/config"
	"github.com/gowok/gowok/optional"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB map[string]*mongo.Client

func NewMongoDB(config map[string]config.MongoDB) (MongoDB, error) {
	mongos := MongoDB{}
	c := context.Background()

	for name, dbC := range config {
		if !dbC.Enabled {
			continue
		}

		opts := options.Client().ApplyURI(dbC.DSN)
		client, err := mongo.Connect(c, opts)
		if err != nil {
			return nil, err
		}

		mongos[name] = client
	}

	return mongos, nil
}

func (d MongoDB) Get(name ...string) optional.Optional[*mongo.Client] {
	n := ""
	if len(name) > 0 {
		n = name[0]
		if db, ok := d[n]; ok {
			return optional.Of(&db)
		}
	}

	if n != "" {
		slog.Info("using default connection", "not_found", n)
	}

	if db, ok := d["default"]; ok {
		return optional.Of(&db)
	}

	return optional.Empty[*mongo.Client]()
}
