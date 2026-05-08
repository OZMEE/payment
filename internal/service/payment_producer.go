package service

import (
	"context"
	"payment/internal/appers"
	"payment/internal/model"
	"payment/pkg/config"
	"strconv"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

type PaymentProducer interface {
	SendPaymentEvent(ctx context.Context, event *model.OutboxEvent) error
}

type PaymentProducerImpl struct {
	producer       *kgo.Client
	log            *zap.Logger
	topic          string
	contextTimeout int32
}

func NewPaymentProducerImpl(cfg config.KafkaConfig, producer *kgo.Client, log *zap.Logger) *PaymentProducerImpl {
	return &PaymentProducerImpl{
		producer:       producer,
		log:            log,
		topic:          cfg.PaymentTopic,
		contextTimeout: cfg.ContextTimeout,
	}
}

func (p *PaymentProducerImpl) SendPaymentEvent(ctx context.Context, event *model.OutboxEvent) error {
	key := strconv.FormatInt(event.ID, 10)

	record := &kgo.Record{
		Topic: p.topic,
		Key:   []byte(key),
		Value: []byte(event.Payload),
	}

	produceCtx, cancel := context.WithTimeout(ctx, time.Duration(p.contextTimeout)*time.Second)
	defer cancel()

	results := p.producer.ProduceSync(produceCtx, record)

	if err := results.FirstErr(); err != nil {
		p.log.Error("failed to send event", zap.Error(err))
		return appers.ErrKafkaSendEvent.Builder().Msg(err.Error()).Build()
	}

	return nil
}
