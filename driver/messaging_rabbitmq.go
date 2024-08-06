package driver

import (
	"github.com/gowok/gowok/driver/messaging"
	"github.com/wagslane/go-rabbitmq"
)

type Messaging interface {
	Publish(topic string, channel string, message messaging.Message) error
	Consume(channel string) (<-chan messaging.Message, error)
	Ack(message messaging.Message) error
	IsAvailable() bool
}

func NewRabbitMQ(dsn string) (*rabbitmq.Conn, error) {
	conn, err := rabbitmq.NewConn(
		dsn,
		rabbitmq.WithConnectionOptionsLogging,
	)

	return conn, err
}

func NewRabbitMQConsumer(
	conn *rabbitmq.Conn,
	queue string,
	routingKey string,
	exchangeName string,
	opts ...func(*rabbitmq.ConsumerOptions),
) (*rabbitmq.Consumer, error) {
	if opts == nil {
		opts = make([]func(*rabbitmq.ConsumerOptions), 0)
	}
	if routingKey != "" {
		opts = append(opts, rabbitmq.WithConsumerOptionsRoutingKey(routingKey))
	}
	if exchangeName != "" {
		opts = append(
			opts,
			rabbitmq.WithConsumerOptionsExchangeName(exchangeName),
			rabbitmq.WithConsumerOptionsExchangeDeclare,
		)
	}
	consumer, err := rabbitmq.NewConsumer(
		conn,
		queue,
		opts...,
	)

	return consumer, err
}

func NewRabbitMQPublisher(
	conn *rabbitmq.Conn,
	exchangeName string,
	opts ...func(*rabbitmq.PublisherOptions),
) (*rabbitmq.Publisher, error) {
	if opts == nil {
		opts = make([]func(*rabbitmq.PublisherOptions), 0)
	}
	opts = append(
		opts,
		rabbitmq.WithPublisherOptionsLogging,
	)
	if exchangeName != "" {
		opts = append(
			opts,
			rabbitmq.WithPublisherOptionsExchangeName(exchangeName),
			rabbitmq.WithPublisherOptionsExchangeDeclare,
		)
	}
	publisher, err := rabbitmq.NewPublisher(
		conn,
		opts...,
	)

	return publisher, err
}
