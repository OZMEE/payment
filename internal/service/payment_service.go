package service

import (
	"context"
	"fmt"
	"payment/internal/model"
	"payment/internal/repository"
)

type PaymentService interface {
	GetAllPayments(ctx context.Context) ([]*model.PaymentDto, error)
	GetPaymentById(ctx context.Context, id string) (*model.PaymentDto, error)
	PostPayment(ctx context.Context, dto *model.PaymentDto) (string, error)
	PutPayment(ctx context.Context, dto *model.PaymentDto, id string) error
	DeletePayment(ctx context.Context, id string) error
}

type PaymentServiceImpl struct {
	repository repository.PaymentRepository
}

func NewPaymentServiceImpl(repository repository.PaymentRepository) *PaymentServiceImpl {
	return &PaymentServiceImpl{repository: repository}
}

func (s PaymentServiceImpl) GetAllPayments(ctx context.Context) ([]*model.PaymentDto, error) {
	listDb, err := s.repository.GetAllPayments(ctx)
	if err != nil {
		return nil, err
	}
	dtos := make([]*model.PaymentDto, 0, len(listDb))
	for _, db := range listDb {
		dtos = append(dtos, mapToDto(db))
	}
	return dtos, nil
}

func (s PaymentServiceImpl) GetPaymentById(ctx context.Context, id string) (*model.PaymentDto, error) {
	if db, ok := payments[id]; !ok {
		db, err := s.repository.GetPaymentById(ctx, id)
		if err != nil {
			return nil, err
		}
		cache(db)
		return mapToDto(db), nil
	} else {
		return mapToDto(&db), nil
	}
}

func (s PaymentServiceImpl) PostPayment(ctx context.Context, dto *model.PaymentDto) (string, error) {
	db, err := s.repository.PostPayment(ctx, dto)
	if err != nil {
		return "", fmt.Errorf("PostPayment: %w", err)
	}
	cache(db)
	return db.ID, nil
}

func (s PaymentServiceImpl) PutPayment(ctx context.Context, dto *model.PaymentDto, id string) error {
	db, err := s.repository.PutPayment(ctx, dto, id)
	if err != nil {
		return fmt.Errorf("PutPayment: %w", err)
	}
	cache(db)
	return nil
}

func (s PaymentServiceImpl) DeletePayment(ctx context.Context, id string) error {
	_, err := s.repository.DeletePayment(ctx, id)
	if err != nil {
		return fmt.Errorf("DeletePayment: %w", err)
	}
	deleteCache(id)
	return nil
}

var payments = map[string]model.PaymentDb{}

func cache(payment *model.PaymentDb) {
	payments[payment.ID] = *payment
}

func deleteCache(id string) {
	delete(payments, id)
}

func mapToDto(db *model.PaymentDb) *model.PaymentDto {
	return &model.PaymentDto{Amount: db.Amount}
}
