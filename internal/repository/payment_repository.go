package repository

import (
	"context"
	"database/sql"
	"fmt"
	"payment/internal/model"
)

type PaymentRepository interface {
	GetAllPayments(ctx context.Context) ([]*model.PaymentDb, error)
	GetPaymentById(ctx context.Context, id string) (*model.PaymentDb, error)
	PostPayment(ctx context.Context, dto *model.PaymentDto) (*model.PaymentDb, error)
	PutPayment(ctx context.Context, payment *model.PaymentDto, id string) (*model.PaymentDb, error)
	DeletePayment(ctx context.Context, id string) (*model.PaymentDb, error)
}

type PaymentRepositoryImpl struct {
	repoSql PaymentRepositorySql
}

func NewPaymentRepositoryImpl(repoSql PaymentRepositorySql) *PaymentRepositoryImpl {
	return &PaymentRepositoryImpl{repoSql: repoSql}
}

func (r *PaymentRepositoryImpl) GetAllPayments(ctx context.Context) ([]*model.PaymentDb, error) {
	rows, err := r.repoSql.selectAllPayments(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*model.PaymentDb, 0)
	for rows.Next() {
		var payment model.PaymentDb
		err := rows.Scan(&payment.ID, &payment.Amount)
		if err != nil {
			return nil, err
		}
		result = append(result, &payment)
	}

	return result, nil
}

func (r *PaymentRepositoryImpl) GetPaymentById(ctx context.Context, id string) (*model.PaymentDb, error) {
	row := r.repoSql.selectPaymentByIdSql(ctx, id)
	return scanRow(row)
}

func (r *PaymentRepositoryImpl) PostPayment(ctx context.Context, dto *model.PaymentDto) (*model.PaymentDb, error) {
	row := r.repoSql.insertIntoPaymentsSql(ctx, dto.Amount)
	return scanRow(row)
}

func (r *PaymentRepositoryImpl) PutPayment(ctx context.Context, dto *model.PaymentDto, id string) (*model.PaymentDb, error) {
	row := r.repoSql.updatePaymentsSql(ctx, dto.Amount, id)
	return scanRow(row)
}

func (r *PaymentRepositoryImpl) DeletePayment(ctx context.Context, id string) (*model.PaymentDb, error) {
	row := r.repoSql.deletePaymentsSql(ctx, id)
	return scanRow(row)
}

func scanRow(rows *sql.Row) (*model.PaymentDb, error) {
	var payment model.PaymentDb
	if err := rows.Scan(&payment.ID, &payment.Amount); err != nil {
		return nil, fmt.Errorf("scan row: %w", err)
	}
	return &payment, nil
}
