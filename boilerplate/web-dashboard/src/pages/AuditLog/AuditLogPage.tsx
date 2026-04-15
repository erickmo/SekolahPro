import { useState } from 'react'
import { Search, RotateCcw } from 'lucide-react'
import { useQuery } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { auditLogService } from '@/services/audit-log.service'
import type { AuditAction, AuditLogFilters } from '@/types/audit-log.types'
import styles from './AuditLogPage.module.css'

// ─── Constants ────────────────────────────────────────────────────────────────

const PAGE_SIZE = 20

const ACTION_OPTIONS: { value: AuditAction | ''; label: string }[] = [
  { value: '', label: 'Semua Aksi' },
  { value: 'create', label: 'Buat' },
  { value: 'update', label: 'Perbarui' },
  { value: 'delete', label: 'Hapus' },
  { value: 'login', label: 'Login' },
  { value: 'logout', label: 'Logout' },
  { value: 'export', label: 'Export' },
  { value: 'view', label: 'Lihat' },
]

const ACTION_LABEL_MAP: Record<AuditAction, string> = {
  create: 'Buat',
  update: 'Perbarui',
  delete: 'Hapus',
  login: 'Login',
  logout: 'Logout',
  export: 'Export',
  view: 'Lihat',
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

function formatTimestamp(ts: string): string {
  return new Intl.DateTimeFormat('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(ts))
}

function getActionClass(action: AuditAction): string {
  switch (action) {
    case 'create': return styles.badgeCreate
    case 'delete': return styles.badgeDelete
    case 'update': return styles.badgeUpdate
    case 'login':
    case 'logout': return styles.badgeAuth
    case 'export':
    case 'view': return styles.badgeNeutral
    default: return styles.badgeNeutral
  }
}

// ─── Skeleton rows ────────────────────────────────────────────────────────────

function SkeletonRows() {
  return (
    <>
      {Array.from({ length: 8 }).map((_, i) => (
        <tr key={i} className={styles.skeletonRow}>
          <td><span className={styles.skeleton} style={{ width: '110px' }} /></td>
          <td>
            <span className={styles.skeleton} style={{ width: '100px' }} />
            <span className={styles.skeleton} style={{ width: '130px', marginTop: '4px' }} />
          </td>
          <td><span className={styles.skeleton} style={{ width: '64px' }} /></td>
          <td><span className={styles.skeleton} style={{ width: '80px' }} /></td>
          <td><span className={styles.skeleton} style={{ width: '200px' }} /></td>
          <td><span className={styles.skeleton} style={{ width: '90px' }} /></td>
        </tr>
      ))}
    </>
  )
}

// ─── Audit log page ───────────────────────────────────────────────────────────

export default function AuditLogPage() {
  const [search, setSearch] = useState('')
  const [action, setAction] = useState<AuditAction | ''>('')
  const [page, setPage] = useState(1)

  const debouncedSearch = useDebounce(search, 300)

  const filters: AuditLogFilters = {
    search: debouncedSearch || undefined,
    action: action || undefined,
    page,
    pageSize: PAGE_SIZE,
  }

  const { data, isLoading } = useQuery({
    queryKey: ['audit-logs', filters],
    queryFn: () => auditLogService.list(filters),
  })

  const logs = data?.data ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0

  function handleReset() {
    setSearch('')
    setAction('')
    setPage(1)
  }

  const hasFilters = search !== '' || action !== ''

  return (
    <div className={styles.root}>
      {/* Page title */}
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Log Aktivitas</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} entri ditemukan
            </p>
          )}
        </div>
      </div>

      {/* Filters */}
      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Cari pengguna, entitas, deskripsi..."
            value={search}
            onChange={(e) => {
              setSearch(e.target.value)
              setPage(1)
            }}
          />
        </div>

        <select
          className={styles.select}
          value={action}
          onChange={(e) => {
            setAction(e.target.value as AuditAction | '')
            setPage(1)
          }}
          aria-label="Filter aksi"
        >
          {ACTION_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {opt.label}
            </option>
          ))}
        </select>

        {hasFilters && (
          <button
            type="button"
            className={styles.resetBtn}
            onClick={handleReset}
          >
            <RotateCcw size={14} />
            Reset Filter
          </button>
        )}
      </div>

      {/* Table card */}
      <div className={styles.tableCard}>
        <div className={styles.tableScroll}>
          <table className={styles.table}>
            <thead>
              <tr>
                <th className={styles.th}>Waktu</th>
                <th className={styles.th}>Pengguna</th>
                <th className={styles.th}>Aksi</th>
                <th className={styles.th}>Entitas</th>
                <th className={styles.th}>Deskripsi</th>
                <th className={styles.th}>IP</th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                <SkeletonRows />
              ) : logs.length === 0 ? (
                <tr>
                  <td colSpan={6} className={styles.emptyCell}>
                    <div className={styles.emptyState}>
                      <p className={styles.emptyTitle}>Tidak ada data</p>
                      <p className={styles.emptySubtitle}>
                        Belum ada log aktivitas yang tercatat.
                      </p>
                    </div>
                  </td>
                </tr>
              ) : (
                logs.map((log) => (
                  <tr key={log.id} className={styles.row}>
                    <td className={styles.td}>
                      <span className={styles.timestamp}>
                        {formatTimestamp(log.timestamp)}
                      </span>
                    </td>
                    <td className={styles.td}>
                      <p className={styles.userName}>{log.userName}</p>
                      <p className={styles.userEmail}>{log.userEmail}</p>
                    </td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getActionClass(log.action)}`}>
                        {ACTION_LABEL_MAP[log.action] ?? log.action}
                      </span>
                    </td>
                    <td className={styles.td}>
                      <span className={styles.entity}>{log.entity}</span>
                      {log.entityId && (
                        <span className={styles.entityId}>#{log.entityId}</span>
                      )}
                    </td>
                    <td className={styles.td}>
                      <span className={styles.description}>{log.description}</span>
                    </td>
                    <td className={styles.td}>
                      <span className={styles.ipAddress}>
                        {log.ipAddress ?? '—'}
                      </span>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination */}
        {!isLoading && totalPages > 1 && (
          <div className={styles.pagination}>
            <button
              type="button"
              className={styles.pageBtn}
              disabled={page <= 1}
              onClick={() => setPage((p) => p - 1)}
            >
              Sebelumnya
            </button>
            <span className={styles.pageInfo}>
              Halaman {page} dari {totalPages}
            </span>
            <button
              type="button"
              className={styles.pageBtn}
              disabled={page >= totalPages}
              onClick={() => setPage((p) => p + 1)}
            >
              Berikutnya
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
