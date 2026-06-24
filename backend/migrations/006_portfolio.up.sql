CREATE TABLE portfolios (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    description TEXT DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    portfolio_id INT NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
    ticker VARCHAR(20) NOT NULL,
    transaction_type VARCHAR(10) NOT NULL,
    shares DECIMAL(15,4) NOT NULL,
    price DECIMAL(15,4) NOT NULL,
    fee DECIMAL(15,4) DEFAULT 0,
    transaction_date DATE NOT NULL,
    notes TEXT DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_txn_portfolio ON transactions(portfolio_id, transaction_date DESC);

CREATE TABLE fee_rebates (
    id SERIAL PRIMARY KEY,
    portfolio_id INT NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
    amount DECIMAL(15,4) NOT NULL,
    rebate_date DATE NOT NULL,
    broker VARCHAR(100) DEFAULT '',
    notes TEXT DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_rebates_portfolio ON fee_rebates(portfolio_id);

CREATE TABLE dividends (
    id SERIAL PRIMARY KEY,
    portfolio_id INT NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
    ticker VARCHAR(20) NOT NULL,
    amount DECIMAL(15,4) NOT NULL,
    dividend_date DATE NOT NULL,
    notes TEXT DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_dividends_portfolio ON dividends(portfolio_id);
