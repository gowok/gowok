package driver

import "github.com/gowok/gowok/driver/messaging"

type Messaging interface {
	Publish(topic string, channel string, message messaging.Message) error
	Consume(channel string) (<-chan messaging.Message, error)
	Ack(message messaging.Message) error
	IsAvailable() bool
}
