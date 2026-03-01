package ports

import "context"

// Message — сообщение брокера без привязки к конкретной реализации.
type Message struct {
	Topic string
	Key   []byte
	Value []byte
}

// MessageHandler — обработчик сообщения брокера.
type MessageHandler func(ctx context.Context, msg Message) error

// BrokerConsumer — порт для потребления сообщений из брокера.
type BrokerConsumer interface {
	RegisterHandler(topic string, h MessageHandler)
	Run(ctx context.Context) error
	Close() error
}

// BrokerProducer — порт для отправки сообщений в брокер.
type BrokerProducer interface {
	SendMessage(ctx context.Context, topic string, key, value []byte) error
	Close() error
}
