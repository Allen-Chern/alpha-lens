package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/alpha-lens/backend/internal/models"
	"github.com/go-chi/chi/v5"
)

const dateLayout = "2006-01-02"

func (h *Handler) ListStocks(w http.ResponseWriter, r *http.Request) {
	market := r.URL.Query().Get("market")
	page := parseIntDefault(r.URL.Query().Get("page"), 1)
	limit := parseIntDefault(r.URL.Query().Get("limit"), 50)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}
	offset := (page - 1) * limit

	var total int
	countQuery := "SELECT COUNT(*) FROM stocks WHERE ($1 = '' OR market = $1)"
	if err := h.DB.QueryRow(countQuery, market).Scan(&total); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	rows, err := h.DB.Query(
		`SELECT ticker, name, market, exchange, market_cap, avg_volume, is_active
		 FROM stocks
		 WHERE ($1 = '' OR market = $1)
		 ORDER BY ticker
		 LIMIT $2 OFFSET $3`,
		market, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	stocks := []models.Stock{}
	for rows.Next() {
		var s models.Stock
		if err := rows.Scan(&s.Ticker, &s.Name, &s.Market, &s.Exchange, &s.MarketCap, &s.AvgVolume, &s.IsActive); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		stocks = append(stocks, s)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"stocks": stocks,
		"total":  total,
		"page":   page,
		"limit":  limit,
	})
}

func (h *Handler) GetStock(w http.ResponseWriter, r *http.Request) {
	ticker := chi.URLParam(r, "ticker")

	var s models.Stock
	err := h.DB.QueryRow(
		`SELECT ticker, name, market, exchange, market_cap, avg_volume, is_active
		 FROM stocks WHERE ticker = $1`, ticker).
		Scan(&s.Ticker, &s.Name, &s.Market, &s.Exchange, &s.MarketCap, &s.AvgVolume, &s.IsActive)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "stock not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.LatestPrice = h.latestPrice(ticker)
	s.LatestScore = h.latestScore(ticker)

	writeJSON(w, http.StatusOK, s)
}

func (h *Handler) latestPrice(ticker string) *models.StockPrice {
	var p models.StockPrice
	var date time.Time
	err := h.DB.QueryRow(
		`SELECT date, open, high, low, close, volume
		 FROM stock_prices WHERE ticker = $1 ORDER BY date DESC LIMIT 1`, ticker).
		Scan(&date, &p.Open, &p.High, &p.Low, &p.Close, &p.Volume)
	if err != nil {
		return nil
	}
	p.Date = date.Format(dateLayout)
	return &p
}

func (h *Handler) latestScore(ticker string) *models.StockScore {
	var sc models.StockScore
	var date time.Time
	err := h.DB.QueryRow(
		`SELECT date, total_score, fundamental_score, chip_score, momentum_score, theme_score, risk_score
		 FROM stock_scores WHERE ticker = $1 ORDER BY date DESC LIMIT 1`, ticker).
		Scan(&date, &sc.TotalScore, &sc.FundamentalScore, &sc.ChipScore, &sc.MomentumScore, &sc.ThemeScore, &sc.RiskScore)
	if err != nil {
		return nil
	}
	sc.Date = date.Format(dateLayout)
	return &sc
}

func (h *Handler) GetStockScore(w http.ResponseWriter, r *http.Request) {
	ticker := chi.URLParam(r, "ticker")

	var scoreID int
	var sc models.StockScore
	var date time.Time
	err := h.DB.QueryRow(
		`SELECT id, date, total_score, fundamental_score, chip_score, momentum_score, theme_score, risk_score
		 FROM stock_scores WHERE ticker = $1 ORDER BY date DESC LIMIT 1`, ticker).
		Scan(&scoreID, &date, &sc.TotalScore, &sc.FundamentalScore, &sc.ChipScore, &sc.MomentumScore, &sc.ThemeScore, &sc.RiskScore)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "score not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	sc.Ticker = ticker
	sc.Date = date.Format(dateLayout)

	rows, err := h.DB.Query(
		`SELECT indicator_name, indicator_value, indicator_score, weight
		 FROM score_breakdown WHERE score_id = $1 ORDER BY id`, scoreID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	sc.Breakdown = []models.ScoreBreakdown{}
	for rows.Next() {
		var b models.ScoreBreakdown
		if err := rows.Scan(&b.IndicatorName, &b.IndicatorValue, &b.IndicatorScore, &b.Weight); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		sc.Breakdown = append(sc.Breakdown, b)
	}

	writeJSON(w, http.StatusOK, sc)
}

func (h *Handler) GetStockReport(w http.ResponseWriter, r *http.Request) {
	ticker := chi.URLParam(r, "ticker")

	var rep models.Report
	var date time.Time
	err := h.DB.QueryRow(
		`SELECT ticker, date, market, file_path, summary
		 FROM reports WHERE ticker = $1 ORDER BY date DESC LIMIT 1`, ticker).
		Scan(&rep.Ticker, &date, &rep.Market, &rep.FilePath, &rep.Summary)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "report not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	rep.Date = date.Format(dateLayout)
	writeJSON(w, http.StatusOK, rep)
}

func (h *Handler) CreateStockReport(w http.ResponseWriter, r *http.Request) {
	ticker := chi.URLParam(r, "ticker")

	var name, market string
	err := h.DB.QueryRow(`SELECT name, market FROM stocks WHERE ticker = $1`, ticker).Scan(&name, &market)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "stock not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	today := time.Now().Format(dateLayout)
	dir := filepath.Join("data", "reports", market, ticker)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	filePath := filepath.Join(dir, today+".md")
	summary := "報告已產生"
	content := fmt.Sprintf("# %s (%s) 投資分析報告\n\n日期：%s\n市場：%s\n\n%s\n", name, ticker, today, market, summary)
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_, err = h.DB.Exec(
		`INSERT INTO reports (ticker, date, market, file_path, summary)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (ticker, date)
		 DO UPDATE SET market = EXCLUDED.market, file_path = EXCLUDED.file_path, summary = EXCLUDED.summary`,
		ticker, today, market, filePath, summary)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.Report{
		Ticker:   ticker,
		Date:     today,
		Market:   market,
		FilePath: filePath,
		Summary:  summary,
	})
}

func (h *Handler) GetStockHistory(w http.ResponseWriter, r *http.Request) {
	ticker := chi.URLParam(r, "ticker")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	rows, err := h.DB.Query(
		`SELECT date, open, high, low, close, volume
		 FROM stock_prices
		 WHERE ticker = $1
		   AND ($2 = '' OR date >= $2::date)
		   AND ($3 = '' OR date <= $3::date)
		 ORDER BY date ASC`,
		ticker, from, to)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	prices := []models.StockPrice{}
	for rows.Next() {
		var p models.StockPrice
		var date time.Time
		if err := rows.Scan(&date, &p.Open, &p.High, &p.Low, &p.Close, &p.Volume); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		p.Date = date.Format(dateLayout)
		prices = append(prices, p)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ticker": ticker,
		"prices": prices,
	})
}

func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}
