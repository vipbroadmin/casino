package outboxpg

import (
	"context"
	"database/sql"
	"time"

	"players_service/internal/infra/postgres"
	playeruc "players_service/internal/usecase/player"

	"github.com/google/uuid"
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

func (r *Repo) FetchUnpublishedByType(ctx context.Context, typ string, limit int) ([]playeruc.OutboxMessage, error) {
	if limit <= 0 {
		limit = 100
	}

	const q = `
SELECT id, aggregate, aggregate_id, type, key, payload, created_at
FROM outbox
WHERE published_at IS NULL AND type = $1
ORDER BY created_at
LIMIT $2
`
	rows, err := r.db.QueryContext(ctx, q, typ, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []playeruc.OutboxMessage
	for rows.Next() {
		var (
			id          uuid.UUID
			aggregate   string
			aggregateID uuid.UUID
			msgType     string
			key         string
			payload     []byte
			createdAt   time.Time
		)
		if err := rows.Scan(&id, &aggregate, &aggregateID, &msgType, &key, &payload, &createdAt); err != nil {
			return nil, err
		}
		res = append(res, playeruc.OutboxMessage{
			ID:          id,
			Aggregate:   aggregate,
			AggregateID: aggregateID,
			Type:        msgType,
			Key:         key,
			Payload:     payload,
			CreatedAt:   createdAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (r *Repo) MarkPublished(ctx context.Context, id uuid.UUID, at time.Time) error {
	const q = `UPDATE outbox SET published_at = $2 WHERE id = $1 AND published_at IS NULL`
	_, err := r.db.ExecContext(ctx, q, id, at)
	return err
}
