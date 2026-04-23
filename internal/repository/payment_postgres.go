package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"payment/internal/appers"
	"payment/internal/model"
	"payment/pkg/db"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type PaymentRepositoryImpl struct {
	db *db.Database
}

func NewPaymentRepositoryImpl(db *db.Database) *PaymentRepositoryImpl {
	return &PaymentRepositoryImpl{db: db}
}

func (r *PaymentRepositoryImpl) GetAllPayments(ctx context.Context) ([]*model.Payment, error) {
	query := "SELECT * FROM payments"

	rows, err := r.db.QueryRows(ctx, query)
	if err != nil {
		return nil, appers.NewSqlExecutions(err.Error())
	}

	res := make([]*model.Payment, 0)
	for rows.Next() {
		var payment model.Payment
		err := rows.Scan(&payment.ID, &payment.PaymentId, &payment.Amount)
		if err != nil {
			return nil, appers.NewSqlExecutions(err.Error())
		}
		res = append(res, &payment)
	}
	return res, nil
}

func (r *PaymentRepositoryImpl) GetPaymentById(ctx context.Context, id int64) (*model.Payment, error) {
	query := "SELECT * FROM payments WHERE id = $1"
	row := r.db.QueryRow(ctx, query, id)
	return scanRow(row)
}

func (r *PaymentRepositoryImpl) PostPayment(ctx context.Context, dto *model.Payment) (*model.Payment, error) {
	query := "INSERT INTO payments (payment_id, amount) VALUES ($1, $2) RETURNING id, payment_id, amount"
	row := r.db.QueryRow(ctx, query, dto.PaymentId, dto.Amount)
	return scanRow(row)
}

func (r *PaymentRepositoryImpl) PutPayment(ctx context.Context, dto *model.Payment, id int64) (*model.Payment, error) {
	query := "UPDATE payments SET amount = $1 WHERE id = $2 RETURNING id, payment_id, amount"
	row := r.db.QueryRow(ctx, query, dto.Amount, id)
	return scanRow(row)
}

func (r *PaymentRepositoryImpl) DeletePayment(ctx context.Context, id int64) (*model.Payment, error) {
	query := "DELETE FROM payments WHERE id = $1 RETURNING id, payment_id, amount"
	row := r.db.QueryRow(ctx, query, id)
	return scanRow(row)
}

func scanRow(rows *sqlx.Row) (*model.Payment, error) {
	var payment model.Payment
	if err := rows.Scan(&payment.ID, &payment.PaymentId, &payment.Amount); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, appers.ErrEventFound.SetMsg(err.Error())
		}
		if pgErr, ok := errors.AsType[*pq.Error](err); ok {
			switch pgErr.Code {
			case "23505":
				return nil, appers.ErrDuplicatePayment.SetMsg(pgErr.Error())
			default:
				return nil, appers.NewSqlExecutions(fmt.Sprintf("column - %s; constraint - %s, err - %s",
					pgErr.Column, pgErr.Constraint, err.Error()))
			}
		}
		return nil, appers.NewSqlExecutions(err.Error())
	}

	return &payment, nil
}
