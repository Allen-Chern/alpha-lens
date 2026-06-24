import type { Config } from 'tailwindcss'

export default {
  darkMode: 'class',
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        background: {
          DEFAULT: '#18181B',
        },
        accent: {
          DEFAULT: '#A78BFA',
          light: '#6366F1',
        },
        rise: '#4ADE80',
        fall: '#F87171',
      },
    },
  },
  plugins: [],
} satisfies Config
