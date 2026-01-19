package playerpg

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"players_service/internal/domain/player"
)

type RequisitesRepo struct {
	db *sql.DB
}

func NewRequisites(db *sql.DB) *RequisitesRepo {
	return &RequisitesRepo{db: db}
}

func (r *RequisitesRepo) GetByPlayerID(ctx context.Context, playerID uuid.UUID) (*player.Requisites, error) {
	ex := pickExecutor(ctx, r.db)
	const q = `
SELECT id, player_id, payment_method_id, form_data, created_at, updated_at
  FROM player_requisites
 WHERE player_id = $1
`
	var (
		req           player.Requisites
		formDataRaw   []byte
		createdAt, updatedAt time.Time
	)
	err := ex.QueryRowContext(ctx, q, playerID).Scan(
		&req.ID,
		&req.PlayerID,
		&req.PaymentMethodID,
		&formDataRaw,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, player.ErrNotFound
		}
		return nil, err
	}

	if len(formDataRaw) > 0 {
		if err := json.Unmarshal(formDataRaw, &req.FormData); err != nil {
			return nil, err
		}
	}
	req.CreatedAt = createdAt
	req.UpdatedAt = updatedAt

	return &req, nil
}

func (r *RequisitesRepo) Upsert(ctx context.Context, req *player.Requisites, now time.Time) error {
	ex := pickExecutor(ctx, r.db)
	formDataJSON, err := json.Marshal(req.FormData)
	if err != nil {
		return err
	}

	const q = `
INSERT INTO player_requisites (id, player_id, payment_method_id, form_data, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (player_id, payment_method_id)
DO UPDATE SET form_data = $4, updated_at = $6
`
	_, err = ex.ExecContext(ctx, q,
		req.ID,
		req.PlayerID,
		req.PaymentMethodID,
		formDataJSON,
		now,
		now,
	)
	return err
}
