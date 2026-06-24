import { cn, scoreBg } from '../lib/utils'

interface Props {
  score: number
  size?: 'sm' | 'md'
}

export default function ScoreBadge({ score, size = 'md' }: Props) {
  return (
    <span
      className={cn(
        'inline-flex items-center justify-center rounded font-mono font-semibold',
        scoreBg(score),
        size === 'sm' ? 'text-xs px-1.5 py-0.5' : 'text-sm px-2 py-1',
      )}
    >
      {score.toFixed(1)}
    </span>
  )
}
