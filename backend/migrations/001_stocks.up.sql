CREATE TABLE stocks (
    id SERIAL PRIMARY KEY,
    ticker VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(200) NOT NULL,
    market VARCHAR(10) NOT NULL,
    exchange VARCHAR(20) DEFAULT '',
    market_cap BIGINT DEFAULT 0,
    avg_volume BIGINT DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_stocks_market ON stocks(market);
CREATE INDEX idx_stocks_ticker ON stocks(ticker);
CREATE INDEX idx_stocks_active ON stocks(is_active);
