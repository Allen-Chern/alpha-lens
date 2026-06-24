import { useState, useCallback } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { api, type Podcast, type PodcastEpisode, type PodcastMention } from '../lib/api'
import { useWebSocket } from '../lib/ws'
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/card'
import { Button } from '../components/ui/button'
import { Badge } from '../components/ui/badge'
import {
  RefreshCw, Trash2, Plus, ExternalLink, Radio, Check,
  ChevronDown, ChevronRight, FileText, Brain, AlertCircle, Loader2,
} from 'lucide-react'

// ─── Helpers ─────────────────────────────────────────────────────────────────

function sentimentBadge(s: string) {
  if (s === 'bullish') return <Badge variant="success">看多</Badge>
  if (s === 'bearish') return <Badge variant="error">看空</Badge>
  return <Badge>中性</Badge>
}

function formatDate(dt: string | null | undefined) {
  if (!dt) return '—'
  return new Date(dt).toLocaleDateString('zh-TW', { timeZone: 'Asia/Taipei', month: 'short', day: 'numeric' })
}

function srcLabel(src: string) {
  if (src === 'yt_captions') return 'YT 字幕'
  if (src === 'whisper_yt' || src === 'whisper_rss') return 'Whisper'
  if (src === 'remote') return '遠端'
  if (src === 'show_notes') return 'show notes'
  return src || '—'
}

// ─── Types ───────────────────────────────────────────────────────────────────

type EpStatus = 'idle' | 'fetching_transcript' | 'analyzing' | 'error' | 'transcript_done' | 'analyzed'

interface EpisodeState {
  status: EpStatus
  message?: string
}

// ─── EpisodeList ─────────────────────────────────────────────────────────────

function EpisodeList({
  podcastId,
  episodeStates,
  onFetchTranscript,
  onAnalyze,
}: {
  podcastId: number
  episodeStates: Record<number, EpisodeState>
  onFetchTranscript: (epId: number) => void
  onAnalyze: (epId: number) => void
}) {
  const { data, isLoading } = useQuery({
    queryKey: ['episodes', podcastId],
    queryFn: () => api.listEpisodes(podcastId),
  })

  if (isLoading) {
    return <p className="text-xs text-zinc-500 py-2 pl-2">載入中...</p>
  }

  const episodes = data?.episodes ?? []
  if (episodes.length === 0) {
    return <p className="text-xs text-zinc-500 py-2 pl-2">尚無集數，先同步</p>
  }

  return (
    <div className="mt-2 space-y-1">
      {episodes.map(ep => (
        <EpisodeRow
          key={ep.id}
          episode={ep}
          state={episodeStates[ep.id]}
          onFetchTranscript={() => onFetchTranscript(ep.id)}
          onAnalyze={() => onAnalyze(ep.id)}
        />
      ))}
    </div>
  )
}

function EpisodeRow({
  episode,
  state,
  onFetchTranscript,
  onAnalyze,
}: {
  episode: PodcastEpisode
  state: EpisodeState | undefined
  onFetchTranscript: () => void
  onAnalyze: () => void
}) {
  const status = state?.status ?? 'idle'
  const isBusy = status === 'fetching_transcript' || status === 'analyzing'

  // Local optimistic transcript/analyzed state driven by WS
  const hasTranscript = episode.has_transcript || status === 'transcript_done' || status === 'analyzed'
  const isAnalyzed = episode.analyzed_at != null || status === 'analyzed'

  return (
    <div className="flex items-start gap-2 px-2 py-2 rounded-md hover:bg-zinc-800/50 transition-colors">
      {/* Status icon */}
      <div className="mt-0.5 flex-shrink-0">
        {status === 'fetching_transcript' || status === 'analyzing' ? (
          <Loader2 size={13} className="text-violet-400 animate-spin" />
        ) : status === 'error' ? (
          <AlertCircle size={13} className="text-red-400" aria-label={state?.message} />
        ) : isAnalyzed ? (
          <Brain size={13} className="text-emerald-400" />
        ) : hasTranscript ? (
          <FileText size={13} className="text-violet-300" />
        ) : (
          <div className="w-3 h-3 rounded-full border border-zinc-700 mt-0.5" />
        )}
      </div>

      {/* Info */}
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-1.5">
          <span className="text-xs text-zinc-200 truncate">{episode.title}</span>
          {episode.episode_url && (
            <a href={episode.episode_url} target="_blank" rel="noopener noreferrer"
               className="text-zinc-600 hover:text-zinc-400 flex-shrink-0">
              <ExternalLink size={10} />
            </a>
          )}
        </div>
        <div className="flex items-center gap-2 mt-0.5">
          <span className="text-zinc-600 text-xs">{formatDate(episode.published_at)}</span>
          {hasTranscript && (
            <span className="text-zinc-500 text-xs">{srcLabel(episode.transcript_src)}</span>
          )}
          {isAnalyzed && episode.mention_count > 0 && (
            <span className="text-emerald-500 text-xs">{episode.mention_count} 標的</span>
          )}
          {status === 'transcript_done' && (
            <span className="text-violet-400 text-xs animate-pulse">逐字稿已就緒</span>
          )}
          {status === 'analyzed' && (
            <span className="text-emerald-400 text-xs animate-pulse">分析完成</span>
          )}
          {status === 'error' && (
            <span className="text-red-400 text-xs truncate max-w-[120px]" title={state?.message}>
              {state?.message}
            </span>
          )}
        </div>
      </div>

      {/* Actions */}
      <div className="flex gap-1 flex-shrink-0">
        <button
          onClick={onFetchTranscript}
          disabled={isBusy}
          title="抓逐字稿"
          className={`px-1.5 py-1 rounded text-xs border transition-colors ${
            isBusy
              ? 'border-zinc-800 text-zinc-600 cursor-not-allowed'
              : 'border-zinc-700 text-zinc-400 hover:border-violet-500 hover:text-violet-300'
          }`}
        >
          <FileText size={11} />
        </button>
        <button
          onClick={onAnalyze}
          disabled={isBusy || !hasTranscript}
          title={hasTranscript ? 'LLM 分析標的' : '先抓逐字稿'}
          className={`px-1.5 py-1 rounded text-xs border transition-colors ${
            isBusy || !hasTranscript
              ? 'border-zinc-800 text-zinc-600 cursor-not-allowed'
              : 'border-zinc-700 text-zinc-400 hover:border-emerald-500 hover:text-emerald-300'
          }`}
        >
          <Brain size={11} />
        </button>
      </div>
    </div>
  )
}

// ─── AddPodcastForm ───────────────────────────────────────────────────────────

function AddPodcastForm({ onSuccess }: { onSuccess: () => void }) {
  const [name, setName] = useState('')
  const [rssUrl, setRssUrl] = useState('')
  const [language, setLanguage] = useState('zh')
  const [open, setOpen] = useState(false)

  const mut = useMutation({
    mutationFn: () => api.addPodcast({ name, rss_url: rssUrl, description: '', language }),
    onSuccess: () => {
      setName(''); setRssUrl(''); setOpen(false)
      onSuccess()
    },
  })

  if (!open) {
    return (
      <Button size="sm" onClick={() => setOpen(true)} className="flex items-center gap-1">
        <Plus size={14} />
        新增播客
      </Button>
    )
  }

  return (
    <div className="p-4 border border-zinc-700 rounded-lg space-y-3 bg-zinc-900">
      <p className="text-sm font-medium text-zinc-300">新增播客</p>
      <input
        className="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-100 placeholder-zinc-500 focus:outline-none focus:border-violet-500"
        placeholder="播客名稱"
        value={name}
        onChange={e => setName(e.target.value)}
      />
      <input
        className="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-100 placeholder-zinc-500 focus:outline-none focus:border-violet-500"
        placeholder="RSS 網址 或 youtube.com/@頻道名稱"
        value={rssUrl}
        onChange={e => setRssUrl(e.target.value)}
      />
      <select
        className="bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-100 focus:outline-none focus:border-violet-500"
        value={language}
        onChange={e => setLanguage(e.target.value)}
      >
        <option value="zh">中文</option>
        <option value="en">English</option>
      </select>
      <div className="flex gap-2">
        <Button size="sm" onClick={() => mut.mutate()} disabled={!name || !rssUrl || mut.isPending}>
          {mut.isPending ? '新增中...' : '確認新增'}
        </Button>
        <Button size="sm" variant="outline" onClick={() => setOpen(false)}>取消</Button>
      </div>
      {mut.isError && (
        <p className="text-xs text-red-400">{(mut.error as Error).message}</p>
      )}
    </div>
  )
}

// ─── PodcastCard ──────────────────────────────────────────────────────────────

function PodcastCard({
  podcast,
  episodeStates,
  onFetchTranscript,
  onAnalyzeEpisode,
}: {
  podcast: Podcast
  episodeStates: Record<number, EpisodeState>
  onFetchTranscript: (epId: number) => void
  onAnalyzeEpisode: (epId: number) => void
}) {
  const [expanded, setExpanded] = useState(false)
  const qc = useQueryClient()

  const syncMut = useMutation({
    mutationFn: () => api.syncPodcast(podcast.id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['mentions'] })
      qc.invalidateQueries({ queryKey: ['episodes', podcast.id] })
    },
  })
  const delMut = useMutation({
    mutationFn: () => api.deletePodcast(podcast.id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['podcasts'] }),
  })

  return (
    <div className="border border-zinc-800 rounded-lg bg-zinc-900 overflow-hidden">
      {/* Header row */}
      <div className="flex items-center justify-between p-3 hover:border-zinc-700 transition-colors">
        <button
          className="flex items-center gap-2 min-w-0 flex-1 text-left"
          onClick={() => setExpanded(v => !v)}
        >
          {expanded
            ? <ChevronDown size={14} className="text-zinc-500 flex-shrink-0" />
            : <ChevronRight size={14} className="text-zinc-500 flex-shrink-0" />}
          <Radio size={14} className="text-violet-400 flex-shrink-0" />
          <div className="min-w-0">
            <p className="text-sm font-medium text-zinc-100 truncate">{podcast.name}</p>
            <p className="text-xs text-zinc-500">
              {podcast.last_synced_at ? `上次同步 ${formatDate(podcast.last_synced_at)}` : '尚未同步'}
            </p>
          </div>
        </button>
        <div className="flex gap-1 flex-shrink-0 ml-2 items-center">
          {syncMut.isSuccess && (
            <span className="text-xs text-emerald-400 mr-1">
              +{syncMut.data.new_episodes} 集
            </span>
          )}
          <Button
            size="sm"
            variant="outline"
            onClick={() => syncMut.mutate()}
            disabled={syncMut.isPending}
            aria-label="同步"
            title="同步 feed"
          >
            <RefreshCw size={13} className={syncMut.isPending ? 'animate-spin' : ''} />
          </Button>
          <Button
            size="sm"
            variant="outline"
            onClick={() => delMut.mutate()}
            disabled={delMut.isPending}
            aria-label="刪除"
            className="text-red-400 hover:text-red-300"
          >
            <Trash2 size={13} />
          </Button>
        </div>
      </div>

      {/* Episode list (expandable) */}
      {expanded && (
        <div className="border-t border-zinc-800 px-1 pb-2">
          <EpisodeList
            podcastId={podcast.id}
            episodeStates={episodeStates}
            onFetchTranscript={onFetchTranscript}
            onAnalyze={onAnalyzeEpisode}
          />
        </div>
      )}
    </div>
  )
}

// ─── MentionRow ───────────────────────────────────────────────────────────────

function MentionRow({ mention }: { mention: PodcastMention }) {
  const [expanded, setExpanded] = useState(false)
  const qc = useQueryClient()
  const adoptMut = useMutation({
    mutationFn: () => api.toggleAdopt(mention.id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['mentions'] }),
  })

  return (
    <>
      <tr
        className="border-b border-zinc-800/50 hover:bg-zinc-800/30 transition-colors cursor-pointer"
        onClick={() => mention.original_quote && setExpanded(v => !v)}
      >
        <td className="px-4 py-3">
          {mention.ticker ? (
            <Link
              to={`/stocks/${mention.ticker}`}
              className="font-mono text-violet-400 hover:text-violet-300 font-medium"
              onClick={e => e.stopPropagation()}
            >
              {mention.ticker}
            </Link>
          ) : (
            <span className="text-zinc-500 text-xs">{mention.ticker_raw}</span>
          )}
        </td>
        <td className="px-4 py-3">{sentimentBadge(mention.sentiment)}</td>
        <td className="px-4 py-3 text-zinc-300 text-sm max-w-xs">
          <div className="flex items-center gap-1">
            {mention.original_quote && (
              expanded
                ? <ChevronDown size={12} className="text-zinc-500 flex-shrink-0" />
                : <ChevronRight size={12} className="text-zinc-500 flex-shrink-0" />
            )}
            <p className="truncate" title={mention.thesis}>{mention.thesis || '—'}</p>
          </div>
        </td>
        <td className="px-4 py-3 text-xs text-zinc-500">
          <div className="flex items-center gap-1">
            <span className="truncate max-w-[140px]" title={mention.episode_title}>
              {mention.episode_title}
            </span>
            {mention.episode_url && (
              <a href={mention.episode_url} target="_blank" rel="noopener noreferrer"
                 className="text-zinc-600 hover:text-zinc-400 flex-shrink-0" aria-label="開啟節目"
                 onClick={e => e.stopPropagation()}>
                <ExternalLink size={11} />
              </a>
            )}
          </div>
          <p className="text-zinc-600">{mention.podcast_name} · {formatDate(mention.published_at)}</p>
        </td>
        <td className="px-4 py-3 text-right">
          <button
            onClick={e => { e.stopPropagation(); adoptMut.mutate() }}
            disabled={adoptMut.isPending}
            aria-label={mention.adopt ? '取消採納' : '採納為訊號'}
            className={`p-1.5 rounded transition-colors ${
              mention.adopt
                ? 'bg-emerald-500/20 text-emerald-400 hover:bg-emerald-500/30'
                : 'text-zinc-600 hover:text-zinc-400 hover:bg-zinc-800'
            }`}
          >
            <Check size={14} />
          </button>
        </td>
      </tr>
      {expanded && mention.original_quote && (
        <tr className="border-b border-zinc-800/30 bg-zinc-900/60">
          <td colSpan={5} className="px-4 py-3">
            <p className="text-xs text-zinc-400 font-medium mb-1">原話</p>
            <blockquote className="text-xs text-zinc-300 leading-relaxed border-l-2 border-violet-600/50 pl-3 italic">
              {mention.original_quote}
            </blockquote>
          </td>
        </tr>
      )}
    </>
  )
}

// ─── InsightsPage ─────────────────────────────────────────────────────────────

export default function InsightsPage() {
  const qc = useQueryClient()
  const [filterSentiment, setFilterSentiment] = useState('')
  const [filterAdopt, setFilterAdopt] = useState(false)

  // Per-episode in-flight status, keyed by episode_id
  const [episodeStates, setEpisodeStates] = useState<Record<number, EpisodeState>>({})

  const setEpState = useCallback((id: number, state: EpisodeState) => {
    setEpisodeStates(prev => ({ ...prev, [id]: state }))
  }, [])

  // WS: receive real-time status updates from backend jobs
  useWebSocket(useCallback((event) => {
    if (event.type !== 'episode_status') return
    const { episode_id, podcast_id, status, message } = event

    setEpState(episode_id, { status, message })

    // Refresh episode list when a job finishes
    if (status === 'transcript_done' || status === 'analyzed' || status === 'error') {
      qc.invalidateQueries({ queryKey: ['episodes', podcast_id] })
      if (status === 'analyzed') {
        qc.invalidateQueries({ queryKey: ['mentions'] })
      }
    }
  }, [qc, setEpState]))

  // Trigger actions (fire-and-forget; WS updates drive the UI)
  const handleFetchTranscript = useCallback((epId: number) => {
    setEpState(epId, { status: 'fetching_transcript' })
    api.fetchTranscript(epId).catch(() => setEpState(epId, { status: 'error', message: '請求失敗' }))
  }, [setEpState])

  const handleAnalyzeEpisode = useCallback((epId: number) => {
    setEpState(epId, { status: 'analyzing' })
    api.analyzeEpisode(epId).catch(() => setEpState(epId, { status: 'error', message: '請求失敗' }))
  }, [setEpState])

  const { data: podcastData } = useQuery({
    queryKey: ['podcasts'],
    queryFn: api.listPodcasts,
  })

  const { data: mentionData, isLoading } = useQuery({
    queryKey: ['mentions', filterSentiment, filterAdopt],
    queryFn: () => api.listMentions({ sentiment: filterSentiment, adopt: filterAdopt }),
  })

  return (
    <div className="p-6 space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-zinc-100">播客洞察</h1>
        <p className="text-zinc-400 text-sm mt-1">追蹤投資播客，以 AI 萃取潛在投資標的</p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left: Podcast list */}
        <div className="space-y-3">
          <Card>
            <CardHeader>
              <CardTitle>追蹤播客</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              <AddPodcastForm onSuccess={() => qc.invalidateQueries({ queryKey: ['podcasts'] })} />
              {(podcastData?.podcasts ?? []).map(p => (
                <PodcastCard
                  key={p.id}
                  podcast={p}
                  episodeStates={episodeStates}
                  onFetchTranscript={handleFetchTranscript}
                  onAnalyzeEpisode={handleAnalyzeEpisode}
                />
              ))}
              {(podcastData?.podcasts ?? []).length === 0 && (
                <p className="text-sm text-zinc-500 text-center py-4">尚無播客，先新增一個</p>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardContent className="pt-4">
              <p className="text-xs text-zinc-500 leading-relaxed">
                支援 RSS 播客網址，以及 YouTube 頻道網址（如 youtube.com/@channelname），會自動轉換。<br />
                同步後展開播客可對每集個別抓逐字稿（<FileText size={10} className="inline" />）或 LLM 分析（<Brain size={10} className="inline" />）。<br />
                <span className="text-emerald-400">✓ 採納</span> 的標的未來將納入 Theme 計分維度。
              </p>
            </CardContent>
          </Card>
        </div>

        {/* Right: Mentions feed */}
        <div className="lg:col-span-2">
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle>投資標的提及</CardTitle>
                <div className="flex items-center gap-2">
                  <select
                    className="bg-zinc-800 border border-zinc-700 rounded px-2 py-1 text-xs text-zinc-300 focus:outline-none"
                    value={filterSentiment}
                    onChange={e => setFilterSentiment(e.target.value)}
                  >
                    <option value="">全部觀點</option>
                    <option value="bullish">看多</option>
                    <option value="bearish">看空</option>
                    <option value="neutral">中性</option>
                  </select>
                  <button
                    onClick={() => setFilterAdopt(v => !v)}
                    className={`px-2 py-1 rounded text-xs border transition-colors ${
                      filterAdopt
                        ? 'border-emerald-500 text-emerald-400 bg-emerald-500/10'
                        : 'border-zinc-700 text-zinc-400 hover:border-zinc-600'
                    }`}
                  >
                    僅顯示已採納
                  </button>
                </div>
              </div>
            </CardHeader>
            <CardContent className="p-0">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-zinc-800">
                    <th className="px-4 py-3 text-left text-zinc-400 w-24">標的</th>
                    <th className="px-4 py-3 text-left text-zinc-400 w-20">觀點</th>
                    <th className="px-4 py-3 text-left text-zinc-400">投資論點</th>
                    <th className="px-4 py-3 text-left text-zinc-400">來源</th>
                    <th className="px-4 py-3 text-right text-zinc-400 w-16">採納</th>
                  </tr>
                </thead>
                <tbody>
                  {isLoading && (
                    <tr>
                      <td colSpan={5} className="px-4 py-8 text-center text-zinc-500">載入中...</td>
                    </tr>
                  )}
                  {!isLoading && (mentionData?.mentions ?? []).length === 0 && (
                    <tr>
                      <td colSpan={5} className="px-4 py-8 text-center text-zinc-500">
                        尚無標的提及，先同步一個播客
                      </td>
                    </tr>
                  )}
                  {(mentionData?.mentions ?? []).map(m => (
                    <MentionRow key={m.id} mention={m} />
                  ))}
                </tbody>
              </table>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}
