-- +goose Up
CREATE TABLE IF NOT EXISTS player_documents (
  id UUID PRIMARY KEY,
  player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
  type TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'pending',
  file_url TEXT,
  metadata JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_player_documents_player_id ON player_documents(player_id);
CREATE INDEX IF NOT EXISTS idx_player_documents_status ON player_documents(status);

-- +goose Down
DROP INDEX IF EXISTS idx_player_documents_status;
DROP INDEX IF EXISTS idx_player_documents_player_id;
DROP TABLE IF EXISTS player_documents;
