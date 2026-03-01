package kafka

import (
	"fmt"
	"time"

	"github.com/IBM/sarama"
)

// Config настройки подключения к Kafka.
type Config struct {
	Brokers []string
	GroupID string

	// Producer
	AckPolicy int16
	RetryMax  int

	// Consumer
	AutoCommitEnable bool
	MaxWaitTime      time.Duration
}

// NewConsumerGroup создаёт sarama.ConsumerGroup по конфигурации
func NewConsumerGroup(cfg *Config) (sarama.ConsumerGroup, error) {
	if len(cfg.Brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers are required")
	}

	sc := sarama.NewConfig()
	sc.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{
		sarama.NewBalanceStrategySticky(),
		sarama.NewBalanceStrategyRoundRobin(),
		sarama.NewBalanceStrategyRange(),
	}
	sc.Consumer.Offsets.Initial = sarama.OffsetNewest
	sc.Consumer.Offsets.AutoCommit.Enable = cfg.AutoCommitEnable
	if cfg.MaxWaitTime > 0 {
		sc.Consumer.MaxWaitTime = cfg.MaxWaitTime
	} else {
		sc.Consumer.MaxWaitTime = 500 * time.Millisecond
	}

	groupID := cfg.GroupID
	if groupID == "" {
		groupID = "default-consumer-group"
	}

	cg, err := sarama.NewConsumerGroup(cfg.Brokers, groupID, sc)
	if err != nil {
		return nil, fmt.Errorf("creating kafka consumer group: %w", err)
	}
	return cg, nil
}

// NewSyncProducer создаёт sarama.SyncProducer по конфигурации.
func NewSyncProducer(cfg *Config) (sarama.SyncProducer, error) {
	if len(cfg.Brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers are required")
	}

	sc := sarama.NewConfig()
	sc.Producer.RequiredAcks = sarama.RequiredAcks(cfg.AckPolicy)
	sc.Producer.Return.Successes = true
	if cfg.RetryMax > 0 {
		sc.Producer.Retry.Max = cfg.RetryMax
	} else {
		sc.Producer.Retry.Max = 5
	}

	sp, err := sarama.NewSyncProducer(cfg.Brokers, sc)
	if err != nil {
		return nil, fmt.Errorf("creating kafka sync producer: %w", err)
	}
	return sp, nil
}
