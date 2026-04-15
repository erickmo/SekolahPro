import { Lock, Building2 } from 'lucide-react'
import styles from './StatusPills.module.css'

export function ReadonlyPill() {
  return (
    <span className={`${styles.pill} ${styles.pillReadonly}`} title="Data ini tidak dapat diedit">
      <Lock size={11} strokeWidth={2.5} />
      Hanya Baca
    </span>
  )
}

export function ManagedByHQPill() {
  return (
    <span
      className={`${styles.pill} ${styles.pillHQ}`}
      title="Data ini dikelola oleh Kantor Pusat"
    >
      <Building2 size={11} strokeWidth={2.5} />
      Dikelola HQ
    </span>
  )
}

interface StatusPillsProps {
  readonly?: boolean
  managedByHQ?: boolean
}

/** Renders both pills inline — pass into header slots */
export function StatusPills({ readonly: isReadonly, managedByHQ }: StatusPillsProps) {
  if (!isReadonly && !managedByHQ) return null
  return (
    <>
      {isReadonly && <ReadonlyPill />}
      {managedByHQ && <ManagedByHQPill />}
    </>
  )
}
