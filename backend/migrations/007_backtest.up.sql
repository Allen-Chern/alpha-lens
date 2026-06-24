CREATE TABLE backtest_runs (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    market VARCHAR(10) NOT NULL DEFAULT 'TW',
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    initial_capital DECIMAL(15,4) NOT NULL DEFAULT 1000000,
    parameters JSONB DEFAULT '{}',
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    error_message TEXT DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);
CREATE INDEX idx_backtest_runs_status ON backtest_runs(status, created_at DESC);

CREATE TABLE backtest_results (
    id SERIAL PRIMARY KEY,
    backtest_id INT NOT NULL REFERENCES backtest_runs(id) ON DELETE CASCADE,
    total_return DECIMAL(10,4) DEFAULT 0,
    max_drawdown DECIMAL(10,4) DEFAULT 0,
    sharpe_ratio DECIMAL(10,4) DEFAULT 0,
    win_rate DECIMAL(10,4) DEFAULT 0,
    total_trades INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE backtest_trades (
    id SERIAL PRIMARY KEY,
    backtest_id INT NOT NULL REFERENCES backtest_runs(id) ON DELETE CASCADE,
    ticker VARCHAR(20) NOT NULL,
    action VARCHAR(10) NOT NULL,
    price DECIMAL(15,4) NOT NULL DEFAULT 0,
    shares DECIMAL(15,4) NOT NULL DEFAULT 0,
    trade_date DATE NOT NULL,
    pnl DECIMAL(15,4) DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_bt_trades_backtest ON backtest_trades(backtest_id, trade_date);
