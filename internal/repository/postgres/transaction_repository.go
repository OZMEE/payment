package postgres

import (
	"payment/pkg/db"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type TransactionalRepository interface {
	RunInTx(f func(tx *sqlx.Tx) (any, error)) (response any, errResponse error)
}

type TransactionalRepositoryImpl struct {
	db  *db.Database
	log *zap.Logger
}

func NewTransactionalRepositoryImpl(db *db.Database, log *zap.Logger) *TransactionalRepositoryImpl {
	return &TransactionalRepositoryImpl{db: db, log: log}
}

func (t *TransactionalRepositoryImpl) RunInTx(f func(tx *sqlx.Tx) (any, error)) (response any, errResponse error) {
	tx, err := t.db.BeginTransaction()
	if err != nil {
		return nil, err
	}

	defer func() {
		if r := recover(); r != nil {
			if err := tx.Rollback(); err != nil {
				t.log.Error("Failed to rollback transaction after panic",
					zap.Any("recovered", r),
					zap.Error(err))
			} else {
				t.log.Error("Transaction rolled back after panic",
					zap.Any("recovered", r))
			}
			panic(r)
		}

		if errResponse != nil {
			if err := tx.Rollback(); err != nil {
				t.log.Error("Failed to rollback transaction", zap.Error(err))
				return
			}
			t.log.Error("Rollback transaction", zap.Error(errResponse))
		} else {
			if errResponse = tx.Commit(); err != nil {
				t.log.Error("Failed to commit transaction", zap.Any("response", response), zap.Error(errResponse))
				return
			}
			if response != nil {
				t.log.Info("Successfully commit transaction", zap.Any("response", response))
			}
		}
	}()

	return f(tx)
}
