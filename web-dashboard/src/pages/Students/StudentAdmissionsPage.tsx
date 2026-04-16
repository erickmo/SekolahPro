import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { studentAdmissionService } from '@/services/student.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { StudentAdmission, AdmissionStatus, AdmissionType } from '@/types/student.types'
import { ADMISSION_STATUS_LABELS, ADMISSION_TYPE_LABELS } from '@/types/student.types'
import styles from './StudentAdmissionsPage.module.css'

const PAGE_SIZE = 20

const STATUS_OPTIONS: { value: AdmissionStatus | ''; label: string }[] = [
  { value: '', label: 'Semua Status' },
  ...Object.entries(ADMISSION_STATUS_LABELS).map(([value, label]) => ({
    value: value as AdmissionStatus,
    label,
  })),
]

const ADMISSION_TYPE_OPTIONS: { value: AdmissionType | ''; label: string }[] = [
  { value: '', label: 'Semua' },
  ...Object.entries(ADMISSION_TYPE_LABELS).map(([value, label]) => ({
    value: value as AdmissionType,
    label,
  })),
]

function getAdmissionStatusClass(status: AdmissionStatus): string {
  switch (status) {
    case 'pending': return styles.badgePending
    case 'document_review': return styles.badgeDocumentReview
    case 'test': return styles.badgeTest
    case 'written_test': return styles.badgeWrittenTest
    case 'interview': return styles.badgeInterview
    case 'accepted': return styles.badgeAccepted
    case 'rejected': return styles.badgeRejected
    default: return styles.badgeDefault
  }
}

function getAdmissionTypeClass(type: AdmissionType): string {
  switch (type) {
    case 'regular': return styles.typeRegular
    case 'pindahan': return styles.typePindahan
    case 'afirmasi': return styles.typeAfirmasi
    case 'prestasi': return styles.typePrestasi
    default: return styles.badgeDefault
  }
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '-'
  try {
    return new Date(dateStr).toLocaleDateString('id-ID', {
      day: 'numeric',
      month: 'long',
      year: 'numeric',
    })
  } catch {
    return dateStr
  }
}

export default function StudentAdmissionsPage() {
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState<AdmissionStatus | ''>('')
  const [admissionType, setAdmissionType] = useState<AdmissionType | ''>('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { registration_number: debouncedSearch } : {}),
    ...(status ? { status } : {}),
    ...(admissionType ? { admission_type: admissionType } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.studentAdmissions, filters],
    queryFn: () => studentAdmissionService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => studentAdmissionService.delete(id),
    onSuccess: () => {
      toast.success('Data pendaftaran berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.studentAdmissions] })
    },
    onError: () => {
      toast.error('Gagal menghapus data pendaftaran')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || status !== '' || admissionType !== ''

  function handleReset() {
    setSearch('')
    setStatus('')
    setAdmissionType('')
    setPage(1)
  }

  function handleDelete(item: StudentAdmission) {
    if (!confirm(`Hapus pendaftaran "${item.registrationNumber}"?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>PPDB - Penerimaan Peserta Didik Baru</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} data ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Tambah Pendaftaran
        </button>
      </div>

      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Cari no. pendaftaran / nama siswa..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          />
        </div>

        <div className={styles.segmentedControl}>
          {ADMISSION_TYPE_OPTIONS.map((opt) => (
            <button
              key={opt.value}
              type="button"
              className={`${styles.segmentBtn} ${admissionType === opt.value ? styles.segmentActive : ''}`}
              onClick={() => { setAdmissionType(opt.value); setPage(1) }}
            >
              {opt.label}
            </button>
          ))}
        </div>

        <select
          className={styles.select}
          value={status}
          onChange={(e) => { setStatus(e.target.value as AdmissionStatus | ''); setPage(1) }}
          aria-label="Filter status"
        >
          {STATUS_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
        {hasFilters && (
          <button type="button" className={styles.resetBtn} onClick={handleReset}>
            <RotateCcw size={14} /> Reset Filter
          </button>
        )}
      </div>

      <div className={styles.tableCard}>
        <div className={styles.tableScroll}>
          <table className={styles.table}>
            <thead>
              <tr>
                <th className={styles.th}>No. Pendaftaran</th>
                <th className={styles.th}>Nama</th>
                <th className={styles.th}>Jalur</th>
                <th className={styles.th}>Tgl Daftar</th>
                <th className={styles.th}>Status</th>
                <th className={styles.th}>Aksi</th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                Array.from({ length: 5 }).map((_, i) => (
                  <tr key={i} className={styles.skeletonRow}>
                    {Array.from({ length: 6 }).map((_, j) => (
                      <td key={j}><span className={styles.skeleton} style={{ width: '80px' }} /></td>
                    ))}
                  </tr>
                ))
              ) : items.length === 0 ? (
                <tr>
                  <td colSpan={6} className={styles.emptyCell}>
                    <div className={styles.emptyState}>
                      <p className={styles.emptyTitle}>Tidak ada data</p>
                      <p className={styles.emptySubtitle}>Belum ada data pendaftaran.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.code}>{item.registrationNumber || '-'}</span></td>
                    <td className={styles.td}><span className={styles.name}>{item.studentName || '-'}</span></td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getAdmissionTypeClass(item.admissionType)}`}>
                        {ADMISSION_TYPE_LABELS[item.admissionType] ?? item.admissionType}
                      </span>
                    </td>
                    <td className={styles.td}>{formatDate(item.registrationDate)}</td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getAdmissionStatusClass(item.status)}`}>
                        {ADMISSION_STATUS_LABELS[item.status] ?? item.status}
                      </span>
                    </td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus ${item.registrationNumber}`}
                      >
                        <Trash2 size={14} />
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {!isLoading && totalPages > 1 && (
          <div className={styles.pagination}>
            <button type="button" className={styles.pageBtn} disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
              Sebelumnya
            </button>
            <span className={styles.pageInfo}>Halaman {page} dari {totalPages}</span>
            <button type="button" className={styles.pageBtn} disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)}>
              Berikutnya
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
