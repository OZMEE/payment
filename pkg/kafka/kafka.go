package kafka

import (
	"fmt"
	"payment/pkg/config"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

func NewProducer(cfg config.KafkaConfig) (*kgo.Client, error) {
	var acks kgo.Acks
	switch cfg.Acks {
	case 0:
		acks = kgo.NoAck()
	case 1:
		acks = kgo.LeaderAck()
	case -1:
		acks = kgo.AllISRAcks()
	}
	opts := []kgo.Opt{
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.DefaultProduceTopic(cfg.Topic),
		kgo.RequiredAcks(acks),
		kgo.ProducerBatchCompression(kgo.SnappyCompression()),
		kgo.ProducerLinger(time.Duration(cfg.LingerMs) * time.Millisecond),
		kgo.ProducerBatchMaxBytes(cfg.BatchSize),

		kgo.RecordRetries(3),
		kgo.RecordDeliveryTimeout(30 * time.Second),
		kgo.MaxBufferedRecords(10_000),
		kgo.MaxBufferedBytes(100 * 1024 * 1024),
	}
	fmt.Println(opts)

	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	return client, nil
}
