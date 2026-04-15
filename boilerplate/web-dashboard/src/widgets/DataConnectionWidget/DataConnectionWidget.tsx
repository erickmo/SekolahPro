import { Link2 } from 'lucide-react'
import { useCompanyPath } from '@/hooks/useCompanyPath'
import { useHQPath } from '@/hooks/useHQPath'
import { useDashboardContext, type DashboardContext } from '@/hooks/useDashboardContext'
import type { FilterTuple } from '@/widgets/DataTable/filter.types'
import { buildFilterQueryString } from '@/widgets/DataTable/filter.utils'
import styles from './DataConnectionWidget.module.css'

// ─── Types ────────────────────────────────────────────────────────────────────

export type { FilterTuple }

export interface DataConnectionItem {
  icon: React.ReactNode
  title: string
  subtitle: string
  /** Legacy: direct click handler. Prefer path + filters instead. */
  onClick?: () => void
  /** Target path for filter-based navigation (e.g. 'master-data/items') */
  path?: string
  /** Structured filters appended as ?filters=[...] query string
   * Format: [field, operator, value] using Frappe-style operators
   * Examples: [['brand', '=', 'Samsung'], ['price', '>', '1000000']]
   * Operators: =, !=, like, in, >, >=, <, <=, between, is
   */
  filters?: FilterTuple[]
}

interface DataConnectionWidgetProps {
  title?: string
  items: DataConnectionItem[]
  /** Explicit dashboard context. If not provided, auto-detects from URL. */
  dashboardContext?: DashboardContext
}

// ─── Component ────────────────────────────────────────────────────────────────

export function DataConnectionWidget({
  title = 'Data Terkait',
  items,
  dashboardContext: explicitContext,
}: DataConnectionWidgetProps) {
  if (items.length === 0) return null

  const detectedContext = useDashboardContext()
  const context = explicitContext ?? detectedContext

  return (
    <div className={styles.wrapper}>
      <div className={styles.header}>
        <Link2 size={16} className={styles.headerIcon} />
        <h2 className={styles.headerTitle}>{title}</h2>
      </div>

      <div className={styles.grid}>
        {items.map((item, idx) => (
          <DataConnectionCard key={idx} {...item} context={context} />
        ))}
      </div>
    </div>
  )
}

// ─── Card ─────────────────────────────────────────────────────────────────────

interface DataConnectionCardProps extends DataConnectionItem {
  context: DashboardContext
}

function DataConnectionCard({ icon, title, subtitle, onClick, path, filters, context }: DataConnectionCardProps) {
  const { nav: companyNav } = useCompanyPath()
  const { nav: hqNav } = useHQPath()

  const nav = context === 'hq' ? hqNav : companyNav

  function handleClick() {
    if (path) {
      // Build URL with structured filters: ?filters=[["field","operator","value"]]
      // Uses Frappe-style operators (=, !=, like, in, >, >=, <, <=)
      const url = buildFilterQueryString(path, filters ?? [])
      nav(url)
    } else if (onClick) {
      onClick()
    }
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault()
      handleClick()
    }
  }

  return (
    <div
      className={styles.card}
      onClick={handleClick}
      onKeyDown={handleKeyDown}
      role="button"
      tabIndex={0}
    >
      <div className={styles.cardIcon}>{icon}</div>
      <div className={styles.cardBody}>
        <span className={styles.cardTitle}>{title}</span>
        <span className={styles.cardSub}>{subtitle}</span>
      </div>
    </div>
  )
}
