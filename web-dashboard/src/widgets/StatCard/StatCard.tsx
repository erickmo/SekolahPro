import { useEffect, useRef } from 'react'
import { TrendingUp, TrendingDown } from 'lucide-react'
import { animate } from 'framer-motion'
import { formatCurrency, formatNumber } from '@/utils/format'
import styles from './StatCard.module.css'

interface StatCardProps {
  title: string
  value: number
  format?: 'currency' | 'number' | 'percent'
  trend?: number
  trendDirection?: 'up' | 'down'
  sparkline?: number[]
  icon?: React.ReactNode
  color?: 'primary' | 'secondary' | 'accent' | 'warning' | 'danger' | 'info'
  animationDelay?: number
}

function Sparkline({ data, color }: { data: number[]; color: string }) {
  const min = Math.min(...data)
  const max = Math.max(...data)
  const range = max - min || 1
  const w = 64
  const h = 24
  const pts = data.map((v, i) => ({
    x: (i / (data.length - 1)) * w,
    y: h - ((v - min) / range) * h,
  }))
  const path = pts.map((p, i) => `${i === 0 ? 'M' : 'L'} ${p.x} ${p.y}`).join(' ')
  const fill = `${path} L ${w} ${h} L 0 ${h} Z`

  return (
    <svg width={w} height={h} viewBox={`0 0 ${w} ${h}`} className={styles.sparkline}>
      <defs>
        <linearGradient id={`sg-${color}`} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor={color} stopOpacity="0.3" />
          <stop offset="100%" stopColor={color} stopOpacity="0" />
        </linearGradient>
      </defs>
      <path d={fill} fill={`url(#sg-${color})`} />
      <path d={path} fill="none" stroke={color} strokeWidth="1.5" strokeLinecap="round" />
    </svg>
  )
}

const COLOR_MAP: Record<string, string> = {
  primary: '#4D2975',
  secondary: '#E9A800',
  accent: '#26B8B0',
  warning: '#D97706',
  danger: '#DC2626',
  info: '#2563EB',
}

export function StatCard({
  title,
  value,
  format = 'number',
  trend,
  trendDirection,
  sparkline,
  icon,
  color = 'primary',
  animationDelay = 0,
}: StatCardProps) {
  const valueRef = useRef<HTMLSpanElement>(null)
  const hexColor = COLOR_MAP[color] ?? COLOR_MAP.primary

  useEffect(() => {
    const el = valueRef.current
    if (!el) return
    const controls = animate(0, value, {
      duration: 1.2,
      ease: 'easeOut',
      delay: animationDelay / 1000,
      onUpdate: (v) => {
        if (format === 'currency') {
          el.textContent = formatCurrency(v, v >= 1_000_000)
        } else if (format === 'percent') {
          el.textContent = `${v.toFixed(1)}%`
        } else {
          el.textContent = formatNumber(Math.round(v))
        }
      },
    })
    return () => controls.stop()
  }, [value, format, animationDelay])

  const isPositiveTrend = trendDirection === 'up'

  return (
    <div className={styles.card} style={{ animationDelay: `${animationDelay}ms` }}>
      <div className={styles.header}>
        {icon && (
          <div className={styles.iconWrap} style={{ background: `${hexColor}15`, color: hexColor }}>
            {icon}
          </div>
        )}
        <span className={styles.title}>{title}</span>
      </div>

      <div className={styles.body}>
        <span ref={valueRef} className={styles.value}>
          {format === 'currency' ? formatCurrency(value, value >= 1_000_000) : formatNumber(value)}
        </span>

        {sparkline && sparkline.length > 1 && (
          <div className={styles.sparklineWrap}>
            <Sparkline data={sparkline} color={hexColor} />
          </div>
        )}
      </div>

      {trend !== undefined && trend !== 0 && (
        <div className={`${styles.trend} ${isPositiveTrend ? styles.up : styles.down}`}>
          {isPositiveTrend ? <TrendingUp size={12} /> : <TrendingDown size={12} />}
          <span>{Math.abs(trend).toFixed(1)}% dari bulan lalu</span>
        </div>
      )}
    </div>
  )
}
