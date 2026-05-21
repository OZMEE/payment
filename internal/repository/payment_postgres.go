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

type PaymentRepository interface {
	GetAllPayments(ctx context.Context) ([]*model.Payment, error)
	GetPaymentById(ctx context.Context, id int64) (*model.Payment, error)
	PostPayment(ctx context.Context, tx *sqlx.Tx, dto *model.Payment) (*model.Payment, error)
	PutPayment(ctx context.Context, payment *model.Payment, id int64) (*model.Payment, error)
	DeletePayment(ctx context.Context, id int64) (*model.Payment, error)
}

type PaymentRepositoryImpl struct {
	db *db.Database
}

func NewPaymentRepositoryImpl(db *db.Database) *PaymentRepositoryImpl {
	return &PaymentRepositoryImpl{db: db}
}

func (r *PaymentRepositoryImpl) GetAllPayments(ctx context.Context) ([]*model.Payment, error) {
	const op = "PaymentRepositoryImpl.GetAllPayments"

	query := "SELECT id, payment_id, amount FROM payments"

	rows, err := r.db.QueryRows(ctx, query)
	if err != nil {
		return nil, appers.ErrSqlExecutions.Builder().Msg(err.Error()).Op(op).Build()
	}

	res := make([]*model.Payment, 0)
	for rows.Next() {
		var payment model.Payment
		err := rows.Scan(&payment.ID, &payment.PaymentId, &payment.Amount)
		if err != nil {
			return nil, appers.ErrSqlExecutions.Builder().Msg(err.Error()).Op(op).Build()
		}
		res = append(res, &payment)
	}
	return res, nil
}

func (r *PaymentRepositoryImpl) GetPaymentById(ctx context.Context, id int64) (*model.Payment, error) {
	const op = "PaymentRepositoryImpl.GetPaymentById"
	query := "SELECT id, payment_id, amount FROM payments WHERE id = $1"
	row := r.db.QueryRow(ctx, query, id)
	return scanPaymentRow(row, op)
}

func (r *PaymentRepositoryImpl) PostPayment(ctx context.Context, tx *sqlx.Tx, dto *model.Payment) (*model.Payment, error) {
	const op = "PaymentRepositoryImpl.PostPayment"
	query := "INSERT INTO payments (payment_id, amount) VALUES ($1, $2) RETURNING id, payment_id, amount"
	row := tx.QueryRowxContext(ctx, query, dto.PaymentId, dto.Amount)
	return scanPaymentRow(row, op)
}

func (r *PaymentRepositoryImpl) PutPayment(ctx context.Context, dto *model.Payment, id int64) (*model.Payment, error) {
	const op = "PaymentRepositoryImpl.PutPayment"
	query := "UPDATE payments SET amount = $1 WHERE id = $2 RETURNING id, payment_id, amount"
	row := r.db.QueryRow(ctx, query, dto.Amount, id)
	return scanPaymentRow(row, op)
}

func (r *PaymentRepositoryImpl) DeletePayment(ctx context.Context, id int64) (*model.Payment, error) {
	const op = "PaymentRepositoryImpl.DeletePayment"
	query := "DELETE FROM payments WHERE id = $1 RETURNING id, payment_id, amount"
	row := r.db.QueryRow(ctx, query, id)
	return scanPaymentRow(row, op)
}

func scanPaymentRow(rows *sqlx.Row, op string) (*model.Payment, error) {
	var payment model.Payment
	if err := rows.Scan(&payment.ID, &payment.PaymentId, &payment.Amount); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, appers.ErrEventFound.Builder().Msg(err.Error()).Build()
		}
		if pgErr, ok := errors.AsType[*pq.Error](err); ok {
			switch pgErr.Code {
			case "23505":
				return nil, appers.ErrDuplicatePayment.Builder().Msg(pgErr.Error()).Build()
			default:
				return nil, appers.ErrSqlExecutions.Builder().Msg(fmt.Sprintf("column - %s; constraint - %s, err - %s",
					pgErr.Column, pgErr.Constraint, err.Error())).Op(op).Build()
			}
		}
		return nil, appers.ErrSqlExecutions.Builder().Msg(err.Error()).Op(op).Build()
	}

	return &payment, nil
}
