// Types
export interface Stock {
  ticker: string
  name: string
  market: string
  exchange: string
  market_cap: number
  avg_volume: number
  is_active: boolean
}

export interface StockPrice {
  date: string
  open: number
  high: number
  low: number
  close: number
  volume: number
}

export interface StockScore {
  date: string
  total_score: number
  fundamental_score: number
  chip_score: number
  momentum_score: number
  theme_score: number
  risk_score: number
}

export interface StockDetail extends Stock {
  latest_price: StockPrice | null
  latest_score: StockScore | null
}

export interface ScoreBreakdown {
  indicator_name: string
  indicator_value: number
  indicator_score: number
  weight: number
}

export interface StockScoreDetail extends StockScore {
  ticker: string
  breakdown: ScoreBreakdown[]
}

export interface ScreenerStock {
  ticker: string
  name: string
  market: string
  total_score: number
  fundamental_score: number
  chip_score: number
  momentum_score: number
  theme_score: number
  risk_score: number
  close: number
  change_pct: number
}

export interface Portfolio {
  id: number
  name: string
  description: string
  created_at: string
}

export interface Transaction {
  id: number
  portfolio_id: number
  ticker: string
  transaction_type: 'BUY' | 'SELL'
  shares: number
  price: number
  fee: number
  transaction_date: string
  notes: string
  created_at: string
}

export interface Holding {
  ticker: string
  name: string
  shares: number
  avg_cost: number
  current_price: number
  market_value: number
  unrealized_pnl: number
  unrealized_pnl_pct: number
}

export interface PnL {
  portfolio_id: number
  holdings: Holding[]
  total_cost: number
  total_value: number
  total_unrealized_pnl: number
  total_realized_pnl: number
  total_dividends: number
  total_rebates: number
  total_fees: number
}

export interface BacktestRun {
  id: number
  name: string
  market: string
  start_date: string
  end_date: string
  initial_capital: number
  parameters: Record<string, unknown>
  status: 'PENDING' | 'RUNNING' | 'DONE' | 'FAILED'
  created_at: string
  completed_at: string | null
  result: {
    total_return: number
    max_drawdown: number
    sharpe_ratio: number
    win_rate: number
    total_trades: number
  } | null
}

export interface PipelineRun {
  id: number
  pipeline_type: string
  started_at: string
  completed_at: string | null
  status: 'RUNNING' | 'SUCCESS' | 'FAILED'
  stocks_processed: number
}

export interface Podcast {
  id: number
  name: string
  rss_url: string
  description: string
  language: string
  is_active: boolean
  last_synced_at: string | null
  created_at: string
}

export interface PodcastEpisode {
  id: number
  title: string
  published_at: string | null
  episode_url: string
  transcript_src: string
  has_transcript: boolean
  analyzed_at: string | null
  mention_count: number
}

export interface PodcastMention {
  id: number
  episode_id: number
  episode_title: string
  episode_url: string
  published_at: string | null
  podcast_id: number
  podcast_name: string
  ticker: string | null
  ticker_raw: string
  sentiment: 'bullish' | 'bearish' | 'neutral'
  confidence: number
  thesis: string
  original_quote: string
  adopt: boolean
  created_at: string
}

// API functions
const BASE = '' // uses Vite proxy in dev

async function apiFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    ...options,
    headers: { 'Content-Type': 'application/json', ...options?.headers },
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(err.error || 'API error')
  }
  return res.json()
}

export const api = {
  // Screener
  screener: (params: {
    market?: string
    min_score?: number
    sort?: string
    order?: string
    page?: number
    limit?: number
  }) => {
    const q = new URLSearchParams()
    if (params.market) q.set('market', params.market)
    if (params.min_score !== undefined) q.set('min_score', String(params.min_score))
    if (params.sort) q.set('sort', params.sort)
    if (params.order) q.set('order', params.order)
    if (params.page) q.set('page', String(params.page))
    if (params.limit) q.set('limit', String(params.limit))
    return apiFetch<{ stocks: ScreenerStock[]; total: number }>(`/api/screener?${q}`)
  },

  // Stocks
  getStock: (ticker: string) => apiFetch<StockDetail>(`/api/stocks/${ticker}`),
  getStockScore: (ticker: string) => apiFetch<StockScoreDetail>(`/api/stocks/${ticker}/score`),
  getStockHistory: (ticker: string, from?: string, to?: string) => {
    const q = new URLSearchParams()
    if (from) q.set('from', from)
    if (to) q.set('to', to)
    return apiFetch<{ ticker: string; prices: StockPrice[] }>(`/api/stocks/${ticker}/history?${q}`)
  },
  generateReport: (ticker: string) =>
    apiFetch<{ ticker: string; date: string; summary: string; file_path?: string }>(
      `/api/stocks/${ticker}/report`,
      { method: 'POST', body: '{}' },
    ),

  // Portfolio
  getPortfolios: () => apiFetch<{ portfolios: Portfolio[] }>('/api/portfolio'),
  getPnL: (portfolio_id: number) => apiFetch<PnL>(`/api/portfolio/pnl?portfolio_id=${portfolio_id}`),
  getTransactions: (portfolio_id: number) =>
    apiFetch<{ transactions: Transaction[] }>(`/api/portfolio/transactions?portfolio_id=${portfolio_id}`),
  addTransaction: (data: Omit<Transaction, 'id' | 'created_at'>) =>
    apiFetch<Transaction>('/api/portfolio/transactions', { method: 'POST', body: JSON.stringify(data) }),
  addRebate: (data: { portfolio_id: number; amount: number; rebate_date: string; broker: string; notes: string }) =>
    apiFetch<unknown>('/api/portfolio/rebates', { method: 'POST', body: JSON.stringify(data) }),
  addDividend: (data: { portfolio_id: number; ticker: string; amount: number; dividend_date: string; notes: string }) =>
    apiFetch<unknown>('/api/portfolio/dividends', { method: 'POST', body: JSON.stringify(data) }),

  // Pipeline
  getPipelineStatus: () => apiFetch<{ runs: PipelineRun[] }>('/api/pipeline/status'),
  triggerPipeline: (pipeline_type: string) =>
    apiFetch<PipelineRun>('/api/pipeline/run', { method: 'POST', body: JSON.stringify({ pipeline_type }) }),

  // Insights
  listPodcasts: () => apiFetch<{ podcasts: Podcast[] }>('/api/insights/podcasts'),
  addPodcast: (data: { name: string; rss_url: string; description: string; language: string }) =>
    apiFetch<{ id: number }>('/api/insights/podcasts', { method: 'POST', body: JSON.stringify(data) }),
  deletePodcast: (id: number) =>
    apiFetch<{ ok: boolean }>(`/api/insights/podcasts/${id}`, { method: 'DELETE' }),
  syncPodcast: (id: number) =>
    apiFetch<{ new_episodes: number; new_mentions: number }>(`/api/insights/podcasts/${id}/sync`, { method: 'POST', body: '{}' }),
  listMentions: (params: { ticker?: string; sentiment?: string; adopt?: boolean }) => {
    const q = new URLSearchParams()
    if (params.ticker) q.set('ticker', params.ticker)
    if (params.sentiment) q.set('sentiment', params.sentiment)
    if (params.adopt) q.set('adopt', 'true')
    return apiFetch<{ mentions: PodcastMention[] }>(`/api/insights/mentions?${q}`)
  },
  toggleAdopt: (id: number) =>
    apiFetch<{ adopt: boolean }>(`/api/insights/mentions/${id}/adopt`, { method: 'PUT', body: '{}' }),
  listEpisodes: (podcastId: number) =>
    apiFetch<{ episodes: PodcastEpisode[] }>(`/api/insights/podcasts/${podcastId}/episodes`),
  fetchTranscript: (episodeId: number) =>
    apiFetch<{ ok: boolean }>(`/api/insights/episodes/${episodeId}/fetch-transcript`, { method: 'POST', body: '{}' }),
  analyzeEpisode: (episodeId: number) =>
    apiFetch<{ ok: boolean }>(`/api/insights/episodes/${episodeId}/analyze`, { method: 'POST', body: '{}' }),

  // Backtest
  getBacktests: () => apiFetch<{ backtests: BacktestRun[] }>('/api/backtest'),
  getBacktest: (id: number) => apiFetch<BacktestRun>(`/api/backtest/${id}`),
  createBacktest: (data: {
    name: string
    market: string
    start_date: string
    end_date: string
    initial_capital: number
    parameters: Record<string, unknown>
  }) => apiFetch<BacktestRun>('/api/backtest', { method: 'POST', body: JSON.stringify(data) }),
}
