-- players
CREATE TABLE IF NOT EXISTS players (
  id              UUID PRIMARY KEY,
  email           TEXT NOT NULL UNIQUE,
  phone           TEXT NULL,

  status          SMALLINT NOT NULL,
  status_reason   TEXT NULL,

  country_code    TEXT NULL,
  locale          TEXT NULL,
  time_zone       TEXT NULL,

  first_name      TEXT NULL,
  last_name       TEXT NULL,
  birth_date      DATE NULL,
  gender          SMALLINT NOT NULL DEFAULT 0,

  registration_ip INET NULL,
  registered_at   TIMESTAMPTZ NULL,
  last_login_at   TIMESTAMPTZ NULL,

  metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,

  version         BIGINT NOT NULL,
  created_at      TIMESTAMPTZ NOT NULL,
  updated_at      TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_players_status ON players(status);
CREATE INDEX IF NOT EXISTS idx_players_registered_at ON players(registered_at);

-- events (audit)
CREATE TABLE IF NOT EXISTS player_status_events (
  id          UUID PRIMARY KEY,
  player_id   UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
  from_status SMALLINT NOT NULL,
  to_status   SMALLINT NOT NULL,
  reason      TEXT NOT NULL,
  actor_type  SMALLINT NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_pse_player_id_created_at ON player_status_events(player_id, created_at DESC);

-- outbox (for Kafka / async publish)
CREATE TABLE IF NOT EXISTS outbox (
  id           UUID PRIMARY KEY,
  aggregate    TEXT NOT NULL,
  aggregate_id UUID NOT NULL,
  type         TEXT NOT NULL,
  key          TEXT NOT NULL,
  payload      JSONB NOT NULL,
  created_at   TIMESTAMPTZ NOT NULL,
  published_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_outbox_published_at ON outbox(published_at);
