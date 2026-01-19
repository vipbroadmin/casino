-- +goose Up
CREATE TABLE IF NOT EXISTS player_requisites (
  id UUID PRIMARY KEY,
  player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
  payment_method_id UUID NOT NULL,
  form_data JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(player_id, payment_method_id)
);

CREATE INDEX IF NOT EXISTS idx_player_requisites_player_id ON player_requisites(player_id);
CREATE INDEX IF NOT EXISTS idx_player_requisites_payment_method_id ON player_requisites(payment_method_id);

-- +goose Down
DROP INDEX IF EXISTS idx_player_requisites_payment_method_id;
DROP INDEX IF EXISTS idx_player_requisites_player_id;
DROP TABLE IF EXISTS player_requisites;
