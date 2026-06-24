import { useEffect, useRef, useCallback } from 'react'

export type EpisodeStatus =
  | 'fetching_transcript'
  | 'transcript_done'
  | 'analyzing'
  | 'analyzed'
  | 'error'

export interface EpisodeStatusEvent {
  type: 'episode_status'
  episode_id: number
  podcast_id: number
  status: EpisodeStatus
  src?: string
  chars?: number
  mention_count?: number
  message?: string
}

export type WSEvent = EpisodeStatusEvent

export function useWebSocket(onMessage: (e: WSEvent) => void) {
  const wsRef = useRef<WebSocket | null>(null)
  const onMessageRef = useRef(onMessage)
  onMessageRef.current = onMessage

  const connect = useCallback(() => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const ws = new WebSocket(`${protocol}//${window.location.host}/api/ws`)
    wsRef.current = ws

    ws.onmessage = (e) => {
      try {
        onMessageRef.current(JSON.parse(e.data) as WSEvent)
      } catch {}
    }

    ws.onclose = () => {
      setTimeout(connect, 2000)
    }

    ws.onerror = () => {
      ws.close()
    }
  }, [])

  useEffect(() => {
    connect()
    return () => {
      const ws = wsRef.current
      if (ws) {
        // Prevent the onclose reconnect loop on intentional teardown
        ws.onclose = null
        ws.close()
      }
    }
  }, [connect])
}
