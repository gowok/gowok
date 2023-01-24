package driver

import (
	"context"

	"github.com/gowok/gowok/driver/database"
)

type Cache interface {
	database.KVReader
	database.KVWriter
	IsAvailable(ctx context.Context) bool
}
