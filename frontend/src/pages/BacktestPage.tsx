import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../lib/api'
import { formatPercent } from '../lib/utils'
import { Card, CardHeader, CardTitle, CardContent } from '../components/ui/card'
import { Button } from '../components/ui/button'
import { Input } from '../components/ui/input'
import { Select } from '../components/ui/select'
import { Badge } from '../components/ui/badge'
import { Plus } from 'lucide-react'

function statusBadge(status: string) {
  if (status === 'DONE') return <Badge variant="success">完成</Badge>
  if (status === 'RUNNING') return <Badge variant="warning">運行中</Badge>
  if (status === 'FAILED') return <Badge variant="error">失敗</Badge>
  return <Badge variant="default">{status}</Badge>
}

export default function BacktestPage() {
  const qc = useQueryClient()
  const [showForm, setShowForm] = useState(false)
  const [selectedId, setSelectedId] = useState<number | null>(null)

  const { data, isLoading } = useQuery({
    queryKey: ['backtests'],
    queryFn: api.getBacktests,
  })

  const { data: detail } = useQuery({
    queryKey: ['backtest', selectedId],
    queryFn: () => api.getBacktest(selectedId!),
    enabled: selectedId !== null,
  })

  const [form, setForm] = useState({
    name: '',
    market: 'TW',
    start_date: '2025-01-01',
    end_date: '2026-01-01',
    initial_capital: '1000000',
    min_score: '70',
    top_n: '10',
  })

  const createMutation = useMutation({
    mutationFn: () =>
      api.createBacktest({
        name: form.name,
        market: form.market,
        start_date: form.start_date,
        end_date: form.end_date,
        initial_capital: Number(form.initial_capital),
        parameters: { min_score: Number(form.min_score), top_n: Number(form.top_n), rebalance: 'monthly' },
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['backtests'] })
      setShowForm(false)
    },
  })

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-zinc-100">策略回測</h1>
          <p className="text-zinc-400 text-sm mt-1">歷史驗證評分策略績效</p>
        </div>
        <Button size="sm" onClick={() => setShowForm((s) => !s)}>
          <Plus size={14} className="mr-1" />
          新建回測
        </Button>
      </div>

      {showForm && (
        <Card>
          <CardHeader>
            <CardTitle>新建回測</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-3 gap-3 mb-3">
              <Input placeholder="策略名稱" value={form.name} onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))} />
              <Select value={form.market} onChange={(e) => setForm((f) => ({ ...f, market: e.target.value }))}>
                <option value="TW">台股</option>
                <option value="US">美股</option>
              </Select>
              <Input
                placeholder="初始資金"
                type="number"
                value={form.initial_capital}
                onChange={(e) => setForm((f) => ({ ...f, initial_capital: e.target.value }))}
              />
              <div>
                <label className="text-xs text-zinc-500 mb-1 block">開始日期</label>
                <Input type="date" value={form.start_date} onChange={(e) => setForm((f) => ({ ...f, start_date: e.target.value }))} />
              </div>
              <div>
                <label className="text-xs text-zinc-500 mb-1 block">結束日期</label>
                <Input type="date" value={form.end_date} onChange={(e) => setForm((f) => ({ ...f, end_date: e.target.value }))} />
              </div>
              <Input
                placeholder="最低評分 (min_score)"
                type="number"
                value={form.min_score}
                onChange={(e) => setForm((f) => ({ ...f, min_score: e.target.value }))}
              />
            </div>
            <div className="flex gap-2">
              <Button size="sm" onClick={() => createMutation.mutate()} disabled={createMutation.isPending}>
                {createMutation.isPending ? '建立中...' : '建立回測'}
              </Button>
              <Button size="sm" variant="ghost" onClick={() => setShowForm(false)}>
                取消
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      <div className="grid grid-cols-3 gap-4">
        {/* Backtest list */}
        <div className="col-span-1 space-y-2">
          <h2 className="text-sm font-medium text-zinc-400 mb-3">回測列表</h2>
          {isLoading && <div className="text-zinc-500 text-sm">載入中...</div>}
          {data?.backtests.map((bt) => (
            <Card
              key={bt.id}
              className={`cursor-pointer hover:border-violet-400/50 transition-colors ${
                selectedId === bt.id ? 'border-violet-400' : ''
              }`}
              onClick={() => setSelectedId(bt.id)}
            >
              <CardContent className="p-3">
                <div className="flex items-center justify-between mb-1">
                  <span className="text-sm font-medium text-zinc-100 truncate">{bt.name}</span>
                  {statusBadge(bt.status)}
                </div>
                <div className="text-xs text-zinc-500">
                  {bt.market} · {bt.start_date} → {bt.end_date}
                </div>
                {bt.result && (
                  <div
                    className={`text-sm font-mono font-semibold mt-1 ${
                      bt.result.total_return >= 0 ? 'text-green-400' : 'text-red-400'
                    }`}
                  >
                    {formatPercent(bt.result.total_return * 100)}
                  </div>
                )}
              </CardContent>
            </Card>
          ))}
          {data?.backtests.length === 0 && <div className="text-zinc-500 text-sm">尚無回測記錄</div>}
        </div>

        {/* Backtest detail */}
        <div className="col-span-2">
          {selectedId && detail ? (
            <div className="space-y-4">
              <Card>
                <CardHeader>
                  <CardTitle>{detail.name}</CardTitle>
                </CardHeader>
                <CardContent>
                  {detail.result ? (
                    <div className="grid grid-cols-4 gap-4">
                      {[
                        {
                          label: '總報酬',
                          value: formatPercent(detail.result.total_return * 100),
                          positive: detail.result.total_return >= 0,
                        },
                        {
                          label: '最大回撤',
                          value: formatPercent(detail.result.max_drawdown * 100),
                          positive: false,
                        },
                        {
                          label: 'Sharpe',
                          value: detail.result.sharpe_ratio.toFixed(2),
                          positive: detail.result.sharpe_ratio >= 1,
                        },
                        {
                          label: '勝率',
                          value: formatPercent(detail.result.win_rate * 100),
                          positive: detail.result.win_rate >= 0.5,
                        },
                      ].map((m) => (
                        <div key={m.label} className="text-center">
                          <div className="text-xs text-zinc-500 mb-1">{m.label}</div>
                          <div className={`text-xl font-mono font-bold ${m.positive ? 'text-green-400' : 'text-red-400'}`}>
                            {m.value}
                          </div>
                        </div>
                      ))}
                    </div>
                  ) : (
                    <div className="text-zinc-500 text-sm">
                      {detail.status === 'PENDING'
                        ? '回測尚未開始'
                        : detail.status === 'RUNNING'
                          ? '回測執行中...'
                          : '尚無結果'}
                    </div>
                  )}
                </CardContent>
              </Card>
            </div>
          ) : (
            <div className="h-48 flex items-center justify-center text-zinc-600 text-sm">選擇左側回測查看詳情</div>
          )}
        </div>
      </div>
    </div>
  )
}
