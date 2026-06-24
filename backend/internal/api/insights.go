package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/alpha-lens/backend/internal/ingestion"
	"github.com/go-chi/chi/v5"
)

// ─── Podcast CRUD ─────────────────────────────────────────────────────────────

func (h *Handler) ListPodcasts(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(`
		SELECT id, name, rss_url, description, language, is_active, last_synced_at, created_at
		FROM podcasts ORDER BY created_at DESC`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	type podcast struct {
		ID          int        `json:"id"`
		Name        string     `json:"name"`
		RSSURL      string     `json:"rss_url"`
		Description string     `json:"description"`
		Language    string     `json:"language"`
		IsActive    bool       `json:"is_active"`
		LastSynced  *time.Time `json:"last_synced_at"`
		CreatedAt   time.Time  `json:"created_at"`
	}

	var podcasts []podcast
	for rows.Next() {
		var p podcast
		var ls sql.NullTime
		if err := rows.Scan(&p.ID, &p.Name, &p.RSSURL, &p.Description,
			&p.Language, &p.IsActive, &ls, &p.CreatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if ls.Valid {
			p.LastSynced = &ls.Time
		}
		podcasts = append(podcasts, p)
	}
	if podcasts == nil {
		podcasts = []podcast{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"podcasts": podcasts})
}

func (h *Handler) AddPodcast(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		RSSURL      string `json:"rss_url"`
		Description string `json:"description"`
		Language    string `json:"language"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.RSSURL == "" {
		writeError(w, http.StatusBadRequest, "name and rss_url are required")
		return
	}
	if req.Language == "" {
		req.Language = "zh"
	}

	// 自動把 YouTube 頻道網址轉成 Atom feed URL
	feedURL, err := ingestion.ResolveYouTubeURL(req.RSSURL)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var id int
	err = h.DB.QueryRow(`
		INSERT INTO podcasts (name, rss_url, description, language)
		VALUES ($1,$2,$3,$4) RETURNING id`,
		req.Name, feedURL, req.Description, req.Language).Scan(&id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": id})
}

func (h *Handler) DeletePodcast(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	h.DB.Exec(`DELETE FROM podcasts WHERE id=$1`, id)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// ─── Sync + analyze ──────────────────────────────────────────────────────────

func (h *Handler) SyncPodcast(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var rssURL string
	if err := h.DB.QueryRow(`SELECT rss_url FROM podcasts WHERE id=$1`, id).Scan(&rssURL); err != nil {
		writeError(w, http.StatusNotFound, "podcast not found")
		return
	}

	// 1. 同步 feed（快，只抓 RSS XML）
	newEps, err := ingestion.SyncPodcastFeed(h.DB, h.Settings.DataDir, id, rssURL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 2. LLM 分析在背景跑，不阻塞 HTTP 回應
	llm := ingestion.NewLLMClient(
		h.Settings.LLMProvider,
		h.Settings.OllamaBaseURL,
		h.Settings.OllamaModel,
		h.Settings.ClaudeAPIKey,
	)
	go ingestion.AnalyzePendingEpisodes(h.DB, h.Settings.DataDir, id, llm, h.Settings.WhisperModel)

	writeJSON(w, http.StatusAccepted, map[string]any{
		"new_episodes": newEps,
		"analyzing":    true,
	})
}

// ─── Mentions ────────────────────────────────────────────────────────────────

func (h *Handler) ListMentions(w http.ResponseWriter, r *http.Request) {
	ticker := r.URL.Query().Get("ticker")
	sentiment := r.URL.Query().Get("sentiment")
	adoptOnly := r.URL.Query().Get("adopt") == "true"

	query := `
		SELECT pm.id, pm.episode_id, pe.title, pe.episode_url, pe.published_at,
		       p.id, p.name, pm.ticker, pm.ticker_raw, pm.sentiment,
		       pm.confidence, pm.thesis, pm.original_quote, pm.adopt, pm.created_at
		FROM podcast_mentions pm
		JOIN podcast_episodes pe ON pe.id = pm.episode_id
		JOIN podcasts p ON p.id = pe.podcast_id
		WHERE 1=1`

	args := []any{}
	n := 1
	if ticker != "" {
		query += ` AND pm.ticker = $` + strconv.Itoa(n)
		args = append(args, ticker)
		n++
	}
	if sentiment != "" {
		query += ` AND pm.sentiment = $` + strconv.Itoa(n)
		args = append(args, sentiment)
		n++
	}
	if adoptOnly {
		query += ` AND pm.adopt = true`
	}
	query += ` ORDER BY pm.created_at DESC LIMIT 100`

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	type mention struct {
		ID            int        `json:"id"`
		EpisodeID     int        `json:"episode_id"`
		EpisodeTitle  string     `json:"episode_title"`
		EpisodeURL    string     `json:"episode_url"`
		PublishedAt   *time.Time `json:"published_at"`
		PodcastID     int        `json:"podcast_id"`
		PodcastName   string     `json:"podcast_name"`
		Ticker        *string    `json:"ticker"`
		TickerRaw     string     `json:"ticker_raw"`
		Sentiment     string     `json:"sentiment"`
		Confidence    float64    `json:"confidence"`
		Thesis        string     `json:"thesis"`
		OriginalQuote string     `json:"original_quote"`
		Adopt         bool       `json:"adopt"`
		CreatedAt     time.Time  `json:"created_at"`
	}

	var mentions []mention
	for rows.Next() {
		var m mention
		var pub sql.NullTime
		if err := rows.Scan(
			&m.ID, &m.EpisodeID, &m.EpisodeTitle, &m.EpisodeURL, &pub,
			&m.PodcastID, &m.PodcastName, &m.Ticker, &m.TickerRaw,
			&m.Sentiment, &m.Confidence, &m.Thesis, &m.OriginalQuote, &m.Adopt, &m.CreatedAt,
		); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if pub.Valid {
			m.PublishedAt = &pub.Time
		}
		mentions = append(mentions, m)
	}
	if mentions == nil {
		mentions = []mention{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"mentions": mentions})
}

func (h *Handler) ToggleAdopt(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var adopt bool
	err = h.DB.QueryRow(
		`UPDATE podcast_mentions SET adopt = NOT adopt WHERE id=$1 RETURNING adopt`, id,
	).Scan(&adopt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"adopt": adopt})
}
