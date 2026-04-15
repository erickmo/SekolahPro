import { useState } from 'react'
import { Link } from 'react-router-dom'
import { ArrowLeft, Save, HelpCircle, X, AlertCircle, ChevronRight, Home } from 'lucide-react'
import { useModuleAccess } from '@/hooks/useModuleAccess'
import { useAutoBreadcrumbs } from '@/hooks/useAutoBreadcrumbs'
import { StatusPills } from '@/widgets/StatusPills/StatusPills'
import styles from './FormPageTemplate.module.css'

// ─── Types ────────────────────────────────────────────────────────────────────

export interface FormPageTab {
  id: string
  label: string
  icon?: React.ReactNode
  content: React.ReactNode
}

export interface FormPageAction {
  label: string
  icon?: React.ReactNode
  onClick: () => void
  disabled?: boolean
  variant?: 'default' | 'warning' | 'success'
}

interface FormPageTemplateProps {
  /** Page title: "Tambah Pelanggan" or "Edit Pelanggan" */
  title: string

  /** Optional icon shown in header left */
  icon?: React.ReactNode
  /** Optional code/subtitle shown above title */
  code?: string
  /** Optional badges shown below title */
  badges?: React.ReactNode

  /** Back button */
  onBack: () => void
  backLabel?: string

  /** Tabs. Single tab = no tab bar chrome, just content inside card. */
  tabs: FormPageTab[]

  /** Form submission */
  onSubmit: (e: React.FormEvent) => void
  onCancel: () => void
  isSubmitting?: boolean
  submitLabel?: string

  /** Extra action buttons rendered before Cancel in the submit bar */
  extraActions?: FormPageAction[]

  /** Server-side error banner shown inside the form */
  serverError?: string

  /** Help modal content — shown when user clicks (?) */
  helpTitle?: string
  helpText?: string

  /** When true: hides submit bar (form is view-only), shows "Hanya Baca" pill */
  readonly?: boolean
  /** When true: shows "Dikelola HQ" pill */
  managedByHQ?: boolean
  /** Content to render above the tabs card */
  beforeTabsContent?: React.ReactNode
}

// ─── Component ────────────────────────────────────────────────────────────────

export function FormPageTemplate({
  title,
  icon,
  code,
  badges,
  onBack,
  backLabel = 'Kembali',
  tabs,
  onSubmit,
  onCancel,
  isSubmitting = false,
  submitLabel = 'Simpan',
  extraActions = [],
  serverError,
  helpTitle,
  helpText,
  readonly: isReadonly = false,
  managedByHQ = false,
  beforeTabsContent,
}: FormPageTemplateProps) {
  const moduleAccess = useModuleAccess()
  const effectiveReadonly = isReadonly || moduleAccess.readonly
  const effectiveManagedByHQ = managedByHQ || moduleAccess.managedByHQ

  const [activeTab, setActiveTab] = useState(tabs[0]?.id ?? '')
  const [showHelp, setShowHelp] = useState(false)

  const activeContent = tabs.find((t) => t.id === activeTab)?.content
  const hasMultipleTabs = tabs.length > 1
  const hasHelp = !!(helpTitle || helpText)
  const autoBreadcrumbs = useAutoBreadcrumbs(title)

  return (
    <div className={styles.page}>
      {/* Header card — breadcrumbs + identity + actions (mirrors DetailPageTemplate) */}
      <div className={styles.header}>
        <div className={styles.breadcrumbRow}>
          <button className={styles.backBtn} onClick={onBack} type="button">
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
                type="button"
                onClick={() => setShowHelp(true)}
                title="Bantuan"
              >
                <HelpCircle size={16} />
              </button>
            )}
          </div>
        </div>
      </div>

      <form onSubmit={onSubmit} noValidate>
        {/* Server error banner */}
        {serverError && (
          <div className={styles.errorBanner}>
            <AlertCircle size={16} />
            <span>{serverError}</span>
          </div>
        )}

        {/* Content above tabs card */}
        {beforeTabsContent}

        {/* Tabs card */}
        <div className={styles.tabsCard}>
          {hasMultipleTabs && (
            <div className={styles.tabList}>
              {tabs.map((tab) => (
                <button
                  key={tab.id}
                  type="button"
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

        {/* Submit bar — hidden when readonly or managed by HQ */}
        {!effectiveReadonly && !effectiveManagedByHQ && <div className={styles.submitBar}>
          {extraActions.length > 0 && (
            <div className={styles.extraActions}>
              {extraActions.map((action) => (
                <button
                  key={action.label}
                  type="button"
                  className={
                    action.variant === 'warning' ? styles.btnWarning
                    : action.variant === 'success' ? styles.btnSuccess
                    : styles.btnSecondary
                  }
                  onClick={action.onClick}
                  disabled={action.disabled ?? isSubmitting}
                >
                  {action.icon}
                  {action.label}
                </button>
              ))}
            </div>
          )}
          <button
            type="button"
            className={styles.btnSecondary}
            onClick={onCancel}
            disabled={isSubmitting}
          >
            Batal
          </button>
          <button
            type="submit"
            className={styles.btnPrimary}
            disabled={isSubmitting}
          >
            <Save size={14} />
            {isSubmitting ? 'Menyimpan...' : submitLabel}
          </button>
        </div>}
      </form>

      {/* Help modal */}
      {showHelp && hasHelp && (
        <div className={styles.overlay} onClick={() => setShowHelp(false)}>
          <div className={styles.helpModal} onClick={(e) => e.stopPropagation()}>
            <div className={styles.helpModalHeader}>
              <div className={styles.helpModalIcon}>
                <HelpCircle size={20} />
              </div>
              <h3 className={styles.helpModalTitle}>{helpTitle ?? title}</h3>
              <button className={styles.helpModalClose} onClick={() => setShowHelp(false)}>
                <X size={16} />
              </button>
            </div>
            <p className={styles.helpModalBody}>{helpText}</p>
          </div>
        </div>
      )}
    </div>
  )
}
