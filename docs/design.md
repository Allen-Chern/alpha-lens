# AlphaLens — 設計文件

> 版本：1.0  
> 日期：2026-06-02

---

## 產品定位

個人使用的投資管理平台，模擬華爾街法人分析框架，協助在台股（主）與美股（輔）找到好標的、管理持倉、並提供調整建議。

---

## 理解摘要

- **使用者**：個人散戶，需要機構級分析深度，不做當沖
- **市場**：台股為主（上市 + 上櫃），美股 S&P 500 為輔
- **策略**：核心長線持倉 + 部分波段，不預設比重
- **更新頻率**：每日盤後，次日早上查看報告
- **語言**：繁體中文

---

## 核心功能

| 功能 | 說明 |
|------|------|
| 主動篩選 | 每日盤後自動評分，輸出值得關注標的清單 |
| 單股查詢 | 輸入代碼，產生完整法人式研究報告 |
| 持倉追蹤 | 手動輸入交易、退傭、現金股利，計算各期損益 |
| 投資建議 | 根據評分變化，提供加碼 / 減碼 / 出場參考 |
| 回測系統 | 個股歷史驗證、策略回測、組合績效分析 |

---

## 評分框架（固定權重）

| 面向 | 權重 | 關鍵指標 |
|------|------|----------|
| 基本面品質 | 30% | 月營收年增率趨勢、EPS 年增率、ROE、毛利率、FCF |
| 籌碼動能 | 25% | 外資/投信買超、融資融券、董監申報、法人均成本帶 |
| 價格動能 | 20% | 相對大盤強度（3M/1M）、均線排列、成交量趨勢 |
| 題材性 | 15% | 主題標籤、新聞情緒分數、產業景氣週期 |
| 風險指標 | 10% | 流動性、Beta、財務槓桿 |

### 基本面子指標（30%）

| 子指標 | 權重 | 說明 |
|--------|------|------|
| 月營收年增率趨勢 | 30% | 連三個月加速成長為強訊號 |
| EPS 年增率 | 25% | 方向比絕對值重要 |
| ROE（近四季平均） | 20% | >15% 且穩定為高分 |
| 毛利率趨勢 | 15% | 擴張代表定價能力改善 |
| 自由現金流（FCF） | 10% | 正值且成長 |

### 籌碼動能子指標（25%）

| 子指標 | 權重 | 說明 |
|--------|------|------|
| 外資近 10 日買超強度 | 25% | 張數 + 連續天數加權 |
| 投信近 10 日買超強度 | 25% | 台股投信動向往往領先股價 |
| 外資＋投信同步買超 | 20% | 同步時給予乘數加分 |
| 融資比率健康度 | 15% | 低融資 = 籌碼乾淨，反向加分 |
| 法人均成本 vs 現價 | 15% | 現價高於均成本 10%+ 為健康 |

### 價格動能子指標（20%）

| 子指標 | 權重 | 說明 |
|--------|------|------|
| 近 3 個月相對大盤強度 | 40% | 核心動能指標 |
| 近 1 個月相對強度 | 25% | 短線動能確認 |
| 均線多頭排列 | 20% | 20MA > 60MA > 120MA |
| 成交量趨勢 | 15% | 量增價漲優於量縮上漲 |

### 題材性子指標（15%）

| 子指標 | 權重 | 說明 |
|--------|------|------|
| 主題標籤匹配度 | 40% | AI、CoWoS、HBM 等當前主流題材 |
| 近期新聞情緒分數 | 35% | LLM 分析近 7 日新聞，-1 到 +1 |
| 產業景氣週期位置 | 25% | 擴張期高分，收縮期扣分 |

### 風險指標子指標（10%）

| 子指標 | 權重 | 說明 |
|--------|------|------|
| 流動性分數 | 40% | 日均量越高分越高 |
| Beta 值 | 30% | 中 Beta（0.8–1.3）得分最高 |
| 財務槓桿 | 30% | 負債比過高扣分 |

---

## 標的池

| 市場 | 篩選條件 | 預估數量 |
|------|----------|----------|
| 台股上市 | 市值 > 50億、均量 > 500萬、排除金融/全額交割/處置股 | 約 400–500 支 |
| 台股上櫃 | 市值 > 20億、均量 > 200萬 | 約 100–150 支 |
| 美股 | S&P 500 成分股 | 500 支 |

---

## 資料來源

| 來源 | 用途 | 費用 |
|------|------|------|
| FinMind 免費層 | 台股股價、財報、法人籌碼、融資券 | $0 |
| MOPS 公開資訊觀測站 | 台股董監持股申報（爬蟲） | $0 |
| yfinance | 美股股價、財報 | $0 |
| SEC EDGAR API | 美股 Form 4 內部人申報 | $0 |
| Capitol Trades | 美國國會議員交易申報 | $0 |
| Ollama + Qwen 2.5 14B | 報告文字產生（測試階段） | $0 |
| Claude API | 報告文字產生（品質升級後） | 按用量 |

---

## 每日管線

### 台股管線（每日 18:00 CST，週一至週五）
### 美股管線（每日 06:00 CST，週二至週六）

```
Stage 1｜資料擷取
  ├── 更新標的池
  ├── 抓取當日 OHLCV 股價
  ├── 抓取法人籌碼
  ├── 抓取財報更新（季報）
  └── 抓取董監申報

Stage 2｜評分計算
  ├── 計算五大面向子分數
  ├── 計算加權綜合分
  └── 寫入 stock_scores + score_breakdown

Stage 3｜報告產生（選擇性）
  ├── 綜合分前 30 名
  ├── 持倉中的股票
  └── 使用者觀察清單

Stage 4｜管線狀態記錄
  └── 寫入 pipeline_runs
```

---

## 資料庫 Schema

### 市場資料群組

```sql
stocks              -- 標的池主檔
stock_prices        -- 每日 OHLCV
fundamentals        -- 季報/年報財務指標
institutional_flows -- 每日法人籌碼
insider_transactions-- 董監/內部人申報
congressional_trades-- 美國國會議員申報
themes              -- 產業題材主題
stock_themes        -- 股票與題材多對多
```

### 評分系統群組

```sql
stock_scores     -- 每日五大面向分數 + 綜合分
score_breakdown  -- 子指標明細
reports          -- 報告 metadata（內容存為 Markdown 檔案）
```

報告檔案路徑規則：`data/reports/{market}/{ticker}/{date}.md`

### 持倉管理群組

```sql
portfolios    -- 投資組合主檔
transactions  -- 買賣交易紀錄
fee_rebates   -- 券商退傭紀錄
dividends     -- 現金股利紀錄
```

### 回測群組

```sql
backtest_runs    -- 回測設定（JSONB 參數）
backtest_results -- 結果摘要（總報酬、最大回撤、Sharpe Ratio）
backtest_trades  -- 虛擬交易明細
```

---

## API 端點

### 標的與篩選
```
GET  /api/stocks
GET  /api/stocks/:ticker
GET  /api/stocks/:ticker/score
GET  /api/stocks/:ticker/report
POST /api/stocks/:ticker/report
GET  /api/stocks/:ticker/history
GET  /api/screener
```

### 持倉管理
```
GET  /api/portfolio
POST /api/portfolio/transactions
GET  /api/portfolio/transactions
POST /api/portfolio/rebates
POST /api/portfolio/dividends
GET  /api/portfolio/pnl
```

### 回測
```
POST /api/backtest
GET  /api/backtest/:id
GET  /api/backtest/:id/trades
GET  /api/backtest
```

### 系統管理
```
GET  /api/pipeline/status
POST /api/pipeline/run
GET  /api/settings
PUT  /api/settings
```

---

## 技術架構

| 項目 | 決定 |
|------|------|
| 後端語言 | Go |
| 排程 | robfig/cron（內建） |
| 資料庫 | PostgreSQL |
| LLM（測試） | Ollama + Qwen 2.5 14B |
| LLM（正式） | Claude API（可配置切換） |
| 前端 | React + Vite |
| UI 元件 | shadcn/ui + Tailwind CSS |
| 部署 | Docker Compose，全本機 |

### 專案結構

```
alpha-lens/
├── backend/
│   ├── cmd/
│   ├── internal/
│   │   ├── ingestion/
│   │   ├── scoring/
│   │   ├── reports/
│   │   ├── portfolio/
│   │   ├── backtest/
│   │   └── api/
│   ├── migrations/
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   ├── pages/
│   │   └── lib/
│   └── Dockerfile
├── data/
│   └── reports/
├── docker-compose.yml
└── docs/
    └── design.md
```

---

## 視覺風格

- 預設深色模式（`#18181B` 背景），支援深淺切換
- 強調色：深色 `#A78BFA`，淺色 `#6366F1`
- 漲：`#4ADE80` / 跌：`#F87171`
- 資訊密度：平衡型（重要資訊優先展示）

---

## 假設

1. FinMind 600 次/天免費層，透過批次策略可覆蓋全標的池
2. MOPS 爬蟲為非官方方式，格式異動時需維護
3. 16GB RAM 在 Qwen 2.5 14B 運行期間，其他服務記憶體需控管
4. Capitol Trades 資料延遲 30–45 天，與申報法規一致

## 明確排除

- 當沖 / 盤中即時資料
- 金融股納入評分池
- 股票股利換算股數

---

## Decision Log

| # | 決策 | 替代方案 | 選擇原因 |
|---|------|----------|----------|
| 1 | 前後端分離 | 模組化單體 | 使用者熟悉 React，UI 彈性較高 |
| 2 | React + Vite | Next.js | 純 SPA 不需 SSR，Vite 更輕量 |
| 3 | Go 後端含排程 | GitHub Actions | 全部 Docker 化，不依賴外部系統 |
| 4 | PostgreSQL | SQLite | 複雜查詢、回測計算、時序資料更穩健 |
| 5 | Ollama / Qwen 2.5 14B | Claude API | 測試階段零成本，保留升級路徑 |
| 6 | 報告存為 Markdown 檔案 | 存入 DB TEXT | DB 輕量，報告可直接閱讀，長期不膨脹 |
| 7 | 固定評分權重 | 可自訂權重 | 模擬法人標準，降低使用複雜度 |
| 8 | 深色現代 + 切換 | 單一淺色 | 使用者偏好深色，保留彈性 |
| 9 | 標的池約 1,100 支 | 全市場 | 符合 FinMind 免費層限制，覆蓋主流標的 |
| 10 | 股利記現金等值 | 換算股數 | 設計簡潔，符合使用者需求 |
