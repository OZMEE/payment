package service

import (
	"context"
	"payment/internal/model"
	"payment/internal/repository"
)

type PaymentService interface {
	GetAllPayments(ctx context.Context) ([]*model.Payment, error)
	GetPaymentById(ctx context.Context, id int64) (*model.Payment, error)
	PostPayment(ctx context.Context, dto *model.PaymentDto) (*model.Payment, error)
	PutPayment(ctx context.Context, dto *model.PaymentDto, id int64) error
	DeletePayment(ctx context.Context, id int64) error
}

type PaymentServiceImpl struct {
	repository      repository.PaymentRepository
	paymentProducer PaymentProducer
}

func NewPaymentServiceImpl(repository repository.PaymentRepository, paymentProducer PaymentProducer) *PaymentServiceImpl {
	return &PaymentServiceImpl{repository: repository, paymentProducer: paymentProducer}
}

func (s PaymentServiceImpl) GetAllPayments(ctx context.Context) ([]*model.Payment, error) {
	payments, err := s.repository.GetAllPayments(ctx)
	if err != nil {
		return nil, err
	}
	return payments, nil
}

func (s PaymentServiceImpl) GetPaymentById(ctx context.Context, id int64) (*model.Payment, error) {
	db, err := s.repository.GetPaymentById(ctx, id)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (s PaymentServiceImpl) PostPayment(ctx context.Context, dto *model.PaymentDto) (*model.Payment, error) {
	db, err := s.repository.PostPayment(ctx, dtoToPayment(dto))
	if err != nil {
		return nil, err
	}

	if err := s.paymentProducer.SendPaymentEvent(ctx, db, "user-id"); err != nil {
		return nil, err
	}

	return db, nil
}

func (s PaymentServiceImpl) PutPayment(ctx context.Context, dto *model.PaymentDto, id int64) error {
	_, err := s.repository.PutPayment(ctx, dtoToPayment(dto), id)
	if err != nil {
		return err
	}
	return nil
}

func (s PaymentServiceImpl) DeletePayment(ctx context.Context, id int64) error {
	_, err := s.repository.DeletePayment(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func dtoToPayment(dto *model.PaymentDto) *model.Payment {
	return &model.Payment{Amount: dto.Amount, PaymentId: dto.PaymentId}
}
