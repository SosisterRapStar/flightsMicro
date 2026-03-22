package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/IBM/sarama"
	log "github.com/sirupsen/logrus"

	"github.com/SosisterRapStar/cliros/message"

	infrakafka "github.com/SosisterRapStar/flights/internal/infrastructure/kafka"
)

type SagaPubsub struct {
	producer      sarama.SyncProducer
	consumerGroup sarama.ConsumerGroup

	mu       sync.RWMutex
	closed   bool
	handlers map[string]func(context.Context, message.Message) error

	consumerWg sync.WaitGroup
}

func NewSagaPubsub(cfg *infrakafka.Config) (*SagaPubsub, error) {
	if cfg == nil {
		return nil, fmt.Errorf("kafka config is required")
	}
	useCfg := cfg
	if cfg.GroupID == "" {
		useCfg = &infrakafka.Config{
			Brokers:          cfg.Brokers,
			GroupID:          "leti-paper-saga",
			AckPolicy:        cfg.AckPolicy,
			RetryMax:         cfg.RetryMax,
			AutoCommitEnable: cfg.AutoCommitEnable,
			MaxWaitTime:      cfg.MaxWaitTime,
		}
	}

	producer, err := infrakafka.NewSyncProducer(useCfg)
	if err != nil {
		return nil, fmt.Errorf("kafka sync producer: %w", err)
	}

	consumerGroup, err := infrakafka.NewConsumerGroup(useCfg)
	if err != nil {
		_ = producer.Close()
		return nil, fmt.Errorf("kafka consumer group: %w", err)
	}

	return &SagaPubsub{
		producer:      producer,
		consumerGroup: consumerGroup,
		handlers:      make(map[string]func(context.Context, message.Message) error),
	}, nil
}

// Publish реализует broker.Publisher. Сериализует message.Message в JSON и отправляет в топик.
func (s *SagaPubsub) Publish(ctx context.Context, topic string, msg message.Message) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return fmt.Errorf("saga pubsub is closed")
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	_, _, err = s.producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(data),
	})
	if err != nil {
		return fmt.Errorf("kafka send: %w", err)
	}
	return nil
}

// Subscribe реализует broker.Subsciber.
// Регистрирует handler для указанного топика; для старта потребления нужно вызвать Run.
func (s *SagaPubsub) Subscribe(ctx context.Context, topic string, handler func(context.Context, message.Message) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return fmt.Errorf("saga pubsub is closed")
	}
	if _, exists := s.handlers[topic]; exists {
		return fmt.Errorf("already subscribed to topic: %s", topic)
	}
	s.handlers[topic] = handler
	return nil
}

// Run запускает consumer group и обрабатывает сообщения для всех подписанных топиков.
// Обычно вызывается один раз на всё приложение в отдельной горутине.
func (s *SagaPubsub) Run(ctx context.Context) error {
	s.mu.RLock()
	topics := make([]string, 0, len(s.handlers))
	for t := range s.handlers {
		topics = append(topics, t)
	}
	s.mu.RUnlock()

	if len(topics) == 0 {
		return nil
	}

	consumer := &consumerAdapter{
		handlers: s.handlers,
		mu:       &s.mu,
	}

	s.consumerWg.Add(1)
	go func() {
		defer s.consumerWg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if err := s.consumerGroup.Consume(ctx, topics, consumer); err != nil {
					return
				}
			}
		}
	}()

	return nil
}

// Close завершает работу consumer-group и закрывает producer.
func (s *SagaPubsub) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	s.mu.Unlock()

	s.consumerWg.Wait()
	_ = s.consumerGroup.Close()
	return s.producer.Close()
}

// consumerAdapter адаптирует сообщения Kafka к message.Message из LETI-paper.
type consumerAdapter struct {
	handlers map[string]func(context.Context, message.Message) error
	mu       *sync.RWMutex
}

var _ sarama.ConsumerGroupHandler = (*consumerAdapter)(nil)

func (c *consumerAdapter) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (c *consumerAdapter) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (c *consumerAdapter) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		c.mu.RLock()
		handler, ok := c.handlers[msg.Topic]
		c.mu.RUnlock()
		if !ok {
			session.MarkMessage(msg, "")
			continue
		}

		var m message.Message
		if err := json.Unmarshal(msg.Value, &m); err != nil {
			preview := string(msg.Value)
			if len(preview) > 500 {
				preview = preview[:500] + "..."
			}
			log.WithError(err).WithFields(log.Fields{
				"topic":         msg.Topic,
				"partition":     msg.Partition,
				"offset":        msg.Offset,
				"value_len":     len(msg.Value),
				"value_preview": preview,
			}).Warn("failed to unmarshal saga message")
			session.MarkMessage(msg, "")
			continue
		}

		if err := handler(context.Background(), m); err != nil {
			continue
		}

		session.MarkMessage(msg, "")
	}
	return nil
}
