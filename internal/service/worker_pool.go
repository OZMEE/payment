package service

import (
	"context"
	"errors"
	"math"
	"math/rand/v2"
	"payment/internal/appers"
	"payment/internal/repository"
	"payment/pkg/config"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type WorkerPool interface {
}

type WorkerPoolImpl struct {
	paymentProducer         PaymentProducer
	outboxRepository        repository.OutboxRepository
	transactionalRepository repository.TransactionalRepository
	workersCount            int
	maxAttempts             int
	timeoutSec              int
	limitEvents             int
	log                     *zap.Logger
}

func NewWorkerPoolImpl(cfg config.WorkerConfig, paymentProducer PaymentProducer,
	outboxRepository repository.OutboxRepository, transactionalRepository repository.TransactionalRepository,
	log *zap.Logger) *WorkerPoolImpl {
	return &WorkerPoolImpl{
		paymentProducer:         paymentProducer,
		outboxRepository:        outboxRepository,
		transactionalRepository: transactionalRepository,
		workersCount:            cfg.Count,
		maxAttempts:             cfg.MaxAttempts,
		timeoutSec:              cfg.TimeoutSec,
		limitEvents:             cfg.LimitEvents,
		log:                     log,
	}
}

func (p *WorkerPoolImpl) Start(ctx context.Context, wg *sync.WaitGroup) {
	for workerId := 0; workerId < p.workersCount; workerId++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := p.worker(ctx, workerId)
			if err != nil {
				p.log.Error("Worker fall", zap.Int("workerId", workerId), zap.Error(err))
			}
		}()
	}
}

func (p *WorkerPoolImpl) worker(ctx context.Context, workerId int) (errResponse error) {
	ticker := time.NewTicker(time.Duration(p.timeoutSec) * time.Second)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:

			_, err := p.transactionalRepository.RunInTx(func(tx *sqlx.Tx) (any, error) {
				outboxEvents, err := p.outboxRepository.GetProcessingOutboxes(ctx, tx, p.limitEvents)
				if err != nil {
					p.logError(err, workerId)
					return nil, err
				}
				if len(outboxEvents) == 0 {
					p.log.Info("Not found events", zap.Int("worker_id", workerId))
					return nil, nil
				}
				p.log.Info("Found events", zap.Any("events", outboxEvents), zap.Int("id worker", workerId))

				for _, event := range outboxEvents {
					if err := p.paymentProducer.SendPaymentEvent(ctx, event); err != nil {
						event.Attempts++
						if event.Attempts < p.maxAttempts {
							event.NextRetryAt = calculateNextRetry(event.Attempts)
						} else {
							event.Status = repository.FailedStatus
						}
					} else {
						event.Status = repository.SuccessStatus
					}
				}

				err = p.outboxRepository.UpdateOutboxes(ctx, tx, outboxEvents)
				if err != nil {
					p.logError(err, workerId)
					return nil, err
				}
				return outboxEvents, nil
			})
			if err != nil {
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
