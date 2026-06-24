CREATE TABLE podcasts (
  id             SERIAL PRIMARY KEY,
  name           TEXT NOT NULL,
  rss_url        TEXT NOT NULL UNIQUE,
  description    TEXT NOT NULL DEFAULT '',
  language       TEXT NOT NULL DEFAULT 'zh',
  is_active      BOOLEAN NOT NULL DEFAULT true,
  last_synced_at TIMESTAMPTZ,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE podcast_episodes (
  id             SERIAL PRIMARY KEY,
  podcast_id     INT NOT NULL REFERENCES podcasts(id) ON DELETE CASCADE,
  guid           TEXT NOT NULL UNIQUE,
  title          TEXT NOT NULL,
  published_at   TIMESTAMPTZ,
  episode_url    TEXT NOT NULL DEFAULT '',
  audio_url      TEXT NOT NULL DEFAULT '',
  video_id       TEXT NOT NULL DEFAULT '',    -- YouTube video ID
  transcript_url TEXT NOT NULL DEFAULT '',    -- podcast:transcript URL
  transcript     TEXT NOT NULL DEFAULT '',    -- show notes（sync 時）or 真實逐字稿（analyze 時）
  transcript_src TEXT NOT NULL DEFAULT '',    -- 'show_notes' | 'yt_captions' | 'remote'
  analyzed_at    TIMESTAMPTZ,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ON podcast_episodes (podcast_id, published_at DESC);

CREATE TABLE podcast_mentions (
  id         SERIAL PRIMARY KEY,
  episode_id INT NOT NULL REFERENCES podcast_episodes(id) ON DELETE CASCADE,
  ticker     TEXT REFERENCES stocks(ticker),
  ticker_raw TEXT NOT NULL,
  sentiment  TEXT NOT NULL DEFAULT 'neutral' CHECK (sentiment IN ('bullish','bearish','neutral')),
  confidence NUMERIC(3,2) NOT NULL DEFAULT 0.5,
  thesis     TEXT NOT NULL DEFAULT '',
  adopt      BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ON podcast_mentions (ticker, created_at DESC);
CREATE INDEX ON podcast_mentions (episode_id);
