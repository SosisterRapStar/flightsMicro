// Package broker предоставляет адаптеры брокера сообщений под интерфейсы LETI-paper
// (Publisher и Subscriber) для тестирования фреймворка в микросервисной архитектуре.
package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/IBM/sarama"

	"github.com/SosisterRapStar/LETI-paper/domain/message"
)

// KafkaConfig настройки подключения к Kafka для адаптера LETI-paper.
type KafkaConfig struct {
	Brokers []string
	GroupID string
}

// Kafka реализует broker.Publisher и broker.Subscriber для Kafka (sarama).
// Используется в тестах LETI-paper вместо NATS.
type Kafka struct {
	producer      sarama.SyncProducer
	consumerGroup sarama.ConsumerGroup
	handlers      map[string]func(context.Context, message.Message) error
	mu            sync.RWMutex
	closed        bool
	consumerCtx   context.Context
	consumerCancel context.CancelFunc
	consumerWg    sync.WaitGroup
}

// NewKafka создаёт клиент Kafka, реализующий интерфейсы LETI-paper.
// Subscribe вызывают до Run. Run запускает consumer group и блокируется до отмены ctx или Close().
func NewKafka(cfg KafkaConfig) (*Kafka, error) {
	if len(cfg.Brokers) == 0 {
		return nil, fmt.Errorf("brokers are required")
	}
	if cfg.GroupID == "" {
		cfg.GroupID = "leti-paper-default"
	}

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}

	producer, err := sarama.NewSyncProducer(cfg.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("kafka sync producer: %w", err)
	}

	consumerGroup, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, config)
	if err != nil {
		_ = producer.Close()
		return nil, fmt.Errorf("kafka consumer group: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Kafka{
		producer:      producer,
		consumerGroup: consumerGroup,
		handlers:      make(map[string]func(context.Context, message.Message) error),
		consumerCtx:   ctx,
		consumerCancel: cancel,
	}, nil
}

// Publish реализует broker.Publisher. Сериализует message в JSON и отправляет в топик.
func (k *Kafka) Publish(ctx context.Context, topic string, msg message.Message) error {
	k.mu.RLock()
	defer k.mu.RUnlock()
	if k.closed {
		return fmt.Errorf("kafka client is closed")
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	_, _, err = k.producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(data),
	})
	if err != nil {
		return fmt.Errorf("kafka send: %w", err)
	}
	return nil
}

// Subscribe реализует broker.Subscriber. Регистрирует обработчик для топика.
// Перед получением сообщений нужно вызвать Run(ctx).
func (k *Kafka) Subscribe(ctx context.Context, topic string, handler func(context.Context, message.Message) error) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	if k.closed {
		return fmt.Errorf("kafka client is closed")
	}
	if _, exists := k.handlers[topic]; exists {
		return fmt.Errorf("already subscribed to topic: %s", topic)
	}
	k.handlers[topic] = handler
	return nil
}

// Close останавливает consumer и закрывает producer. Реализует broker.Subscriber.
func (k *Kafka) Close() error {
	k.mu.Lock()
	if k.closed {
		k.mu.Unlock()
		return nil
	}
	k.closed = true
	k.consumerCancel()
	k.mu.Unlock()

	k.consumerWg.Wait()
	_ = k.consumerGroup.Close()
	return k.producer.Close()
}

// Run запускает consumer group для всех зарегистрированных топиков. Блокируется до отмены ctx или Close().
// Вызывать после одного или нескольких Subscribe.
func (k *Kafka) Run(ctx context.Context) error {
	k.mu.RLock()
	topics := make([]string, 0, len(k.handlers))
	for t := range k.handlers {
		topics = append(topics, t)
	}
	k.mu.RUnlock()

	if len(topics) == 0 {
		return nil
	}

	consumer := &kafkaConsumer{handlers: k.handlers, mu: &k.mu}
	runCtx := ctx
	if runCtx == nil {
		runCtx = k.consumerCtx
	}

	k.consumerWg.Add(1)
	go func() {
		defer k.consumerWg.Done()
		for {
			select {
			case <-k.consumerCtx.Done():
				return
			case <-runCtx.Done():
				return
			default:
				if err := k.consumerGroup.Consume(runCtx, topics, consumer); err != nil {
					return
				}
			}
		}
	}()
	return nil
}

type kafkaConsumer struct {
	handlers map[string]func(context.Context, message.Message) error
	mu      *sync.RWMutex
}

func (c *kafkaConsumer) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (c *kafkaConsumer) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (c *kafkaConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
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
			session.MarkMessage(msg, "")
			continue
		}
		if err := handler(context.Background(), m); err != nil {
			// Не коммитим offset при ошибке — Kafka повторно доставит сообщение
			continue
		}
		session.MarkMessage(msg, "")
	}
	return nil
}
