package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/alpha-lens/backend/internal/ingestion"
	ws "github.com/alpha-lens/backend/internal/ws"
	"github.com/go-chi/chi/v5"
)

type episodeResponse struct {
	ID            int        `json:"id"`
	Title         string     `json:"title"`
	PublishedAt   *time.Time `json:"published_at"`
	EpisodeURL    string     `json:"episode_url"`
	TranscriptSrc string     `json:"transcript_src"`
	HasTranscript bool       `json:"has_transcript"`
	AnalyzedAt    *time.Time `json:"analyzed_at"`
	MentionCount  int        `json:"mention_count"`
}

func (h *Handler) ListEpisodes(w http.ResponseWriter, r *http.Request) {
	podcastID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	rows, err := h.DB.Query(`
		SELECT e.id, e.title, e.published_at, e.episode_url,
		       e.transcript_src, e.transcript_path, e.transcript,
		       e.analyzed_at, COUNT(m.id) AS mention_count
		FROM podcast_episodes e
		LEFT JOIN podcast_mentions m ON m.episode_id = e.id
		WHERE e.podcast_id = $1
		GROUP BY e.id
		ORDER BY e.published_at DESC
		LIMIT 100`, podcastID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	var episodes []episodeResponse
	for rows.Next() {
		var ep episodeResponse
		var transcriptPath, transcriptLegacy string
		if err := rows.Scan(
			&ep.ID, &ep.Title, &ep.PublishedAt, &ep.EpisodeURL,
			&ep.TranscriptSrc, &transcriptPath, &transcriptLegacy,
			&ep.AnalyzedAt, &ep.MentionCount,
		); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		ep.HasTranscript = transcriptPath != "" || transcriptLegacy != ""
		episodes = append(episodes, ep)
	}
	if episodes == nil {
		episodes = []episodeResponse{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"episodes": episodes})
}

func (h *Handler) FetchTranscript(w http.ResponseWriter, r *http.Request) {
	epID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var podcastID int
	if err := h.DB.QueryRow(`SELECT podcast_id FROM podcast_episodes WHERE id=$1`, epID).
		Scan(&podcastID); err != nil {
		writeError(w, http.StatusNotFound, "episode not found")
		return
	}

	go func() {
		progress := func(e ingestion.ProgressEvent) {
			h.Hub.Broadcast(ws.Event{
				Type:      "episode_status",
				EpisodeID: epID,
				PodcastID: podcastID,
				Status:    e.Status,
				Src:       e.Src,
				Chars:     e.Chars,
				Message:   e.Message,
			})
		}
		if err := ingestion.FetchEpisodeTranscript(
			h.DB, h.Settings.DataDir, epID, h.Settings.WhisperModel, progress,
		); err != nil {
			h.Hub.Broadcast(ws.Event{
				Type:      "episode_status",
				EpisodeID: epID,
				PodcastID: podcastID,
				Status:    "error",
				Message:   err.Error(),
			})
		}
	}()

	writeJSON(w, http.StatusAccepted, map[string]any{"ok": true})
}

func (h *Handler) AnalyzeEpisode(w http.ResponseWriter, r *http.Request) {
	epID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var podcastID int
	if err := h.DB.QueryRow(`SELECT podcast_id FROM podcast_episodes WHERE id=$1`, epID).
		Scan(&podcastID); err != nil {
		writeError(w, http.StatusNotFound, "episode not found")
		return
	}

	llm := ingestion.NewLLMClient(
		h.Settings.LLMProvider,
		h.Settings.OllamaBaseURL,
		h.Settings.OllamaModel,
		h.Settings.ClaudeAPIKey,
	)

	go func() {
		progress := func(e ingestion.ProgressEvent) {
			h.Hub.Broadcast(ws.Event{
				Type:         "episode_status",
				EpisodeID:    epID,
				PodcastID:    podcastID,
				Status:       e.Status,
				Src:          e.Src,
				MentionCount: e.MentionCount,
				Message:      e.Message,
			})
		}
		if err := ingestion.AnalyzeSingleEpisode(
			h.DB, h.Settings.DataDir, epID, llm, progress,
		); err != nil {
			h.Hub.Broadcast(ws.Event{
				Type:      "episode_status",
				EpisodeID: epID,
				PodcastID: podcastID,
				Status:    "error",
				Message:   err.Error(),
			})
		}
	}()

	writeJSON(w, http.StatusAccepted, map[string]any{"ok": true})
}
