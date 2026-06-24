package ingestion

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// ─── RSS / Atom feed types ───────────────────────────────────────────────────

type rssFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Items []rssItem `xml:"item"`
}

type rssItem struct {
	GUID        string       `xml:"guid"`
	Title       string       `xml:"title"`
	PubDate     string       `xml:"pubDate"`
	Link        string       `xml:"link"`
	Description string       `xml:"description"`
	Encoded     string       `xml:"http://purl.org/rss/1.0/modules/content/ encoded"`
	Enclosure   rssEncl      `xml:"enclosure"`
	Transcript  rssTranscript `xml:"https://podcastindex.org/namespace/1.0 transcript"`
}

type rssTranscript struct {
	URL  string `xml:"url,attr"`
	Type string `xml:"type,attr"`
}

type rssEncl struct {
	URL string `xml:"url,attr"`
}

type atomFeed struct {
	XMLName xml.Name    `xml:"http://www.w3.org/2005/Atom feed"`
	Entries []atomEntry `xml:"http://www.w3.org/2005/Atom entry"`
}

type atomEntry struct {
	VideoID   string   `xml:"http://www.youtube.com/xml/schemas/2015 videoId"`
	Title     string   `xml:"http://www.w3.org/2005/Atom title"`
	Published string   `xml:"http://www.w3.org/2005/Atom published"`
	Link      atomLink `xml:"http://www.w3.org/2005/Atom link"`
	Summary   string   `xml:"http://www.w3.org/2005/Atom summary"`
	Content   string   `xml:"http://www.w3.org/2005/Atom content"`
	MediaDesc string   `xml:"http://search.yahoo.com/mrss/ description"`
	// MediaTitle not needed
}

type atomLink struct {
	Href string `xml:"href,attr"`
}

// ─── Internal episode data ───────────────────────────────────────────────────

type episodeData struct {
	GUID          string
	Title         string
	PublishedAt   time.Time
	EpisodeURL    string
	AudioURL      string
	Content       string // show notes（sync 時存入 DB）
	VideoID       string // YouTube videoId（有的話用 yt-dlp 抓字幕）
	TranscriptURL string // podcast:transcript URL
}

// ─── Feed fetch + parse ──────────────────────────────────────────────────────

func fetchFeed(feedURL string) ([]episodeData, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "AlphaLens/1.0 (+https://alphalens.local)")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Try RSS first
	var rss rssFeed
	if err := xml.Unmarshal(body, &rss); err == nil && len(rss.Channel.Items) > 0 {
		return rssItems(rss.Channel.Items), nil
	}

	// Try Atom (YouTube)
	var atom atomFeed
	if err := xml.Unmarshal(body, &atom); err == nil && len(atom.Entries) > 0 {
		return atomEntries(atom.Entries), nil
	}

	return nil, fmt.Errorf("unrecognised feed format")
}

func rssItems(items []rssItem) []episodeData {
	eps := make([]episodeData, 0, len(items))
	for _, it := range items {
		content := it.Encoded
		if content == "" {
			content = it.Description
		}
		guid := it.GUID
		if guid == "" {
			guid = it.Link
		}
		pub, _ := time.Parse(time.RFC1123Z, it.PubDate)
		if pub.IsZero() {
			pub, _ = time.Parse(time.RFC1123, it.PubDate)
		}
		eps = append(eps, episodeData{
			GUID:          guid,
			Title:         it.Title,
			PublishedAt:   pub,
			EpisodeURL:    it.Link,
			AudioURL:      it.Enclosure.URL,
			Content:       stripHTML(content),
			TranscriptURL: it.Transcript.URL,
		})
	}
	return eps
}

func atomEntries(entries []atomEntry) []episodeData {
	eps := make([]episodeData, 0, len(entries))
	for _, e := range entries {
		content := e.MediaDesc
		if content == "" {
			content = e.Content
		}
		if content == "" {
			content = e.Summary
		}
		guid := e.VideoID
		if guid == "" {
			guid = e.Link.Href
		}
		pub, _ := time.Parse(time.RFC3339, e.Published)
		episodeURL := e.Link.Href
		if e.VideoID != "" && episodeURL == "" {
			episodeURL = "https://www.youtube.com/watch?v=" + e.VideoID
		}
		eps = append(eps, episodeData{
			GUID:        guid,
			Title:       e.Title,
			PublishedAt: pub,
			EpisodeURL:  episodeURL,
			Content:     stripHTML(content),
			VideoID:     e.VideoID,
		})
	}
	return eps
}

// stripHTML removes HTML tags and collapses whitespace.
func stripHTML(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
			b.WriteRune(' ')
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

// ─── Public pipeline functions ───────────────────────────────────────────────

// SyncPodcastFeed fetches the RSS/Atom feed and inserts new episodes.
// Returns count of newly added episodes.
func SyncPodcastFeed(db *sql.DB, dataDir string, podcastID int, rssURL string) (int, error) {
	episodes, err := fetchFeed(rssURL)
	if err != nil {
		return 0, fmt.Errorf("fetch feed %s: %w", rssURL, err)
	}

	// 首次 sync 只取最新 5 集；後續 sync 靠 ON CONFLICT DO NOTHING 去重
	var existing int
	db.QueryRow(`SELECT COUNT(*) FROM podcast_episodes WHERE podcast_id=$1`, podcastID).Scan(&existing)
	if existing == 0 && len(episodes) > 3 {
		episodes = episodes[:3]
		log.Printf("[podcast] first sync: limiting to 3 most recent episodes")
	}

	count := 0
	for _, ep := range episodes {
		var id int
		err := db.QueryRow(`
			INSERT INTO podcast_episodes
				(podcast_id, guid, title, published_at, episode_url, audio_url,
				 video_id, transcript_url, transcript_src)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,'show_notes')
			ON CONFLICT (guid) DO NOTHING
			RETURNING id`,
			podcastID, ep.GUID, ep.Title, ep.PublishedAt,
			ep.EpisodeURL, ep.AudioURL,
			ep.VideoID, ep.TranscriptURL,
		).Scan(&id)
		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			log.Printf("[podcast] insert episode %q: %v", ep.Title, err)
			continue
		}
		if ep.Content != "" {
			path, werr := WriteTranscript(dataDir, podcastID, id, ep.Content)
			if werr != nil {
				log.Printf("[podcast] write show notes for ep %d: %v", id, werr)
			} else {
				db.Exec(`UPDATE podcast_episodes SET transcript_path=$1 WHERE id=$2`, path, id)
			}
		}
		count++
	}

	db.Exec(`UPDATE podcasts SET last_synced_at=NOW() WHERE id=$1`, podcastID)
	log.Printf("[podcast] sync podcast %d: %d new episodes", podcastID, count)
	return count, nil
}

// AnalyzePendingEpisodes fetches real transcripts (yt captions / Whisper) then
// sends content to LLM to extract stock mentions. Returns total mention count.
func AnalyzePendingEpisodes(db *sql.DB, dataDir string, podcastID int, llm *LLMClient, whisperModel string) (int, error) {
	rows, err := db.Query(`
		SELECT id, title, video_id, transcript_url, audio_url, transcript_path, transcript, transcript_src
		FROM podcast_episodes
		WHERE podcast_id=$1 AND analyzed_at IS NULL
		ORDER BY published_at DESC LIMIT 50`, podcastID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	type ep struct {
		ID             int
		Title          string
		VideoID        string
		TranscriptURL  string
		AudioURL       string
		TranscriptPath string
		Transcript     string // legacy fallback
		TranscriptSrc  string
	}
	var episodes []ep
	for rows.Next() {
		var e ep
		if err := rows.Scan(&e.ID, &e.Title, &e.VideoID, &e.TranscriptURL,
			&e.AudioURL, &e.TranscriptPath, &e.Transcript, &e.TranscriptSrc); err != nil {
			return 0, err
		}
		episodes = append(episodes, e)
	}
	rows.Close()

	log.Printf("[podcast] analyzing %d episodes (provider=%s model=%s whisper=%q)",
		len(episodes), llm.Provider(), llm.Model(), whisperModel)

	total := 0
	for _, e := range episodes {
		// 讀 transcript：優先從檔案，fallback 舊 DB 欄位
		content := e.Transcript
		if e.TranscriptPath != "" {
			if text, err := ReadTranscript(dataDir, e.TranscriptPath); err != nil {
				log.Printf("[podcast] episode %d: read transcript file: %v", e.ID, err)
			} else {
				content = text
			}
		}

		// 嘗試取得真實逐字稿（show notes 準確率太低）
		if e.TranscriptSrc == "show_notes" {
			t0 := time.Now()
			ep := episodeData{
				VideoID:       e.VideoID,
				TranscriptURL: e.TranscriptURL,
				AudioURL:      e.AudioURL,
			}
			if real, src := fetchTranscript(ep, whisperModel); real != "" {
				content = real
				path, werr := WriteTranscript(dataDir, podcastID, e.ID, content)
				if werr != nil {
					log.Printf("[podcast] episode %d: write transcript file: %v", e.ID, werr)
					db.Exec(`UPDATE podcast_episodes SET transcript=$1, transcript_src=$2 WHERE id=$3`,
						content, src, e.ID)
				} else {
					db.Exec(`UPDATE podcast_episodes SET transcript_path=$1, transcript='', transcript_src=$2 WHERE id=$3`,
						path, src, e.ID)
				}
				log.Printf("[podcast] episode %d: transcript via %s (%d chars, %.1fs)",
					e.ID, src, len(content), time.Since(t0).Seconds())
			} else {
				log.Printf("[podcast] episode %d: no real transcript, using show notes (%d chars)", e.ID, len(content))
			}
		}

		if content == "" {
			log.Printf("[podcast] episode %d %q: no content, skip", e.ID, e.Title)
			db.Exec(`UPDATE podcast_episodes SET analyzed_at=NOW() WHERE id=$1`, e.ID)
			continue
		}

		log.Printf("[podcast] LLM episode %d %q (src=%s %d chars)",
			e.ID, e.Title, e.TranscriptSrc, len(content))
		t1 := time.Now()
		mentions, err := llm.ExtractMentions(content)
		if err != nil {
			log.Printf("[podcast] LLM episode %d error (%.1fs): %v", e.ID, time.Since(t1).Seconds(), err)
		} else {
			log.Printf("[podcast] LLM episode %d done: %d mentions (%.1fs)", e.ID, len(mentions), time.Since(t1).Seconds())
		}

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
				e.ID, m.Ticker, m.TickerRaw, m.Sentiment, m.Confidence, m.Thesis, m.OriginalQuote)
		}

		db.Exec(`UPDATE podcast_episodes SET analyzed_at=NOW() WHERE id=$1`, e.ID)
		total += len(mentions)
		log.Printf("[podcast] episode %d analyzed: %d mentions", e.ID, len(mentions))
	}
	return total, nil
}
