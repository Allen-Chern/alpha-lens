import { cn } from '../../lib/utils'
import { ButtonHTMLAttributes } from 'react'

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'default' | 'outline' | 'ghost' | 'destructive'
  size?: 'sm' | 'md' | 'lg'
}

export function Button({ className, variant = 'default', size = 'md', ...props }: ButtonProps) {
  return (
    <button
      className={cn(
        'inline-flex items-center justify-center rounded-md font-medium transition-colors focus-visible:outline-none disabled:pointer-events-none disabled:opacity-50',
        {
          'bg-violet-500 text-white hover:bg-violet-600': variant === 'default',
          'border border-zinc-700 bg-transparent text-zinc-100 hover:bg-zinc-800': variant === 'outline',
          'bg-transparent text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800': variant === 'ghost',
          'bg-red-500 text-white hover:bg-red-600': variant === 'destructive',
        },
        {
          'text-xs px-2.5 py-1.5 h-8': size === 'sm',
          'text-sm px-4 py-2 h-9': size === 'md',
          'text-base px-6 py-2.5 h-11': size === 'lg',
        },
        className,
      )}
      {...props}
    />
  )
}
