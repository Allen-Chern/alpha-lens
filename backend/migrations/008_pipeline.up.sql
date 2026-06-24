CREATE TABLE pipeline_runs (
    id SERIAL PRIMARY KEY,
    pipeline_type VARCHAR(50) NOT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    status VARCHAR(20) NOT NULL DEFAULT 'RUNNING',
    error_message TEXT DEFAULT '',
    stocks_processed INT DEFAULT 0
);
CREATE INDEX idx_pipeline_type_date ON pipeline_runs(pipeline_type, started_at DESC);
