import { Check } from 'lucide-react'
import styles from './Stepper.module.css'

export interface StepItem {
  key: string
  label: string
  description?: string
}

interface StepperProps {
  steps: StepItem[]
  activeStep: number   // 0-based index
  orientation?: 'horizontal' | 'vertical'
}

type StepStatus = 'completed' | 'active' | 'pending'

function getStatus(index: number, activeStep: number): StepStatus {
  if (index < activeStep) return 'completed'
  if (index === activeStep) return 'active'
  return 'pending'
}

export function Stepper({ steps, activeStep, orientation = 'horizontal' }: StepperProps) {
  return (
    <ol className={`${styles.stepper} ${styles[orientation]}`} aria-label="Progress langkah">
      {steps.map((step, index) => {
        const status = getStatus(index, activeStep)
        const isLast = index === steps.length - 1

        return (
          <li key={step.key} className={`${styles.step} ${styles[status]}`}>
            <div className={styles.stepHead}>
              <div className={styles.circle} aria-current={status === 'active' ? 'step' : undefined}>
                {status === 'completed' ? (
                  <Check size={14} strokeWidth={3} />
                ) : (
                  <span className={styles.number}>{index + 1}</span>
                )}
              </div>

              {!isLast && <div className={styles.connector} aria-hidden="true" />}
            </div>

            <div className={styles.stepBody}>
              <span className={styles.label}>{step.label}</span>
              {step.description && (
                <span className={styles.description}>{step.description}</span>
              )}
            </div>
          </li>
        )
      })}
    </ol>
  )
}

// ─── Stepper Navigation Buttons ───────────────────────────────────────────────

interface StepperActionsProps {
  activeStep: number
  totalSteps: number
  onBack: () => void
  onNext: () => void
  onFinish: () => void
  nextLabel?: string
  backLabel?: string
  finishLabel?: string
  isLoading?: boolean
  canProceed?: boolean
}

export function StepperActions({
  activeStep,
  totalSteps,
  onBack,
  onNext,
  onFinish,
  nextLabel = 'Lanjut',
  backLabel = 'Kembali',
  finishLabel = 'Selesai',
  isLoading = false,
  canProceed = true,
}: StepperActionsProps) {
  const isFirst = activeStep === 0
  const isLast = activeStep === totalSteps - 1

  return (
    <div className={styles.actions}>
      <button
        type="button"
        className={styles.btnBack}
        onClick={onBack}
        disabled={isFirst || isLoading}
      >
        {backLabel}
      </button>
      <button
        type="button"
        className={styles.btnNext}
        onClick={isLast ? onFinish : onNext}
        disabled={!canProceed || isLoading}
      >
        {isLoading ? 'Memproses...' : isLast ? finishLabel : nextLabel}
      </button>
    </div>
  )
}
