ALTER TABLE podcast_episodes
  ADD COLUMN IF NOT EXISTS transcript_path TEXT NOT NULL DEFAULT '';
