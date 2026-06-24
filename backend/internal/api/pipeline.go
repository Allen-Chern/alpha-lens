package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/alpha-lens/backend/internal/ingestion"
	"github.com/alpha-lens/backend/internal/models"
)

func (h *Handler) PipelineStatus(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(
		`SELECT id, pipeline_type, started_at, completed_at, status, stocks_processed
		 FROM pipeline_runs ORDER BY started_at DESC LIMIT 20`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	runs := []models.PipelineRun{}
	for rows.Next() {
		var run models.PipelineRun
		var completedAt sql.NullTime
		if err := rows.Scan(&run.ID, &run.PipelineType, &run.StartedAt, &completedAt, &run.Status, &run.StocksProcessed); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if completedAt.Valid {
			run.CompletedAt = &completedAt.Time
		}
		runs = append(runs, run)
	}
	writeJSON(w, http.StatusOK, map[string]any{"runs": runs})
}

func (h *Handler) PipelineRun(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PipelineType string `json:"pipeline_type"`
		Date         string `json:"date"` // 可選，YYYY-MM-DD，省略時用今日
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.PipelineType == "" {
		writeError(w, http.StatusBadRequest, "pipeline_type is required")
		return
	}

	var run models.PipelineRun
	err := h.DB.QueryRow(
		`INSERT INTO pipeline_runs (pipeline_type, status, stocks_processed)
		 VALUES ($1, 'RUNNING', 0)
		 RETURNING id, pipeline_type, started_at, status, stocks_processed`,
		req.PipelineType).
		Scan(&run.ID, &run.PipelineType, &run.StartedAt, &run.Status, &run.StocksProcessed)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	go func(id int, pipelineType, date string) {
		var processed int
		var runErr error

		switch pipelineType {
		case "TW_DAILY":
			if date != "" {
				processed, runErr = ingestion.RunTWDailyForDate(h.DB, h.Settings.FinMindToken, date)
			} else {
				processed, runErr = ingestion.RunTWDaily(h.DB, h.Settings.FinMindToken)
			}
		default:
			log.Printf("[pipeline] %s not yet implemented", pipelineType)
			time.Sleep(2 * time.Second)
			processed = 0
		}

		if runErr != nil {
			log.Printf("[pipeline] %s error: %v", pipelineType, runErr)
			h.DB.Exec(
				`UPDATE pipeline_runs SET status='FAILED', error_message=$1, completed_at=NOW() WHERE id=$2`,
				runErr.Error(), id)
			return
		}
		h.DB.Exec(
			`UPDATE pipeline_runs SET status='SUCCESS', stocks_processed=$1, completed_at=NOW() WHERE id=$2`,
			processed, id)
	}(run.ID, req.PipelineType, req.Date)

	writeJSON(w, http.StatusOK, run)
}
