package kafka

import (
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
		kgo.DefaultProduceTopic(cfg.CommitTopic),
		kgo.RequiredAcks(acks),
		kgo.ProducerBatchCompression(kgo.SnappyCompression()),
		kgo.ProducerLinger(time.Duration(cfg.LingerMs) * time.Millisecond),
		kgo.ProducerBatchMaxBytes(cfg.BatchSize),

		kgo.RecordRetries(int(cfg.RecordRetries)),
		kgo.RecordDeliveryTimeout(time.Duration(cfg.RecordDeliveryTimeout) * time.Second),
		kgo.MaxBufferedRecords(int(cfg.MaxBufferedBytes)),
		kgo.MaxBufferedBytes(int(cfg.MaxBufferedRecords)),
		kgo.DialTimeout(time.Duration(cfg.DialTimeout) * time.Second), // Таймаут на установление соединения
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	return client, nil
}
