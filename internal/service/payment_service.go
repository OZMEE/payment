package service

import (
	"context"
	"payment/internal/model"
	"payment/internal/repository"

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
	paymentRepository repository.PaymentRepository
	outboxRepository  repository.OutboxRepository
	log               *zap.Logger
}

func NewPaymentServiceImpl(repository repository.PaymentRepository, outboxRepository repository.OutboxRepository, log *zap.Logger) *PaymentServiceImpl {
	return &PaymentServiceImpl{paymentRepository: repository, outboxRepository: outboxRepository, log: log}
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
	tx, err := s.paymentRepository.BeginTransaction()
	if err != nil {
		return nil, err
	}

	defer func() {
		if r := recover(); r != nil {
			if err := tx.Rollback(); err != nil {
				s.log.Error("Failed to rollback transaction after panic",
					zap.Any("recovered", r),
					zap.Error(err))
			} else {
				s.log.Error("Transaction rolled back after panic",
					zap.Any("recovered", r))
			}
			panic(r)
		}

		if errResponse != nil {
			if err := tx.Rollback(); err != nil {
				s.log.Error("Failed to rollback transaction", zap.Error(err))
				return
			}
			s.log.Error("Rollback transaction", zap.Error(errResponse))
		} else {
			if errResponse = tx.Commit(); err != nil {
				s.log.Error("Failed to commit transaction", zap.Any("payment", response), zap.Error(errResponse))
				return
			}
			s.log.Info("Successfully commit transaction", zap.Any("payment", response))
		}
	}()

	payment, err := s.paymentRepository.PostPayment(ctx, tx, dtoToPayment(dto))
	if err != nil {
		return nil, err
	}

	if err = s.outboxRepository.CreateOutbox(ctx, tx, payment); err != nil {
		return nil, err
	}

	return payment, nil
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
