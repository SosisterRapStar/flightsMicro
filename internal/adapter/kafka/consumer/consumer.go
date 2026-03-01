package consumer

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/IBM/sarama"

	"github.com/SosisterRapStar/LETI-PaperTestMicroservices/internal/domain/ports"
)

// Consumer реализует ports.BrokerConsumer поверх sarama.ConsumerGroup.
type Consumer struct {
	router map[string]ports.MessageHandler
	mu     sync.RWMutex
	client sarama.ConsumerGroup
	topics []string
}

// New создаёт Consumer. Принимает уже готовый sarama.ConsumerGroup
// (создаётся через infrastructure/kafka.NewConsumerGroup) и список топиков.
func New(client sarama.ConsumerGroup, topics []string) (*Consumer, error) {
	if client == nil {
		return nil, fmt.Errorf("consumer group client is required")
	}
	if len(topics) == 0 {
		return nil, fmt.Errorf("topics are required")
	}
	return &Consumer{
		router: make(map[string]ports.MessageHandler, 10),
		client: client,
		topics: topics,
	}, nil
}

// RegisterHandler реализует ports.BrokerConsumer.
func (c *Consumer) RegisterHandler(topic string, h ports.MessageHandler) {
	c.mu.Lock()
	c.router[topic] = h
	c.mu.Unlock()
}

func (c *Consumer) getRouter() map[string]ports.MessageHandler {
	c.mu.RLock()
	defer c.mu.RUnlock()
	cp := make(map[string]ports.MessageHandler, len(c.router))
	for k, v := range c.router {
		cp[k] = v
	}
	return cp
}

// Run реализует ports.BrokerConsumer. Блокируется до отмены ctx.
func (c *Consumer) Run(ctx context.Context) error {
	handler := &consumeHandler{consumer: c}
	for {
		if err := c.client.Consume(ctx, c.topics, handler); err != nil {
			if err == sarama.ErrClosedConsumerGroup {
				return nil
			}
			log.Printf("consumer error: %v", err)
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

// Close реализует ports.BrokerConsumer.
func (c *Consumer) Close() error {
	return c.client.Close()
}

var _ ports.BrokerConsumer = (*Consumer)(nil)

type consumeHandler struct {
	consumer *Consumer
}

func (h *consumeHandler) Setup(sarama.ConsumerGroupSession) error {
	log.Println("kafka consumer: new session created")
	return nil
}

func (h *consumeHandler) Cleanup(sarama.ConsumerGroupSession) error {
	log.Println("kafka consumer: group rebalancing")
	return nil
}

func (h *consumeHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	router := h.consumer.getRouter()
	for msg := range claim.Messages() {
		if session.Context().Err() != nil {
			return nil
		}
		handler, ok := router[msg.Topic]
		if !ok {
			log.Printf("no handler for topic %s, timestamp %s", msg.Topic, msg.Timestamp)
			session.MarkMessage(msg, "")
			continue
		}
		pmsg := ports.Message{
			Topic: msg.Topic,
			Key:   msg.Key,
			Value: msg.Value,
		}
		if err := handler(context.Background(), pmsg); err != nil {
			log.Printf("handler error for topic %s: %v", msg.Topic, err)
			continue
		}
		session.MarkMessage(msg, "")
	}
	return nil
}
