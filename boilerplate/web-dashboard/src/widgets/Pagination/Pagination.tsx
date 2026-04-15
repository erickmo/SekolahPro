import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from 'lucide-react'
import styles from './Pagination.module.css'

interface PaginationProps {
  page: number
  pageSize: number
  total: number
  onPageChange: (page: number) => void
  onPageSizeChange?: (size: number) => void
  pageSizeOptions?: number[]
  showSummary?: boolean
}

const DEFAULT_PAGE_SIZES = [10, 20, 50, 100]
const MAX_PAGES_SHOWN = 5

function getPageNumbers(current: number, total: number): (number | '...')[] {
  if (total <= MAX_PAGES_SHOWN) return Array.from({ length: total }, (_, i) => i + 1)

  const pages: (number | '...')[] = []
  const half = Math.floor(MAX_PAGES_SHOWN / 2)
  let start = Math.max(2, current - half)
  let end = Math.min(total - 1, current + half)

  if (current - half <= 2) end = Math.min(total - 1, MAX_PAGES_SHOWN - 1)
  if (current + half >= total - 1) start = Math.max(2, total - MAX_PAGES_SHOWN + 2)

  pages.push(1)
  if (start > 2) pages.push('...')
  for (let i = start; i <= end; i++) pages.push(i)
  if (end < total - 1) pages.push('...')
  pages.push(total)

  return pages
}

export function Pagination({
  page,
  pageSize,
  total,
  onPageChange,
  onPageSizeChange,
  pageSizeOptions = DEFAULT_PAGE_SIZES,
  showSummary = true,
}: PaginationProps) {
  const totalPages = Math.max(1, Math.ceil(total / pageSize))
  const from = Math.min((page - 1) * pageSize + 1, total)
  const to = Math.min(page * pageSize, total)
  const pages = getPageNumbers(page, totalPages)

  return (
    <div className={styles.pagination}>
      {showSummary && (
        <span className={styles.summary}>
          {total === 0 ? 'Tidak ada data' : `${from}–${to} dari ${total} data`}
        </span>
      )}

      <div className={styles.controls}>
        {/* First */}
        <button
          type="button"
          className={styles.btn}
          onClick={() => onPageChange(1)}
          disabled={page === 1}
          aria-label="Halaman pertama"
        >
          <ChevronsLeft size={14} />
        </button>

        {/* Prev */}
        <button
          type="button"
          className={styles.btn}
          onClick={() => onPageChange(page - 1)}
          disabled={page === 1}
          aria-label="Halaman sebelumnya"
        >
          <ChevronLeft size={14} />
        </button>

        {/* Page numbers */}
        {pages.map((p, i) =>
          p === '...' ? (
            <span key={`ellipsis-${i}`} className={styles.ellipsis}>…</span>
          ) : (
            <button
              key={p}
              type="button"
              className={`${styles.btn} ${p === page ? styles.active : ''}`}
              onClick={() => onPageChange(p as number)}
              aria-label={`Halaman ${p}`}
              aria-current={p === page ? 'page' : undefined}
            >
              {p}
            </button>
          ),
        )}

        {/* Next */}
        <button
          type="button"
          className={styles.btn}
          onClick={() => onPageChange(page + 1)}
          disabled={page === totalPages}
          aria-label="Halaman berikutnya"
        >
          <ChevronRight size={14} />
        </button>

        {/* Last */}
        <button
          type="button"
          className={styles.btn}
          onClick={() => onPageChange(totalPages)}
          disabled={page === totalPages}
          aria-label="Halaman terakhir"
        >
          <ChevronsRight size={14} />
        </button>
      </div>

      {onPageSizeChange && (
        <select
          className={styles.sizeSelect}
          value={pageSize}
          onChange={(e) => onPageSizeChange(Number(e.target.value))}
          aria-label="Jumlah per halaman"
        >
          {pageSizeOptions.map((s) => (
            <option key={s} value={s}>{s} / halaman</option>
          ))}
        </select>
      )}
    </div>
  )
}
