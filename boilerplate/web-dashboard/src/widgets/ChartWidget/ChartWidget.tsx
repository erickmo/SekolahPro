import {
  AreaChart, Area,
  BarChart, Bar,
  LineChart, Line,
  PieChart, Pie, Cell,
  ComposedChart,
  XAxis, YAxis, CartesianGrid, Tooltip, Legend,
  ResponsiveContainer,
} from 'recharts'
import { ChartCard } from '@/widgets/ChartCard/ChartCard'
import styles from './ChartWidget.module.css'

// ─── Types ────────────────────────────────────────────────────────────────────

export type ChartType = 'area' | 'bar' | 'line' | 'pie' | 'donut' | 'composed'

export interface SeriesConfig {
  key: string
  /** Display name in legend and tooltip */
  name?: string
  color?: string
  /** Use gradient fill (area / bar) */
  gradient?: boolean
  /** Dashed stroke */
  dashed?: boolean
  /** For composed chart: which element to render. Default: 'bar' */
  as?: 'bar' | 'area' | 'line'
  /** Stack group ID */
  stackId?: string
  /** Show data dots. Default: false for area, true for line */
  dot?: boolean
}

export interface ChartConfig {
  type: ChartType
  data: Record<string, unknown>[]
  series: SeriesConfig[]
  /** x-axis data key. Default: 'name' */
  xKey?: string
  /** Chart body height in px. Default: 220 */
  height?: number
  /** Show legend. Default: auto (true when series.length > 1) */
  legend?: boolean
  /** Show background grid. Default: true */
  grid?: boolean
  /** Tooltip value formatter */
  valueFormatter?: (value: unknown) => string
  /** x-axis tick formatter */
  xFormatter?: (value: unknown) => string
  /** y-axis tick formatter */
  yFormatter?: (value: unknown) => string
  /** Pie/donut: key used as slice label. Default: 'name' */
  labelKey?: string
  /** Donut inner radius. Default: 55 */
  innerRadius?: number
  /** Stack bars/areas automatically */
  stacked?: boolean
}

export interface ChartWidgetProps {
  title: string
  subtitle?: string
  actions?: React.ReactNode
  config: ChartConfig
  className?: string
  loading?: boolean
  emptyMessage?: string
}

// ─── Default palette ──────────────────────────────────────────────────────────

const PALETTE = ['#4D2975', '#E9A800', '#26B8B0', '#8B4FCF', '#2563EB', '#16A34A', '#DC2626']

const TOOLTIP_STYLE = {
  background: 'var(--color-surface-elevated)',
  border: '1px solid var(--color-border)',
  borderRadius: '8px',
  fontSize: '12px',
  color: 'var(--color-text-primary)',
}

const AXIS_TICK = { fontSize: 11, fill: 'var(--color-text-tertiary)' }

// ─── Helpers ──────────────────────────────────────────────────────────────────

function color(s: SeriesConfig, idx: number): string {
  return s.color ?? PALETTE[idx % PALETTE.length]
}

function gradId(key: string): string {
  return `cw_g_${key.replace(/[^a-z0-9]/gi, '_')}`
}

function tooltipFmt(fmt?: (v: unknown) => string) {
  if (!fmt) return undefined
  return (v: unknown, name: string): [string, string] => [fmt(v), name]
}

// ─── Shared renderers ─────────────────────────────────────────────────────────

function Gradients({ series }: { series: SeriesConfig[] }) {
  const withGrad = series.filter((s) => s.gradient)
  if (withGrad.length === 0) return null
  return (
    <defs>
      {withGrad.map((s, i) => {
        const c = color(s, i)
        return (
          <linearGradient key={s.key} id={gradId(s.key)} x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%" stopColor={c} stopOpacity={0.35} />
            <stop offset="95%" stopColor={c} stopOpacity={0} />
          </linearGradient>
        )
      })}
    </defs>
  )
}

function CartesianBase({ cfg }: { cfg: ChartConfig }) {
  const showLegend = cfg.legend ?? cfg.series.length > 1
  const yWidth = cfg.yFormatter ? 72 : 48
  return (
    <>
      {cfg.grid !== false && (
        <CartesianGrid strokeDasharray="3 3" stroke="var(--color-border-subtle)" />
      )}
      <XAxis
        dataKey={cfg.xKey ?? 'name'}
        tick={AXIS_TICK}
        tickFormatter={cfg.xFormatter as ((v: unknown) => string) | undefined}
        axisLine={{ stroke: 'var(--color-border-subtle)' }}
        tickLine={false}
      />
      <YAxis
        tick={AXIS_TICK}
        tickFormatter={cfg.yFormatter as ((v: unknown) => string) | undefined}
        axisLine={false}
        tickLine={false}
        width={yWidth}
      />
      <Tooltip formatter={tooltipFmt(cfg.valueFormatter) as never} contentStyle={TOOLTIP_STYLE} />
      {showLegend && <Legend wrapperStyle={{ fontSize: '12px' }} />}
    </>
  )
}

// ─── Chart type renderers ─────────────────────────────────────────────────────

function AreaRenderer({ cfg }: { cfg: ChartConfig }) {
  return (
    <AreaChart data={cfg.data} margin={{ top: 8, right: 8, left: 0, bottom: 0 }}>
      <Gradients series={cfg.series} />
      <CartesianBase cfg={cfg} />
      {cfg.series.map((s, i) => {
        const c = color(s, i)
        return (
          <Area
            key={s.key} type="monotone" dataKey={s.key} name={s.name ?? s.key}
            stroke={c} strokeWidth={2}
            fill={s.gradient ? `url(#${gradId(s.key)})` : 'transparent'}
            strokeDasharray={s.dashed ? '5 3' : undefined}
            dot={s.dot ? { r: 3, fill: c } : false}
            stackId={s.stackId ?? (cfg.stacked ? 'stack' : undefined)}
          />
        )
      })}
    </AreaChart>
  )
}

function BarRenderer({ cfg }: { cfg: ChartConfig }) {
  return (
    <BarChart data={cfg.data} margin={{ top: 8, right: 8, left: 0, bottom: 0 }}>
      <Gradients series={cfg.series} />
      <CartesianBase cfg={cfg} />
      {cfg.series.map((s, i) => {
        const c = color(s, i)
        const isStacked = !!(s.stackId ?? cfg.stacked)
        return (
          <Bar
            key={s.key} dataKey={s.key} name={s.name ?? s.key}
            fill={s.gradient ? `url(#${gradId(s.key)})` : c}
            radius={isStacked ? 0 : [3, 3, 0, 0]}
            stackId={s.stackId ?? (cfg.stacked ? 'stack' : undefined)}
          />
        )
      })}
    </BarChart>
  )
}

function LineRenderer({ cfg }: { cfg: ChartConfig }) {
  return (
    <LineChart data={cfg.data} margin={{ top: 8, right: 8, left: 0, bottom: 0 }}>
      <CartesianBase cfg={cfg} />
      {cfg.series.map((s, i) => {
        const c = color(s, i)
        return (
          <Line
            key={s.key} type="monotone" dataKey={s.key} name={s.name ?? s.key}
            stroke={c} strokeWidth={2}
            strokeDasharray={s.dashed ? '5 3' : undefined}
            dot={s.dot !== false ? { r: 3, fill: c } : false}
            activeDot={{ r: 5 }}
          />
        )
      })}
    </LineChart>
  )
}

function PieRenderer({ cfg, isDonut }: { cfg: ChartConfig; isDonut: boolean }) {
  const dataKey = cfg.series[0]?.key ?? 'value'
  const nameKey = cfg.labelKey ?? 'name'
  const inner = isDonut ? (cfg.innerRadius ?? 55) : 0
  const outer = inner + 40
  const showLegend = cfg.legend ?? true
  return (
    <PieChart>
      <Pie
        data={cfg.data} cx="50%" cy="45%"
        innerRadius={inner} outerRadius={outer}
        paddingAngle={isDonut ? 3 : 1}
        dataKey={dataKey} nameKey={nameKey}
      >
        {cfg.data.map((entry, i) => (
          <Cell key={`c-${i}`} fill={(entry.color as string) ?? PALETTE[i % PALETTE.length]} />
        ))}
      </Pie>
      <Tooltip formatter={tooltipFmt(cfg.valueFormatter) as never} contentStyle={TOOLTIP_STYLE} />
      {showLegend && <Legend wrapperStyle={{ fontSize: '12px' }} />}
    </PieChart>
  )
}

function ComposedRenderer({ cfg }: { cfg: ChartConfig }) {
  return (
    <ComposedChart data={cfg.data} margin={{ top: 8, right: 8, left: 0, bottom: 0 }}>
      <Gradients series={cfg.series} />
      <CartesianBase cfg={cfg} />
      {cfg.series.map((s, i) => {
        const c = color(s, i)
        if (s.as === 'area') {
          return (
            <Area
              key={s.key} type="monotone" dataKey={s.key} name={s.name ?? s.key}
              stroke={c} strokeWidth={2}
              fill={s.gradient ? `url(#${gradId(s.key)})` : 'transparent'}
              strokeDasharray={s.dashed ? '5 3' : undefined}
              dot={s.dot ? { r: 3, fill: c } : false}
            />
          )
        }
        if (s.as === 'line') {
          return (
            <Line
              key={s.key} type="monotone" dataKey={s.key} name={s.name ?? s.key}
              stroke={c} strokeWidth={2}
              strokeDasharray={s.dashed ? '5 3' : undefined}
              dot={s.dot !== false ? { r: 3, fill: c } : false}
              activeDot={{ r: 5 }}
            />
          )
        }
        return (
          <Bar
            key={s.key} dataKey={s.key} name={s.name ?? s.key}
            fill={s.gradient ? `url(#${gradId(s.key)})` : c}
            radius={[3, 3, 0, 0]}
          />
        )
      })}
    </ComposedChart>
  )
}

// ─── Route to renderer ────────────────────────────────────────────────────────

function renderChart(cfg: ChartConfig): React.ReactElement {
  switch (cfg.type) {
    case 'area':     return <AreaRenderer cfg={cfg} />
    case 'bar':      return <BarRenderer cfg={cfg} />
    case 'line':     return <LineRenderer cfg={cfg} />
    case 'pie':      return <PieRenderer cfg={cfg} isDonut={false} />
    case 'donut':    return <PieRenderer cfg={cfg} isDonut={true} />
    case 'composed': return <ComposedRenderer cfg={cfg} />
  }
}

// ─── ChartWidget ──────────────────────────────────────────────────────────────

export function ChartWidget({
  title,
  subtitle,
  actions,
  config,
  className,
  loading = false,
  emptyMessage = 'Tidak ada data',
}: ChartWidgetProps) {
  const height = config.height ?? 220
  const isEmpty = !loading && config.data.length === 0

  return (
    <ChartCard title={title} subtitle={subtitle} actions={actions} className={className}>
      {loading ? (
        <div className={styles.skeleton} style={{ height }} />
      ) : isEmpty ? (
        <div className={styles.empty} style={{ height }}>
          <span>{emptyMessage}</span>
        </div>
      ) : (
        <ResponsiveContainer width="100%" height={height}>
          {renderChart(config)}
        </ResponsiveContainer>
      )}
    </ChartCard>
  )
}
