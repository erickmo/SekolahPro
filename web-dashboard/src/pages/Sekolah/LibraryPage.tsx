import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { libraryBookService } from '@/services/sprint7.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { LibraryBook, BookCategory } from '@/types/sprint7.types'
import { BOOK_CATEGORY_LABELS, BOOK_LANGUAGE_LABELS } from '@/types/sprint7.types'
import styles from './LibraryPage.module.css'

const PAGE_SIZE = 20

const CATEGORY_OPTIONS: { value: BookCategory | ''; label: string }[] = [
  { value: '', label: 'Semua Kategori' },
  ...Object.entries(BOOK_CATEGORY_LABELS).map(([value, label]) => ({
    value: value as BookCategory,
    label,
  })),
]

function getCategoryClass(category: BookCategory): string {
  const map: Record<BookCategory, string> = {
    fiction: styles.badgeFiction,
    nonfiction: styles.badgeNonfiction,
    textbook: styles.badgeTextbook,
    reference: styles.badgeReference,
    kitab_kuning: styles.badgeKitabKuning,
    al_quran: styles.badgeAlQuran,
    hadits: styles.badgeHadits,
    fiqh: styles.badgeFiqh,
    aqidah: styles.badgeAqidah,
    science: styles.badgeScience,
    literature: styles.badgeLiterature,
    other: styles.badgeOther,
  }
  return map[category] ?? styles.badgeOther
}

export default function LibraryPage() {
  const [search, setSearch] = useState('')
  const [category, setCategory] = useState<BookCategory | ''>('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { q: debouncedSearch } : {}),
    ...(category ? { category } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.libraryBooks, filters],
    queryFn: () => libraryBookService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => libraryBookService.delete(id),
    onSuccess: () => {
      toast.success('Data buku berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.libraryBooks] })
    },
    onError: () => {
      toast.error('Gagal menghapus data buku')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || category !== ''

  function handleReset() {
    setSearch('')
    setCategory('')
    setPage(1)
  }

  function handleDelete(item: LibraryBook) {
    if (!confirm(`Hapus buku "${item.title}"?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Perpustakaan</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} data ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Tambah Buku
        </button>
      </div>

      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Cari judul atau penulis..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          />
        </div>
        <select
          className={styles.select}
          value={category}
          onChange={(e) => { setCategory(e.target.value as BookCategory | ''); setPage(1) }}
          aria-label="Filter kategori"
        >
          {CATEGORY_OPTIONS.map((opt) => (
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
                <th className={styles.th}>Judul</th>
                <th className={styles.th}>Penulis</th>
                <th className={styles.th}>ISBN</th>
                <th className={styles.th}>Kategori</th>
                <th className={styles.th}>Bahasa</th>
                <th className={styles.th}>Klasifikasi DDC</th>
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
                    <div>
                      <p className={styles.emptyTitle}>Tidak ada data</p>
                      <p className={styles.emptySubtitle}>Belum ada data buku perpustakaan.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.name}>{item.title || '-'}</span></td>
                    <td className={styles.td}>{item.author || '-'}</td>
                    <td className={styles.td}><span className={styles.code}>{item.isbn || '-'}</span></td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getCategoryClass(item.category)}`}>
                        {BOOK_CATEGORY_LABELS[item.category] ?? item.category}
                      </span>
                    </td>
                    <td className={styles.td}>{BOOK_LANGUAGE_LABELS[item.language] ?? item.language}</td>
                    <td className={styles.td}><span className={styles.code}>{item.ddcClassification || '-'}</span></td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus buku ${item.title}`}
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
