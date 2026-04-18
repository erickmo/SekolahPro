import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { facilityBookingService } from '@/services/sprint7.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { FacilityBooking, BookingStatus, FacilityType, RequesterType } from '@/types/sprint7.types'
import {
  BOOKING_STATUS_LABELS,
  FACILITY_TYPE_LABELS,
  REQUESTER_TYPE_LABELS,
} from '@/types/sprint7.types'
import styles from './FacilityBookingPage.module.css'

const PAGE_SIZE = 20

const STATUS_OPTIONS: { value: BookingStatus | ''; label: string }[] = [
  { value: '', label: 'Semua Status' },
  ...Object.entries(BOOKING_STATUS_LABELS).map(([value, label]) => ({
    value: value as BookingStatus,
    label,
  })),
]

const FACILITY_TYPE_OPTIONS: { value: FacilityType | ''; label: string }[] = [
  { value: '', label: 'Semua Fasilitas' },
  ...Object.entries(FACILITY_TYPE_LABELS).map(([value, label]) => ({
    value: value as FacilityType,
    label,
  })),
]

function getStatusBadgeClass(status: BookingStatus): string {
  switch (status) {
    case 'pending': return styles.badgeYellow
    case 'approved': return styles.badgeGreen
    case 'rejected': return styles.badgeRed
    case 'cancelled': return styles.badgeGray
    case 'completed': return styles.badgeBlue
    default: return styles.badgeDefault
  }
}

function getRequesterBadgeClass(type: RequesterType): string {
  switch (type) {
    case 'teacher': return styles.badgeBlue
    case 'student': return styles.badgeGreen
    case 'staff': return styles.badgeYellow
    case 'external': return styles.badgeDefault
    case 'admin': return styles.badgeGray
    default: return styles.badgeDefault
  }
}

function getFacilityTypeBadgeClass(type: string): string {
  switch (type) {
    case 'classroom': return styles.badgeBlue
    case 'lab': return styles.badgeGreen
    case 'library': return styles.badgeYellow
    case 'hall': return styles.badgeDefault
    case 'meeting_room': return styles.badgeGray
    case 'sports_field': return styles.badgeGreen
    case 'mosque': return styles.badgeDefault
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

export default function FacilityBookingPage() {
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState<BookingStatus | ''>('')
  const [facilityType, setFacilityType] = useState<FacilityType | ''>('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { q: debouncedSearch } : {}),
    ...(status ? { status } : {}),
    ...(facilityType ? { facility_type: facilityType } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.facilityBookings, filters],
    queryFn: () => facilityBookingService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => facilityBookingService.delete(id),
    onSuccess: () => {
      toast.success('Data booking berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.facilityBookings] })
    },
    onError: () => {
      toast.error('Gagal menghapus data booking')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || status !== '' || facilityType !== ''

  function handleReset() {
    setSearch('')
    setStatus('')
    setFacilityType('')
    setPage(1)
  }

  function handleDelete(item: FacilityBooking) {
    if (!confirm(`Hapus booking "${item.facilityName}" tanggal ${formatDate(item.bookingDate)}?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Pemesanan Fasilitas</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} data ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Buat Booking
        </button>
      </div>

      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Cari fasilitas atau tujuan..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          />
        </div>
        <select
          className={styles.select}
          value={status}
          onChange={(e) => { setStatus(e.target.value as BookingStatus | ''); setPage(1) }}
          aria-label="Filter status"
        >
          {STATUS_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
        <select
          className={styles.select}
          value={facilityType}
          onChange={(e) => { setFacilityType(e.target.value as FacilityType | ''); setPage(1) }}
          aria-label="Filter tipe fasilitas"
        >
          {FACILITY_TYPE_OPTIONS.map((opt) => (
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
                <th className={styles.th}>Fasilitas</th>
                <th className={styles.th}>Tipe Fasilitas</th>
                <th className={styles.th}>Tanggal</th>
                <th className={styles.th}>Waktu</th>
                <th className={styles.th}>Pemohon</th>
                <th className={styles.th}>Tujuan</th>
                <th className={styles.th}>Status</th>
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
                    <div>
                      <p className={styles.emptyTitle}>Tidak ada data</p>
                      <p className={styles.emptySubtitle}>Belum ada data pemesanan fasilitas.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.name}>{item.facilityName || '-'}</span></td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getFacilityTypeBadgeClass(item.facilityType)}`}>
                        {FACILITY_TYPE_LABELS[item.facilityType as keyof typeof FACILITY_TYPE_LABELS] ?? item.facilityType}
                      </span>
                    </td>
                    <td className={styles.td}>{formatDate(item.bookingDate)}</td>
                    <td className={styles.td}>{item.startTime} - {item.endTime}</td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getRequesterBadgeClass(item.requesterType)}`}>
                        {REQUESTER_TYPE_LABELS[item.requesterType] ?? item.requesterType}
                      </span>
                    </td>
                    <td className={styles.td}>{item.purpose || '-'}</td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getStatusBadgeClass(item.status)}`}>
                        {BOOKING_STATUS_LABELS[item.status] ?? item.status}
                      </span>
                    </td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus booking ${item.facilityName}`}
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
