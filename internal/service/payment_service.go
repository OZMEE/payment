package service

import (
	"context"
	"payment/internal/appers"
	"payment/internal/model"
	"payment/internal/repository/postgres"
	"payment/internal/repository/redis"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type PaymentService interface {
	GetAllPayments(ctx context.Context) ([]*model.Payment, error)
	GetPaymentById(ctx context.Context, id int64) (*model.Payment, error)
	PostPayment(ctx context.Context, dto *model.PaymentDto) (*model.Payment, error)
	PutPayment(ctx context.Context, dto *model.PaymentDto, id int64) error
	DeletePayment(ctx context.Context, id int64) error
}

type PaymentServiceImpl struct {
	paymentRepository       postgres.PaymentRepository
	outboxRepository        postgres.OutboxRepository
	cacheRepository         redis.CacheRepository
	transactionalRepository postgres.TransactionalRepository
	log                     *zap.Logger
}

func NewPaymentServiceImpl(repository postgres.PaymentRepository, outboxRepository postgres.OutboxRepository,
	cacheRepository redis.CacheRepository, transactionalRepository postgres.TransactionalRepository,
	log *zap.Logger) *PaymentServiceImpl {
	return &PaymentServiceImpl{
		paymentRepository:       repository,
		outboxRepository:        outboxRepository,
		cacheRepository:         cacheRepository,
		transactionalRepository: transactionalRepository,
		log:                     log}
}

func (s PaymentServiceImpl) GetAllPayments(ctx context.Context) ([]*model.Payment, error) {
	payments, err := s.paymentRepository.GetAllPayments(ctx)
	if err != nil {
		return nil, err
	}
	return payments, nil
}

func (s PaymentServiceImpl) GetPaymentById(ctx context.Context, id int64) (*model.Payment, error) {
	var payment *model.Payment

	if cached, err := s.cacheRepository.GetPayment(ctx, id); err == nil && cached != nil {
		payment = cached
	} else {
		payment, err = s.paymentRepository.GetPaymentById(ctx, id)
		if err != nil {
			return nil, err
		}
	}

	return payment, nil
}

func (s PaymentServiceImpl) PostPayment(ctx context.Context, dto *model.PaymentDto) (response *model.Payment, errResponse error) {
	const op = "PaymentServiceImpl.PostPayment"

	entity, err := s.transactionalRepository.RunInTx(func(tx *sqlx.Tx) (any, error) {

		payment, err := s.paymentRepository.PostPayment(ctx, tx, dtoToPayment(dto))
		if err != nil {
			return nil, err
		}

		if err = s.outboxRepository.CreateOutbox(ctx, tx, payment); err != nil {
			return nil, err
		}

		return payment, nil
	})
	if err != nil {
		return nil, err
	}

	if payment, ok := entity.(*model.Payment); !ok {
		return nil, appers.ErrTypeAssertion.Builder().Op(op).Build()
	} else {
		if err := s.cacheRepository.SetPayment(ctx, payment); err != nil {
			//Если не получилось захешировать новый payment, ошибку не кидаем, т.к. транзакция закомичена и запись достанется просто не из redis, а из postgres
			s.log.Warn("Error caching payment", zap.Error(err))
		}
		return payment, nil
	}
}

func (s PaymentServiceImpl) PutPayment(ctx context.Context, dto *model.PaymentDto, id int64) error {
	const op = "PaymentServiceImpl.PutPayment"

	_, err := s.transactionalRepository.RunInTx(func(tx *sqlx.Tx) (any, error) {
		payment, err := s.paymentRepository.PutPayment(ctx, tx, dtoToPayment(dto), id)
		if err != nil {
			return nil, err
		}

		err = s.cacheRepository.SetPayment(ctx, payment)
		if err != nil {
			//Если не получилось обновить существующую запись кидаем ошибку и делаем роллбек, чтобы не было расхождение у postgres и redis
			return nil, err
		}
		return payment, nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s PaymentServiceImpl) DeletePayment(ctx context.Context, id int64) error {
	const op = "PaymentServiceImpl.DeletePayment"

	_, err := s.transactionalRepository.RunInTx(func(tx *sqlx.Tx) (any, error) {
		payment, err := s.paymentRepository.DeletePayment(ctx, tx, id)
		if err != nil {
			return nil, err
		}

		err = s.cacheRepository.DelPayment(ctx, id)
		if err != nil {
			//Если не получилось удалить запись и в бд и в redis - откат, чтобы нельзя было из redis достать несуществующую запись
			return nil, err
		}

		return payment, nil
	})
	if err != nil {
		return err
	}

	return nil
}

func dtoToPayment(dto *model.PaymentDto) *model.Payment {
	return &model.Payment{Amount: dto.Amount, PaymentId: dto.PaymentId}
}
