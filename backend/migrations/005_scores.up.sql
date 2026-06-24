CREATE TABLE stock_scores (
    id SERIAL PRIMARY KEY,
    ticker VARCHAR(20) NOT NULL REFERENCES stocks(ticker) ON DELETE CASCADE,
    date DATE NOT NULL,
    total_score DECIMAL(5,2) DEFAULT 0,
    fundamental_score DECIMAL(5,2) DEFAULT 0,
    chip_score DECIMAL(5,2) DEFAULT 0,
    momentum_score DECIMAL(5,2) DEFAULT 0,
    theme_score DECIMAL(5,2) DEFAULT 0,
    risk_score DECIMAL(5,2) DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(ticker, date)
);
CREATE INDEX idx_scores_date_total ON stock_scores(date DESC, total_score DESC);
CREATE INDEX idx_scores_ticker ON stock_scores(ticker, date DESC);

CREATE TABLE score_breakdown (
    id SERIAL PRIMARY KEY,
    score_id INT NOT NULL REFERENCES stock_scores(id) ON DELETE CASCADE,
    indicator_name VARCHAR(100) NOT NULL,
    indicator_value DECIMAL(15,6) DEFAULT 0,
    indicator_score DECIMAL(5,2) DEFAULT 0,
    weight DECIMAL(5,4) DEFAULT 0
);
CREATE INDEX idx_breakdown_score_id ON score_breakdown(score_id);

CREATE TABLE reports (
    id SERIAL PRIMARY KEY,
    ticker VARCHAR(20) NOT NULL REFERENCES stocks(ticker) ON DELETE CASCADE,
    date DATE NOT NULL,
    market VARCHAR(10) NOT NULL DEFAULT '',
    file_path VARCHAR(500) DEFAULT '',
    summary TEXT DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(ticker, date)
);
