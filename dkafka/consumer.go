package dkafka

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/Shopify/sarama"
	"github.com/dup2X/gopkg/logger"
)

type Consumer struct {
	cli     sarama.ConsumerGroup
	ctx     context.Context
	cancel  context.CancelFunc
	ready   chan bool
	handler func(topic string, payload []byte) error
	lag     int64
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
	logger.Debugf(nil, logger.DLTagUndefined, "host:%v", strings.Split(cfg.Brokers, ","))
	//client, err := sarama.NewConsumerGroup(strings.Split(cfg.Brokers, ","), cfg.Group, config)
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

func SetConsumerHandler(c *Consumer, proc func(topic string, payload []byte) error) {
	c.handler = proc
}

func ServeLoop(c *Consumer, cfg *Config) {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	ctx := c.ctx
	go func() {
		defer wg.Done()
		for {
			if err := c.cli.Consume(ctx, strings.Split(cfg.Topics, ","), c); err != nil {
				logger.Errorf(nil, logger.DLTagUndefined, "Error from consumer: %v", err)
			}
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
	//close(consumer.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		atomic.AddInt64(&consumer.lag, 1)
		err := consumer.handler(message.Topic, message.Value)
		if err == nil {
			logger.Infof(nil, logger.DLTagUndefined, "Message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic)
		} else {
			logger.Warnf(nil, logger.DLTagUndefined, "handle msg failed,err:%v", err)
		}
		atomic.AddInt64(&consumer.lag, -1)
		session.MarkMessage(message, "")
	}

	return nil
}
