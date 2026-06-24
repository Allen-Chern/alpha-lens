package ingestion

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const finmindBaseURL = "https://api.finmindtrade.com/api/v4/data"

type Client struct {
	token      string
	httpClient *http.Client
}

func NewClient(token string) *Client {
	return &Client{
		token:      token,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

type finmindResponse struct {
	Status int                      `json:"status"`
	Msg    string                   `json:"msg"`
	Data   []map[string]interface{} `json:"data"`
}

func (c *Client) fetch(dataset, dataID, startDate, endDate string) ([]map[string]interface{}, error) {
	params := url.Values{}
	params.Set("dataset", dataset)
	params.Set("start_date", startDate)
	params.Set("end_date", endDate)
	params.Set("token", c.token)
	if dataID != "" {
		params.Set("data_id", dataID)
	}

	resp, err := c.httpClient.Get(finmindBaseURL + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("finmind %s: %w", dataset, err)
	}
	defer resp.Body.Close()

	var result finmindResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("finmind %s decode: %w", dataset, err)
	}
	if result.Status != 200 {
		return nil, fmt.Errorf("finmind %s: status %d: %s", dataset, result.Status, result.Msg)
	}
	return result.Data, nil
}

// ─── typed helpers ─────────────────────────────────────────────────────────

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getFloat(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok {
		switch n := v.(type) {
		case float64:
			return n
		case int:
			return float64(n)
		}
	}
	return 0
}

func getInt64(m map[string]interface{}, key string) int64 {
	return int64(getFloat(m, key))
}

// ─── dataset fetchers ───────────────────────────────────────────────────────

type PriceRow struct {
	Date   string
	Ticker string
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume int64
}

// FetchPrices 抓取指定日期的股價。tickers 為空時嘗試 bulk（需付費層），
// 非空時逐支呼叫（免費層可用）。
func (c *Client) FetchPrices(date string, tickers []string) ([]PriceRow, error) {
	if len(tickers) == 0 {
		return c.fetchPricesDataID("", date)
	}
	var all []PriceRow
	for _, t := range tickers {
		rows, err := c.fetchPricesDataID(t, date)
		if err != nil {
			return nil, err
		}
		all = append(all, rows...)
	}
	return all, nil
}

func (c *Client) fetchPricesDataID(dataID, date string) ([]PriceRow, error) {
	raw, err := c.fetch("TaiwanStockPrice", dataID, date, date)
	if err != nil {
		return nil, err
	}
	rows := make([]PriceRow, 0, len(raw))
	for _, m := range raw {
		rows = append(rows, PriceRow{
			Date:   getString(m, "date"),
			Ticker: getString(m, "stock_id"),
			Open:   getFloat(m, "open"),
			High:   getFloat(m, "max"),
			Low:    getFloat(m, "min"),
			Close:  getFloat(m, "close"),
			Volume: getInt64(m, "Trading_Volume"),
		})
	}
	return rows, nil
}

type InstFlowRow struct {
	Date          string
	Ticker        string
	ForeignBuy    int64
	ForeignSell   int64
	TrustBuy      int64
	TrustSell     int64
	MarginBalance int64
	ShortBalance  int64
}

// FetchInstitutional 抓取法人買賣超並歸類為外資/投信。
// tickers 非空時逐支呼叫（免費層），空時 bulk 呼叫（需付費層）。
func (c *Client) FetchInstitutional(date string, tickers []string) (map[string]*InstFlowRow, error) {
	var raw []map[string]interface{}
	if len(tickers) == 0 {
		var err error
		raw, err = c.fetch("TaiwanStockInstitutionalInvestorsBuySell", "", date, date)
		if err != nil {
			return nil, err
		}
	} else {
		for _, t := range tickers {
			rows, err := c.fetch("TaiwanStockInstitutionalInvestorsBuySell", t, date, date)
			if err != nil {
				return nil, err
			}
			raw = append(raw, rows...)
		}
	}

	flows := make(map[string]*InstFlowRow)
	for _, m := range raw {
		ticker := getString(m, "stock_id")
		name := getString(m, "name")
		buy := getInt64(m, "buy")
		sell := getInt64(m, "sell")

		if _, ok := flows[ticker]; !ok {
			flows[ticker] = &InstFlowRow{Date: date, Ticker: ticker}
		}
		row := flows[ticker]

		switch {
		case containsAny(name, "外陸資", "外資自營商"):
			row.ForeignBuy += buy
			row.ForeignSell += sell
		case containsAny(name, "投信"):
			row.TrustBuy += buy
			row.TrustSell += sell
		}
	}
	return flows, nil
}

// FetchMargin 抓取融資融券餘額並合併進 flows map。
func (c *Client) FetchMargin(date string, tickers []string, flows map[string]*InstFlowRow) error {
	var raw []map[string]interface{}
	if len(tickers) == 0 {
		var err error
		raw, err = c.fetch("TaiwanStockMarginPurchaseShortSale", "", date, date)
		if err != nil {
			return err
		}
	} else {
		for _, t := range tickers {
			rows, err := c.fetch("TaiwanStockMarginPurchaseShortSale", t, date, date)
			if err != nil {
				return err
			}
			raw = append(raw, rows...)
		}
	}
	for _, m := range raw {
		ticker := getString(m, "stock_id")
		if _, ok := flows[ticker]; !ok {
			flows[ticker] = &InstFlowRow{Date: date, Ticker: ticker}
		}
		flows[ticker].MarginBalance = getInt64(m, "MarginPurchaseTodayBalance")
		flows[ticker].ShortBalance = getInt64(m, "ShortSaleTodayBalance")
	}
	return nil
}

type RevenueRow struct {
	Date   string // YYYY-MM-01
	Ticker string
	Period string // YYYY-MM
	Amount int64
}

// FetchRevenuePerTicker 用免費層逐支呼叫抓月營收（最近 N 個月）
func (c *Client) FetchRevenuePerTicker(tickers []string, months int) ([]RevenueRow, error) {
	// 計算 start_date：N 個月前的第一天
	startDate := time.Now().AddDate(0, -months, 0).Format("2006-01") + "-01"
	endDate := time.Now().Format("2006-01") + "-01"

	var all []RevenueRow
	for _, t := range tickers {
		raw, err := c.fetch("TaiwanStockMonthRevenue", t, startDate, endDate)
		if err != nil {
			continue // non-fatal per ticker
		}
		for _, m := range raw {
			d := getString(m, "date") // YYYY-MM-01
			period := ""
			if len(d) >= 7 {
				period = d[:7]
			}
			amt := getInt64(m, "revenue")
			if amt <= 0 {
				continue
			}
			all = append(all, RevenueRow{
				Date:   d,
				Ticker: t,
				Period: period,
				Amount: amt,
			})
		}
	}
	return all, nil
}


func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if len(s) >= len(sub) {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
}
