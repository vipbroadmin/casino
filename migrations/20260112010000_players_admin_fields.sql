-- +goose Up
ALTER TABLE players
  ADD COLUMN IF NOT EXISTS login TEXT,
  ADD COLUMN IF NOT EXISTS password_hash TEXT,
  ADD COLUMN IF NOT EXISTS currency TEXT,
  ADD COLUMN IF NOT EXISTS is_banned BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS level INTEGER NOT NULL DEFAULT 1;

CREATE INDEX IF NOT EXISTS idx_players_login ON players(login);
CREATE INDEX IF NOT EXISTS idx_players_country_code ON players(country_code);
CREATE INDEX IF NOT EXISTS idx_players_currency ON players(currency);

-- +goose Down
DROP INDEX IF EXISTS idx_players_currency;
DROP INDEX IF EXISTS idx_players_country_code;
DROP INDEX IF EXISTS idx_players_login;

ALTER TABLE players
  DROP COLUMN IF EXISTS level,
  DROP COLUMN IF EXISTS is_banned,
  DROP COLUMN IF EXISTS currency,
  DROP COLUMN IF EXISTS password_hash,
  DROP COLUMN IF EXISTS login;

