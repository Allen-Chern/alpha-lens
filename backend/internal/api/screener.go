package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/alpha-lens/backend/internal/models"
)

var sortFields = map[string]string{
	"total_score":       "sc.total_score",
	"fundamental_score": "sc.fundamental_score",
	"chip_score":        "sc.chip_score",
	"momentum_score":    "sc.momentum_score",
	"theme_score":       "sc.theme_score",
	"risk_score":        "sc.risk_score",
	"close":             "p_today.close",
}

func (h *Handler) Screener(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	market := q.Get("market")
	minScore := parseFloatDefault(q.Get("min_score"), 0)
	sortKey := q.Get("sort")
	if sortKey == "" {
		sortKey = "total_score"
	}
	sortCol, ok := sortFields[sortKey]
	if !ok {
		sortCol = "sc.total_score"
	}
	order := "DESC"
	if q.Get("order") == "asc" {
		order = "ASC"
	}
	page := parseIntDefault(q.Get("page"), 1)
	limit := parseIntDefault(q.Get("limit"), 30)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 30
	}
	offset := (page - 1) * limit

	var total int
	countQuery := `
		SELECT COUNT(*)
		FROM stocks s
		JOIN stock_scores sc ON sc.ticker = s.ticker
		WHERE sc.date = (SELECT MAX(date) FROM stock_scores)
		  AND sc.total_score >= $1
		  AND ($2 = '' OR s.market = $2)`
	if err := h.DB.QueryRow(countQuery, minScore, market).Scan(&total); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	query := fmt.Sprintf(`
		SELECT
			s.ticker, s.name, s.market,
			sc.total_score, sc.fundamental_score, sc.chip_score,
			sc.momentum_score, sc.theme_score, sc.risk_score,
			COALESCE(p_today.close, 0) AS close,
			COALESCE((p_today.close - p_prev.close) / NULLIF(p_prev.close, 0) * 100, 0) AS change_pct
		FROM stocks s
		JOIN stock_scores sc ON sc.ticker = s.ticker
		LEFT JOIN LATERAL (
			SELECT close, date FROM stock_prices p
			WHERE p.ticker = s.ticker ORDER BY p.date DESC LIMIT 1
		) p_today ON true
		LEFT JOIN LATERAL (
			SELECT close FROM stock_prices p
			WHERE p.ticker = s.ticker AND p.date < p_today.date
			ORDER BY p.date DESC LIMIT 1
		) p_prev ON true
		WHERE sc.date = (SELECT MAX(date) FROM stock_scores)
		  AND sc.total_score >= $1
		  AND ($2 = '' OR s.market = $2)
		ORDER BY %s %s
		LIMIT $3 OFFSET $4`, sortCol, order)

	rows, err := h.DB.Query(query, minScore, market, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	results := []models.ScreenerResult{}
	for rows.Next() {
		var sr models.ScreenerResult
		if err := rows.Scan(
			&sr.Ticker, &sr.Name, &sr.Market,
			&sr.TotalScore, &sr.FundamentalScore, &sr.ChipScore,
			&sr.MomentumScore, &sr.ThemeScore, &sr.RiskScore,
			&sr.Close, &sr.ChangePct,
		); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		results = append(results, sr)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"stocks": results,
		"total":  total,
	})
}

func parseFloatDefault(s string, def float64) float64 {
	if s == "" {
		return def
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return def
	}
	return v
}
