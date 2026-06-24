CREATE TABLE stock_prices (
    id SERIAL PRIMARY KEY,
    ticker VARCHAR(20) NOT NULL REFERENCES stocks(ticker) ON DELETE CASCADE,
    date DATE NOT NULL,
    open DECIMAL(15,4) DEFAULT 0,
    high DECIMAL(15,4) DEFAULT 0,
    low DECIMAL(15,4) DEFAULT 0,
    close DECIMAL(15,4) DEFAULT 0,
    volume BIGINT DEFAULT 0,
    adj_close DECIMAL(15,4) DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(ticker, date)
);
CREATE INDEX idx_prices_ticker_date ON stock_prices(ticker, date DESC);

CREATE TABLE fundamentals (
    id SERIAL PRIMARY KEY,
    ticker VARCHAR(20) NOT NULL REFERENCES stocks(ticker) ON DELETE CASCADE,
    period_type VARCHAR(10) NOT NULL,
    period VARCHAR(10) NOT NULL,
    revenue BIGINT DEFAULT 0,
    revenue_yoy DECIMAL(10,4) DEFAULT 0,
    eps DECIMAL(10,4) DEFAULT 0,
    eps_yoy DECIMAL(10,4) DEFAULT 0,
    roe DECIMAL(10,4) DEFAULT 0,
    gross_margin DECIMAL(10,4) DEFAULT 0,
    fcf BIGINT DEFAULT 0,
    debt_ratio DECIMAL(10,4) DEFAULT 0,
    reported_at DATE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(ticker, period_type, period)
);
CREATE INDEX idx_fundamentals_ticker ON fundamentals(ticker, period_type, period DESC);
