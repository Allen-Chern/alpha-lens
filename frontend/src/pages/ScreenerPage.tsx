import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { api } from '../lib/api'
import { formatNumber, formatPercent } from '../lib/utils'
import ScoreBadge from '../components/ScoreBadge'

type Market = '' | 'TW' | 'US'

const MARKETS: { label: string; value: Market }[] = [
  { label: '全部', value: '' },
  { label: '台股', value: 'TW' },
  { label: '美股', value: 'US' },
]

export default function ScreenerPage() {
  const navigate = useNavigate()
  const [market, setMarket] = useState<Market>('')
  const [sort, setSort] = useState('total_score')
  const [order, setOrder] = useState<'asc' | 'desc'>('desc')

  const { data, isLoading, error } = useQuery({
    queryKey: ['screener', market, sort, order],
    queryFn: () => api.screener({ market: market || undefined, sort, order }),
  })

  function handleSort(field: string) {
    if (sort === field) {
      setOrder((o) => (o === 'desc' ? 'asc' : 'desc'))
    } else {
      setSort(field)
      setOrder('desc')
    }
  }

  const columns = [
    { key: 'ticker', label: '代碼', sortable: false },
    { key: 'name', label: '公司', sortable: false },
    { key: 'total_score', label: '綜合分', sortable: true },
    { key: 'fundamental_score', label: '基本面', sortable: true },
    { key: 'chip_score', label: '籌碼', sortable: true },
    { key: 'momentum_score', label: '動能', sortable: true },
    { key: 'theme_score', label: '題材', sortable: true },
    { key: 'risk_score', label: '風險', sortable: true },
    { key: 'close', label: '收盤', sortable: true },
    { key: 'change_pct', label: '漲跌%', sortable: false },
  ]

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-zinc-100">選股篩選</h1>
        <p className="text-zinc-400 text-sm mt-1">每日盤後評分排行</p>
      </div>

      {/* Market filter */}
      <div className="flex gap-2 mb-4">
        {MARKETS.map((m) => (
          <button
            key={m.value}
            onClick={() => setMarket(m.value)}
            className={`px-4 py-1.5 rounded-full text-sm font-medium transition-colors ${
              market === m.value ? 'bg-violet-500 text-white' : 'bg-zinc-800 text-zinc-400 hover:text-zinc-100'
            }`}
          >
            {m.label}
          </button>
        ))}
        {data && <span className="ml-auto text-sm text-zinc-500 self-center">共 {data.total} 支</span>}
      </div>

      {/* Table */}
      <div className="rounded-lg border border-zinc-800 overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-zinc-800 bg-zinc-900">
              {columns.map((col) => (
                <th
                  key={col.key}
                  className={`px-4 py-3 text-left text-zinc-400 font-medium whitespace-nowrap ${
                    col.sortable ? 'cursor-pointer hover:text-zinc-100 select-none' : ''
                  } ${col.key === sort ? 'text-violet-400' : ''}`}
                  onClick={() => col.sortable && handleSort(col.key)}
                >
                  {col.label}
                  {col.sortable && col.key === sort && <span className="ml-1">{order === 'desc' ? '▼' : '▲'}</span>}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {isLoading &&
              Array.from({ length: 8 }).map((_, i) => (
                <tr key={i} className="border-b border-zinc-800/50">
                  {columns.map((c) => (
                    <td key={c.key} className="px-4 py-3">
                      <div className="h-4 rounded bg-zinc-800 animate-pulse w-16" />
                    </td>
                  ))}
                </tr>
              ))}
            {error && (
              <tr>
                <td colSpan={columns.length} className="px-4 py-8 text-center text-red-400">
                  載入失敗：{(error as Error).message}
                </td>
              </tr>
            )}
            {data?.stocks.map((stock) => (
              <tr
                key={stock.ticker}
                className="border-b border-zinc-800/50 hover:bg-zinc-800/50 cursor-pointer transition-colors"
                onClick={() => navigate(`/stocks/${stock.ticker}`)}
              >
                <td className="px-4 py-3 font-mono font-semibold text-violet-400">{stock.ticker}</td>
                <td className="px-4 py-3 text-zinc-100">{stock.name}</td>
                <td className="px-4 py-3">
                  <ScoreBadge score={stock.total_score} />
                </td>
                <td className="px-4 py-3">
                  <ScoreBadge score={stock.fundamental_score} size="sm" />
                </td>
                <td className="px-4 py-3">
                  <ScoreBadge score={stock.chip_score} size="sm" />
                </td>
                <td className="px-4 py-3">
                  <ScoreBadge score={stock.momentum_score} size="sm" />
                </td>
                <td className="px-4 py-3">
                  <ScoreBadge score={stock.theme_score} size="sm" />
                </td>
                <td className="px-4 py-3">
                  <ScoreBadge score={stock.risk_score} size="sm" />
                </td>
                <td className="px-4 py-3 font-mono text-zinc-100">{formatNumber(stock.close)}</td>
                <td
                  className={`px-4 py-3 font-mono font-medium ${
                    stock.change_pct >= 0 ? 'text-green-400' : 'text-red-400'
                  }`}
                >
                  {formatPercent(stock.change_pct)}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
