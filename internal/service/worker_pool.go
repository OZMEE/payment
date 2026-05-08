package service

import (
	"context"
	"errors"
	"math"
	"math/rand/v2"
	"payment/internal/appers"
	"payment/internal/repository"
	"payment/pkg/config"
	"time"

	"go.uber.org/zap"
)

type WorkerPool interface {
}

type WorkerPoolImpl struct {
	paymentProducer  PaymentProducer
	outboxRepository repository.OutboxRepository
	workersCount     int
	maxAttempts      int
	timeoutSec       int
	limitEvents      int
	log              *zap.Logger
}

func NewWorkerPoolImpl(cfg config.WorkerConfig, paymentProducer PaymentProducer, outboxRepository repository.OutboxRepository, log *zap.Logger) *WorkerPoolImpl {
	return &WorkerPoolImpl{
		paymentProducer:  paymentProducer,
		outboxRepository: outboxRepository,
		workersCount:     cfg.Count,
		maxAttempts:      cfg.MaxAttempts,
		timeoutSec:       cfg.TimeoutSec,
		limitEvents:      cfg.LimitEvents,
		log:              log,
	}
}

func (p *WorkerPoolImpl) Start(ctx context.Context) {
	for workerId := 0; workerId < p.workersCount; workerId++ {
		go func() {
			err := p.worker(ctx, workerId)
			if err != nil {
				p.log.Error("Worker fall", zap.Int("workerId", workerId), zap.Error(err))
			}
		}()
	}
}

func (p *WorkerPoolImpl) worker(ctx context.Context, workerId int) error {
	ticker := time.NewTicker(time.Duration(p.timeoutSec) * time.Second)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			outboxEvents, err := p.outboxRepository.GetPendingOutboxes(ctx, p.limitEvents)
			if err != nil {
				p.logError(err, workerId)
				return err
			}
			if len(outboxEvents) == 0 {
				p.log.Info("Not found events", zap.Int("worker_id", workerId))
				continue
			}
			p.log.Info("Found events", zap.Any("events", outboxEvents), zap.Int("id worker", workerId))

			for _, event := range outboxEvents {
				if err := p.paymentProducer.SendPaymentEvent(ctx, event); err != nil {
					if event.Attempts < p.maxAttempts {
						event.Attempts++
						event.NextRetryAt = calculateNextRetry(event.Attempts)
						event.Status = repository.PendingStatus
					} else {
						event.Status = repository.FailedStatus
					}
				} else {
					event.Status = repository.SuccessStatus
				}
			}

			err = p.outboxRepository.UpdateOutboxes(ctx, outboxEvents)
			if err != nil {
				p.logError(err, workerId)
				return err
			}
		}
	}
	return nil
}

func calculateNextRetry(attempt int) time.Time {
	baseDelay := 1 * time.Second
	maxDelay := 30 * time.Second

	// Экспоненциальный рост: 1s, 2s, 4s, 8s...
	delay := float64(baseDelay) * math.Pow(2, float64(attempt))

	if delay > float64(maxDelay) {
		delay = float64(maxDelay)
	}

	// Jitter: рандомизация +-20% от delay, чтобы воркеры не просыпались одновременно
	jitter := delay * 0.2 * (rand.Float64()*2 - 1)
	finalDelay := time.Duration(delay + jitter)

	return time.Now().Add(finalDelay)
}

func (p *WorkerPoolImpl) logError(err error, workerId int) {
	var errResp *appers.ErrorResp
	if !errors.As(err, &errResp) {
		p.log.Error(err.Error(),
			zap.Int("workerId", workerId))
		return
	}

	p.log.Error(errResp.Error(), zap.Reflect("error", errResp))
}
