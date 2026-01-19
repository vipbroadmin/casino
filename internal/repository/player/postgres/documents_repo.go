package playerpg

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"players_service/internal/domain/player"
)

type DocumentsRepo struct {
	db *sql.DB
}

func NewDocuments(db *sql.DB) *DocumentsRepo {
	return &DocumentsRepo{db: db}
}

func (r *DocumentsRepo) GetByPlayerID(ctx context.Context, playerID uuid.UUID) ([]*player.Document, error) {
	ex := pickExecutor(ctx, r.db)
	const q = `
SELECT id, player_id, type, status, file_url, metadata, created_at, updated_at
  FROM player_documents
 WHERE player_id = $1
 ORDER BY created_at DESC
`
	rows, err := ex.QueryContext(ctx, q, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []*player.Document
	for rows.Next() {
		var (
			d              player.Document
			fileURL        sql.NullString
			metadataRaw    []byte
			createdAt, updatedAt time.Time
		)
		if err := rows.Scan(
			&d.ID,
			&d.PlayerID,
			&d.Type,
			&d.Status,
			&fileURL,
			&metadataRaw,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, err
		}

		if fileURL.Valid {
			d.FileURL = fileURL.String
		}
		if len(metadataRaw) > 0 {
			if err := json.Unmarshal(metadataRaw, &d.Metadata); err != nil {
				return nil, err
			}
		}
		d.CreatedAt = createdAt
		d.UpdatedAt = updatedAt

		docs = append(docs, &d)
	}

	return docs, rows.Err()
}

func (r *DocumentsRepo) GetByID(ctx context.Context, id uuid.UUID) (*player.Document, error) {
	ex := pickExecutor(ctx, r.db)
	const q = `
SELECT id, player_id, type, status, file_url, metadata, created_at, updated_at
  FROM player_documents
 WHERE id = $1
`
	var (
		d              player.Document
		fileURL        sql.NullString
		metadataRaw    []byte
		createdAt, updatedAt time.Time
	)
	err := ex.QueryRowContext(ctx, q, id).Scan(
		&d.ID,
		&d.PlayerID,
		&d.Type,
		&d.Status,
		&fileURL,
		&metadataRaw,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, player.ErrNotFound
		}
		return nil, err
	}

	if fileURL.Valid {
		d.FileURL = fileURL.String
	}
	if len(metadataRaw) > 0 {
		if err := json.Unmarshal(metadataRaw, &d.Metadata); err != nil {
			return nil, err
		}
	}
	d.CreatedAt = createdAt
	d.UpdatedAt = updatedAt

	return &d, nil
}

func (r *DocumentsRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status player.DocumentStatus, now time.Time) error {
	ex := pickExecutor(ctx, r.db)
	const q = `UPDATE player_documents SET status=$2, updated_at=$3 WHERE id=$1`
	_, err := ex.ExecContext(ctx, q, id, status.String(), now)
	return err
}
