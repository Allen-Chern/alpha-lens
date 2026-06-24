CREATE TABLE themes (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE stock_themes (
    stock_id INT NOT NULL REFERENCES stocks(id) ON DELETE CASCADE,
    theme_id INT NOT NULL REFERENCES themes(id) ON DELETE CASCADE,
    PRIMARY KEY (stock_id, theme_id)
);
