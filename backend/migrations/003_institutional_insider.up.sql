CREATE TABLE institutional_flows (
    id SERIAL PRIMARY KEY,
    ticker VARCHAR(20) NOT NULL REFERENCES stocks(ticker) ON DELETE CASCADE,
    date DATE NOT NULL,
    foreign_buy BIGINT DEFAULT 0,
    foreign_sell BIGINT DEFAULT 0,
    trust_buy BIGINT DEFAULT 0,
    trust_sell BIGINT DEFAULT 0,
    margin_balance BIGINT DEFAULT 0,
    short_balance BIGINT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(ticker, date)
);
CREATE INDEX idx_inst_flows_ticker_date ON institutional_flows(ticker, date DESC);

CREATE TABLE insider_transactions (
    id SERIAL PRIMARY KEY,
    ticker VARCHAR(20) NOT NULL REFERENCES stocks(ticker) ON DELETE CASCADE,
    market VARCHAR(10) NOT NULL DEFAULT '',
    insider_name VARCHAR(200) DEFAULT '',
    transaction_type VARCHAR(20) DEFAULT '',
    shares BIGINT DEFAULT 0,
    price DECIMAL(15,4) DEFAULT 0,
    transaction_date DATE,
    reported_at DATE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_insider_ticker ON insider_transactions(ticker, transaction_date DESC);

CREATE TABLE congressional_trades (
    id SERIAL PRIMARY KEY,
    ticker VARCHAR(20) NOT NULL REFERENCES stocks(ticker) ON DELETE CASCADE,
    trader_name VARCHAR(200) DEFAULT '',
    transaction_type VARCHAR(20) DEFAULT '',
    amount_range VARCHAR(50) DEFAULT '',
    transaction_date DATE,
    reported_at DATE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_congress_ticker ON congressional_trades(ticker, transaction_date DESC);
