import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../lib/api'
import { Card, CardHeader, CardTitle, CardContent } from '../components/ui/card'
import { Button } from '../components/ui/button'
import { Badge } from '../components/ui/badge'
import { Play, RefreshCw } from 'lucide-react'

function statusBadge(status: string) {
  if (status === 'SUCCESS') return <Badge variant="success">成功</Badge>
  if (status === 'RUNNING') return <Badge variant="warning">運行中</Badge>
  if (status === 'FAILED') return <Badge variant="error">失敗</Badge>
  return <Badge>{status}</Badge>
}

function formatDateTime(dt: string | null) {
  if (!dt) return '—'
  return new Date(dt).toLocaleString('zh-TW', { timeZone: 'Asia/Taipei' })
}

export default function PipelinePage() {
  const qc = useQueryClient()

  const { data, isLoading, refetch } = useQuery({
    queryKey: ['pipeline-status'],
    queryFn: api.getPipelineStatus,
    refetchInterval: 5000,
  })

  const triggerMutation = useMutation({
    mutationFn: (type: string) => api.triggerPipeline(type),
    onSuccess: () => {
      setTimeout(() => qc.invalidateQueries({ queryKey: ['pipeline-status'] }), 500)
    },
  })

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-zinc-100">資料管線</h1>
          <p className="text-zinc-400 text-sm mt-1">每日盤後自動執行，可手動觸發</p>
        </div>
        <div className="flex gap-2">
          <Button size="sm" variant="outline" onClick={() => refetch()} aria-label="重新整理">
            <RefreshCw size={14} />
          </Button>
          <Button
            size="sm"
            onClick={() => triggerMutation.mutate('TW_DAILY')}
            disabled={triggerMutation.isPending}
            className="flex items-center gap-1"
          >
            <Play size={14} />
            台股管線
          </Button>
          <Button
            size="sm"
            variant="outline"
            onClick={() => triggerMutation.mutate('US_DAILY')}
            disabled={triggerMutation.isPending}
            className="flex items-center gap-1"
          >
            <Play size={14} />
            美股管線
          </Button>
        </div>
      </div>

      {triggerMutation.isSuccess && (
        <div className="p-3 rounded-lg bg-yellow-400/10 text-yellow-400 text-sm">
          管線已啟動 (ID: {triggerMutation.data.id})，5 秒後自動更新狀態
        </div>
      )}

      <Card>
        <CardHeader>
          <CardTitle>執行記錄</CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-zinc-800">
                <th className="px-4 py-3 text-left text-zinc-400">ID</th>
                <th className="px-4 py-3 text-left text-zinc-400">類型</th>
                <th className="px-4 py-3 text-left text-zinc-400">開始時間</th>
                <th className="px-4 py-3 text-left text-zinc-400">完成時間</th>
                <th className="px-4 py-3 text-left text-zinc-400">狀態</th>
                <th className="px-4 py-3 text-right text-zinc-400">處理股票數</th>
              </tr>
            </thead>
            <tbody>
              {isLoading && (
                <tr>
                  <td colSpan={6} className="px-4 py-8 text-center text-zinc-500">
                    載入中...
                  </td>
                </tr>
              )}
              {data?.runs.length === 0 && (
                <tr>
                  <td colSpan={6} className="px-4 py-8 text-center text-zinc-500">
                    尚無執行記錄
                  </td>
                </tr>
              )}
              {data?.runs.map((r) => (
                <tr key={r.id} className="border-b border-zinc-800/50 hover:bg-zinc-800/30">
                  <td className="px-4 py-3 font-mono text-zinc-500">#{r.id}</td>
                  <td className="px-4 py-3 font-mono text-zinc-300">{r.pipeline_type}</td>
                  <td className="px-4 py-3 text-zinc-400 text-xs">{formatDateTime(r.started_at)}</td>
                  <td className="px-4 py-3 text-zinc-400 text-xs">{formatDateTime(r.completed_at)}</td>
                  <td className="px-4 py-3">{statusBadge(r.status)}</td>
                  <td className="px-4 py-3 text-right font-mono text-zinc-300">{r.stocks_processed}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </CardContent>
      </Card>
    </div>
  )
}
