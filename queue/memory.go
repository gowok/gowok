package queue

import (
	"context"
	"errors"
	"strconv"
	"sync"
)

var (
	ErrClosed         = errors.New("queue: closed")
	ErrUnknownReceipt = errors.New("queue: unknown receipt")
)

type memory struct {
	mu       sync.Mutex
	queues   map[string][]Delivery
	inflight map[string]memoryDelivery
	notify   map[string]chan struct{}
	closed   bool
	nextID   uint64
}

type memoryDelivery struct {
	delivery Delivery
	topic    string
}

func Memory() *memory {
	return &memory{
		queues:   make(map[string][]Delivery),
		inflight: make(map[string]memoryDelivery),
		notify:   make(map[string]chan struct{}),
	}
}

func (m *memory) Enqueue(ctx context.Context, topic string, message Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return ErrClosed
	}

	m.queues[topic] = append(m.queues[topic], Delivery{Message: copyMessage(message)})
	m.signal(topic)
	return nil
}

func (m *memory) Receive(ctx context.Context, topic string) (Delivery, error) {
	for {
		if err := ctx.Err(); err != nil {
			return Delivery{}, err
		}

		m.mu.Lock()
		if m.closed {
			m.mu.Unlock()
			return Delivery{}, ErrClosed
		}

		if len(m.queues[topic]) > 0 {
			delivery := m.queues[topic][0]
			m.queues[topic][0] = Delivery{}
			m.queues[topic] = m.queues[topic][1:]

			m.nextID++
			delivery.Receipt = strconv.FormatUint(m.nextID, 10)
			m.inflight[delivery.Receipt] = memoryDelivery{delivery: delivery, topic: topic}

			m.mu.Unlock()
			return delivery, nil
		}

		notify := m.notify[topic]
		if notify == nil {
			notify = make(chan struct{})
			m.notify[topic] = notify
		}
		m.mu.Unlock()

		select {
		case <-notify:
		case <-ctx.Done():
			return Delivery{}, ctx.Err()
		}
	}
}

func (m *memory) Ack(ctx context.Context, receipt string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.inflight[receipt]; !ok {
		return ErrUnknownReceipt
	}

	delete(m.inflight, receipt)
	return nil
}

func (m *memory) Nack(ctx context.Context, receipt string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	item, ok := m.inflight[receipt]
	if !ok {
		return ErrUnknownReceipt
	}

	delete(m.inflight, receipt)
	item.delivery.Receipt = ""
	m.queues[item.topic] = append(m.queues[item.topic], item.delivery)
	m.signal(item.topic)
	return nil
}

func (m *memory) Close(_ context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return nil
	}

	m.closed = true
	m.queues = make(map[string][]Delivery)
	m.inflight = make(map[string]memoryDelivery)

	for topic, notify := range m.notify {
		close(notify)
		delete(m.notify, topic)
	}

	return nil
}

func (m *memory) signal(topic string) {
	if notify, ok := m.notify[topic]; ok {
		close(notify)
		delete(m.notify, topic)
	}
}

func copyMessage(message Message) Message {
	if message.Payload != nil {
		payload := make([]byte, len(message.Payload))
		copy(payload, message.Payload)
		message.Payload = payload
	}
	if message.Headers != nil {
		headers := make(map[string]string, len(message.Headers))
		for key, value := range message.Headers {
			headers[key] = value
		}
		message.Headers = headers
	}
	return message
}

var _ Queue = (*memory)(nil)
