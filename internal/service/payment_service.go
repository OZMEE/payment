package service

import (
	"context"
	"payment/internal/appers"
	"payment/internal/model"
	"payment/internal/repository"

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
	paymentRepository       repository.PaymentRepository
	outboxRepository        repository.OutboxRepository
	transactionalRepository repository.TransactionalRepository
	log                     *zap.Logger
}

func NewPaymentServiceImpl(repository repository.PaymentRepository, outboxRepository repository.OutboxRepository,
	transactionalRepository repository.TransactionalRepository, log *zap.Logger) *PaymentServiceImpl {
	return &PaymentServiceImpl{
		paymentRepository:       repository,
		outboxRepository:        outboxRepository,
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
	payment, err := s.paymentRepository.GetPaymentById(ctx, id)
	if err != nil {
		return nil, err
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
		return payment, nil
	}
}

func (s PaymentServiceImpl) PutPayment(ctx context.Context, dto *model.PaymentDto, id int64) error {
	_, err := s.paymentRepository.PutPayment(ctx, dtoToPayment(dto), id)
	if err != nil {
		return err
	}
	return nil
}

func (s PaymentServiceImpl) DeletePayment(ctx context.Context, id int64) error {
	_, err := s.paymentRepository.DeletePayment(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func dtoToPayment(dto *model.PaymentDto) *model.Payment {
	return &model.Payment{Amount: dto.Amount, PaymentId: dto.PaymentId}
}
