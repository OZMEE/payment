package service

import (
	"context"
	"encoding/json"
	"payment/internal/appers"
	"payment/internal/model"
	"payment/pkg/config"
	"strconv"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

type PaymentProducer interface {
	SendPaymentEvent(ctx context.Context, payment *model.Payment, userId string) error
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

func (p *PaymentProducerImpl) SendPaymentEvent(ctx context.Context, payment *model.Payment, userId string) error {
	event := model.PaymentEvent{
		UserId:    userId,
		PaymentId: payment.PaymentId,
		Amount:    payment.Amount,
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	key := strconv.FormatInt(payment.ID, 10)

	record := &kgo.Record{
		Topic: p.topic,
		Key:   []byte(key),
		Value: payload,
	}

	produceCtx, cancel := context.WithTimeout(ctx, time.Duration(p.contextTimeout)*time.Second)
	defer cancel()

	results := p.producer.ProduceSync(produceCtx, record)

	if err := results.FirstErr(); err != nil {
		p.log.Error("failed to send event", zap.Error(err))
		return appers.ErrKafkaSendEvent.SetMsg(err.Error())
	}

	return nil
}
