-- Stocks (5 TW, 3 US)
INSERT INTO stocks (ticker, name, market, exchange, market_cap, avg_volume) VALUES
('2330', '台積電', 'TW', 'TWSE', 18500000000000, 25000000),
('2317', '鴻海', 'TW', 'TWSE', 1500000000000, 40000000),
('2454', '聯發科', 'TW', 'TWSE', 2100000000000, 8000000),
('2308', '台達電', 'TW', 'TWSE', 650000000000, 5000000),
('2382', '廣達', 'TW', 'TWSE', 730000000000, 9000000),
('AAPL', 'Apple Inc.', 'US', 'NASDAQ', 3200000000000, 55000000),
('NVDA', 'NVIDIA Corporation', 'US', 'NASDAQ', 3400000000000, 200000000),
('MSFT', 'Microsoft Corporation', 'US', 'NASDAQ', 3100000000000, 22000000);

-- Themes
INSERT INTO themes (name, description) VALUES
('AI', '人工智慧相關'),
('CoWoS', '先進封裝 CoWoS 供應鏈'),
('HBM', '高頻寬記憶體'),
('半導體', '半導體製造與設計'),
('伺服器', 'AI 伺服器及資料中心'),
('電動車', '電動車及零組件');

-- Stock <-> Theme links
INSERT INTO stock_themes (stock_id, theme_id)
SELECT s.id, t.id FROM stocks s, themes t
WHERE s.ticker = '2330' AND t.name IN ('AI', 'CoWoS', 'HBM', '半導體');

INSERT INTO stock_themes (stock_id, theme_id)
SELECT s.id, t.id FROM stocks s, themes t
WHERE s.ticker = '2454' AND t.name IN ('AI', '半導體');

INSERT INTO stock_themes (stock_id, theme_id)
SELECT s.id, t.id FROM stocks s, themes t
WHERE s.ticker = '2317' AND t.name IN ('伺服器');

INSERT INTO stock_themes (stock_id, theme_id)
SELECT s.id, t.id FROM stocks s, themes t
WHERE s.ticker = '2382' AND t.name IN ('AI', '伺服器');

INSERT INTO stock_themes (stock_id, theme_id)
SELECT s.id, t.id FROM stocks s, themes t
WHERE s.ticker = 'NVDA' AND t.name IN ('AI', 'HBM', '半導體');

INSERT INTO stock_themes (stock_id, theme_id)
SELECT s.id, t.id FROM stocks s, themes t
WHERE s.ticker = 'AAPL' AND t.name IN ('AI');

INSERT INTO stock_themes (stock_id, theme_id)
SELECT s.id, t.id FROM stocks s, themes t
WHERE s.ticker = 'MSFT' AND t.name IN ('AI', '伺服器');

-- Stock prices: 3 days x 8 stocks (adj_close = close)
INSERT INTO stock_prices (ticker, date, open, high, low, close, volume, adj_close) VALUES
('2330', '2026-06-17', 995, 1010, 988, 1005, 22000000, 1005),
('2330', '2026-06-18', 1005, 1015, 998, 1008, 19000000, 1008),
('2330', '2026-06-19', 1000, 1022, 992, 1015, 26000000, 1015),
('2317', '2026-06-17', 192, 197, 191, 195, 38000000, 195),
('2317', '2026-06-18', 195, 199, 193, 197, 35000000, 197),
('2317', '2026-06-19', 197, 201, 194, 199, 42000000, 199),
('2454', '2026-06-17', 1185, 1205, 1178, 1198, 7500000, 1198),
('2454', '2026-06-18', 1198, 1215, 1190, 1210, 8200000, 1210),
('2454', '2026-06-19', 1210, 1228, 1203, 1222, 9100000, 1222),
('2308', '2026-06-17', 415, 422, 412, 418, 4800000, 418),
('2308', '2026-06-18', 418, 425, 415, 422, 5100000, 422),
('2308', '2026-06-19', 422, 430, 419, 427, 5500000, 427),
('2382', '2026-06-17', 375, 382, 372, 378, 8500000, 378),
('2382', '2026-06-18', 378, 386, 375, 383, 9200000, 383),
('2382', '2026-06-19', 383, 391, 380, 388, 10000000, 388),
('AAPL', '2026-06-17', 208, 212, 206, 210, 50000000, 210),
('AAPL', '2026-06-18', 210, 214, 209, 212, 51000000, 212),
('AAPL', '2026-06-19', 210, 215, 208, 213, 52000000, 213),
('NVDA', '2026-06-17', 130, 136, 129, 134, 190000000, 134),
('NVDA', '2026-06-18', 134, 138, 132, 136, 192000000, 136),
('NVDA', '2026-06-19', 133, 141, 131, 138, 195000000, 138),
('MSFT', '2026-06-17', 410, 418, 408, 414, 20000000, 414),
('MSFT', '2026-06-18', 414, 420, 412, 417, 20500000, 417),
('MSFT', '2026-06-19', 415, 423, 412, 419, 21000000, 419);

-- Fundamentals (latest quarter per stock)
INSERT INTO fundamentals (ticker, period_type, period, revenue, revenue_yoy, eps, eps_yoy, roe, gross_margin, fcf, debt_ratio, reported_at) VALUES
('2330', 'Q', '2026Q1', 800000000000, 0.42, 13.5, 0.38, 28.5, 0.55, 350000000000, 0.28, '2026-04-15'),
('2317', 'Q', '2026Q1', 1600000000000, 0.12, 2.8, 0.10, 11.5, 0.07, 80000000000, 0.55, '2026-04-20'),
('2454', 'Q', '2026Q1', 140000000000, 0.30, 18.2, 0.25, 22.0, 0.48, 40000000000, 0.20, '2026-04-18'),
('2308', 'Q', '2026Q1', 110000000000, 0.18, 5.1, 0.15, 19.0, 0.30, 18000000000, 0.35, '2026-04-22'),
('2382', 'Q', '2026Q1', 380000000000, 0.35, 4.5, 0.40, 18.5, 0.09, 12000000000, 0.40, '2026-04-25'),
('AAPL', 'Q', '2026Q1', 95000000000, 0.06, 1.65, 0.08, 145.0, 0.46, 25000000000, 0.65, '2026-05-01'),
('NVDA', 'Q', '2026Q1', 60000000000, 0.78, 0.95, 0.85, 95.0, 0.75, 28000000000, 0.20, '2026-05-22'),
('MSFT', 'Q', '2026Q1', 65000000000, 0.16, 3.10, 0.18, 38.0, 0.69, 22000000000, 0.30, '2026-04-28');

-- Institutional flows (latest date per TW stock)
INSERT INTO institutional_flows (ticker, date, foreign_buy, foreign_sell, trust_buy, trust_sell, margin_balance, short_balance) VALUES
('2330', '2026-06-19', 60000, 15000, 18000, 6000, 120000, 12000),
('2317', '2026-06-19', 45000, 30000, 8000, 9000, 200000, 25000),
('2454', '2026-06-19', 22000, 12000, 9000, 4000, 60000, 5000),
('2308', '2026-06-19', 15000, 9000, 5000, 3000, 40000, 4000),
('2382', '2026-06-19', 28000, 14000, 7000, 3500, 70000, 8000);

-- Stock scores for 2026-06-19
INSERT INTO stock_scores (ticker, date, total_score, fundamental_score, chip_score, momentum_score, theme_score, risk_score) VALUES
('2330', '2026-06-19', 88.5, 90.0, 85.5, 92.0, 88.0, 85.5),
('2317', '2026-06-19', 72.0, 68.5, 75.0, 70.0, 72.0, 78.5),
('2454', '2026-06-19', 83.0, 85.0, 80.5, 85.5, 80.0, 78.0),
('2308', '2026-06-19', 75.5, 78.0, 72.5, 74.0, 68.5, 82.0),
('2382', '2026-06-19', 79.5, 80.0, 78.5, 82.0, 78.0, 76.5),
('AAPL', '2026-06-19', 71.0, 75.0, 68.0, 70.5, 65.0, 80.5),
('NVDA', '2026-06-19', 91.5, 88.5, 90.0, 95.5, 95.0, 86.0),
('MSFT', '2026-06-19', 78.5, 82.5, 75.0, 76.0, 72.5, 82.0);

-- Score breakdown: 19 indicators per stock
WITH s AS (SELECT id FROM stock_scores WHERE ticker='2330' AND date='2026-06-19')
INSERT INTO score_breakdown (score_id, indicator_name, indicator_value, indicator_score, weight) VALUES
((SELECT id FROM s), 'revenue_yoy_trend', 0.42, 95.0, 0.09),
((SELECT id FROM s), 'eps_yoy', 0.38, 90.0, 0.075),
((SELECT id FROM s), 'roe', 28.5, 92.0, 0.06),
((SELECT id FROM s), 'gross_margin_trend', 0.55, 88.0, 0.045),
((SELECT id FROM s), 'fcf_growth', 1, 85.0, 0.03),
((SELECT id FROM s), 'foreign_net_buy_10d', 45000, 88.0, 0.0625),
((SELECT id FROM s), 'trust_net_buy_10d', 12000, 82.0, 0.0625),
((SELECT id FROM s), 'foreign_trust_sync', 1, 90.0, 0.05),
((SELECT id FROM s), 'margin_health', 0.12, 85.0, 0.0375),
((SELECT id FROM s), 'inst_cost_vs_price', 0.15, 88.0, 0.0375),
((SELECT id FROM s), 'momentum_3m', 0.22, 92.0, 0.08),
((SELECT id FROM s), 'momentum_1m', 0.08, 88.0, 0.05),
((SELECT id FROM s), 'ma_alignment', 1, 95.0, 0.04),
((SELECT id FROM s), 'volume_trend', 0.18, 85.0, 0.03),
((SELECT id FROM s), 'theme_score', 0.88, 88.0, 0.15),
((SELECT id FROM s), 'news_sentiment', 0.72, 86.0, 0.0),
((SELECT id FROM s), 'liquidity_score', 95.0, 95.0, 0.04),
((SELECT id FROM s), 'beta', 1.05, 88.0, 0.03),
((SELECT id FROM s), 'financial_leverage', 0.28, 85.0, 0.03);

WITH s AS (SELECT id FROM stock_scores WHERE ticker='2317' AND date='2026-06-19')
INSERT INTO score_breakdown (score_id, indicator_name, indicator_value, indicator_score, weight) VALUES
((SELECT id FROM s), 'revenue_yoy_trend', 0.12, 65.0, 0.09),
((SELECT id FROM s), 'eps_yoy', 0.10, 62.0, 0.075),
((SELECT id FROM s), 'roe', 11.5, 60.0, 0.06),
((SELECT id FROM s), 'gross_margin_trend', 0.07, 58.0, 0.045),
((SELECT id FROM s), 'fcf_growth', 1, 70.0, 0.03),
((SELECT id FROM s), 'foreign_net_buy_10d', 15000, 75.0, 0.0625),
((SELECT id FROM s), 'trust_net_buy_10d', -1000, 70.0, 0.0625),
((SELECT id FROM s), 'foreign_trust_sync', 0, 72.0, 0.05),
((SELECT id FROM s), 'margin_health', 0.18, 76.0, 0.0375),
((SELECT id FROM s), 'inst_cost_vs_price', 0.05, 74.0, 0.0375),
((SELECT id FROM s), 'momentum_3m', 0.10, 70.0, 0.08),
((SELECT id FROM s), 'momentum_1m', 0.03, 68.0, 0.05),
((SELECT id FROM s), 'ma_alignment', 1, 72.0, 0.04),
((SELECT id FROM s), 'volume_trend', 0.08, 70.0, 0.03),
((SELECT id FROM s), 'theme_score', 0.72, 72.0, 0.15),
((SELECT id FROM s), 'news_sentiment', 0.60, 70.0, 0.0),
((SELECT id FROM s), 'liquidity_score', 90.0, 90.0, 0.04),
((SELECT id FROM s), 'beta', 1.15, 78.0, 0.03),
((SELECT id FROM s), 'financial_leverage', 0.55, 78.0, 0.03);

WITH s AS (SELECT id FROM stock_scores WHERE ticker='2454' AND date='2026-06-19')
INSERT INTO score_breakdown (score_id, indicator_name, indicator_value, indicator_score, weight) VALUES
((SELECT id FROM s), 'revenue_yoy_trend', 0.30, 87.0, 0.09),
((SELECT id FROM s), 'eps_yoy', 0.25, 84.0, 0.075),
((SELECT id FROM s), 'roe', 22.0, 85.0, 0.06),
((SELECT id FROM s), 'gross_margin_trend', 0.48, 82.0, 0.045),
((SELECT id FROM s), 'fcf_growth', 1, 80.0, 0.03),
((SELECT id FROM s), 'foreign_net_buy_10d', 18000, 82.0, 0.0625),
((SELECT id FROM s), 'trust_net_buy_10d', 9000, 80.0, 0.0625),
((SELECT id FROM s), 'foreign_trust_sync', 1, 84.0, 0.05),
((SELECT id FROM s), 'margin_health', 0.10, 80.0, 0.0375),
((SELECT id FROM s), 'inst_cost_vs_price', 0.12, 80.0, 0.0375),
((SELECT id FROM s), 'momentum_3m', 0.18, 86.0, 0.08),
((SELECT id FROM s), 'momentum_1m', 0.06, 84.0, 0.05),
((SELECT id FROM s), 'ma_alignment', 1, 88.0, 0.04),
((SELECT id FROM s), 'volume_trend', 0.14, 82.0, 0.03),
((SELECT id FROM s), 'theme_score', 0.80, 80.0, 0.15),
((SELECT id FROM s), 'news_sentiment', 0.68, 80.0, 0.0),
((SELECT id FROM s), 'liquidity_score', 85.0, 85.0, 0.04),
((SELECT id FROM s), 'beta', 1.10, 80.0, 0.03),
((SELECT id FROM s), 'financial_leverage', 0.20, 82.0, 0.03);

WITH s AS (SELECT id FROM stock_scores WHERE ticker='2308' AND date='2026-06-19')
INSERT INTO score_breakdown (score_id, indicator_name, indicator_value, indicator_score, weight) VALUES
((SELECT id FROM s), 'revenue_yoy_trend', 0.18, 78.0, 0.09),
((SELECT id FROM s), 'eps_yoy', 0.15, 76.0, 0.075),
((SELECT id FROM s), 'roe', 19.0, 78.0, 0.06),
((SELECT id FROM s), 'gross_margin_trend', 0.30, 75.0, 0.045),
((SELECT id FROM s), 'fcf_growth', 1, 76.0, 0.03),
((SELECT id FROM s), 'foreign_net_buy_10d', 9000, 73.0, 0.0625),
((SELECT id FROM s), 'trust_net_buy_10d', 4000, 72.0, 0.0625),
((SELECT id FROM s), 'foreign_trust_sync', 1, 74.0, 0.05),
((SELECT id FROM s), 'margin_health', 0.14, 73.0, 0.0375),
((SELECT id FROM s), 'inst_cost_vs_price', 0.08, 72.0, 0.0375),
((SELECT id FROM s), 'momentum_3m', 0.12, 74.0, 0.08),
((SELECT id FROM s), 'momentum_1m', 0.04, 73.0, 0.05),
((SELECT id FROM s), 'ma_alignment', 1, 76.0, 0.04),
((SELECT id FROM s), 'volume_trend', 0.10, 72.0, 0.03),
((SELECT id FROM s), 'theme_score', 0.685, 68.5, 0.15),
((SELECT id FROM s), 'news_sentiment', 0.62, 72.0, 0.0),
((SELECT id FROM s), 'liquidity_score', 80.0, 80.0, 0.04),
((SELECT id FROM s), 'beta', 1.08, 82.0, 0.03),
((SELECT id FROM s), 'financial_leverage', 0.35, 82.0, 0.03);

WITH s AS (SELECT id FROM stock_scores WHERE ticker='2382' AND date='2026-06-19')
INSERT INTO score_breakdown (score_id, indicator_name, indicator_value, indicator_score, weight) VALUES
((SELECT id FROM s), 'revenue_yoy_trend', 0.35, 82.0, 0.09),
((SELECT id FROM s), 'eps_yoy', 0.40, 84.0, 0.075),
((SELECT id FROM s), 'roe', 18.5, 78.0, 0.06),
((SELECT id FROM s), 'gross_margin_trend', 0.09, 76.0, 0.045),
((SELECT id FROM s), 'fcf_growth', 1, 78.0, 0.03),
((SELECT id FROM s), 'foreign_net_buy_10d', 14000, 79.0, 0.0625),
((SELECT id FROM s), 'trust_net_buy_10d', 6000, 78.0, 0.0625),
((SELECT id FROM s), 'foreign_trust_sync', 1, 80.0, 0.05),
((SELECT id FROM s), 'margin_health', 0.13, 78.0, 0.0375),
((SELECT id FROM s), 'inst_cost_vs_price', 0.11, 78.0, 0.0375),
((SELECT id FROM s), 'momentum_3m', 0.16, 82.0, 0.08),
((SELECT id FROM s), 'momentum_1m', 0.05, 80.0, 0.05),
((SELECT id FROM s), 'ma_alignment', 1, 84.0, 0.04),
((SELECT id FROM s), 'volume_trend', 0.13, 80.0, 0.03),
((SELECT id FROM s), 'theme_score', 0.78, 78.0, 0.15),
((SELECT id FROM s), 'news_sentiment', 0.66, 78.0, 0.0),
((SELECT id FROM s), 'liquidity_score', 82.0, 82.0, 0.04),
((SELECT id FROM s), 'beta', 1.12, 76.0, 0.03),
((SELECT id FROM s), 'financial_leverage', 0.40, 76.0, 0.03);

WITH s AS (SELECT id FROM stock_scores WHERE ticker='AAPL' AND date='2026-06-19')
INSERT INTO score_breakdown (score_id, indicator_name, indicator_value, indicator_score, weight) VALUES
((SELECT id FROM s), 'revenue_yoy_trend', 0.06, 72.0, 0.09),
((SELECT id FROM s), 'eps_yoy', 0.08, 74.0, 0.075),
((SELECT id FROM s), 'roe', 145.0, 80.0, 0.06),
((SELECT id FROM s), 'gross_margin_trend', 0.46, 78.0, 0.045),
((SELECT id FROM s), 'fcf_growth', 1, 76.0, 0.03),
((SELECT id FROM s), 'insider_net_buy', -5000, 66.0, 0.0625),
((SELECT id FROM s), 'congress_net_buy', 1, 70.0, 0.0625),
((SELECT id FROM s), 'inst_holding_change', 0.02, 68.0, 0.05),
((SELECT id FROM s), 'short_interest', 0.012, 70.0, 0.0375),
((SELECT id FROM s), 'inst_cost_vs_price', 0.10, 68.0, 0.0375),
((SELECT id FROM s), 'momentum_3m', 0.08, 70.0, 0.08),
((SELECT id FROM s), 'momentum_1m', 0.03, 71.0, 0.05),
((SELECT id FROM s), 'ma_alignment', 1, 72.0, 0.04),
((SELECT id FROM s), 'volume_trend', 0.06, 70.0, 0.03),
((SELECT id FROM s), 'theme_score', 0.65, 65.0, 0.15),
((SELECT id FROM s), 'news_sentiment', 0.58, 70.0, 0.0),
((SELECT id FROM s), 'liquidity_score', 95.0, 95.0, 0.04),
((SELECT id FROM s), 'beta', 1.20, 80.0, 0.03),
((SELECT id FROM s), 'financial_leverage', 0.65, 80.0, 0.03);

WITH s AS (SELECT id FROM stock_scores WHERE ticker='NVDA' AND date='2026-06-19')
INSERT INTO score_breakdown (score_id, indicator_name, indicator_value, indicator_score, weight) VALUES
((SELECT id FROM s), 'revenue_yoy_trend', 0.78, 98.0, 0.09),
((SELECT id FROM s), 'eps_yoy', 0.85, 96.0, 0.075),
((SELECT id FROM s), 'roe', 95.0, 92.0, 0.06),
((SELECT id FROM s), 'gross_margin_trend', 0.75, 90.0, 0.045),
((SELECT id FROM s), 'fcf_growth', 1, 88.0, 0.03),
((SELECT id FROM s), 'insider_net_buy', 8000, 88.0, 0.0625),
((SELECT id FROM s), 'congress_net_buy', 1, 90.0, 0.0625),
((SELECT id FROM s), 'inst_holding_change', 0.05, 92.0, 0.05),
((SELECT id FROM s), 'short_interest', 0.015, 88.0, 0.0375),
((SELECT id FROM s), 'inst_cost_vs_price', 0.20, 90.0, 0.0375),
((SELECT id FROM s), 'momentum_3m', 0.35, 96.0, 0.08),
((SELECT id FROM s), 'momentum_1m', 0.12, 95.0, 0.05),
((SELECT id FROM s), 'ma_alignment', 1, 98.0, 0.04),
((SELECT id FROM s), 'volume_trend', 0.25, 92.0, 0.03),
((SELECT id FROM s), 'theme_score', 0.95, 95.0, 0.15),
((SELECT id FROM s), 'news_sentiment', 0.85, 92.0, 0.0),
((SELECT id FROM s), 'liquidity_score', 98.0, 98.0, 0.04),
((SELECT id FROM s), 'beta', 1.45, 86.0, 0.03),
((SELECT id FROM s), 'financial_leverage', 0.20, 86.0, 0.03);

WITH s AS (SELECT id FROM stock_scores WHERE ticker='MSFT' AND date='2026-06-19')
INSERT INTO score_breakdown (score_id, indicator_name, indicator_value, indicator_score, weight) VALUES
((SELECT id FROM s), 'revenue_yoy_trend', 0.16, 82.0, 0.09),
((SELECT id FROM s), 'eps_yoy', 0.18, 84.0, 0.075),
((SELECT id FROM s), 'roe', 38.0, 85.0, 0.06),
((SELECT id FROM s), 'gross_margin_trend', 0.69, 84.0, 0.045),
((SELECT id FROM s), 'fcf_growth', 1, 82.0, 0.03),
((SELECT id FROM s), 'insider_net_buy', 3000, 74.0, 0.0625),
((SELECT id FROM s), 'congress_net_buy', 1, 76.0, 0.0625),
((SELECT id FROM s), 'inst_holding_change', 0.03, 75.0, 0.05),
((SELECT id FROM s), 'short_interest', 0.010, 76.0, 0.0375),
((SELECT id FROM s), 'inst_cost_vs_price', 0.13, 75.0, 0.0375),
((SELECT id FROM s), 'momentum_3m', 0.12, 76.0, 0.08),
((SELECT id FROM s), 'momentum_1m', 0.04, 75.0, 0.05),
((SELECT id FROM s), 'ma_alignment', 1, 78.0, 0.04),
((SELECT id FROM s), 'volume_trend', 0.09, 74.0, 0.03),
((SELECT id FROM s), 'theme_score', 0.725, 72.5, 0.15),
((SELECT id FROM s), 'news_sentiment', 0.70, 78.0, 0.0),
((SELECT id FROM s), 'liquidity_score', 92.0, 92.0, 0.04),
((SELECT id FROM s), 'beta', 0.95, 82.0, 0.03),
((SELECT id FROM s), 'financial_leverage', 0.30, 82.0, 0.03);

-- Portfolio
INSERT INTO portfolios (name, description) VALUES ('主要投資組合', '個人主要股票投資組合');
