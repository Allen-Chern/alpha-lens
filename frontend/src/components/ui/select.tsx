import { cn } from '../../lib/utils'
import { SelectHTMLAttributes } from 'react'

export function Select({ className, ...props }: SelectHTMLAttributes<HTMLSelectElement>) {
  return (
    <select
      className={cn(
        'flex h-9 w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-1 text-sm text-zinc-100 focus:outline-none focus:ring-1 focus:ring-violet-400',
        className,
      )}
      {...props}
    />
  )
}
