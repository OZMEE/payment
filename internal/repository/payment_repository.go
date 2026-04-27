package repository

import (
	"context"
	"payment/internal/model"
)

type PaymentRepository interface {
	GetAllPayments(ctx context.Context) ([]*model.Payment, error)
	GetPaymentById(ctx context.Context, id int64) (*model.Payment, error)
	PostPayment(ctx context.Context, dto *model.Payment) (*model.Payment, error)
	PutPayment(ctx context.Context, payment *model.Payment, id int64) (*model.Payment, error)
	DeletePayment(ctx context.Context, id int64) (*model.Payment, error)
}
