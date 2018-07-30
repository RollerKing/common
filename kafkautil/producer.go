package kafkautil

import (
	"github.com/Shopify/sarama"
)

type SimpleProducer struct {
	sarama.SyncProducer
}

func NewProducer(brokers []string) (*SimpleProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Retry.Max = 2
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Flush.MaxMessages = 1
	config.Producer.Return.Successes = true
	p, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}
	return &SimpleProducer{SyncProducer: p}, nil
}

func (sp *SimpleProducer) Send(topic, partitionKey string, data []byte) error {
	pm := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(partitionKey),
		Value: sarama.ByteEncoder(data),
	}
	_, _, err := sp.SendMessage(pm)
	return err
}
