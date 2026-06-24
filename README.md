# Alpha Lens

個人投資研究平台，整合台股與美股的多因子評分、播客洞察與投資組合管理。

## 功能

### 選股器
對追蹤中的股票進行多因子評分（0–100），維度包含：

| 維度 | 說明 |
|---|---|
| 基本面 | 營收/EPS YoY、ROE、毛利率、FCF |
| 籌碼 | 外資/投信淨買超、融資健康度、法人成本 |
| 動能 | 3M/1M 報酬、均線排列、量能趨勢 |
| 主題 | 股票與 AI、半導體、電動車等主題的關聯強度 |
| 風險 | Beta、財務槓桿、流動性 |

### 播客洞察
追蹤投資播客（RSS 或 YouTube 頻道），以 AI 自動萃取潛在投資標的：

1. **同步** — 抓取 RSS/Atom feed，取得最新集數
2. **抓稿** — 依優先序取得逐字稿：YouTube 自動字幕 → `podcast:transcript` URL → Whisper 語音辨識
3. **分析** — Claude API（Haiku 4.5）讀取完整逐字稿（無需分 chunk，支援至 200K tokens），萃取：
   - 提及的投資標的與代碼
   - 看多 / 看空 / 中性觀點
   - 投資論點摘要
   - 講者原話（verbatim quote）
4. **採納** — 標記認同的標的，日後納入 Theme 計分

前後端透過 WebSocket 即時推送抓稿與分析進度。

### 投資組合
記錄買賣交易、股利、手續費退佣，計算持倉市值與損益。

### 回測
對多因子評分設定參數，以歷史資料驗證選股策略績效。

### 資料 Pipeline
台股每日資料透過 [FinMind](https://finmindtrade.com/) API 擷取，包含股價、三大法人、融資融券。

---

## 技術架構

```
frontend/   React 18 + TypeScript + Vite + Tailwind CSS + React Query
backend/    Go 1.25 + chi + pgx + golang-migrate
            ├── internal/api/        HTTP & WebSocket handlers
            ├── internal/ingestion/  資料擷取 + LLM 分析
            └── migrations/          Postgres schema（13 個 migration）
data/       transcript 檔案（gzip 壓縮，volume 對應 /app/data）
```

### 關鍵技術決策
- **Transcript 存檔**：逐字稿存為 `data/transcripts/{podcast_id}/{episode_id}.txt.gz`，DB 只存相對路徑，避免大型 TEXT 欄位膨脹
- **WebSocket Hub**：gorilla/websocket，單一 Hub 廣播至所有連線的前端 tab
- **無 chunk 分析**：Claude 200K context 完整讀取逐字稿，保留跨段落的上下文，`original_quote` 精度更高

---

## 本機開發

### 前置需求
- [mise](https://mise.jdx.dev/)（管理 Go 版本）
- Node.js 20+
- Docker（跑 Postgres）

### 啟動步驟

```bash
# 1. 複製環境變數範本並填入
cp .env.example .env

# 2. 啟動 Postgres
docker compose -f docker-compose.dev.yml up -d

# 3. 啟動後端（自動執行 migration + seed）
cd backend
mise exec -- go run ./cmd/server

# 4. 啟動前端（另開終端機）
cd frontend
npm install
npm run dev
```

瀏覽器開啟 `http://localhost:5173`。

### 環境變數

| 變數 | 說明 | 預設值 |
|---|---|---|
| `POSTGRES_*` | DB 連線設定 | — |
| `LLM_PROVIDER` | `claude` 或 `ollama` | `ollama` |
| `CLAUDE_API_KEY` | Anthropic API Key（使用 Claude 時需設定）| — |
| `OLLAMA_BASE_URL` | Ollama 端點 | `http://localhost:11434` |
| `OLLAMA_MODEL` | 模型名稱 | `qwen2.5:14b` |
| `WHISPER_MODEL` | Whisper 模型（`small`/`medium`/`large-v3`），空字串停用 | — |
| `FINMIND_TOKEN` | FinMind API token | — |
| `DATA_DIR` | Transcript 等大型檔案的根目錄 | `/app/data` |

### Production 部署

```bash
docker compose up -d
```

後端 image 包含 `yt-dlp`（YouTube 字幕）與 Whisper Python 腳本；`./data` 目錄 mount 至 `/app/data`。

---

## API 端點（摘要）

```
GET  /api/screener                        多因子選股
GET  /api/stocks/:ticker                  個股資料
GET  /api/insights/podcasts               播客清單
POST /api/insights/podcasts/:id/sync      同步並分析播客
GET  /api/insights/podcasts/:id/episodes  集數清單與處理狀態
POST /api/insights/episodes/:id/fetch-transcript  抓逐字稿（非同步）
POST /api/insights/episodes/:id/analyze   LLM 分析（非同步）
GET  /api/insights/mentions               投資標的提及清單
GET  /api/ws                              WebSocket（進度推播）
```
