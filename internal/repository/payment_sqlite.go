package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"payment/internal/appers"
	"payment/internal/model"
	"time"

	_ "modernc.org/sqlite"
)

type PaymentRepositorySqlImpl struct {
	db *sql.DB
}

type PaymentRepository interface {
	GetAllPayments(ctx context.Context) ([]*model.Payment, error)
	GetPaymentById(ctx context.Context, id int64) (*model.Payment, error)
	PostPayment(ctx context.Context, dto *model.Payment) (*model.Payment, error)
	PutPayment(ctx context.Context, payment *model.Payment, id int64) (*model.Payment, error)
	DeletePayment(ctx context.Context, id int64) (*model.Payment, error)
}

func NewPaymentRepositorySqlImpl() *PaymentRepositorySqlImpl {
	dsn := "db.sqlite"

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(2 * time.Hour)

	if err := initTables(ctx, db); err != nil {
		log.Fatalf("init tables: %v", err)
	}
	fmt.Println("init tables successfully")
	return &PaymentRepositorySqlImpl{db: db}
}

func initTables(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS payments (
		    id INTEGER PRIMARY KEY AUTOINCREMENT,
		    amount INTEGER
		);
	`)
	if err != nil {
		return appers.NewSqlExecutions(err.Error())
	}
	return nil
}

func (r *PaymentRepositorySqlImpl) GetAllPayments(ctx context.Context) ([]*model.Payment, error) {
	query := "SELECT id, amount FROM payments;"
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, appers.NewSqlExecutions(err.Error())
	}

	res := make([]*model.Payment, 0)
	for rows.Next() {
		var payment model.Payment
		err := rows.Scan(&payment.ID, &payment.Amount)
		if err != nil {
			return nil, appers.NewSqlExecutions(err.Error())
		}
		res = append(res, &payment)
	}

	return res, nil
}

func (r *PaymentRepositorySqlImpl) GetPaymentById(ctx context.Context, id int64) (*model.Payment, error) {
	query := "SELECT id, amount FROM payments WHERE id = $1"
	row := r.db.QueryRowContext(ctx, query, id)
	return scanRow(row)
}

func (r *PaymentRepositorySqlImpl) PostPayment(ctx context.Context, dto *model.Payment) (*model.Payment, error) {
	query := "INSERT INTO payments (amount) VALUES (?) RETURNING id, amount"
	row := r.db.QueryRowContext(ctx, query, dto.Amount)
	return scanRow(row)
}

func (r *PaymentRepositorySqlImpl) PutPayment(ctx context.Context, dto *model.Payment, id int64) (*model.Payment, error) {
	query := "UPDATE payments SET amount = ? WHERE id = ? RETURNING id, amount"
	row := r.db.QueryRowContext(ctx, query, dto.Amount, id)
	return scanRow(row)
}

func (r *PaymentRepositorySqlImpl) DeletePayment(ctx context.Context, id int64) (*model.Payment, error) {
	query := "DELETE FROM payments WHERE id = ? RETURNING id, amount"
	row := r.db.QueryRowContext(ctx, query, id)
	return scanRow(row)
}

func scanRow(rows *sql.Row) (*model.Payment, error) {
	var payment model.Payment
	if err := rows.Scan(&payment.ID, &payment.Amount); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, appers.ErrEventFound
		}
		return nil, appers.NewSqlExecutions(err.Error())
	}

	return &payment, nil
}
