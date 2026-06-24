package ingestion

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// ProgressEvent carries state from a background job to the caller for broadcasting.
type ProgressEvent struct {
	Status       string
	Src          string
	Message      string
	Chars        int
	MentionCount int
}

type ProgressFunc func(ProgressEvent)

// FetchEpisodeTranscript fetches a real transcript for a single episode (yt-dlp / Whisper / remote)
// and persists it to a file. Existing transcript is overwritten.
func FetchEpisodeTranscript(db *sql.DB, dataDir string, episodeID int, whisperModel string, progress ProgressFunc) error {
	var podcastID int
	var ep episodeData
	err := db.QueryRow(`
		SELECT podcast_id, video_id, transcript_url, audio_url
		FROM podcast_episodes WHERE id=$1`, episodeID,
	).Scan(&podcastID, &ep.VideoID, &ep.TranscriptURL, &ep.AudioURL)
	if err != nil {
		return fmt.Errorf("episode %d not found: %w", episodeID, err)
	}

	progress(ProgressEvent{Status: "fetching_transcript"})

	text, src := fetchTranscript(ep, whisperModel)
	if text == "" {
		return fmt.Errorf("no transcript available (tried yt-dlp, remote url, whisper)")
	}

	path, err := WriteTranscript(dataDir, podcastID, episodeID, text)
	if err != nil {
		return fmt.Errorf("write transcript: %w", err)
	}

	db.Exec(`UPDATE podcast_episodes SET transcript_path=$1, transcript='', transcript_src=$2 WHERE id=$3`,
		path, src, episodeID)

	progress(ProgressEvent{Status: "transcript_done", Src: src, Chars: len(text)})
	return nil
}

// AnalyzeSingleEpisode runs LLM mention extraction on a single episode.
// Existing mentions are cleared before inserting new ones.
func AnalyzeSingleEpisode(db *sql.DB, dataDir string, episodeID int, llm *LLMClient, progress ProgressFunc) error {
	var transcriptPath, transcript, transcriptSrc string
	err := db.QueryRow(`
		SELECT transcript_path, transcript, transcript_src
		FROM podcast_episodes WHERE id=$1`, episodeID,
	).Scan(&transcriptPath, &transcript, &transcriptSrc)
	if err != nil {
		return fmt.Errorf("episode %d not found: %w", episodeID, err)
	}

	content := transcript
	if transcriptPath != "" {
		if text, rerr := ReadTranscript(dataDir, transcriptPath); rerr != nil {
			log.Printf("[episode] read transcript file ep %d: %v", episodeID, rerr)
		} else {
			content = text
		}
	}

	if content == "" {
		return fmt.Errorf("no transcript to analyze")
	}

	progress(ProgressEvent{Status: "analyzing", Src: transcriptSrc})

	t0 := time.Now()
	mentions, err := llm.ExtractMentions(content)
	if err != nil {
		return fmt.Errorf("LLM: %w", err)
	}
	log.Printf("[episode] ep %d: %d mentions (%.1fs)", episodeID, len(mentions), time.Since(t0).Seconds())

	db.Exec(`DELETE FROM podcast_mentions WHERE episode_id=$1`, episodeID)
	for _, m := range mentions {
		if m.Sentiment == "" {
			m.Sentiment = "neutral"
		}
		if m.Confidence <= 0 {
			m.Confidence = 0.5
		}
		db.Exec(`
			INSERT INTO podcast_mentions (episode_id, ticker, ticker_raw, sentiment, confidence, thesis, original_quote)
			VALUES ($1, NULLIF($2,''), $3, $4, $5, $6, $7)`,
			episodeID, m.Ticker, m.TickerRaw, m.Sentiment, m.Confidence, m.Thesis, m.OriginalQuote)
	}
	db.Exec(`UPDATE podcast_episodes SET analyzed_at=NOW() WHERE id=$1`, episodeID)

	progress(ProgressEvent{Status: "analyzed", MentionCount: len(mentions)})
	return nil
}
