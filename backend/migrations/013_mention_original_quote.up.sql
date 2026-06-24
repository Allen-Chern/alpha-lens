ALTER TABLE podcast_mentions
  ADD COLUMN IF NOT EXISTS original_quote TEXT NOT NULL DEFAULT '';
