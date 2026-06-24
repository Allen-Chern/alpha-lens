import { useParams } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../lib/api'
import { formatNumber, formatPercent, formatLargeNumber, scoreColor } from '../lib/utils'
import ScoreBadge from '../components/ScoreBadge'
import { Card, CardHeader, CardTitle, CardContent } from '../components/ui/card'
import { Button } from '../components/ui/button'
import { Badge } from '../components/ui/badge'
import { LineChart, Line, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts'
import { FileText } from 'lucide-react'

const DIMENSIONS = [
  { key: 'fundamental_score', label: '基本面', weight: '30%' },
  { key: 'chip_score', label: '籌碼動能', weight: '25%' },
  { key: 'momentum_score', label: '價格動能', weight: '20%' },
  { key: 'theme_score', label: '題材性', weight: '15%' },
  { key: 'risk_score', label: '風險指標', weight: '10%' },
] as const

export default function StockDetailPage() {
  const { ticker } = useParams<{ ticker: string }>()
  const qc = useQueryClient()

  const { data: stock, isLoading } = useQuery({
    queryKey: ['stock', ticker],
    queryFn: () => api.getStock(ticker!),
    enabled: !!ticker,
  })

  const { data: score } = useQuery({
    queryKey: ['stock-score', ticker],
    queryFn: () => api.getStockScore(ticker!),
    enabled: !!ticker,
  })

  const { data: history } = useQuery({
    queryKey: ['stock-history', ticker],
    queryFn: () => api.getStockHistory(ticker!, undefined, undefined),
    enabled: !!ticker,
  })

  const reportMutation = useMutation({
    mutationFn: () => api.generateReport(ticker!),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['stock', ticker] }),
  })

  if (isLoading) {
    return (
      <div className="p-6">
        <div className="h-8 w-48 rounded bg-zinc-800 animate-pulse mb-4" />
        <div className="h-4 w-32 rounded bg-zinc-800 animate-pulse" />
      </div>
    )
  }

  if (!stock) return <div className="p-6 text-red-400">找不到股票</div>

  const price = stock.latest_price
  const latestScore = stock.latest_score

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <div className="flex items-center gap-3">
            <h1 className="text-3xl font-bold font-mono text-violet-400">{stock.ticker}</h1>
            <span className="text-xl text-zinc-100">{stock.name}</span>
            <Badge variant={stock.market === 'TW' ? 'default' : 'outline'}>{stock.market}</Badge>
          </div>
          <div className="flex items-center gap-4 mt-2">
            {price && (
              <>
                <span className="text-2xl font-mono font-bold text-zinc-100">{formatNumber(price.close)}</span>
                <span
                  className={`text-lg font-mono ${price.close > price.open ? 'text-green-400' : 'text-red-400'}`}
                >
                  {formatPercent(((price.close - price.open) / price.open) * 100)}
                </span>
              </>
            )}
            <span className="text-sm text-zinc-500">市值 {formatLargeNumber(stock.market_cap)}</span>
          </div>
        </div>
        <Button
          onClick={() => reportMutation.mutate()}
          disabled={reportMutation.isPending}
          variant="outline"
          className="flex items-center gap-2"
        >
          <FileText size={16} />
          {reportMutation.isPending ? '產生中...' : '產生報告'}
        </Button>
      </div>

      {/* Score cards */}
      {latestScore && (
        <div>
          <div className="flex items-center gap-3 mb-3">
            <span className="text-zinc-400 text-sm">綜合評分</span>
            <ScoreBadge score={latestScore.total_score} />
          </div>
          <div className="grid grid-cols-5 gap-3">
            {DIMENSIONS.map((d) => {
              const s = latestScore[d.key]
              return (
                <Card key={d.key}>
                  <CardContent className="p-4">
                    <div className="text-xs text-zinc-500 mb-1">{d.label}</div>
                    <div className="text-xs text-zinc-600 mb-2">權重 {d.weight}</div>
                    <div className={`text-2xl font-mono font-bold ${scoreColor(s)}`}>{s.toFixed(1)}</div>
                    <div className="mt-2 h-1.5 rounded-full bg-zinc-800">
                      <div className="h-full rounded-full bg-violet-400" style={{ width: `${s}%` }} />
                    </div>
                  </CardContent>
                </Card>
              )
            })}
          </div>
        </div>
      )}

      {/* Price chart */}
      {history && history.prices.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>價格走勢</CardTitle>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={200}>
              <LineChart data={history.prices}>
                <XAxis dataKey="date" tick={{ fill: '#71717a', fontSize: 11 }} />
                <YAxis tick={{ fill: '#71717a', fontSize: 11 }} domain={['auto', 'auto']} />
                <Tooltip
                  contentStyle={{ background: '#18181b', border: '1px solid #3f3f46', borderRadius: '6px' }}
                  labelStyle={{ color: '#a1a1aa' }}
                  itemStyle={{ color: '#a78bfa' }}
                />
                <Line type="monotone" dataKey="close" stroke="#a78bfa" dot={false} strokeWidth={2} />
              </LineChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>
      )}

      {/* Score breakdown */}
      {score && score.breakdown.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>評分明細</CardTitle>
          </CardHeader>
          <CardContent className="p-0">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-zinc-800">
                  <th className="px-4 py-2 text-left text-zinc-400">指標</th>
                  <th className="px-4 py-2 text-right text-zinc-400">數值</th>
                  <th className="px-4 py-2 text-right text-zinc-400">分數</th>
                  <th className="px-4 py-2 text-right text-zinc-400">權重</th>
                </tr>
              </thead>
              <tbody>
                {score.breakdown.map((b, i) => (
                  <tr key={i} className="border-b border-zinc-800/50 hover:bg-zinc-800/30">
                    <td className="px-4 py-2 text-zinc-300 font-mono text-xs">{b.indicator_name}</td>
                    <td className="px-4 py-2 text-right font-mono text-zinc-400">
                      {typeof b.indicator_value === 'number' ? b.indicator_value.toFixed(4) : b.indicator_value}
                    </td>
                    <td className="px-4 py-2 text-right">
                      <ScoreBadge score={b.indicator_score} size="sm" />
                    </td>
                    <td className="px-4 py-2 text-right text-zinc-500">{(b.weight * 100).toFixed(1)}%</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </CardContent>
        </Card>
      )}

      {reportMutation.isSuccess && (
        <div className="p-3 rounded-lg bg-green-400/10 text-green-400 text-sm">
          ✓ 報告已產生：{reportMutation.data.file_path || ''}
        </div>
      )}
    </div>
  )
}
