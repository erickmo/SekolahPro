import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { userService } from '@/services/user.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import { ROLE_LABELS, USER_STATUS_LABELS } from '@/types/user.types'
import type { User } from '@/types/user.types'
import styles from './UsersPage.module.css'

// ── Constants ───────────────────────────────────────────────────────────────

const PAGE_SIZE = 20
const DATE_LOCALE = 'id-ID'

const STATUS_OPTIONS: { value: string; label: string }[] = [
  { value: '', label: 'Semua Status' },
  { value: 'active', label: 'Aktif' },
  { value: 'inactive', label: 'Nonaktif' },
]

const TABLE_COLUMN_COUNT = 7

// ── Helpers ─────────────────────────────────────────────────────────────────

function getStatusClass(isActive: boolean): string {
  return isActive ? styles.badgeActive : styles.badgeInactive
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString(DATE_LOCALE, {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  })
}

// ── Component ───────────────────────────────────────────────────────────────

export default function UsersPage() {
  const [search, setSearch] = useState('')
  const [statusFilter, setStatusFilter] = useState('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { search: debouncedSearch } : {}),
    ...(statusFilter ? { is_active: statusFilter === 'active' } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.users, filters],
    queryFn: () => userService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => userService.delete(id),
    onSuccess: () => {
      toast.success('Pengguna berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.users] })
    },
    onError: () => {
      toast.error('Gagal menghapus pengguna')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || statusFilter !== ''

  function handleReset() {
    setSearch('')
    setStatusFilter('')
    setPage(1)
  }

  function handleDelete(user: User) {
    if (!confirm(`Hapus pengguna "${user.fullName}"?`)) return
    deleteMutation.mutate(user.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Pengguna</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString(DATE_LOCALE)} data ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Tambah Pengguna
        </button>
      </div>

      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Cari pengguna..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          />
        </div>
        <select
          className={styles.select}
          value={statusFilter}
          onChange={(e) => { setStatusFilter(e.target.value); setPage(1) }}
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
                <th className={styles.th}>Email</th>
                <th className={styles.th}>Nama Lengkap</th>
                <th className={styles.th}>Role</th>
                <th className={styles.th}>Status</th>
                <th className={styles.th}>Telepon</th>
                <th className={styles.th}>Tanggal Daftar</th>
                <th className={styles.th}>Aksi</th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                renderSkeleton()
              ) : items.length === 0 ? (
                renderEmptyState()
              ) : (
                items.map(renderRow)
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

  // ── Sub-renders (kept under 40 lines each) ────────────────────────────────

  function renderSkeleton() {
    return Array.from({ length: 5 }).map((_, i) => (
      <tr key={i} className={styles.skeletonRow}>
        {Array.from({ length: TABLE_COLUMN_COUNT }).map((_, j) => (
          <td key={j}><span className={styles.skeleton} style={{ width: '80px' }} /></td>
        ))}
      </tr>
    ))
  }

  function renderEmptyState() {
    return (
      <tr>
        <td colSpan={TABLE_COLUMN_COUNT} className={styles.emptyCell}>
          <div className={styles.emptyState}>
            <p className={styles.emptyTitle}>Tidak ada data</p>
            <p className={styles.emptySubtitle}>Belum ada pengguna yang terdaftar.</p>
          </div>
        </td>
      </tr>
    )
  }

  function renderRow(user: User) {
    const statusLabel = user.isActive ? USER_STATUS_LABELS.active : USER_STATUS_LABELS.inactive
    return (
      <tr key={user.id} className={styles.row}>
        <td className={styles.td}><span className={styles.emailCell}>{user.email}</span></td>
        <td className={styles.td}><span className={styles.nameCell}>{user.fullName}</span></td>
        <td className={styles.td}>
          {user.roles.map((role) => (
            <span key={role} className={`${styles.badge} ${styles.badgeRole}`}>
              {ROLE_LABELS[role] ?? role}
            </span>
          ))}
        </td>
        <td className={styles.td}>
          <span className={`${styles.badge} ${getStatusClass(user.isActive)}`}>
            {statusLabel}
          </span>
        </td>
        <td className={styles.td}>{user.phone || '-'}</td>
        <td className={styles.td}>{formatDate(user.createdAt)}</td>
        <td className={styles.td}>
          <div className={styles.actionGroup}>
            <button
              type="button"
              className={styles.deleteBtn}
              onClick={() => handleDelete(user)}
              disabled={deleteMutation.isPending}
              aria-label={`Hapus ${user.fullName}`}
            >
              <Trash2 size={14} />
            </button>
          </div>
        </td>
      </tr>
    )
  }
}
