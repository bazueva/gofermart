package db

import (
	"context"
	"database/sql"

	"github.com/bazueva/gofermart/internal/interfaces"
)

type SQLDBWrapper struct {
	*sql.DB
}

func NewSQLDBWrapper(db *sql.DB) interfaces.DB {
	return &SQLDBWrapper{db}
}

func (w *SQLDBWrapper) BeginTx(ctx context.Context, opts *sql.TxOptions) (interfaces.Tx, error) {
	tx, err := w.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return tx, nil
}
