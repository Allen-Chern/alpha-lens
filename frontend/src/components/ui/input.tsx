import { cn } from '../../lib/utils'
import { InputHTMLAttributes } from 'react'

export function Input({ className, ...props }: InputHTMLAttributes<HTMLInputElement>) {
  return (
    <input
      className={cn(
        'flex h-9 w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-1 text-sm text-zinc-100 placeholder:text-zinc-500 focus:outline-none focus:ring-1 focus:ring-violet-400 disabled:opacity-50',
        className,
      )}
      {...props}
    />
  )
}
