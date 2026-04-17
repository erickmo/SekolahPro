import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { healthRecordService } from '@/services/sprint5.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { HealthRecord, HealthRecordType } from '@/types/sprint5.types'
import { HEALTH_RECORD_TYPE_LABELS } from '@/types/sprint5.types'
import styles from './HealthRecordPage.module.css'

const PAGE_SIZE = 20

const RECORD_TYPE_OPTIONS: { value: HealthRecordType | ''; label: string }[] = [
  { value: '', label: 'Semua Jenis Catatan' },
  ...Object.entries(HEALTH_RECORD_TYPE_LABELS).map(([value, label]) => ({
    value: value as HealthRecordType,
    label,
  })),
]

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

export default function HealthRecordPage() {
  const [search, setSearch] = useState('')
  const [recordType, setRecordType] = useState<HealthRecordType | ''>('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { student_name: debouncedSearch } : {}),
    ...(recordType ? { record_type: recordType } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.healthRecord, filters],
    queryFn: () => healthRecordService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => healthRecordService.delete(id),
    onSuccess: () => {
      toast.success('Catatan kesehatan berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.healthRecord] })
    },
    onError: () => {
      toast.error('Gagal menghapus catatan kesehatan')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || recordType !== ''

  function handleReset() {
    setSearch('')
    setRecordType('')
    setPage(1)
  }

  function handleDelete(item: HealthRecord) {
    if (!confirm(`Hapus catatan kesehatan "${item.studentName}"?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Catatan Kesehatan</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} data ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Tambah Catatan
        </button>
      </div>

      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Cari nama siswa..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          />
        </div>
        <select
          className={styles.select}
          value={recordType}
          onChange={(e) => { setRecordType(e.target.value as HealthRecordType | ''); setPage(1) }}
          aria-label="Filter jenis catatan"
        >
          {RECORD_TYPE_OPTIONS.map((opt) => (
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
                <th className={styles.th}>Siswa</th>
                <th className={styles.th}>NIS</th>
                <th className={styles.th}>Jenis</th>
                <th className={styles.th}>Tanggal</th>
                <th className={styles.th}>Diagnosis</th>
                <th className={styles.th}>Tindakan</th>
                <th className={styles.th}>Dokter</th>
                <th className={styles.th}>Aksi</th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                Array.from({ length: 5 }).map((_, i) => (
                  <tr key={i} className={styles.skeletonRow}>
                    {Array.from({ length: 8 }).map((_, j) => (
                      <td key={j}><span className={styles.skeleton} style={{ width: '80px' }} /></td>
                    ))}
                  </tr>
                ))
              ) : items.length === 0 ? (
                <tr>
                  <td colSpan={8} className={styles.emptyCell}>
                    <div className={styles.emptyState}>
                      <p className={styles.emptyTitle}>Tidak ada data</p>
                      <p className={styles.emptySubtitle}>Belum ada catatan kesehatan.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.name}>{item.studentName || '-'}</span></td>
                    <td className={styles.td}><span className={styles.code}>{item.studentNis || '-'}</span></td>
                    <td className={styles.td}>
                      <span className={styles.badgeType}>
                        {HEALTH_RECORD_TYPE_LABELS[item.recordType] ?? item.recordType}
                      </span>
                    </td>
                    <td className={styles.td}>{formatDate(item.recordDate)}</td>
                    <td className={styles.td}>{item.diagnosis || '-'}</td>
                    <td className={styles.td}>{item.treatment || '-'}</td>
                    <td className={styles.td}>{item.doctorName || '-'}</td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus catatan kesehatan ${item.studentName}`}
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
