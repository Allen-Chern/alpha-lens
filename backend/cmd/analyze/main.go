// analyze: 從 DB 讀取指定 episode 的 transcript，執行 LLM 標的萃取並印出結果。
//
// 用法：
//
//	go run ./cmd/analyze <episode_id>
//
// 環境變數與正式服務相同（POSTGRES_*, LLM_PROVIDER, OLLAMA_MODEL, OLLAMA_BASE_URL, CLAUDE_API_KEY）。
package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/alpha-lens/backend/internal/db"
	"github.com/alpha-lens/backend/internal/ingestion"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: analyze <episode_id>\n")
		os.Exit(1)
	}
	episodeID, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid episode_id: %v\n", err)
		os.Exit(1)
	}

	conn, err := db.Connect(buildDatabaseURL())
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer conn.Close()

	var title, transcriptPath, transcriptLegacy, src string
	err = conn.QueryRow(
		`SELECT title, COALESCE(transcript_path,''), COALESCE(transcript,''), COALESCE(transcript_src,'')
		 FROM podcast_episodes WHERE id=$1`,
		episodeID,
	).Scan(&title, &transcriptPath, &transcriptLegacy, &src)
	if err != nil {
		log.Fatalf("episode %d not found: %v", episodeID, err)
	}

	dataDir := getenv("DATA_DIR", "/app/data")
	transcript := transcriptLegacy
	if transcriptPath != "" {
		text, rerr := ingestion.ReadTranscript(dataDir, transcriptPath)
		if rerr != nil {
			log.Fatalf("read transcript file: %v", rerr)
		}
		transcript = text
	}

	fmt.Printf("Episode %d: %q\n", episodeID, title)
	fmt.Printf("Transcript src: %s  length: %d chars\n\n", src, len(transcript))

	if transcript == "" {
		fmt.Println("transcript is empty — nothing to analyze")
		os.Exit(0)
	}

	llm := ingestion.NewLLMClient(
		getenv("LLM_PROVIDER", "ollama"),
		getenv("OLLAMA_BASE_URL", "http://localhost:11434"),
		getenv("OLLAMA_MODEL", "deepseek-r1:8b"),
		os.Getenv("CLAUDE_API_KEY"),
	)
	fmt.Printf("LLM: provider=%s model=%s\n\n", llm.Provider(), llm.Model())

	mentions, err := llm.ExtractMentions(transcript)
	if err != nil {
		log.Fatalf("LLM error: %v", err)
	}

	if len(mentions) == 0 {
		fmt.Println("no mentions found")
		return
	}
	fmt.Printf("Found %d mention(s):\n\n", len(mentions))
	for _, m := range mentions {
		ticker := m.Ticker
		if ticker == "" {
			ticker = "(unknown)"
		}
		fmt.Printf("  %-8s  %-10s  conf=%.2f  raw=%q\n    thesis: %s\n\n",
			ticker, m.Sentiment, m.Confidence, m.TickerRaw, m.Thesis)
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func buildDatabaseURL() string {
	user := getenv("POSTGRES_USER", "alphalens")
	password := getenv("POSTGRES_PASSWORD", "alphalens")
	host := getenv("POSTGRES_HOST", "localhost")
	port := getenv("POSTGRES_PORT", "5432")
	dbname := getenv("POSTGRES_DB", "alphalens")
	sslmode := getenv("POSTGRES_SSLMODE", "disable")
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, dbname, sslmode)
}
