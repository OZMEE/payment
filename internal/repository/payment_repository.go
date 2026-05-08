package repository

import (
	"context"
	"payment/internal/model"

	"github.com/jmoiron/sqlx"
)

type PaymentRepository interface {
	GetAllPayments(ctx context.Context) ([]*model.Payment, error)
	GetPaymentById(ctx context.Context, id int64) (*model.Payment, error)
	PostPayment(ctx context.Context, dto *model.Payment, tx *sqlx.Tx) (*model.Payment, error)
	PutPayment(ctx context.Context, payment *model.Payment, id int64) (*model.Payment, error)
	DeletePayment(ctx context.Context, id int64) (*model.Payment, error)
	BeginTransaction(ctx context.Context) (*sqlx.Tx, error)
}
