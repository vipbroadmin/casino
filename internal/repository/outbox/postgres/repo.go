package outboxpg

import (
	"context"
	"database/sql"

	"players_service/internal/infra/postgres"
	playeruc "players_service/internal/usecase/player"
)

// Repo uses Outbox pattern table "outbox".
// It supports running inside UnitOfWork transaction (sql.Tx stored in context).
type Repo struct {
	db *sql.DB
}

func New(db *sql.DB) *Repo { return &Repo{db: db} }

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

func (r *Repo) Enqueue(ctx context.Context, msg playeruc.OutboxMessage) error {
	ex := pickExecutor(ctx, r.db)

	const q = `
INSERT INTO outbox (
  id, aggregate, aggregate_id, type, key, payload, created_at, published_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,NULL)
`
	_, err := ex.ExecContext(ctx, q,
		msg.ID, msg.Aggregate, msg.AggregateID, msg.Type, msg.Key, msg.Payload, msg.CreatedAt,
	)
	return err
}
