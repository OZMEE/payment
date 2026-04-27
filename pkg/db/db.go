package db

import (
	"context"
	"database/sql"
	"fmt"
	"payment/pkg/config"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"

	_ "github.com/lib/pq"
)

type DB interface {
	QueryRow(ctx context.Context, query string, args ...any) *sqlx.Row
	QueryRows(ctx context.Context, query string, args ...any) (*sqlx.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	Close() error
}

type Database struct {
	db *sqlx.DB
}

func (d *Database) QueryRow(ctx context.Context, query string, args ...any) *sqlx.Row {
	return d.db.QueryRowxContext(ctx, query, args...)
}

func (d *Database) QueryRows(ctx context.Context, query string, args ...any) (*sqlx.Rows, error) {
	return d.db.QueryxContext(ctx, query, args...)
}

func (d *Database) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return d.db.ExecContext(ctx, query, args...)
}

func (d *Database) Close() error {
	return d.db.Close()
}

func New(cfg config.DatabaseConfig) (*Database, error) {
	db, err := sqlx.Open(cfg.Driver, cfg.DNS())
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	if err := db.Ping(); err != nil {
		return nil, err
	}
	fmt.Println("successfully connected to db")

	err = goose.SetDialect(cfg.Driver)
	if err != nil {
		return nil, err
	}
	if err := goose.Up(db.DB, cfg.MigrationsPath); err != nil {
		return nil, err
	}

	return &Database{db: db}, nil
}
