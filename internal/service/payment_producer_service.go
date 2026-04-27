package service

import (
	"context"
	"encoding/json"
	"payment/internal/appers"
	"payment/internal/model"
	"strconv"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

type PaymentProducer interface {
	SendPaymentEvent(ctx context.Context, payment *model.Payment, userId string) error
}

type PaymentEvent struct {
	UserId    string `json:"user_id"`
	PaymentId int64  `json:"payment_id"`
	Amount    int64  `json:"amount"`
}

type PaymentProducerImpl struct {
	producer *kgo.Client
	log      *zap.Logger
}

func NewPaymentProducerImpl(producer *kgo.Client, log *zap.Logger) *PaymentProducerImpl {
	return &PaymentProducerImpl{
		producer: producer,
		log:      log,
	}
}

func (p *PaymentProducerImpl) SendPaymentEvent(ctx context.Context, payment *model.Payment, userId string) error {
	event := PaymentEvent{
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
		Topic: "payment-events",
		Key:   []byte(key),
		Value: payload,
	}

	produceCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	done := make(chan error, 1)

	p.producer.Produce(produceCtx, record, func(r *kgo.Record, err error) {
		done <- err
	})

	select {
	case err := <-done:
		if err != nil {
			p.log.Error("failed to send event", zap.Error(err))
			return appers.ErrKafkaSendEvent.SetMsg(err.Error())
		}
		p.log.Info("event sent", zap.Int64("payment_id", payment.PaymentId))
		return nil
	case <-ctx.Done():
		p.log.Warn("request context canceled, but produce may still complete",
			zap.Int64("payment_id", payment.PaymentId))

		return nil
	case <-time.After(5 * time.Second):
		return appers.ErrKafkaProduceTimeout
	}
}
