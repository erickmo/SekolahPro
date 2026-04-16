import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { studentService } from '@/services/student.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { Student, StudentStatus, StudentGender } from '@/types/student.types'
import { STUDENT_STATUS_LABELS, STUDENT_GENDER_LABELS } from '@/types/student.types'
import styles from './StudentsPage.module.css'

const PAGE_SIZE = 20

const STATUS_OPTIONS: { value: StudentStatus | ''; label: string }[] = [
  { value: '', label: 'Semua Status' },
  ...Object.entries(STUDENT_STATUS_LABELS).map(([value, label]) => ({
    value: value as StudentStatus,
    label,
  })),
]

const GENDER_OPTIONS: { value: StudentGender | ''; label: string }[] = [
  { value: '', label: 'Semua Jenis Kelamin' },
  ...Object.entries(STUDENT_GENDER_LABELS).map(([value, label]) => ({
    value: value as StudentGender,
    label,
  })),
]

function getStatusClass(status: StudentStatus): string {
  switch (status) {
    case 'active': return styles.badgeActive
    case 'inactive': return styles.badgeInactive
    case 'graduated': return styles.badgeGraduated
    case 'transferred': return styles.badgeTransferred
    case 'dropped_out': return styles.badgeDroppedOut
    default: return styles.badgeDefault
  }
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '-'
  try {
    const d = new Date(dateStr)
    const places = d.toLocaleDateString('id-ID', {
      day: 'numeric',
      month: 'long',
      year: 'numeric',
    })
    return places
  } catch {
    return dateStr
  }
}

export default function StudentsPage() {
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState<StudentStatus | ''>('')
  const [gender, setGender] = useState<StudentGender | ''>('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { full_name: debouncedSearch } : {}),
    ...(status ? { status } : {}),
    ...(gender ? { gender } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.students, filters],
    queryFn: () => studentService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => studentService.delete(id),
    onSuccess: () => {
      toast.success('Siswa berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.students] })
    },
    onError: () => {
      toast.error('Gagal menghapus siswa')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || status !== '' || gender !== ''

  function handleReset() {
    setSearch('')
    setStatus('')
    setGender('')
    setPage(1)
  }

  function handleDelete(item: Student) {
    if (!confirm(`Hapus siswa "${item.fullName}"?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Data Siswa</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} data ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Tambah Siswa
        </button>
      </div>

      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Cari nama / NIS..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          />
        </div>
        <select
          className={styles.select}
          value={status}
          onChange={(e) => { setStatus(e.target.value as StudentStatus | ''); setPage(1) }}
          aria-label="Filter status"
        >
          {STATUS_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
        <select
          className={styles.select}
          value={gender}
          onChange={(e) => { setGender(e.target.value as StudentGender | ''); setPage(1) }}
          aria-label="Filter jenis kelamin"
        >
          {GENDER_OPTIONS.map((opt) => (
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
                <th className={styles.th}>NIS</th>
                <th className={styles.th}>Nama</th>
                <th className={styles.th}>Jenis Kelamin</th>
                <th className={styles.th}>Tempat / Tgl Lahir</th>
                <th className={styles.th}>Agama</th>
                <th className={styles.th}>Status</th>
                <th className={styles.th}>Aksi</th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                Array.from({ length: 5 }).map((_, i) => (
                  <tr key={i} className={styles.skeletonRow}>
                    {Array.from({ length: 7 }).map((_, j) => (
                      <td key={j}><span className={styles.skeleton} style={{ width: '80px' }} /></td>
                    ))}
                  </tr>
                ))
              ) : items.length === 0 ? (
                <tr>
                  <td colSpan={7} className={styles.emptyCell}>
                    <div className={styles.emptyState}>
                      <p className={styles.emptyTitle}>Tidak ada data</p>
                      <p className={styles.emptySubtitle}>Belum ada siswa yang terdaftar.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.code}>{item.nis || '-'}</span></td>
                    <td className={styles.td}><span className={styles.name}>{item.fullName}</span></td>
                    <td className={styles.td}>{STUDENT_GENDER_LABELS[item.gender] ?? item.gender}</td>
                    <td className={styles.td}>
                      {item.birthPlace ? `${item.birthPlace}, ` : ''}{formatDate(item.birthDate)}
                    </td>
                    <td className={styles.td}>{item.religion ? item.religion.charAt(0).toUpperCase() + item.religion.slice(1) : '-'}</td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getStatusClass(item.status)}`}>
                        {STUDENT_STATUS_LABELS[item.status] ?? item.status}
                      </span>
                    </td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus ${item.fullName}`}
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
