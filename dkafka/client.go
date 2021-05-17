package dkafka

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/Shopify/sarama"
	"github.com/dup2X/gopkg/logger"
)

type Consumer struct {
	cli    sarama.ConsumerGroup
	ctx    context.Context
	cancel context.CancelFunc
	ready  chan bool
}

func NewConsumer(cfg *Config) (*Consumer, error) {
	config := sarama.NewConfig()
	ver, err := sarama.ParseKafkaVersion(cfg.Version)
	if err != nil {
		return nil, err
	}
	config.Version = ver

	switch cfg.Assignor {
	case "sticky":
		config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategySticky
	case "roundrobin":
		config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	case "range":
		config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	default:
		return nil, fmt.Errorf("invalid assignor %s", cfg.Assignor)
	}

	if cfg.Oldest {
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	}
	ctx, cancel := context.WithCancel(context.Background())
	client, err := sarama.NewConsumerGroup(strings.Split(cfg.Brokers, ","), cfg.Group, config)
	if err != nil {
		return nil, err
	}
	c := &Consumer{
		cli:    client,
		ctx:    ctx,
		cancel: cancel,
	}
	return c, nil
}

func ServeLoop(c *Consumer, cfg *Config) {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	ctx := c.ctx
	go func() {
		defer wg.Done()
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := c.cli.Consume(ctx, strings.Split(cfg.Topics, ","), c); err != nil {
				logger.Errorf(nil, logger.DLTagUndefined, "Error from consumer: %v", err)
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
			c.ready = make(chan bool)
		}
	}()

	<-c.ready // Await till the consumer has been set up
	logger.Infof(nil, logger.DLTagUndefined, "Sarama consumer up and running!...")

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-ctx.Done():
		logger.Infof(nil, logger.DLTagUndefined, "terminating: context cancelled...")
	case <-sigterm:
		logger.Infof(nil, logger.DLTagUndefined, "terminating: via signal!...")
	}
	c.cancel()
	wg.Wait()
	if err := c.cli.Close(); err != nil {
		logger.Errorf(nil, logger.DLTagUndefined, "Error closing client: %v", err)
	}
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(consumer.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/master/consumer_group.go#L27-L29
	for message := range claim.Messages() {
		logger.Infof(nil, logger.DLTagUndefined, "Message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic)
		session.MarkMessage(message, "")
	}

	return nil
}
