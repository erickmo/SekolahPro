import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { leaveRequestService } from '@/services/sprint6.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { LeaveRequest, LeaveRequestStatus } from '@/types/sprint6.types'
import { LEAVE_REQUEST_STATUS_LABELS } from '@/types/sprint6.types'
import styles from './LeaveManagementPage.module.css'

const PAGE_SIZE = 20

const STATUS_OPTIONS: { value: LeaveRequestStatus | ''; label: string }[] = [
  { value: '', label: 'Semua Status' },
  ...Object.entries(LEAVE_REQUEST_STATUS_LABELS).map(([value, label]) => ({
    value: value as LeaveRequestStatus,
    label,
  })),
]

function getStatusClass(status: LeaveRequestStatus): string {
  switch (status) {
    case 'approved': return styles.badgeApproved
    case 'rejected': return styles.badgeRejected
    case 'pending_approval': return styles.badgePending
    case 'submitted': return styles.badgeSubmitted
    case 'cancelled': return styles.badgeCancelled
    case 'completed': return styles.badgeCompleted
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

export default function LeaveManagementPage() {
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState<LeaveRequestStatus | ''>('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { teacher_name: debouncedSearch } : {}),
    ...(status ? { status } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.leaveRequest, filters],
    queryFn: () => leaveRequestService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => leaveRequestService.delete(id),
    onSuccess: () => {
      toast.success('Pengajuan cuti berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.leaveRequest] })
    },
    onError: () => {
      toast.error('Gagal menghapus pengajuan cuti')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || status !== ''

  function handleReset() {
    setSearch('')
    setStatus('')
    setPage(1)
  }

  function handleDelete(item: LeaveRequest) {
    if (!confirm(`Hapus pengajuan cuti "${item.teacherName}" periode ${formatDate(item.startDate)} - ${formatDate(item.endDate)}?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Manajemen Cuti & Izin</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} data ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Ajukan Cuti
        </button>
      </div>

      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Cari nama guru..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          />
        </div>
        <select
          className={styles.select}
          value={status}
          onChange={(e) => { setStatus(e.target.value as LeaveRequestStatus | ''); setPage(1) }}
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
                <th className={styles.th}>Guru</th>
                <th className={styles.th}>NIP</th>
                <th className={styles.th}>Jenis Cuti</th>
                <th className={styles.th}>Tanggal Mulai</th>
                <th className={styles.th}>Tanggal Selesai</th>
                <th className={styles.th}>Hari</th>
                <th className={styles.th}>Status</th>
                <th className={styles.th}>Alasan</th>
                <th className={styles.th}>Aksi</th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                Array.from({ length: 5 }).map((_, i) => (
                  <tr key={i} className={styles.skeletonRow}>
                    {Array.from({ length: 9 }).map((_, j) => (
                      <td key={j}><span className={styles.skeleton} style={{ width: '80px' }} /></td>
                    ))}
                  </tr>
                ))
              ) : items.length === 0 ? (
                <tr>
                  <td colSpan={9} className={styles.emptyCell}>
                    <div className={styles.emptyState}>
                      <p className={styles.emptyTitle}>Tidak ada data</p>
                      <p className={styles.emptySubtitle}>Belum ada pengajuan cuti atau izin.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.name}>{item.teacherName || '-'}</span></td>
                    <td className={styles.td}><span className={styles.code}>{item.teacherNip || '-'}</span></td>
                    <td className={styles.td}>{item.leaveTypeName || '-'}</td>
                    <td className={styles.td}>{formatDate(item.startDate)}</td>
                    <td className={styles.td}>{formatDate(item.endDate)}</td>
                    <td className={styles.td}>{item.totalDays}</td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getStatusClass(item.status)}`}>
                        {LEAVE_REQUEST_STATUS_LABELS[item.status] ?? item.status}
                      </span>
                    </td>
                    <td className={styles.td}>{item.reason || '-'}</td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus cuti ${item.teacherName}`}
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
