import { Check, X } from 'lucide-react'
import styles from './ProgressWidget.module.css'

// ─── Types ────────────────────────────────────────────────────────────────────

export interface ProgressStep {
  /** Harus cocok dengan nilai currentStatus */
  id: string
  label: string
}

export interface TerminalStatus {
  label: string
  /** 'danger' = dibatalkan/ditolak, 'warning' = kadaluarsa */
  variant: 'danger' | 'warning'
}

interface ProgressWidgetProps {
  /** Urutan langkah-langkah dalam alur dokumen */
  steps: ProgressStep[]
  /** Status dokumen saat ini — harus cocok dengan salah satu step.id */
  currentStatus: string
  /**
   * Isi jika dokumen berada di status terminal negatif (dibatalkan, ditolak, kadaluarsa).
   * Akan ditampilkan sebagai node akhir dengan warna merah/kuning setelah langkah terakhir.
   */
  terminalStatus?: TerminalStatus
  className?: string
}

type StepState = 'completed' | 'current' | 'pending'

// ─── Component ────────────────────────────────────────────────────────────────

export function ProgressWidget({
  steps,
  currentStatus,
  terminalStatus,
  className,
}: ProgressWidgetProps) {
  const currentIndex = steps.findIndex((s) => s.id === currentStatus)
  const isTerminal = !!terminalStatus

  return (
    <div className={`${styles.root} ${className ?? ''}`} role="list" aria-label="Status dokumen">
      {steps.map((step, idx) => {
        const state: StepState =
          idx < currentIndex ? 'completed'
          : idx === currentIndex ? 'current'
          : 'pending'

        const isLast = idx === steps.length - 1
        const showConnector = !isLast || isTerminal

        return (
          <div key={step.id} className={styles.stepSlot} role="listitem">
            <_StepNode step={step} state={state} />
            {showConnector && (
              <_Connector
                filled={state === 'completed' || (state === 'current' && !isLast)}
              />
            )}
          </div>
        )
      })}

      {isTerminal && (
        <div className={styles.stepSlot} role="listitem">
          <_TerminalNode status={terminalStatus!} />
        </div>
      )}
    </div>
  )
}

// ─── Sub-components ───────────────────────────────────────────────────────────

interface StepNodeProps {
  step: ProgressStep
  state: StepState
}

function _StepNode({ step, state }: StepNodeProps) {
  return (
    <div className={styles.nodeWrapper}>
      <div className={`${styles.node} ${styles[`node_${state}`]}`} aria-hidden="true">
        {state === 'completed' && <Check size={12} strokeWidth={3} />}
      </div>
      <span
        className={`${styles.label} ${styles[`label_${state}`]}`}
        title={step.label}
      >
        {step.label}
      </span>
    </div>
  )
}

interface TerminalNodeProps {
  status: TerminalStatus
}

function _TerminalNode({ status }: TerminalNodeProps) {
  return (
    <div className={styles.nodeWrapper}>
      <div
        className={`${styles.node} ${styles.node_terminal} ${styles[`node_terminal_${status.variant}`]}`}
        aria-hidden="true"
      >
        <X size={12} strokeWidth={3} />
      </div>
      <span
        className={`${styles.label} ${styles[`label_terminal_${status.variant}`]}`}
        title={status.label}
      >
        {status.label}
      </span>
    </div>
  )
}

interface ConnectorProps {
  filled: boolean
}

function _Connector({ filled }: ConnectorProps) {
  return (
    <div className={`${styles.connector} ${filled ? styles.connectorFilled : ''}`} aria-hidden="true" />
  )
}
