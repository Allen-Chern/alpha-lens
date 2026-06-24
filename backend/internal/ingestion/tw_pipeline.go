package ingestion

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// RunTWDaily 執行台股每日資料擷取，回傳寫入筆數
func RunTWDaily(db *sql.DB, token string) (int, error) {
	loc, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		// Alpine 預設無 timezone 資料，fallback 到 UTC+8
		loc = time.FixedZone("CST", 8*60*60)
	}
	date := time.Now().In(loc).Format("2006-01-02")
	return RunTWDailyForDate(db, token, date)
}

type twStock struct {
	Ticker   string
	Exchange string
}

// RunTWDailyForDate 可指定日期，方便補跑歷史資料
func RunTWDailyForDate(db *sql.DB, token string, date string) (int, error) {
	client := NewClient(token)
	log.Printf("[TW pipeline] date=%s start", date)

	stocks, err := activeTWStocks(db)
	if err != nil {
		return 0, fmt.Errorf("get tickers: %w", err)
	}
	if len(stocks) == 0 {
		log.Printf("[TW pipeline] no active TW stocks in DB")
		return 0, nil
	}
	tickers := make([]string, len(stocks))
	for i, s := range stocks {
		tickers[i] = s.Ticker
	}
	log.Printf("[TW pipeline] fetching %d tickers", len(tickers))

	// ── Stage 1a: 股價 ───────────────────────────────────────────────────
	prices, err := client.FetchPrices(date, tickers)
	if err != nil {
		return 0, fmt.Errorf("fetch prices: %w", err)
	}
	if len(prices) == 0 {
		log.Printf("[TW pipeline] date=%s no price data (market closed?)", date)
		return 0, nil
	}
	log.Printf("[TW pipeline] prices: %d rows", len(prices))

	written, err := upsertPrices(db, prices)
	if err != nil {
		return 0, fmt.Errorf("upsert prices: %w", err)
	}
	log.Printf("[TW pipeline] prices upserted: %d", written)

	// ── Stage 1b: 法人籌碼 + 融資融券 ───────────────────────────────────
	flows, err := client.FetchInstitutional(date, tickers)
	if err != nil {
		log.Printf("[TW pipeline] institutional fetch error (non-fatal): %v", err)
	} else {
		log.Printf("[TW pipeline] institutional: %d stocks", len(flows))
		if err := client.FetchMargin(date, tickers, flows); err != nil {
			log.Printf("[TW pipeline] margin fetch error (non-fatal): %v", err)
		}
		if err := upsertInstFlows(db, flows); err != nil {
			return 0, fmt.Errorf("upsert inst flows: %w", err)
		}
	}

	// ── Stage 1c: 月營收 ────────────────────────────────────────────────
	// 優先 MOPS（台灣 IP 環境），fallback 到 FinMind 免費層（逐支呼叫）
	var revenueRows []RevenueRow
	mopsRows, mopsErr := FetchMOPSRevenueRecent(2)
	if mopsErr == nil && len(mopsRows) > 0 {
		tickerSet := make(map[string]bool, len(tickers))
		for _, t := range tickers {
			tickerSet[t] = true
		}
		for _, r := range mopsRows {
			if tickerSet[r.Ticker] {
				revenueRows = append(revenueRows, r)
			}
		}
		log.Printf("[TW pipeline] mops revenue: %d rows", len(revenueRows))
	} else {
		if mopsErr != nil {
			log.Printf("[TW pipeline] mops unavailable (%v), using FinMind", mopsErr)
		}
		revenueRows, _ = client.FetchRevenuePerTicker(tickers, 2)
		log.Printf("[TW pipeline] finmind revenue: %d rows", len(revenueRows))
	}
	if len(revenueRows) > 0 {
		if err := upsertRevenue(db, revenueRows); err != nil {
			log.Printf("[TW pipeline] upsert revenue error (non-fatal): %v", err)
		}
	}

	log.Printf("[TW pipeline] date=%s done, %d stocks", date, written)
	return written, nil
}

func activeTWStocks(db *sql.DB) ([]twStock, error) {
	rows, err := db.Query(
		`SELECT ticker, exchange FROM stocks WHERE market='TW' AND is_active=true ORDER BY ticker`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stocks []twStock
	for rows.Next() {
		var s twStock
		if err := rows.Scan(&s.Ticker, &s.Exchange); err != nil {
			return nil, err
		}
		stocks = append(stocks, s)
	}
	return stocks, rows.Err()
}

// ─── DB upsert helpers ──────────────────────────────────────────────────────

func upsertPrices(db *sql.DB, rows []PriceRow) (int, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO stock_prices (ticker, date, open, high, low, close, volume)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (ticker, date) DO UPDATE SET
			open   = EXCLUDED.open,
			high   = EXCLUDED.high,
			low    = EXCLUDED.low,
			close  = EXCLUDED.close,
			volume = EXCLUDED.volume`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	count := 0
	for _, r := range rows {
		if _, err := stmt.Exec(r.Ticker, r.Date, r.Open, r.High, r.Low, r.Close, r.Volume); err != nil {
			// FK 失敗代表不在標的池，跳過
			continue
		}
		count++
	}
	return count, tx.Commit()
}

func upsertInstFlows(db *sql.DB, flows map[string]*InstFlowRow) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO institutional_flows
			(ticker, date, foreign_buy, foreign_sell, trust_buy, trust_sell, margin_balance, short_balance)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (ticker, date) DO UPDATE SET
			foreign_buy    = EXCLUDED.foreign_buy,
			foreign_sell   = EXCLUDED.foreign_sell,
			trust_buy      = EXCLUDED.trust_buy,
			trust_sell     = EXCLUDED.trust_sell,
			margin_balance = EXCLUDED.margin_balance,
			short_balance  = EXCLUDED.short_balance`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, r := range flows {
		if _, err := stmt.Exec(r.Ticker, r.Date,
			r.ForeignBuy, r.ForeignSell,
			r.TrustBuy, r.TrustSell,
			r.MarginBalance, r.ShortBalance); err != nil {
			continue // 不在標的池，跳過
		}
	}
	return tx.Commit()
}

func upsertRevenue(db *sql.DB, rows []RevenueRow) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO fundamentals (ticker, period_type, period, revenue, reported_at)
		VALUES ($1, 'monthly', $2, $3, $4)
		ON CONFLICT (ticker, period_type, period) DO UPDATE SET
			revenue     = EXCLUDED.revenue,
			reported_at = EXCLUDED.reported_at`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, r := range rows {
		if _, err := stmt.Exec(r.Ticker, r.Period, r.Amount, r.Date); err != nil {
			continue
		}
	}
	return tx.Commit()
}
