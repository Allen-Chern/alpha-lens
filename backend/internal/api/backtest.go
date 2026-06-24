package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/alpha-lens/backend/internal/models"
	"github.com/go-chi/chi/v5"
)

type createBacktestReq struct {
	Name           string         `json:"name"`
	Market         string         `json:"market"`
	StartDate      string         `json:"start_date"`
	EndDate        string         `json:"end_date"`
	InitialCapital float64        `json:"initial_capital"`
	Parameters     map[string]any `json:"parameters"`
}

func (h *Handler) CreateBacktest(w http.ResponseWriter, r *http.Request) {
	var req createBacktestReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Market == "" {
		req.Market = "TW"
	}
	if req.Parameters == nil {
		req.Parameters = map[string]any{}
	}
	paramsJSON, err := json.Marshal(req.Parameters)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid parameters")
		return
	}

	var bt models.BacktestRun
	var created time.Time
	err = h.DB.QueryRow(
		`INSERT INTO backtest_runs (name, market, start_date, end_date, initial_capital, parameters, status)
		 VALUES ($1, $2, $3, $4, $5, $6, 'PENDING')
		 RETURNING id, status, created_at`,
		req.Name, req.Market, req.StartDate, req.EndDate, req.InitialCapital, paramsJSON).
		Scan(&bt.ID, &bt.Status, &created)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	bt.Name = req.Name
	bt.Market = req.Market
	bt.StartDate = req.StartDate
	bt.EndDate = req.EndDate
	bt.InitialCapital = req.InitialCapital
	bt.CreatedAt = created
	writeJSON(w, http.StatusCreated, map[string]any{
		"id":              bt.ID,
		"name":            bt.Name,
		"market":          bt.Market,
		"start_date":      bt.StartDate,
		"end_date":        bt.EndDate,
		"initial_capital": bt.InitialCapital,
		"status":          bt.Status,
		"created_at":      bt.CreatedAt,
	})
}

func (h *Handler) GetBacktest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var bt models.BacktestRun
	var startDate, endDate time.Time
	var paramsJSON []byte
	var completedAt sql.NullTime
	err := h.DB.QueryRow(
		`SELECT id, name, market, start_date, end_date, initial_capital, parameters, status, created_at, completed_at
		 FROM backtest_runs WHERE id = $1`, id).
		Scan(&bt.ID, &bt.Name, &bt.Market, &startDate, &endDate, &bt.InitialCapital, &paramsJSON, &bt.Status, &bt.CreatedAt, &completedAt)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "backtest not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	bt.StartDate = startDate.Format(dateLayout)
	bt.EndDate = endDate.Format(dateLayout)
	if len(paramsJSON) > 0 {
		json.Unmarshal(paramsJSON, &bt.Parameters)
	}
	if completedAt.Valid {
		bt.CompletedAt = &completedAt.Time
	}
	bt.Result = h.backtestResult(bt.ID)

	writeJSON(w, http.StatusOK, bt)
}

func (h *Handler) backtestResult(backtestID int) *models.BacktestResult {
	var res models.BacktestResult
	err := h.DB.QueryRow(
		`SELECT total_return, max_drawdown, sharpe_ratio, win_rate, total_trades
		 FROM backtest_results WHERE backtest_id = $1 ORDER BY id DESC LIMIT 1`, backtestID).
		Scan(&res.TotalReturn, &res.MaxDrawdown, &res.SharpeRatio, &res.WinRate, &res.TotalTrades)
	if err != nil {
		return nil
	}
	return &res
}

func (h *Handler) GetBacktestTrades(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	rows, err := h.DB.Query(
		`SELECT id, ticker, action, price, shares, trade_date, pnl
		 FROM backtest_trades WHERE backtest_id = $1 ORDER BY trade_date ASC, id ASC`, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	trades := []models.BacktestTrade{}
	for rows.Next() {
		var t models.BacktestTrade
		var date time.Time
		if err := rows.Scan(&t.ID, &t.Ticker, &t.Action, &t.Price, &t.Shares, &date, &t.PnL); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		t.TradeDate = date.Format(dateLayout)
		trades = append(trades, t)
	}
	writeJSON(w, http.StatusOK, map[string]any{"trades": trades})
}

func (h *Handler) ListBacktests(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(
		`SELECT id, name, market, start_date, end_date, initial_capital, status, created_at, completed_at
		 FROM backtest_runs ORDER BY created_at DESC, id DESC`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	backtests := []models.BacktestRun{}
	for rows.Next() {
		var bt models.BacktestRun
		var startDate, endDate time.Time
		var completedAt sql.NullTime
		if err := rows.Scan(&bt.ID, &bt.Name, &bt.Market, &startDate, &endDate, &bt.InitialCapital, &bt.Status, &bt.CreatedAt, &completedAt); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		bt.StartDate = startDate.Format(dateLayout)
		bt.EndDate = endDate.Format(dateLayout)
		if completedAt.Valid {
			bt.CompletedAt = &completedAt.Time
		}
		bt.Result = h.backtestResult(bt.ID)
		backtests = append(backtests, bt)
	}
	writeJSON(w, http.StatusOK, map[string]any{"backtests": backtests})
}
