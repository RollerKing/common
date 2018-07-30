package kafkautil

import (
	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type SimpleConsumer struct {
	brokers []string
	service string
	topics  []string
	canexit chan struct{}
}

type MessageHandle interface {
	Message(topic string, partitionKey string, data []byte) error
	Error(err error)
}

func NewConsumer(brokers []string, service string, topics ...string) *SimpleConsumer {
	return &SimpleConsumer{
		brokers: brokers,
		service: service,
		topics:  topics,
		canexit: make(chan struct{}, 1),
	}
}

func (sc *SimpleConsumer) PollMessage(handler MessageHandle) error {
	config := cluster.NewConfig()
	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = true
	config.Consumer.Offsets.CommitInterval = 1 * time.Second
	config.Consumer.Offsets.Initial = sarama.OffsetNewest //初始从最新的offset开始

	c, err := cluster.NewConsumer(sc.brokers, sc.service, sc.topics, config)
	if err != nil {
		return err
	}

	defer c.Close()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGABRT, syscall.SIGALRM, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)
	for {
		select {
		case msg, more := <-c.Messages():
			if more {
				for {
					if herr := handler.Message(msg.Topic, string(msg.Key), msg.Value); herr == nil {
						c.MarkOffset(msg, "")
						break
					}
					time.Sleep(1 * time.Second)
				}
			}
		case err, more := <-c.Errors():
			if more {
				handler.Error(err)
			}
		case <-c.Notifications():
		case <-signals:
			return nil
		}
	}
}
