package playerpg

import (
	"context"
	"database/sql"

	"players_service/internal/infra/postgres"
)

// executor is implemented by *sql.DB and *sql.Tx
type executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func pickExecutor(ctx context.Context, db *sql.DB) executor {
	if tx, ok := postgres.TxFromContext(ctx); ok {
		return tx
	}
	return db
}
