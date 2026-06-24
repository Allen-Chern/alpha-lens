package models

import "time"

type Stock struct {
	Ticker      string      `json:"ticker"`
	Name        string      `json:"name"`
	Market      string      `json:"market"`
	Exchange    string      `json:"exchange"`
	MarketCap   int64       `json:"market_cap"`
	AvgVolume   int64       `json:"avg_volume"`
	IsActive    bool        `json:"is_active"`
	LatestPrice *StockPrice `json:"latest_price,omitempty"`
	LatestScore *StockScore `json:"latest_score,omitempty"`
}

type StockPrice struct {
	Date   string  `json:"date"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume int64   `json:"volume"`
}

type StockScore struct {
	Ticker           string           `json:"ticker,omitempty"`
	Date             string           `json:"date"`
	TotalScore       float64          `json:"total_score"`
	FundamentalScore float64          `json:"fundamental_score"`
	ChipScore        float64          `json:"chip_score"`
	MomentumScore    float64          `json:"momentum_score"`
	ThemeScore       float64          `json:"theme_score"`
	RiskScore        float64          `json:"risk_score"`
	Breakdown        []ScoreBreakdown `json:"breakdown,omitempty"`
}

type ScoreBreakdown struct {
	IndicatorName  string  `json:"indicator_name"`
	IndicatorValue float64 `json:"indicator_value"`
	IndicatorScore float64 `json:"indicator_score"`
	Weight         float64 `json:"weight"`
}

type Report struct {
	Ticker   string `json:"ticker"`
	Date     string `json:"date"`
	Market   string `json:"market"`
	FilePath string `json:"file_path"`
	Summary  string `json:"summary"`
}

type ScreenerResult struct {
	Ticker           string  `json:"ticker"`
	Name             string  `json:"name"`
	Market           string  `json:"market"`
	TotalScore       float64 `json:"total_score"`
	FundamentalScore float64 `json:"fundamental_score"`
	ChipScore        float64 `json:"chip_score"`
	MomentumScore    float64 `json:"momentum_score"`
	ThemeScore       float64 `json:"theme_score"`
	RiskScore        float64 `json:"risk_score"`
	Close            float64 `json:"close"`
	ChangePct        float64 `json:"change_pct"`
}

type Portfolio struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type Transaction struct {
	ID              int       `json:"id"`
	PortfolioID     int       `json:"portfolio_id"`
	Ticker          string    `json:"ticker"`
	TransactionType string    `json:"transaction_type"`
	Shares          float64   `json:"shares"`
	Price           float64   `json:"price"`
	Fee             float64   `json:"fee"`
	TransactionDate string    `json:"transaction_date"`
	Notes           string    `json:"notes"`
	CreatedAt       time.Time `json:"created_at"`
}

type FeeRebate struct {
	ID          int       `json:"id"`
	PortfolioID int       `json:"portfolio_id"`
	Amount      float64   `json:"amount"`
	RebateDate  string    `json:"rebate_date"`
	Broker      string    `json:"broker"`
	Notes       string    `json:"notes"`
	CreatedAt   time.Time `json:"created_at"`
}

type Dividend struct {
	ID           int       `json:"id"`
	PortfolioID  int       `json:"portfolio_id"`
	Ticker       string    `json:"ticker"`
	Amount       float64   `json:"amount"`
	DividendDate string    `json:"dividend_date"`
	Notes        string    `json:"notes"`
	CreatedAt    time.Time `json:"created_at"`
}

type Holding struct {
	Ticker          string  `json:"ticker"`
	Name            string  `json:"name"`
	Shares          float64 `json:"shares"`
	AvgCost         float64 `json:"avg_cost"`
	CurrentPrice    float64 `json:"current_price"`
	MarketValue     float64 `json:"market_value"`
	UnrealizedPnL   float64 `json:"unrealized_pnl"`
	UnrealizedPnLPct float64 `json:"unrealized_pnl_pct"`
}

type PnLResult struct {
	PortfolioID        int       `json:"portfolio_id"`
	Holdings           []Holding `json:"holdings"`
	TotalCost          float64   `json:"total_cost"`
	TotalValue         float64   `json:"total_value"`
	TotalUnrealizedPnL float64   `json:"total_unrealized_pnl"`
	TotalRealizedPnL   float64   `json:"total_realized_pnl"`
	TotalDividends     float64   `json:"total_dividends"`
	TotalRebates       float64   `json:"total_rebates"`
	TotalFees          float64   `json:"total_fees"`
}

type BacktestRun struct {
	ID             int             `json:"id"`
	Name           string          `json:"name"`
	Market         string          `json:"market"`
	StartDate      string          `json:"start_date"`
	EndDate        string          `json:"end_date"`
	InitialCapital float64         `json:"initial_capital"`
	Parameters     map[string]any  `json:"parameters,omitempty"`
	Status         string          `json:"status"`
	Result         *BacktestResult `json:"result"`
	CreatedAt      time.Time       `json:"created_at"`
	CompletedAt    *time.Time      `json:"completed_at"`
}

type BacktestResult struct {
	TotalReturn float64 `json:"total_return"`
	MaxDrawdown float64 `json:"max_drawdown"`
	SharpeRatio float64 `json:"sharpe_ratio"`
	WinRate     float64 `json:"win_rate"`
	TotalTrades int     `json:"total_trades"`
}

type BacktestTrade struct {
	ID        int     `json:"id"`
	Ticker    string  `json:"ticker"`
	Action    string  `json:"action"`
	Price     float64 `json:"price"`
	Shares    float64 `json:"shares"`
	TradeDate string  `json:"trade_date"`
	PnL       float64 `json:"pnl"`
}

type PipelineRun struct {
	ID              int        `json:"id"`
	PipelineType    string     `json:"pipeline_type"`
	StartedAt       time.Time  `json:"started_at"`
	CompletedAt     *time.Time `json:"completed_at"`
	Status          string     `json:"status"`
	StocksProcessed int        `json:"stocks_processed"`
}

type Settings struct {
	LLMProvider   string `json:"llm_provider"`
	OllamaModel   string `json:"ollama_model"`
	OllamaBaseURL string `json:"ollama_base_url"`
	WhisperModel  string `json:"whisper_model"` // tiny/base/small/medium/large-v3，空字串=停用
	TZ            string `json:"tz"`
	DataDir       string `json:"-"` // transcript 等大型檔案的根目錄，對應 volume ./data:/app/data
	ClaudeAPIKey  string `json:"-"`
	FinMindToken  string `json:"-"`
}
