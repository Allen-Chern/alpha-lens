ALTER TABLE podcast_episodes
  DROP COLUMN IF EXISTS video_id,
  DROP COLUMN IF EXISTS transcript_url,
  DROP COLUMN IF EXISTS transcript_src;
