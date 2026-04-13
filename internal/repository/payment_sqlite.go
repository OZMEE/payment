package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

type PaymentRepositorySql interface {
	selectAllPayments(ctx context.Context) (*sql.Rows, error)
	selectPaymentByIdSql(ctx context.Context, id string) *sql.Row
	insertIntoPaymentsSql(ctx context.Context, amount int) *sql.Row
	updatePaymentsSql(ctx context.Context, amount int, id string) *sql.Row
	deletePaymentsSql(ctx context.Context, id string) *sql.Row
}

type PaymentRepositorySqlImpl struct {
	db *sql.DB
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
		return fmt.Errorf("create payments table: %v", err)
	}
	return nil
}

func (r *PaymentRepositorySqlImpl) selectAllPayments(ctx context.Context) (*sql.Rows, error) {
	query := "SELECT id, amount FROM payments;"
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *PaymentRepositorySqlImpl) selectPaymentByIdSql(ctx context.Context, id string) *sql.Row {
	query := "SELECT id, amount FROM payments WHERE id = $1;"
	row := r.db.QueryRowContext(ctx, query, id)
	return row
}

func (r *PaymentRepositorySqlImpl) insertIntoPaymentsSql(ctx context.Context, amount int) *sql.Row {
	query := "INSERT INTO payments (amount) VALUES (?) RETURNING id, amount"
	row := r.db.QueryRowContext(ctx, query, amount)
	return row
}

func (r *PaymentRepositorySqlImpl) updatePaymentsSql(ctx context.Context, amount int, id string) *sql.Row {
	query := "UPDATE payments SET amount = ? WHERE id = ? RETURNING id, amount"
	row := r.db.QueryRowContext(ctx, query, amount, id)
	return row
}

func (r *PaymentRepositorySqlImpl) deletePaymentsSql(ctx context.Context, id string) *sql.Row {
	query := "DELETE FROM payments WHERE id = ? RETURNING id, amount"
	row := r.db.QueryRowContext(ctx, query, id)
	return row
}
