import { useState, useEffect } from 'react'
import { X } from 'lucide-react'
import { FilterPanel } from './FilterPanel'
import type { ActiveFilter, FilterDef } from './filter.types'
import styles from './FilterDialog.module.css'

interface FilterDialogProps {
  defs: FilterDef[]
  activeFilters: ActiveFilter[]
  onSave: (filters: ActiveFilter[]) => void
  onClose: (saved: boolean) => void
}

export function FilterDialog({ defs, activeFilters, onSave, onClose }: FilterDialogProps) {
  const [pending, setPending] = useState<ActiveFilter[]>(() => activeFilters)

  // Sinkronisasi pending saat activeFilters berubah (mis. dari URL query string)
  // HANYA sinkronisasi saat activeFilters berubah, bukan saat pending berubah
  useEffect(() => {
    // Cek apakah activeFilters berbeda dari pending
    const isSame =
      activeFilters.length === pending.length &&
      activeFilters.every(
        (af, i) =>
          af.id === pending[i]?.id &&
          af.key === pending[i]?.key &&
          af.operator === pending[i]?.operator &&
          JSON.stringify(af.value) === JSON.stringify(pending[i]?.value),
      )
    if (!isSame) {
      setPending(activeFilters)
    }
  }, [activeFilters])

  function handleSave() {
    onSave(pending)
    onClose(true)
  }

  return (
    <div className={styles.overlay} onClick={() => onClose(false)}>
      <div className={styles.dialog} onClick={(e) => e.stopPropagation()}>
        <div className={styles.header}>
          <h3 className={styles.title}>Filter Data</h3>
          <button className={styles.closeBtn} onClick={() => onClose(false)} title="Tutup">
            <X size={16} />
          </button>
        </div>
        <div className={styles.body}>
          <FilterPanel
            defs={defs}
            activeFilters={pending}
            onChange={setPending}
            className={styles.panelBare}
          />
        </div>
        <div className={styles.footer}>
          <div className={styles.footerActions}>
            <button className={styles.btnSecondary} onClick={() => onClose(false)}>Batal</button>
            <button className={styles.btnPrimary} onClick={handleSave}>
              Tambahkan Filter
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
