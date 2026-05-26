package consumer

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/IBM/sarama"

	"github.com/zxCroshka/wb-test/internal/aggregator"
	"github.com/zxCroshka/wb-test/internal/dto"
)

type ConsumerConfig struct {
	Brokers []string
	Topic   string
	GroupID string
}

type Consumer struct {
	config     ConsumerConfig
	consumer   sarama.ConsumerGroup
	aggregator *aggregator.Aggregator
	logger     *slog.Logger
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	ready      chan bool
}

func NewConsumer(
	config ConsumerConfig,
	agg *aggregator.Aggregator,
	logger *slog.Logger,
) (*Consumer, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V2_6_0_0
	saramaConfig.Consumer.Return.Errors = true
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	saramaConfig.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{
		sarama.NewBalanceStrategyRoundRobin(),
	}
	saramaConfig.Consumer.Group.Session.Timeout = 10 * time.Second
	saramaConfig.Consumer.Group.Heartbeat.Interval = 3 * time.Second

	consumer, err := sarama.NewConsumerGroup(config.Brokers, config.GroupID, saramaConfig)
	if err != nil {
		return nil, err
	}

	logger.Info("kafka consumer group created",
		slog.String("topic", config.Topic),
		slog.String("group_id", config.GroupID),
	)

	return &Consumer{
		config:     config,
		consumer:   consumer,
		aggregator: agg,
		logger:     logger.WithGroup("kafka-consumer"),
		ready:      make(chan bool),
	}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	handler := &consumerGroupHandler{
		aggregator: c.aggregator,
		logger:     c.logger,
		ready:      c.ready,
	}

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if err := c.consumer.Consume(ctx, []string{c.config.Topic}, handler); err != nil {
					c.logger.Error("consumer error", slog.String("error", err.Error()))
					time.Sleep(time.Second)
				}
			}
		}
	}()

	<-c.ready
	c.logger.Info("consumer group ready",
		slog.String("topic", c.config.Topic),
		slog.String("group_id", c.config.GroupID),
	)

	return nil
}

func (c *Consumer) Stop() {
	c.logger.Info("stopping kafka consumer")
	if c.cancel != nil {
		c.cancel()
	}
	c.wg.Wait()
	if err := c.consumer.Close(); err != nil {
		c.logger.Error("error closing consumer", slog.String("error", err.Error()))
	}
	c.logger.Info("kafka consumer stopped")
}

type consumerGroupHandler struct {
	aggregator *aggregator.Aggregator
	logger     *slog.Logger
	ready      chan bool
}

func (h *consumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	h.logger.Info("consumer group setup")
	select {
	case <-h.ready:
	default:	
		close(h.ready)
	}
	return nil
}

func (h *consumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	h.logger.Info("consumer group cleanup")
	return nil
}

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		ctx := session.Context()
		h.handleMessage(ctx, msg)
		session.MarkMessage(msg, "")
	}
	return nil
}

func (h *consumerGroupHandler) handleMessage(ctx context.Context, msg *sarama.ConsumerMessage) {
	h.logger.Info(">>> RAW MESSAGE RECEIVED", slog.String("raw", string(msg.Value)))

	var event dto.SearchEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		h.logger.Error("failed to unmarshal",
			slog.String("error", err.Error()),
			slog.String("raw", string(msg.Value)),
		)
		return
	}

	if err := event.Validate(); err != nil {
		h.logger.Warn("invalid event",
			slog.String("error", err.Error()),
			slog.String("query", event.Query),
			slog.String("user_id", event.UserID),
		)
		return
	}

	h.aggregator.ProcessEvent(ctx, event.Query, event.UserID)

	h.logger.Debug("event processed",
		slog.String("query", event.Query),
		slog.String("user_id", event.UserID),
		slog.Int64("partition", int64(msg.Partition)),
		slog.Int64("offset", msg.Offset),
	)
}
