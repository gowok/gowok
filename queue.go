package gowok

import (
	"github.com/gowok/fp/some"
	"github.com/gowok/gowok/queue"
)

func Queue(name ...string) some.Some[queue.Queue] {
	return queue.Client(name...)
}
