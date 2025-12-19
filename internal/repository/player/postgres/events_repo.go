package playerpg

import (
	"context"
	"database/sql"

	"players_service/internal/domain/player"
)

type EventsRepo struct {
	db *sql.DB
}

func NewEvents(db *sql.DB) *EventsRepo { return &EventsRepo{db: db} }

func (r *EventsRepo) Append(ctx context.Context, ev player.PlayerStatusEvent) error {
	ex := pickExecutor(ctx, r.db)

	const q = `
INSERT INTO player_status_events (
  id, player_id, from_status, to_status, reason, actor_type, created_at
) VALUES ($1,$2,$3,$4,$5,$6,$7)
`
	_, err := ex.ExecContext(ctx, q,
		ev.ID, ev.PlayerID, int16(ev.From), int16(ev.To), ev.Reason, int16(ev.ActorType), ev.CreatedAt,
	)
	return err
}
