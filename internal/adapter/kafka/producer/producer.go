package producer

import (
	"context"
	"fmt"
	"sync"

	"github.com/IBM/sarama"

	"github.com/SosisterRapStar/LETI-PaperTestMicroservices/internal/domain/ports"
)

// Producer реализует ports.BrokerProducer поверх sarama.SyncProducer.
type Producer struct {
	sp     sarama.SyncProducer
	mu     sync.RWMutex
	closed bool
}

// New создаёт Producer. Принимает уже готовый sarama.SyncProducer
// (создаётся через infrastructure/kafka.NewSyncProducer).
func New(sp sarama.SyncProducer) (*Producer, error) {
	if sp == nil {
		return nil, fmt.Errorf("sync producer is required")
	}
	return &Producer{sp: sp}, nil
}

// SendMessage реализует ports.BrokerProducer.
func (p *Producer) SendMessage(ctx context.Context, topic string, key, value []byte) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.closed {
		return fmt.Errorf("kafka producer is closed")
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(value),
	}
	if len(key) > 0 {
		msg.Key = sarama.ByteEncoder(key)
	}

	_, _, err := p.sp.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}
	return nil
}

// Close реализует ports.BrokerProducer.
func (p *Producer) Close() error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.closed = true
	p.mu.Unlock()
	return p.sp.Close()
}

var _ ports.BrokerProducer = (*Producer)(nil)
