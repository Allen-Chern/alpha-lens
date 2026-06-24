import { Outlet, NavLink } from 'react-router-dom'
import { BarChart3, TrendingUp, Briefcase, Activity, FlaskConical, Radio, Sun, Moon } from 'lucide-react'
import { useState, useEffect } from 'react'

const navItems = [
  { to: '/screener', icon: BarChart3, label: '選股篩選' },
  { to: '/insights', icon: Radio, label: '播客洞察' },
  { to: '/portfolio', icon: Briefcase, label: '持倉管理' },
  { to: '/pipeline', icon: Activity, label: '資料管線' },
  { to: '/backtest', icon: FlaskConical, label: '策略回測' },
]

export default function Layout() {
  const [dark, setDark] = useState(true)

  useEffect(() => {
    document.documentElement.classList.toggle('dark', dark)
  }, [dark])

  return (
    <div className="flex h-screen bg-zinc-950 dark:bg-zinc-950 text-zinc-100">
      {/* Sidebar */}
      <aside className="w-56 flex-shrink-0 border-r border-zinc-800 flex flex-col">
        <div className="p-4 border-b border-zinc-800">
          <div className="flex items-center gap-2">
            <TrendingUp className="text-violet-400" size={22} />
            <span className="font-bold text-lg text-violet-400">AlphaLens</span>
          </div>
          <p className="text-xs text-zinc-500 mt-1">法人視角投資平台</p>
        </div>
        <nav className="flex-1 p-3 space-y-1">
          {navItems.map(({ to, icon: Icon, label }) => (
            <NavLink
              key={to}
              to={to}
              className={({ isActive }) =>
                `flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors ${
                  isActive
                    ? 'bg-violet-400/10 text-violet-400 font-medium'
                    : 'text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800'
                }`
              }
            >
              <Icon size={16} />
              {label}
            </NavLink>
          ))}
        </nav>
        <div className="p-3 border-t border-zinc-800">
          <button
            onClick={() => setDark((d) => !d)}
            className="flex items-center gap-2 text-sm text-zinc-400 hover:text-zinc-100 transition-colors w-full px-3 py-2 rounded-lg hover:bg-zinc-800"
            aria-label="切換主題"
          >
            {dark ? <Sun size={16} /> : <Moon size={16} />}
            {dark ? '淺色模式' : '深色模式'}
          </button>
        </div>
      </aside>

      {/* Main content */}
      <main className="flex-1 overflow-y-auto">
        <Outlet />
      </main>
    </div>
  )
}
