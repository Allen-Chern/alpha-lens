import { cn } from '../../lib/utils'
import { HTMLAttributes } from 'react'

interface BadgeProps extends HTMLAttributes<HTMLSpanElement> {
  variant?: 'default' | 'success' | 'warning' | 'error' | 'outline'
}

export function Badge({ className, variant = 'default', ...props }: BadgeProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center rounded px-2 py-0.5 text-xs font-medium',
        {
          'bg-zinc-700 text-zinc-300': variant === 'default',
          'bg-green-400/10 text-green-400': variant === 'success',
          'bg-yellow-400/10 text-yellow-400': variant === 'warning',
          'bg-red-400/10 text-red-400': variant === 'error',
          'border border-zinc-700 text-zinc-400': variant === 'outline',
        },
        className,
      )}
      {...props}
    />
  )
}
