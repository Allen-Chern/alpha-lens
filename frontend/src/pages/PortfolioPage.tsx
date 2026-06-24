import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../lib/api'
import { formatNumber, formatPercent, formatLargeNumber } from '../lib/utils'
import { Card, CardHeader, CardTitle, CardContent } from '../components/ui/card'
import { Button } from '../components/ui/button'
import { Input } from '../components/ui/input'
import { Select } from '../components/ui/select'
import { Badge } from '../components/ui/badge'
import { Plus } from 'lucide-react'

// Default to portfolio ID 1 (the seed portfolio)
const PORTFOLIO_ID = 1

interface SummaryCard {
  label: string
  value: string
  sub?: string
  positive?: boolean
}

export default function PortfolioPage() {
  const qc = useQueryClient()
  const [showTxnForm, setShowTxnForm] = useState(false)
  const [showDivForm, setShowDivForm] = useState(false)
  const [showRebateForm, setShowRebateForm] = useState(false)

  const { data: pnl, isLoading } = useQuery({
    queryKey: ['pnl', PORTFOLIO_ID],
    queryFn: () => api.getPnL(PORTFOLIO_ID),
  })

  const { data: txnData } = useQuery({
    queryKey: ['transactions', PORTFOLIO_ID],
    queryFn: () => api.getTransactions(PORTFOLIO_ID),
  })

  // Add transaction form state
  const [txn, setTxn] = useState({
    ticker: '',
    transaction_type: 'BUY',
    shares: '',
    price: '',
    fee: '',
    transaction_date: '',
    notes: '',
  })
  const addTxn = useMutation({
    mutationFn: () =>
      api.addTransaction({
        portfolio_id: PORTFOLIO_ID,
        ticker: txn.ticker.toUpperCase(),
        transaction_type: txn.transaction_type as 'BUY' | 'SELL',
        shares: Number(txn.shares),
        price: Number(txn.price),
        fee: Number(txn.fee) || 0,
        transaction_date: txn.transaction_date,
        notes: txn.notes,
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['pnl'] })
      qc.invalidateQueries({ queryKey: ['transactions'] })
      setShowTxnForm(false)
      setTxn({ ticker: '', transaction_type: 'BUY', shares: '', price: '', fee: '', transaction_date: '', notes: '' })
    },
  })

  // Add dividend form state
  const [div, setDiv] = useState({ ticker: '', amount: '', dividend_date: '', notes: '' })
  const addDiv = useMutation({
    mutationFn: () =>
      api.addDividend({
        portfolio_id: PORTFOLIO_ID,
        ticker: div.ticker.toUpperCase(),
        amount: Number(div.amount),
        dividend_date: div.dividend_date,
        notes: div.notes,
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['pnl'] })
      setShowDivForm(false)
    },
  })

  // Add rebate form state
  const [rebate, setRebate] = useState({ amount: '', rebate_date: '', broker: '', notes: '' })
  const addRebate = useMutation({
    mutationFn: () =>
      api.addRebate({
        portfolio_id: PORTFOLIO_ID,
        amount: Number(rebate.amount),
        rebate_date: rebate.rebate_date,
        broker: rebate.broker,
        notes: rebate.notes,
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['pnl'] })
      setShowRebateForm(false)
    },
  })

  const summaryCards: SummaryCard[] = pnl
    ? [
        { label: '總市值', value: formatLargeNumber(pnl.total_value), sub: `成本 ${formatLargeNumber(pnl.total_cost)}` },
        { label: '未實現損益', value: formatNumber(pnl.total_unrealized_pnl), positive: pnl.total_unrealized_pnl >= 0 },
        { label: '現金股利', value: formatNumber(pnl.total_dividends), positive: true },
        { label: '手續費退傭', value: formatNumber(pnl.total_rebates), positive: true },
        { label: '總手續費', value: formatNumber(pnl.total_fees), positive: false },
      ]
    : []

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-zinc-100">持倉管理</h1>
          <p className="text-zinc-400 text-sm mt-1">主要投資組合</p>
        </div>
        <div className="flex gap-2">
          <Button size="sm" variant="outline" onClick={() => setShowDivForm((s) => !s)}>
            + 股利
          </Button>
          <Button size="sm" variant="outline" onClick={() => setShowRebateForm((s) => !s)}>
            + 退傭
          </Button>
          <Button size="sm" onClick={() => setShowTxnForm((s) => !s)}>
            <Plus size={14} className="mr-1" />
            交易
          </Button>
        </div>
      </div>

      {/* Summary cards */}
      <div className="grid grid-cols-5 gap-3">
        {summaryCards.map((c) => (
          <Card key={c.label}>
            <CardContent className="p-4">
              <div className="text-xs text-zinc-500 mb-1">{c.label}</div>
              <div
                className={`text-lg font-mono font-semibold ${
                  c.positive === undefined ? 'text-zinc-100' : c.positive ? 'text-green-400' : 'text-red-400'
                }`}
              >
                {c.value}
              </div>
              {c.sub && <div className="text-xs text-zinc-600 mt-0.5">{c.sub}</div>}
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Add transaction form */}
      {showTxnForm && (
        <Card>
          <CardHeader>
            <CardTitle>新增交易</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-6 gap-3">
              <Input placeholder="代碼" value={txn.ticker} onChange={(e) => setTxn((t) => ({ ...t, ticker: e.target.value }))} />
              <Select
                value={txn.transaction_type}
                onChange={(e) => setTxn((t) => ({ ...t, transaction_type: e.target.value }))}
              >
                <option value="BUY">買進</option>
                <option value="SELL">賣出</option>
              </Select>
              <Input placeholder="股數" type="number" value={txn.shares} onChange={(e) => setTxn((t) => ({ ...t, shares: e.target.value }))} />
              <Input placeholder="成交價" type="number" value={txn.price} onChange={(e) => setTxn((t) => ({ ...t, price: e.target.value }))} />
              <Input placeholder="手續費" type="number" value={txn.fee} onChange={(e) => setTxn((t) => ({ ...t, fee: e.target.value }))} />
              <Input type="date" value={txn.transaction_date} onChange={(e) => setTxn((t) => ({ ...t, transaction_date: e.target.value }))} />
            </div>
            <div className="flex gap-2 mt-3">
              <Button size="sm" onClick={() => addTxn.mutate()} disabled={addTxn.isPending}>
                {addTxn.isPending ? '儲存中...' : '確認'}
              </Button>
              <Button size="sm" variant="ghost" onClick={() => setShowTxnForm(false)}>
                取消
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Add dividend form */}
      {showDivForm && (
        <Card>
          <CardHeader>
            <CardTitle>新增現金股利</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-4 gap-3">
              <Input placeholder="代碼" value={div.ticker} onChange={(e) => setDiv((d) => ({ ...d, ticker: e.target.value }))} />
              <Input placeholder="金額" type="number" value={div.amount} onChange={(e) => setDiv((d) => ({ ...d, amount: e.target.value }))} />
              <Input type="date" value={div.dividend_date} onChange={(e) => setDiv((d) => ({ ...d, dividend_date: e.target.value }))} />
              <Input placeholder="備註" value={div.notes} onChange={(e) => setDiv((d) => ({ ...d, notes: e.target.value }))} />
            </div>
            <div className="flex gap-2 mt-3">
              <Button size="sm" onClick={() => addDiv.mutate()} disabled={addDiv.isPending}>
                確認
              </Button>
              <Button size="sm" variant="ghost" onClick={() => setShowDivForm(false)}>
                取消
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Add rebate form */}
      {showRebateForm && (
        <Card>
          <CardHeader>
            <CardTitle>新增退傭</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-4 gap-3">
              <Input placeholder="金額" type="number" value={rebate.amount} onChange={(e) => setRebate((r) => ({ ...r, amount: e.target.value }))} />
              <Input type="date" value={rebate.rebate_date} onChange={(e) => setRebate((r) => ({ ...r, rebate_date: e.target.value }))} />
              <Input placeholder="券商" value={rebate.broker} onChange={(e) => setRebate((r) => ({ ...r, broker: e.target.value }))} />
              <Input placeholder="備註" value={rebate.notes} onChange={(e) => setRebate((r) => ({ ...r, notes: e.target.value }))} />
            </div>
            <div className="flex gap-2 mt-3">
              <Button size="sm" onClick={() => addRebate.mutate()} disabled={addRebate.isPending}>
                確認
              </Button>
              <Button size="sm" variant="ghost" onClick={() => setShowRebateForm(false)}>
                取消
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Holdings table */}
      <Card>
        <CardHeader>
          <CardTitle>持股明細</CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-zinc-800">
                <th className="px-4 py-3 text-left text-zinc-400">代碼</th>
                <th className="px-4 py-3 text-left text-zinc-400">名稱</th>
                <th className="px-4 py-3 text-right text-zinc-400">股數</th>
                <th className="px-4 py-3 text-right text-zinc-400">成本均價</th>
                <th className="px-4 py-3 text-right text-zinc-400">現價</th>
                <th className="px-4 py-3 text-right text-zinc-400">市值</th>
                <th className="px-4 py-3 text-right text-zinc-400">未實現損益</th>
                <th className="px-4 py-3 text-right text-zinc-400">損益%</th>
              </tr>
            </thead>
            <tbody>
              {isLoading && (
                <tr>
                  <td colSpan={8} className="px-4 py-8 text-center text-zinc-500">
                    載入中...
                  </td>
                </tr>
              )}
              {pnl?.holdings.length === 0 && (
                <tr>
                  <td colSpan={8} className="px-4 py-8 text-center text-zinc-500">
                    尚無持股，請新增交易
                  </td>
                </tr>
              )}
              {pnl?.holdings.map((h) => (
                <tr key={h.ticker} className="border-b border-zinc-800/50 hover:bg-zinc-800/30">
                  <td className="px-4 py-3 font-mono text-violet-400 font-semibold">{h.ticker}</td>
                  <td className="px-4 py-3 text-zinc-100">{h.name}</td>
                  <td className="px-4 py-3 text-right font-mono text-zinc-300">{formatNumber(h.shares, 0)}</td>
                  <td className="px-4 py-3 text-right font-mono text-zinc-300">{formatNumber(h.avg_cost)}</td>
                  <td className="px-4 py-3 text-right font-mono text-zinc-100">{formatNumber(h.current_price)}</td>
                  <td className="px-4 py-3 text-right font-mono text-zinc-100">{formatLargeNumber(h.market_value)}</td>
                  <td
                    className={`px-4 py-3 text-right font-mono font-semibold ${
                      h.unrealized_pnl >= 0 ? 'text-green-400' : 'text-red-400'
                    }`}
                  >
                    {formatNumber(h.unrealized_pnl)}
                  </td>
                  <td
                    className={`px-4 py-3 text-right font-mono font-semibold ${
                      h.unrealized_pnl_pct >= 0 ? 'text-green-400' : 'text-red-400'
                    }`}
                  >
                    {formatPercent(h.unrealized_pnl_pct)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </CardContent>
      </Card>

      {/* Recent transactions */}
      {txnData && txnData.transactions.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>交易紀錄</CardTitle>
          </CardHeader>
          <CardContent className="p-0">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-zinc-800">
                  <th className="px-4 py-2 text-left text-zinc-400">日期</th>
                  <th className="px-4 py-2 text-left text-zinc-400">代碼</th>
                  <th className="px-4 py-2 text-left text-zinc-400">類型</th>
                  <th className="px-4 py-2 text-right text-zinc-400">股數</th>
                  <th className="px-4 py-2 text-right text-zinc-400">價格</th>
                  <th className="px-4 py-2 text-right text-zinc-400">手續費</th>
                </tr>
              </thead>
              <tbody>
                {txnData.transactions.map((t) => (
                  <tr key={t.id} className="border-b border-zinc-800/50">
                    <td className="px-4 py-2 text-zinc-400 text-xs">{t.transaction_date}</td>
                    <td className="px-4 py-2 font-mono text-violet-400">{t.ticker}</td>
                    <td className="px-4 py-2">
                      <Badge variant={t.transaction_type === 'BUY' ? 'success' : 'error'}>
                        {t.transaction_type === 'BUY' ? '買進' : '賣出'}
                      </Badge>
                    </td>
                    <td className="px-4 py-2 text-right font-mono text-zinc-300">{formatNumber(t.shares, 0)}</td>
                    <td className="px-4 py-2 text-right font-mono text-zinc-300">{formatNumber(t.price)}</td>
                    <td className="px-4 py-2 text-right font-mono text-zinc-500">{formatNumber(t.fee)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
