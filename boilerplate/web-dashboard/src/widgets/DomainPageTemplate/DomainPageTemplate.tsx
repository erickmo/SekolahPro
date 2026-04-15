import { useState, useMemo } from 'react'
import { Link } from 'react-router-dom'
import {
  ArrowLeft, Check, X, HelpCircle, AlertTriangle,
  Printer, FileDown, ExternalLink, Link2, Clock,
} from 'lucide-react'
import { StatusPills } from '@/widgets/StatusPills/StatusPills'
import { formatRelative } from '@/utils/format'
import styles from './DomainPageTemplate.module.css'

// ─── Types ────────────────────────────────────────────────────────────────────

export interface DomainPageTab {
  id: string
  label: string
  icon?: React.ReactNode
  content: React.ReactNode
}

export interface DomainStatusStep {
  key: string
  label: string
}

export interface DomainSummaryItem {
  label: string
  value: React.ReactNode
  variant?: 'default' | 'money' | 'success' | 'warning' | 'danger'
}

export interface DomainDetailAction {
  label: string
  icon?: React.ReactNode
  onClick: () => void | Promise<void>
  variant?: 'default' | 'primary' | 'warning' | 'success' | 'danger'
  disabled?: boolean
  confirmDialog?: {
    title: string
    body: React.ReactNode
    confirmLabel?: string
    variant?: 'danger' | 'warning'
  }
}

export interface DomainRelatedDoc {
  code: string
  title: string
  href: string
  badge?: React.ReactNode
}

export interface DomainRelatedDocGroup {
  label: string
  items: DomainRelatedDoc[]
}

export interface DomainActivityEntry {
  id?: string
  actor: string
  initials: string
  action: string
  timestamp: string
  note?: string
}

interface DomainPageTemplateProps {
  onBack: () => void
  backLabel?: string

  icon?: React.ReactNode
  code?: string
  title: string
  badges?: React.ReactNode

  /** Horizontal step flow bar shown below header card */
  statusFlow?: { steps: DomainStatusStep[]; current: string }

  /** Compact metric strip shown below status flow */
  summaryItems?: DomainSummaryItem[]

  tabs: DomainPageTab[]

  /** Map of status → actions. Only current status's actions are rendered. */
  currentStatus?: string
  statusActions?: Partial<Record<string, DomainDetailAction[]>>

  onPrint?: () => void
  onExportPdf?: () => void

  /** Auto-injects "Dokumen Terkait" tab at the end */
  relatedDocs?: DomainRelatedDocGroup[]

  /** Auto-injects "Riwayat" tab at the very end */
  activityLog?: DomainActivityEntry[]

  helpTitle?: string
  helpText?: string

  readonly?: boolean
  managedByHQ?: boolean
  intercompany?: boolean
}

// ─── Status Flow Bar ──────────────────────────────────────────────────────────

function StatusFlowBar({ steps, current }: { steps: DomainStatusStep[]; current: string }) {
  const currentIdx = Math.max(0, steps.findIndex((s) => s.key === current))

  return (
    <div className={styles.statusFlow}>
      {steps.map((step, idx) => {
        const isPast = idx < currentIdx
        const isCurrent = idx === currentIdx
        const isLast = idx === steps.length - 1
        return (
          <div key={step.key} className={styles.stepWrapper}>
            <div className={styles.stepItem}>
              <div
                className={`${styles.stepCircle} ${
                  isPast ? styles.stepPast
                  : isCurrent ? styles.stepCurrent
                  : styles.stepFuture
                }`}
              >
                {isPast && <Check size={11} strokeWidth={3} />}
                {isCurrent && <span className={styles.stepDot} />}
              </div>
              <span
                className={`${styles.stepLabel} ${
                  !isPast && !isCurrent ? styles.stepLabelMuted : ''
                }`}
              >
                {step.label}
              </span>
            </div>
            {!isLast && (
              <div
                className={`${styles.connector} ${
                  isPast ? styles.connectorFilled : styles.connectorEmpty
                }`}
              />
            )}
          </div>
        )
      })}
    </div>
  )
}

// ─── Summary Strip ────────────────────────────────────────────────────────────

const SUMMARY_VALUE_CLASS: Record<string, string> = {
  money: styles.summaryMoney,
  success: styles.summarySuccess,
  warning: styles.summaryWarning,
  danger: styles.summaryDanger,
}

function SummaryStrip({ items }: { items: DomainSummaryItem[] }) {
  return (
    <div className={styles.summaryStrip}>
      {items.map((item, idx) => (
        <div key={idx} className={styles.summaryItem}>
          <span className={styles.summaryLabel}>{item.label}</span>
          <span
            className={`${styles.summaryValue} ${
              item.variant ? (SUMMARY_VALUE_CLASS[item.variant] ?? '') : ''
            }`}
          >
            {item.value}
          </span>
          {idx < items.length - 1 && <div className={styles.summarySep} />}
        </div>
      ))}
    </div>
  )
}

// ─── Related Docs Tab Content ─────────────────────────────────────────────────

function RelatedDocsContent({ groups }: { groups: DomainRelatedDocGroup[] }) {
  if (groups.length === 0) {
    return <p className={styles.emptyTabText}>Tidak ada dokumen terkait</p>
  }
  return (
    <div className={styles.relatedGroups}>
      {groups.map((group) => (
        <div key={group.label} className={styles.relatedGroup}>
          <p className={styles.relatedGroupLabel}>{group.label}</p>
          {group.items.length === 0 ? (
            <p className={styles.relatedEmpty}>Tidak ada</p>
          ) : (
            <div className={styles.relatedList}>
              {group.items.map((doc) => (
                <Link key={doc.code} to={doc.href} className={styles.relatedItem}>
                  <span className={styles.relatedCode}>{doc.code}</span>
                  <span className={styles.relatedTitle}>{doc.title}</span>
                  {doc.badge && <span className={styles.relatedBadge}>{doc.badge}</span>}
                  <ExternalLink size={12} className={styles.relatedIcon} />
                </Link>
              ))}
            </div>
          )}
        </div>
      ))}
    </div>
  )
}

// ─── Activity Timeline Tab Content ────────────────────────────────────────────

function ActivityTimelineContent({ entries }: { entries: DomainActivityEntry[] }) {
  if (entries.length === 0) {
    return <p className={styles.emptyTabText}>Belum ada riwayat aktivitas</p>
  }
  return (
    <div className={styles.timeline}>
      {entries.map((entry, idx) => (
        <div key={entry.id ?? idx} className={styles.timelineItem}>
          <div className={styles.timelineLeft}>
            <div className={styles.timelineAvatar}>{entry.initials}</div>
            {idx < entries.length - 1 && <div className={styles.timelineLine} />}
          </div>
          <div className={styles.timelineContent}>
            <p className={styles.timelineHeader}>
              <span className={styles.timelineActor}>{entry.actor}</span>
              <span className={styles.timelineAction}>{entry.action}</span>
              <span className={styles.timelineTime}>{formatRelative(entry.timestamp)}</span>
            </p>
            {entry.note && <p className={styles.timelineNote}>{entry.note}</p>}
          </div>
        </div>
      ))}
    </div>
  )
}

// ─── Confirm Dialog ───────────────────────────────────────────────────────────

interface ConfirmDialogProps {
  action: DomainDetailAction
  isLoading: boolean
  onConfirm: () => void
  onCancel: () => void
}

function ConfirmDialog({ action, isLoading, onConfirm, onCancel }: ConfirmDialogProps) {
  const cfg = action.confirmDialog!
  const isDanger = cfg.variant === 'danger'
  return (
    <div className={styles.overlay} onClick={() => !isLoading && onCancel()}>
      <div className={styles.confirmDialog} onClick={(e) => e.stopPropagation()}>
        <div className={`${styles.confirmIcon} ${isDanger ? styles.confirmIconDanger : styles.confirmIconWarning}`}>
          <AlertTriangle size={22} />
        </div>
        <h3 className={styles.confirmTitle}>{cfg.title}</h3>
        <div className={styles.confirmBody}>{cfg.body}</div>
        <div className={styles.confirmActions}>
          <button className={styles.btnSecondary} onClick={onCancel} disabled={isLoading}>
            Batal
          </button>
          <button
            className={isDanger ? styles.btnDanger : styles.btnWarning}
            onClick={onConfirm}
            disabled={isLoading}
          >
            {isLoading ? 'Memproses...' : (cfg.confirmLabel ?? 'Ya, Lanjutkan')}
          </button>
        </div>
      </div>
    </div>
  )
}

// ─── Help Modal ───────────────────────────────────────────────────────────────

function HelpModal({ title, text, onClose }: { title: string; text: string; onClose: () => void }) {
  return (
    <div className={styles.overlay} onClick={onClose}>
      <div className={styles.helpModal} onClick={(e) => e.stopPropagation()}>
        <div className={styles.helpModalHeader}>
          <div className={styles.helpModalIcon}><HelpCircle size={18} /></div>
          <h3 className={styles.helpModalTitle}>{title}</h3>
          <button className={styles.helpModalClose} onClick={onClose}><X size={16} /></button>
        </div>
        <p className={styles.helpModalBody}>{text}</p>
      </div>
    </div>
  )
}

// ─── Main Component ───────────────────────────────────────────────────────────

export function DomainPageTemplate({
  onBack,
  backLabel = 'Kembali',
  icon,
  code,
  title,
  badges,
  statusFlow,
  summaryItems,
  tabs,
  currentStatus,
  statusActions,
  onPrint,
  onExportPdf,
  relatedDocs,
  activityLog,
  helpTitle,
  helpText,
  readonly: isReadonly = false,
  managedByHQ = false,
  intercompany = false,
}: DomainPageTemplateProps) {
  const [activeTab, setActiveTab] = useState(tabs[0]?.id ?? '')
  const [showHelp, setShowHelp] = useState(false)
  const [pendingAction, setPendingAction] = useState<DomainDetailAction | null>(null)
  const [isActionLoading, setIsActionLoading] = useState(false)

  const visibleActions: DomainDetailAction[] = useMemo(() => {
    if (isReadonly || managedByHQ || !statusActions || !currentStatus) return []
    return statusActions[currentStatus] ?? []
  }, [isReadonly, statusActions, currentStatus])

  const allTabs: DomainPageTab[] = useMemo(() => {
    const extras: DomainPageTab[] = []
    if (relatedDocs?.length) {
      extras.push({
        id: '__related',
        label: 'Dokumen Terkait',
        icon: <Link2 size={14} />,
        content: <RelatedDocsContent groups={relatedDocs} />,
      })
    }
    if (activityLog?.length) {
      extras.push({
        id: '__activity',
        label: 'Riwayat',
        icon: <Clock size={14} />,
        content: <ActivityTimelineContent entries={activityLog} />,
      })
    }
    return [...tabs, ...extras]
  }, [tabs, relatedDocs, activityLog])

  const hasHelp = !!(helpTitle || helpText)
  const activeContent = allTabs.find((t) => t.id === activeTab)?.content
  const hasMultipleTabs = allTabs.length > 1

  async function handleActionClick(action: DomainDetailAction) {
    if (action.confirmDialog) {
      setPendingAction(action)
      return
    }
    await action.onClick()
  }

  async function handleConfirm() {
    if (!pendingAction) return
    setIsActionLoading(true)
    try {
      await pendingAction.onClick()
    } finally {
      setIsActionLoading(false)
      setPendingAction(null)
    }
  }

  return (
    <div className={styles.page}>
      {/* Top bar */}
      <div className={styles.topBar}>
        <button className={styles.backBtn} onClick={onBack}>
          <ArrowLeft size={16} />
          {backLabel}
        </button>
        <div className={styles.topBarRight}>
          {onPrint && (
            <button className={styles.iconBtn} onClick={onPrint} title="Cetak">
              <Printer size={16} />
            </button>
          )}
          {onExportPdf && (
            <button className={styles.iconBtn} onClick={onExportPdf} title="Ekspor PDF">
              <FileDown size={16} />
            </button>
          )}
          {hasHelp && (
            <button className={styles.iconBtn} onClick={() => setShowHelp(true)} title="Bantuan">
              <HelpCircle size={16} />
            </button>
          )}
          {visibleActions.length > 0 && (
            <div className={styles.actions}>
              {visibleActions.map((action) => (
                <button
                  key={action.label}
                  className={
                    action.variant === 'danger' ? styles.btnDanger
                    : action.variant === 'primary' ? styles.btnPrimary
                    : action.variant === 'warning' ? styles.btnWarning
                    : action.variant === 'success' ? styles.btnSuccess
                    : styles.btnSecondary
                  }
                  onClick={() => handleActionClick(action)}
                  disabled={action.disabled ?? isActionLoading}
                >
                  {action.icon}
                  {action.label}
                </button>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Header card */}
      <div className={styles.card}>
        <div className={styles.cardHeader}>
          {icon && <div className={styles.itemIcon}>{icon}</div>}
          <div className={styles.itemMeta}>
            {code && <span className={styles.itemCode}>{code}</span>}
            <h1 className={styles.itemTitle}>{title}</h1>
            {(badges || isReadonly || managedByHQ || intercompany) && (
              <div className={styles.badges}>
                {badges}
                <StatusPills readonly={isReadonly} managedByHQ={managedByHQ} />
                {intercompany && (
                  <span className={styles.pillIntercompany}>Intercompany</span>
                )}
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Status flow bar */}
      {statusFlow && (
        <StatusFlowBar steps={statusFlow.steps} current={statusFlow.current} />
      )}

      {/* Summary strip */}
      {summaryItems && summaryItems.length > 0 && (
        <SummaryStrip items={summaryItems} />
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
        <div className={styles.tabContent}>{activeContent}</div>
      </div>

      {/* Confirm dialog */}
      {pendingAction?.confirmDialog && (
        <ConfirmDialog
          action={pendingAction}
          isLoading={isActionLoading}
          onConfirm={handleConfirm}
          onCancel={() => setPendingAction(null)}
        />
      )}

      {/* Help modal */}
      {showHelp && hasHelp && (
        <HelpModal
          title={helpTitle ?? title}
          text={helpText ?? ''}
          onClose={() => setShowHelp(false)}
        />
      )}
    </div>
  )
}
