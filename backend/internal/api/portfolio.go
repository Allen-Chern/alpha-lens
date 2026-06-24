package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/alpha-lens/backend/internal/models"
)

func (h *Handler) ListPortfolios(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(`SELECT id, name, description, created_at FROM portfolios ORDER BY id`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	portfolios := []models.Portfolio{}
	for rows.Next() {
		var p models.Portfolio
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		portfolios = append(portfolios, p)
	}
	writeJSON(w, http.StatusOK, map[string]any{"portfolios": portfolios})
}

func (h *Handler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var t models.Transaction
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var created time.Time
	err := h.DB.QueryRow(
		`INSERT INTO transactions (portfolio_id, ticker, transaction_type, shares, price, fee, transaction_date, notes)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, created_at`,
		t.PortfolioID, t.Ticker, t.TransactionType, t.Shares, t.Price, t.Fee, t.TransactionDate, t.Notes).
		Scan(&t.ID, &created)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	t.CreatedAt = created
	writeJSON(w, http.StatusCreated, t)
}

func (h *Handler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	portfolioID := r.URL.Query().Get("portfolio_id")
	if portfolioID == "" {
		writeError(w, http.StatusBadRequest, "portfolio_id is required")
		return
	}

	rows, err := h.DB.Query(
		`SELECT id, portfolio_id, ticker, transaction_type, shares, price, fee, transaction_date, notes, created_at
		 FROM transactions WHERE portfolio_id = $1 ORDER BY transaction_date DESC, id DESC`, portfolioID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	txns := []models.Transaction{}
	for rows.Next() {
		var t models.Transaction
		var date time.Time
		if err := rows.Scan(&t.ID, &t.PortfolioID, &t.Ticker, &t.TransactionType, &t.Shares, &t.Price, &t.Fee, &date, &t.Notes, &t.CreatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		t.TransactionDate = date.Format(dateLayout)
		txns = append(txns, t)
	}
	writeJSON(w, http.StatusOK, map[string]any{"transactions": txns})
}

func (h *Handler) CreateRebate(w http.ResponseWriter, r *http.Request) {
	var rb models.FeeRebate
	if err := json.NewDecoder(r.Body).Decode(&rb); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	var created time.Time
	err := h.DB.QueryRow(
		`INSERT INTO fee_rebates (portfolio_id, amount, rebate_date, broker, notes)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`,
		rb.PortfolioID, rb.Amount, rb.RebateDate, rb.Broker, rb.Notes).
		Scan(&rb.ID, &created)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	rb.CreatedAt = created
	writeJSON(w, http.StatusCreated, rb)
}

func (h *Handler) CreateDividend(w http.ResponseWriter, r *http.Request) {
	var d models.Dividend
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	var created time.Time
	err := h.DB.QueryRow(
		`INSERT INTO dividends (portfolio_id, ticker, amount, dividend_date, notes)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`,
		d.PortfolioID, d.Ticker, d.Amount, d.DividendDate, d.Notes).
		Scan(&d.ID, &created)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	d.CreatedAt = created
	writeJSON(w, http.StatusCreated, d)
}

type position struct {
	shares  float64
	avgCost float64
}

func (h *Handler) GetPnL(w http.ResponseWriter, r *http.Request) {
	portfolioID := r.URL.Query().Get("portfolio_id")
	if portfolioID == "" {
		writeError(w, http.StatusBadRequest, "portfolio_id is required")
		return
	}

	rows, err := h.DB.Query(
		`SELECT ticker, transaction_type, shares, price, fee
		 FROM transactions WHERE portfolio_id = $1 ORDER BY transaction_date ASC, id ASC`, portfolioID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	positions := map[string]*position{}
	order := []string{}
	var totalRealized, totalFees float64

	for rows.Next() {
		var ticker, txnType string
		var shares, price, fee float64
		if err := rows.Scan(&ticker, &txnType, &shares, &price, &fee); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		totalFees += fee
		p, exists := positions[ticker]
		if !exists {
			p = &position{}
			positions[ticker] = p
			order = append(order, ticker)
		}
		if txnType == "SELL" {
			totalRealized += (price - p.avgCost) * shares
			p.shares -= shares
			if p.shares <= 0 {
				p.shares = 0
				p.avgCost = 0
			}
		} else {
			newTotal := p.shares + shares
			if newTotal > 0 {
				p.avgCost = (p.avgCost*p.shares + price*shares) / newTotal
			}
			p.shares = newTotal
		}
	}

	result := models.PnLResult{
		PortfolioID:      atoiDefault(portfolioID, 0),
		Holdings:         []models.Holding{},
		TotalRealizedPnL: totalRealized,
		TotalFees:        totalFees,
	}

	for _, ticker := range order {
		p := positions[ticker]
		if p.shares <= 0 {
			continue
		}
		var name string
		h.DB.QueryRow(`SELECT name FROM stocks WHERE ticker = $1`, ticker).Scan(&name)

		var current float64
		h.DB.QueryRow(`SELECT close FROM stock_prices WHERE ticker = $1 ORDER BY date DESC LIMIT 1`, ticker).Scan(&current)

		costBasis := p.shares * p.avgCost
		marketValue := p.shares * current
		unrealized := marketValue - costBasis
		var unrealizedPct float64
		if costBasis != 0 {
			unrealizedPct = unrealized / costBasis * 100
		}

		result.Holdings = append(result.Holdings, models.Holding{
			Ticker:           ticker,
			Name:             name,
			Shares:           p.shares,
			AvgCost:          p.avgCost,
			CurrentPrice:     current,
			MarketValue:      marketValue,
			UnrealizedPnL:    unrealized,
			UnrealizedPnLPct: unrealizedPct,
		})
		result.TotalCost += costBasis
		result.TotalValue += marketValue
		result.TotalUnrealizedPnL += unrealized
	}

	h.DB.QueryRow(`SELECT COALESCE(SUM(amount), 0) FROM dividends WHERE portfolio_id = $1`, portfolioID).Scan(&result.TotalDividends)
	h.DB.QueryRow(`SELECT COALESCE(SUM(amount), 0) FROM fee_rebates WHERE portfolio_id = $1`, portfolioID).Scan(&result.TotalRebates)

	writeJSON(w, http.StatusOK, result)
}

func atoiDefault(s string, def int) int {
	return parseIntDefault(s, def)
}
