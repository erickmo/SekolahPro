import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { teacherService } from '@/services/teacher.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { Teacher, TeacherRole, EmployeeType } from '@/types/teacher.types'
import { EMPLOYEE_TYPE_LABELS, ROLE_LABELS, STATUS_LABELS } from '@/types/teacher.types'
import styles from '../AcademicYears/AcademicYearsPage.module.css'

const PAGE_SIZE = 20

const ROLE_OPTIONS: { value: TeacherRole | ''; label: string }[] = [
  { value: '', label: 'Semua Peran' },
  ...Object.entries(ROLE_LABELS).map(([value, label]) => ({ value: value as TeacherRole, label })),
]

const EMP_TYPE_OPTIONS: { value: EmployeeType | ''; label: string }[] = [
  { value: '', label: 'Semua Jenis' },
  ...Object.entries(EMPLOYEE_TYPE_LABELS).map(([value, label]) => ({ value: value as EmployeeType, label })),
]

export default function TeachersPage() {
  const [search, setSearch] = useState('')
  const [role, setRole] = useState<TeacherRole | ''>('')
  const [empType, setEmpType] = useState<EmployeeType | ''>('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { full_name: debouncedSearch } : {}),
    ...(role ? { role } : {}),
    ...(empType ? { employee_type: empType } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.teachers, filters],
    queryFn: () => teacherService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => teacherService.delete(id),
    onSuccess: () => {
      toast.success('Guru/staff berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.teachers] })
    },
    onError: () => toast.error('Gagal menghapus guru/staff'),
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || role !== '' || empType !== ''

  function handleReset() { setSearch(''); setRole(''); setEmpType(''); setPage(1) }
  function handleDelete(item: Teacher) {
    if (!confirm(`Hapus "${item.fullName}"?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Guru & Staff</h1>
          {!isLoading && <p className={styles.totalInfo}>{total.toLocaleString('id-ID')} data ditemukan</p>}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Tambah Guru/Staff
        </button>
      </div>

      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input type="text" className={styles.searchInput} placeholder="Cari nama guru..." value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }} />
        </div>
        <select className={styles.select} value={role} onChange={(e) => { setRole(e.target.value as TeacherRole | ''); setPage(1) }}>
          {ROLE_OPTIONS.map((opt) => <option key={opt.value} value={opt.value}>{opt.label}</option>)}
        </select>
        <select className={styles.select} value={empType} onChange={(e) => { setEmpType(e.target.value as EmployeeType | ''); setPage(1) }}>
          {EMP_TYPE_OPTIONS.map((opt) => <option key={opt.value} value={opt.value}>{opt.label}</option>)}
        </select>
        {hasFilters && (
          <button type="button" className={styles.resetBtn} onClick={handleReset}><RotateCcw size={14} /> Reset Filter</button>
        )}
      </div>

      <div className={styles.tableCard}>
        <div className={styles.tableScroll}>
          <table className={styles.table}>
            <thead>
              <tr>
                <th className={styles.th}>Nama</th>
                <th className={styles.th}>NIP</th>
                <th className={styles.th}>Jenis</th>
                <th className={styles.th}>Peran</th>
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
                <tr><td colSpan={6} className={styles.emptyCell}>
                  <div className={styles.emptyState}>
                    <p className={styles.emptyTitle}>Tidak ada data</p>
                    <p className={styles.emptySubtitle}>Belum ada guru/staff yang terdaftar.</p>
                  </div>
                </td></tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.name}>{item.fullName}</span></td>
                    <td className={styles.td}><span className={styles.code}>{item.nip || '—'}</span></td>
                    <td className={styles.td}>{EMPLOYEE_TYPE_LABELS[item.employeeType] ?? item.employeeType}</td>
                    <td className={styles.td}>{ROLE_LABELS[item.role] ?? item.role}</td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${item.status === 'active' ? styles.badgeActive : styles.badgeClosed}`}>
                        {STATUS_LABELS[item.status] ?? item.status}
                      </span>
                    </td>
                    <td className={styles.td}>
                      <button type="button" className={styles.deleteBtn} onClick={() => handleDelete(item)} disabled={deleteMutation.isPending}>
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
            <button type="button" className={styles.pageBtn} disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>Sebelumnya</button>
            <span className={styles.pageInfo}>Halaman {page} dari {totalPages}</span>
            <button type="button" className={styles.pageBtn} disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)}>Berikutnya</button>
          </div>
        )}
      </div>
    </div>
  )
}
