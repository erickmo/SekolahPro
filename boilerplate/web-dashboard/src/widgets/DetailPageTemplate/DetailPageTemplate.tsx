import { useState, useRef, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { ArrowLeft, HelpCircle, ChevronRight, Home, Link2, MoreHorizontal } from 'lucide-react'
import { StatusPills } from '@/widgets/StatusPills/StatusPills'
import { ProgressWidget, type ProgressStep, type TerminalStatus } from '@/widgets/ProgressWidget/ProgressWidget'
import { DataConnectionWidget, type DataConnectionItem, type FilterTuple } from '@/widgets/DataConnectionWidget/DataConnectionWidget'
import { useAutoBreadcrumbs } from '@/hooks/useAutoBreadcrumbs'
import { useDashboardContext, type DashboardContext } from '@/hooks/useDashboardContext'
import { useModuleAccess } from '@/hooks/useModuleAccess'
import styles from './DetailPageTemplate.module.css'

export type { DataConnectionItem, FilterTuple, DetailPageSidebar }

// ─── Types ────────────────────────────────────────────────────────────────────

export interface DetailPageTab {
  id: string
  label: string
  icon?: React.ReactNode
  content: React.ReactNode
}

export interface DetailPageSidebar {
  content: React.ReactNode
  /** Width of sidebar in pixels. Default: 360 */
  width?: number
  /** Hide sidebar on screens smaller than this breakpoint. Default: 1024 */
  breakpoint?: number
}

export interface DetailPageAction {
  label: string
  icon?: React.ReactNode
  onClick: () => void
  variant?: 'default' | 'primary' | 'warning' | 'success' | 'danger'
  disabled?: boolean
  overflowOnly?: boolean // Force this action to appear in dropdown only
}

export interface HelpModalContent {
  title: string
  text: string
}

interface DetailPageTemplateProps {
  /** Back button label and callback */
  onBack?: () => void
  backLabel?: string

  /** Header card */
  icon?: React.ReactNode
  code?: string
  title: string
  badges?: React.ReactNode

  /** Tabs. Single tab = no tab bar chrome, just content. */
  tabs: DetailPageTab[]

  /** Action buttons (Edit, Hapus, etc.) */
  actions?: DetailPageAction[]

  /** Help modal content and callback — help modal is managed at page scope */
  helpTitle?: string
  helpText?: string
  onShowHelp?: (content: HelpModalContent) => void

  /** When true: hides action buttons, shows "Hanya Baca" pill in header card */
  readonly?: boolean
  /** When true: shows "Dikelola HQ" pill in header card */
  managedByHQ?: boolean

  /** Optional progress stepper shown below the title card */
  progress?: {
    steps: ProgressStep[]
    currentStatus: string
    terminalStatus?: TerminalStatus
  }

  /** When provided, appends a built-in "Koneksi" tab using DataConnectionWidget */
  connections?: {
    title?: string
    items: DataConnectionItem[]
  }

  /** Explicit dashboard context for DataConnectionWidget. If not provided, auto-detects from URL. */
  dashboardContext?: DashboardContext

  /** Right sidebar content */
  sidebar?: DetailPageSidebar
}

// ─── Component ────────────────────────────────────────────────────────────────

export function DetailPageTemplate({
  onBack,
  backLabel = 'Kembali',
  icon,
  code,
  title,
  badges,
  tabs,
  actions = [],
  helpTitle,
  helpText,
  onShowHelp,
  readonly: isReadonly = false,
  managedByHQ = false,
  progress,
  connections,
  dashboardContext: explicitContext,
  sidebar,
}: DetailPageTemplateProps) {
  const moduleAccess = useModuleAccess()
  const effectiveReadonly = isReadonly || moduleAccess.readonly
  const effectiveManagedByHQ = managedByHQ || moduleAccess.managedByHQ

  const detectedContext = useDashboardContext()
  const context = explicitContext ?? detectedContext

  const koneksiTab: DetailPageTab = {
    id: '__koneksi__',
    label: 'Koneksi',
    icon: <Link2 size={14} />,
    content: (
      <DataConnectionWidget
        title={connections?.title}
        items={connections?.items ?? []}
        dashboardContext={context}
      />
    ),
  }

  const allTabs = [...tabs, koneksiTab]

  const [activeTab, setActiveTab] = useState(tabs[0]?.id ?? '__log__')
  const [overflowOpen, setOverflowOpen] = useState(false)
  const overflowRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (overflowRef.current && !overflowRef.current.contains(e.target as Node)) {
        setOverflowOpen(false)
      }
    }
    if (overflowOpen) document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [overflowOpen])

  const activeContent = allTabs.find((t) => t.id === activeTab)?.content
  const hasMultipleTabs = allTabs.length > 1
  const hasHelp = !!(helpTitle || helpText)
  const autoBreadcrumbs = useAutoBreadcrumbs(title)
  const handleBack = onBack ?? (() => window.history.back())

  // ─── Action split: visible vs overflow ────────────────────────────────────
  const overflowOnlyActions = actions.filter((a) => a.overflowOnly)
  const regularActions = actions.filter((a) => !a.overflowOnly)

  const dangerActions = regularActions.filter((a) => a.variant === 'danger')
  const nonDangerActions = regularActions.filter((a) => a.variant !== 'danger')

  let visibleActions: DetailPageAction[]
  let overflowActions: DetailPageAction[] = [...overflowOnlyActions]

  if (dangerActions.length === 0 && regularActions.length <= 2) {
    visibleActions = regularActions
  } else if (nonDangerActions.length <= 2) {
    visibleActions = nonDangerActions
    overflowActions = [...overflowActions, ...dangerActions]
  } else {
    visibleActions = nonDangerActions.slice(0, 2)
    overflowActions = [...overflowActions, ...nonDangerActions.slice(2), ...dangerActions]
  }

  return (
    <div className={styles.page}>
      {/* Header card — breadcrumbs + identity + actions (mirrors PageHeader style) */}
      <div className={styles.header}>
        <div className={styles.breadcrumbRow}>
          <button className={styles.backBtn} onClick={handleBack}>
            <ArrowLeft size={14} />
            {backLabel}
          </button>
          <nav className={styles.breadcrumbs} aria-label="Breadcrumb">
            {autoBreadcrumbs.map((crumb, i) => (
              <span key={i} className={styles.crumbItem}>
                {i === 0 && <Home size={12} />}
                {crumb.href ? (
                  <Link to={crumb.href} className={styles.crumbLink}>{crumb.label}</Link>
                ) : (
                  <span className={styles.crumbCurrent}>{crumb.label}</span>
                )}
                {i < autoBreadcrumbs.length - 1 && (
                  <ChevronRight size={12} className={styles.separator} />
                )}
              </span>
            ))}
          </nav>
        </div>

        <div className={styles.headerContent}>
          <div className={styles.headerLeft}>
            {icon && <div className={styles.itemIcon}>{icon}</div>}
            <div className={styles.itemMeta}>
              {code && <span className={styles.itemCode}>{code}</span>}
              <h1 className={styles.itemTitle}>{title}</h1>
              {(badges || effectiveReadonly || effectiveManagedByHQ) && (
                <div className={styles.badges}>
                  {badges}
                  <StatusPills readonly={effectiveReadonly} managedByHQ={effectiveManagedByHQ} />
                </div>
              )}
            </div>
          </div>

          <div className={styles.headerRight}>
            {hasHelp && (
              <button
                className={styles.helpBtn}
                onClick={() => onShowHelp?.({ title: helpTitle ?? title, text: helpText ?? '' })}
                title="Bantuan"
              >
                <HelpCircle size={16} />
              </button>
            )}
            {actions.length > 0 && !effectiveReadonly && !effectiveManagedByHQ && (
              <div className={styles.actions}>
                {visibleActions.map((action) => (
                  <button
                    key={action.label}
                    className={
                      action.variant === 'primary' ? styles.btnPrimary
                      : action.variant === 'warning' ? styles.btnWarning
                      : action.variant === 'success' ? styles.btnSuccess
                      : styles.btnSecondary
                    }
                    onClick={action.onClick}
                    disabled={action.disabled}
                  >
                    {action.icon}
                    {action.label}
                  </button>
                ))}
                {overflowActions.length > 0 && (
                  <div className={styles.overflowWrap} ref={overflowRef}>
                    <button
                      className={`${styles.btnSecondary} ${styles.overflowBtn}`}
                      onClick={() => setOverflowOpen((v) => !v)}
                      title="Aksi lainnya"
                    >
                      <MoreHorizontal size={15} />
                    </button>
                    {overflowOpen && (
                      <div className={styles.dropdown}>
                        {overflowActions.map((action, i) => {
                          const isDanger = action.variant === 'danger'
                          const prevIsDanger = i > 0 && overflowActions[i - 1].variant === 'danger'
                          const showDivider = isDanger && !prevIsDanger && i > 0
                          return (
                            <div key={action.label}>
                              {showDivider && <div className={styles.dropdownDivider} />}
                              <button
                                className={`${styles.dropdownItem} ${isDanger ? styles.dropdownItemDanger : ''}`}
                                onClick={() => { setOverflowOpen(false); action.onClick() }}
                                disabled={action.disabled}
                              >
                                {action.icon}
                                {action.label}
                              </button>
                            </div>
                          )
                        })}
                      </div>
                    )}
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Progress stepper */}
      {progress && (
        <div className={styles.progressCard}>
          <ProgressWidget
            steps={progress.steps}
            currentStatus={progress.currentStatus}
            terminalStatus={progress.terminalStatus}
          />
        </div>
      )}

      {/* Tabs card */}
      <div className={styles.tabsCard}>
        {hasMultipleTabs && (
          <div className={styles.tabList}>
            {allTabs.map((tab) => (
              <button
                key={tab.id}
                className={`${styles.tab} ${activeTab === tab.id ? styles.tabActive : ''}`}
                onClick={() => setActiveTab(tab.id)}
              >
                {tab.icon}
                {tab.label}
              </button>
            ))}
          </div>
        )}
        <div className={styles.tabContentWrapper}>
          <div className={styles.tabContent}>{activeContent}</div>
          {sidebar && (
            <div
              className={styles.sidebar}
              data-breakpoint={sidebar.breakpoint ?? 1024}
              style={{
                width: `${sidebar.width ?? 360}px`,
              }}
            >
              {sidebar.content}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
