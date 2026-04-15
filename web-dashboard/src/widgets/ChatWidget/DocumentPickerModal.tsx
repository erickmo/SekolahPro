import { useState, useEffect } from 'react'
import { createPortal } from 'react-dom'
import {
  X,
  ShoppingCart,
  Package,
  FileText,
  Receipt,
  Truck,
  ClipboardList,
  FileSearch,
  CreditCard,
  Search,
} from 'lucide-react'
import { formatCurrency } from '@/utils/format'
import type { DocumentModule, LinkedDocument } from './chat.types'
import styles from './DocumentPickerModal.module.css'

const MODULE_LABELS: Record<DocumentModule, string> = {
  PO: 'Purchase Order',
  SO: 'Sales Order',
  INV: 'Faktur Jual',
  PINV: 'Faktur Beli',
  DO: 'Surat Jalan',
  PR: 'Permintaan Beli',
  RFQ: 'RFQ',
  PAYMENT: 'Pembayaran',
}

const MODULE_ICON: Record<DocumentModule, React.ReactNode> = {
  PO: <ShoppingCart size={14} />,
  SO: <Package size={14} />,
  INV: <FileText size={14} />,
  PINV: <Receipt size={14} />,
  DO: <Truck size={14} />,
  PR: <ClipboardList size={14} />,
  RFQ: <FileSearch size={14} />,
  PAYMENT: <CreditCard size={14} />,
}

const ALL_MODULES = Object.keys(MODULE_LABELS) as DocumentModule[]

const STATUS_LABEL: Record<string, string> = {
  confirmed: 'Confirmed',
  draft: 'Draft',
  sent: 'Terkirim',
  verified: 'Verified',
  approved: 'Approved',
  in_transit: 'Dalam Perjalanan',
  delivered: 'Diterima',
  completed: 'Selesai',
}

interface DocumentPickerModalProps {
  alreadyLinked: string[]
  onSelect: (doc: LinkedDocument) => void
  onClose: () => void
}

export function DocumentPickerModal({ alreadyLinked, onSelect, onClose }: DocumentPickerModalProps) {
  const [search, setSearch] = useState('')
  const [activeModule, setActiveModule] = useState<DocumentModule | 'ALL'>('ALL')

  useEffect(() => {
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }
    window.addEventListener('keydown', handleKey)
    return () => window.removeEventListener('keydown', handleKey)
  }, [onClose])

  const documents: LinkedDocument[] = []

  const filtered = documents.filter((doc) => {
    const matchModule = activeModule === 'ALL' || doc.module === activeModule
    const q = search.toLowerCase()
    const matchSearch =
      !q ||
      doc.code.toLowerCase().includes(q) ||
      (doc.counterpartyName?.toLowerCase().includes(q) ?? false)
    return matchModule && matchSearch
  })

  const modal = (
    <div className={styles.overlay} onClick={onClose}>
      <div className={styles.dialog} onClick={(e) => e.stopPropagation()}>
        {/* Header */}
        <div className={styles.header}>
          <span className={styles.headerTitle}>Lampirkan Dokumen</span>
          <button className={styles.closeBtn} onClick={onClose} aria-label="Tutup">
            <X size={15} />
          </button>
        </div>

        {/* Search */}
        <div className={styles.searchRow}>
          <Search size={14} className={styles.searchIcon} />
          <input
            className={styles.searchInput}
            placeholder="Cari nomor dokumen atau nama pihak..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            autoFocus
          />
        </div>

        {/* Module filter chips */}
        <div className={styles.filters}>
          <button
            className={`${styles.chip} ${activeModule === 'ALL' ? styles.chipActive : ''}`}
            onClick={() => setActiveModule('ALL')}
          >
            Semua
          </button>
          {ALL_MODULES.map((mod) => (
            <button
              key={mod}
              className={`${styles.chip} ${activeModule === mod ? styles.chipActive : ''}`}
              onClick={() => setActiveModule(mod)}
            >
              {mod}
            </button>
          ))}
        </div>

        {/* Document list */}
        <div className={styles.list}>
          {filtered.length === 0 ? (
            <p className={styles.empty}>Tidak ada dokumen ditemukan</p>
          ) : (
            filtered.map((doc) => {
              const isLinked = alreadyLinked.includes(doc.id)
              return (
                <div
                  key={doc.id}
                  className={`${styles.row} ${isLinked ? styles.rowLinked : ''}`}
                >
                  <span className={styles.moduleIcon}>{MODULE_ICON[doc.module]}</span>
                  <div className={styles.docInfo}>
                    <span className={styles.docCode}>{doc.code}</span>
                    {doc.counterpartyName && (
                      <span className={styles.docParty}>{doc.counterpartyName}</span>
                    )}
                  </div>
                  <div className={styles.docRight}>
                    {doc.total !== undefined && (
                      <span className={styles.docTotal}>{formatCurrency(doc.total, true)}</span>
                    )}
                    {doc.status && (
                      <span className={styles.docStatus}>
                        {STATUS_LABEL[doc.status] ?? doc.status}
                      </span>
                    )}
                  </div>
                  <button
                    className={`${styles.selectBtn} ${isLinked ? styles.selectBtnLinked : ''}`}
                    onClick={() => !isLinked && onSelect(doc)}
                    disabled={isLinked}
                  >
                    {isLinked ? 'Terpilih' : 'Pilih'}
                  </button>
                </div>
              )
            })
          )}
        </div>
      </div>
    </div>
  )

  return createPortal(modal, document.body)
}
