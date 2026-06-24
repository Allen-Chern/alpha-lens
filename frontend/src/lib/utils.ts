import { type ClassValue, clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatNumber(n: number, decimals = 2): string {
  return n.toLocaleString('zh-TW', { minimumFractionDigits: decimals, maximumFractionDigits: decimals })
}

export function formatPercent(n: number): string {
  const sign = n >= 0 ? '+' : ''
  return `${sign}${n.toFixed(2)}%`
}

export function formatLargeNumber(n: number): string {
  if (Math.abs(n) >= 1e12) return `${(n / 1e12).toFixed(2)}兆`
  if (Math.abs(n) >= 1e8) return `${(n / 1e8).toFixed(2)}億`
  if (Math.abs(n) >= 1e4) return `${(n / 1e4).toFixed(2)}萬`
  return n.toLocaleString()
}

export function scoreColor(score: number): string {
  if (score >= 80) return 'text-green-400'
  if (score >= 60) return 'text-yellow-400'
  return 'text-red-400'
}

export function scoreBg(score: number): string {
  if (score >= 80) return 'bg-green-400/10 text-green-400'
  if (score >= 60) return 'bg-yellow-400/10 text-yellow-400'
  return 'bg-red-400/10 text-red-400'
}
