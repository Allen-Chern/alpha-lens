package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/alpha-lens/backend/internal/api"
	"github.com/alpha-lens/backend/internal/db"
	"github.com/alpha-lens/backend/internal/models"
)

func main() {
	databaseURL := buildDatabaseURL()
	port := getenv("PORT", "8080")

	migrationsPath := getenv("MIGRATIONS_PATH", "./migrations")
	if err := db.RunMigrations(databaseURL, migrationsPath); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}
	log.Println("migrations applied")

	conn, err := db.Connect(databaseURL)
	if err != nil {
		log.Fatalf("db connection failed: %v", err)
	}
	defer conn.Close()

	settings := &models.Settings{
		LLMProvider:   getenv("LLM_PROVIDER", "ollama"),
		OllamaModel:   getenv("OLLAMA_MODEL", "qwen2.5:14b"),
		ClaudeAPIKey:  os.Getenv("CLAUDE_API_KEY"),
		OllamaBaseURL: getenv("OLLAMA_BASE_URL", "http://localhost:11434"),
		WhisperModel:  os.Getenv("WHISPER_MODEL"), // 空字串=停用；建議 small 或 medium
		TZ:            getenv("TZ", "Asia/Taipei"),
		DataDir:       getenv("DATA_DIR", "/app/data"),
		FinMindToken:  os.Getenv("FINMIND_TOKEN"),
	}

	router := api.NewRouter(conn, settings)
	log.Printf("listening on :%s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal(err)
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
