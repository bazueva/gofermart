package db

import (
	"context"

	"github.com/bazueva/gofermart/internal/interfaces"
)

type txContextKey struct{}

func WithTx(ctx context.Context, tx interfaces.Tx) context.Context {
	return context.WithValue(ctx, txContextKey{}, tx)
}

func TxFromContext(ctx context.Context) (interfaces.Tx, bool) {
	tx, ok := ctx.Value(txContextKey{}).(interfaces.Tx)

	return tx, ok
}
