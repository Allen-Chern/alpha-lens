import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import Layout from './components/Layout'
import ScreenerPage from './pages/ScreenerPage'
import StockDetailPage from './pages/StockDetailPage'
import PortfolioPage from './pages/PortfolioPage'
import PipelinePage from './pages/PipelinePage'
import BacktestPage from './pages/BacktestPage'
import InsightsPage from './pages/InsightsPage'

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index element={<Navigate to="/screener" replace />} />
          <Route path="screener" element={<ScreenerPage />} />
          <Route path="stocks/:ticker" element={<StockDetailPage />} />
          <Route path="portfolio" element={<PortfolioPage />} />
          <Route path="pipeline" element={<PipelinePage />} />
          <Route path="backtest" element={<BacktestPage />} />
          <Route path="insights" element={<InsightsPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}
