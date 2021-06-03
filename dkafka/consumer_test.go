package dkafka

import (
	"fmt"
	"testing"
)

func TestConsumer(t *testing.T) {
	cfg := &Config{
		Brokers:  "192.168.49.208:9092",
		Group:    "t1",
		Version:  "2.1.1",
		Topics:   "test",
		Assignor: "range",
		Oldest:   true,
		Verbose:  false,
	}

	var fn = func(topic string, payload []byte) error {
		fmt.Printf("handler %s msg, payload:%s\n", topic, string(payload))
		return nil
	}
	c, err := NewConsumer(cfg)
	fmt.Printf("----%v\n", err)
	fmt.Printf("cli----%v\n", c.cli)
	if err != nil {
		t.Fatalf("err:%v\n", err)
		t.FailNow()
	}
	SetConsumerHandler(c, fn)
	ServeLoop(c, cfg)
}
