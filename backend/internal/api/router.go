package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/alpha-lens/backend/internal/models"
	ws "github.com/alpha-lens/backend/internal/ws"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	DB       *sql.DB
	Settings *models.Settings
	Hub      *ws.Hub
}

func NewRouter(db *sql.DB, settings *models.Settings) http.Handler {
	h := &Handler{DB: db, Settings: settings, Hub: ws.NewHub()}

	r := chi.NewRouter()
	r.Use(corsMiddleware)

	r.Get("/api/stocks", h.ListStocks)
	r.Get("/api/stocks/{ticker}", h.GetStock)
	r.Get("/api/stocks/{ticker}/score", h.GetStockScore)
	r.Get("/api/stocks/{ticker}/report", h.GetStockReport)
	r.Post("/api/stocks/{ticker}/report", h.CreateStockReport)
	r.Get("/api/stocks/{ticker}/history", h.GetStockHistory)

	r.Get("/api/screener", h.Screener)

	r.Get("/api/portfolio", h.ListPortfolios)
	r.Get("/api/portfolio/transactions", h.ListTransactions)
	r.Post("/api/portfolio/transactions", h.CreateTransaction)
	r.Post("/api/portfolio/rebates", h.CreateRebate)
	r.Post("/api/portfolio/dividends", h.CreateDividend)
	r.Get("/api/portfolio/pnl", h.GetPnL)

	r.Get("/api/backtest", h.ListBacktests)
	r.Post("/api/backtest", h.CreateBacktest)
	r.Get("/api/backtest/{id}", h.GetBacktest)
	r.Get("/api/backtest/{id}/trades", h.GetBacktestTrades)

	r.Get("/api/pipeline/status", h.PipelineStatus)
	r.Post("/api/pipeline/run", h.PipelineRun)

	r.Get("/api/settings", h.GetSettings)
	r.Put("/api/settings", h.UpdateSettings)

	r.Get("/api/insights/podcasts", h.ListPodcasts)
	r.Post("/api/insights/podcasts", h.AddPodcast)
	r.Delete("/api/insights/podcasts/{id}", h.DeletePodcast)
	r.Post("/api/insights/podcasts/{id}/sync", h.SyncPodcast)
	r.Get("/api/insights/podcasts/{id}/episodes", h.ListEpisodes)
	r.Get("/api/insights/mentions", h.ListMentions)
	r.Put("/api/insights/mentions/{id}/adopt", h.ToggleAdopt)
	r.Post("/api/insights/episodes/{id}/fetch-transcript", h.FetchTranscript)
	r.Post("/api/insights/episodes/{id}/analyze", h.AnalyzeEpisode)

	r.Get("/api/ws", h.ServeWS)

	return r
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
