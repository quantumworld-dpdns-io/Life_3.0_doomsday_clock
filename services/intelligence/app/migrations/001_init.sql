CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS raw_articles (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  url          TEXT UNIQUE NOT NULL,
  title        TEXT,
  body         TEXT,
  published_at TIMESTAMPTZ,
  source       TEXT,
  ingested_at  TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS scenario_signals (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  article_id UUID REFERENCES raw_articles(id) ON DELETE CASCADE,
  scenario   SMALLINT NOT NULL CHECK (scenario BETWEEN 1 AND 12),
  confidence REAL NOT NULL CHECK (confidence BETWEEN 0 AND 1),
  reasoning  TEXT,
  created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_signals_created ON scenario_signals (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_signals_scenario ON scenario_signals (scenario);
