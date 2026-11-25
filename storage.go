package gowok

import "github.com/gowok/gowok/storage"

func Storage(name ...string) storage.Storage {
	return storage.Client(name...)
}
