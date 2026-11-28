package queue

import (
	"context"
	"sync"
	"time"

	"github.com/gowok/fp/some"
)

type Message struct {
	ID        string
	Payload   []byte
	Headers   map[string]string
	CreatedAt time.Time
}

type Delivery struct {
	Message Message
	Receipt string
}

type Queue interface {
	Enqueue(ctx context.Context, topic string, message Message) error
	Receive(ctx context.Context, topic string) (Delivery, error)
	Ack(ctx context.Context, receipt string) error
	Nack(ctx context.Context, receipt string) error
	Close(ctx context.Context) error
}

var pool sync.Map

func Client(name ...string) some.Some[Queue] {
	clientName := "default"
	if len(name) > 0 && name[0] != "" {
		clientName = name[0]
	}

	value, ok := pool.Load(clientName)
	if !ok {
		return some.Empty[Queue]()
	}

	return some.Of(value.(Queue))
}

func Register(name string, client Queue) {
	if name == "" {
		panic("queue: name is empty")
	}
	if client == nil {
		panic("queue: client is nil")
	}

	pool.Store(name, client)
}
