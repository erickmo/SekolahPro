import { useState, useCallback, useEffect, useRef } from 'react'
import { createPortal } from 'react-dom'
import { ChevronUp, ChevronDown, ChevronsUpDown, Search, X, Columns3, Download, ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight, SlidersHorizontal } from 'lucide-react'
import { EmptyState } from '@/widgets/EmptyState/EmptyState'
import { exportToCSV, exportToJSON, exportToPDF } from '@/utils/export'
import type { ExportFormat } from '@/utils/export'
import { cn } from '@/utils/cn'
import { FilterDialog } from './FilterDialog'
import { InlineFilter } from './InlineFilter'
import type { ActiveFilter, FilterDef } from './filter.types'
import { OPERATOR_LABELS } from './filter.utils'
import styles from './DataTable.module.css'

export type { FilterDef, ActiveFilter }
export type { FilterOperator, FilterFieldType } from './filter.types'

export interface ColumnDef<T> {
  key: string
  /** API sort param name. Defaults to `key` when not set. Use this when the
   *  display key (camelCase) differs from the DB column name (snake_case). */
  sortKey?: string
  header: string
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  render?: (value: any, row: T, index: number) => React.ReactNode
  sortable?: boolean
  width?: number | string
  align?: 'left' | 'center' | 'right'
  visible?: boolean
  locked?: boolean
}

export interface RowActionDef<T> {
  key: string
  label: string
  icon?: React.ReactNode
  variant?: 'default' | 'danger'
  visible?: (row: T) => boolean
  onClick: (row: T) => void
}

interface SortState {
  key: string
  order: 'asc' | 'desc'
}

interface PaginationState {
  page: number
  pageSize: number
  total: number
}

interface DataTableProps<T> {
  columns: ColumnDef<T>[]
  data: T[]
  rowKey?: keyof T | ((row: T) => string)
  isLoading?: boolean
  pagination?: PaginationState
  onPageChange?: (page: number) => void
  onPageSizeChange?: (size: number) => void
  sort?: SortState
  onSortChange?: (sort: SortState | undefined) => void
  search?: string
  onSearchChange?: (search: string) => void
  searchPlaceholder?: string
  rowActions?: RowActionDef<T>[]
  onRowClick?: (row: T) => void
  exportFilename?: string
  showExport?: boolean
  emptyTitle?: string
  emptyDescription?: string
  selectable?: boolean
  selected?: string[]
  onSelectionChange?: (ids: string[]) => void
  filterDefs?: FilterDef[]
  activeFilters?: ActiveFilter[]
  onActiveFiltersChange?: (filters: ActiveFilter[]) => void
  /** Mode pencarian: 'button' (tekan tombol / Enter) atau 'debounce' (otomatis). Default: 'button'. */
  searchMode?: 'button' | 'debounce'
  /** Delay debounce dalam ms. Hanya berlaku untuk searchMode='debounce'. Default: 400. */
  searchDebounceMs?: number
  /** Tampilkan baris inline filter (field + operator + nilai) di bawah toolbar. */
  inlineFilter?: boolean
  /** Custom export handler untuk PDF. Jika disediakan, akan dipanggil untuk ekspor PDF. */
  onExportPDF?: (filename: string, data: T[]) => void
}

const PAGE_SIZE_OPTIONS = [50, 100, 250, 500, 1000]

function SkeletonRow({ cols, selectable }: { cols: number; selectable?: boolean }) {
  return (
    <tr className={styles.skeletonRow}>
      {selectable && <td><div className={styles.skeletonCell} /></td>}
      {Array.from({ length: cols }).map((_, i) => (
        <td key={i}>
          <div className={styles.skeletonCell} />
        </td>
      ))}
    </tr>
  )
}

function formatPillLabel(filter: ActiveFilter, defs: FilterDef[]): string {
  const def = defs.find((d) => d.key === filter.key)
  const defLabel = def?.label ?? filter.key
  const opLabel = OPERATOR_LABELS[filter.operator] ?? filter.operator

  const v = filter.value
  let valStr: string

  if (Array.isArray(v)) {
    const [a, b] = v as string[]
    valStr = a && b ? `${a} – ${b}` : a || b || ''
  } else if (typeof v === 'boolean') {
    valStr = v ? 'Ya' : 'Tidak'
  } else if (typeof v === 'object' && v !== null && 'label' in (v as object)) {
    valStr = (v as { label: string }).label
  } else if (def?.type === 'select' && def.options) {
    const opt = def.options.find((o) => o.value === String(v))
    valStr = opt?.label ?? String(v ?? '')
  } else {
    valStr = String(v ?? '')
  }

  if (!valStr) return defLabel
  return `${defLabel} ${opLabel} ${valStr}`
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function DataTable<T extends Record<string, any>>({
  columns: initialColumns,
  data,
  rowKey = 'id' as keyof T,
  isLoading,
  pagination,
  onPageChange,
  onPageSizeChange,
  sort,
  onSortChange,
  search = '',
  onSearchChange,
  searchPlaceholder = 'Cari...',
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  rowActions: _rowActions,
  onRowClick,
  exportFilename = 'export',
  showExport = true,
  emptyTitle = 'Tidak ada data',
  emptyDescription = 'Belum ada data yang cocok.',
  selectable,
  selected = [],
  onSelectionChange,
  filterDefs,
  activeFilters = [],
  onActiveFiltersChange,
  searchMode = 'button',
  searchDebounceMs = 400,
  inlineFilter,
  onExportPDF,
}: DataTableProps<T>) {
  const [pendingSearch, setPendingSearch] = useState(search)
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  // Sync input value ketika parent mengubah search dari luar (mis. clear)
  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    if (searchMode === 'button') setPendingSearch(search)
  }, [search, searchMode])

  // Auto-search dengan debounce
  useEffect(() => {
    if (searchMode !== 'debounce' || !onSearchChange) return
    if (debounceRef.current) clearTimeout(debounceRef.current)
    debounceRef.current = setTimeout(() => {
      onSearchChange(pendingSearch)
    }, searchDebounceMs)
    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current)
    }
  }, [pendingSearch, searchMode, searchDebounceMs, onSearchChange])

  // localFilters: draft state ditampilkan sebagai pills, belum dikirim ke API.
  // Baru dikirim ke onActiveFiltersChange saat user klik Cari.
  const [localFilters, setLocalFilters] = useState<ActiveFilter[]>(() => activeFilters)
  const [showFilterDialog, setShowFilterDialog] = useState(false)
  // Flag untuk mencegah sinkronisasi saat dialog ditutup setelah user menyimpan filter
  const [skipSync, setSkipSync] = useState(false)

  // Sinkronisasi localFilters saat activeFilters berubah dari luar (mis. dari URL query string)
  // HANYA sinkronisasi saat activeFilters berubah, bukan saat localFilters berubah
  useEffect(() => {
    if (showFilterDialog || skipSync) {
      // Jangan override saat dialog filter sedang terbuka atau saat user baru saja menyimpan
      if (skipSync) setSkipSync(false)
      return
    }
    // Cek apakah activeFilters berbeda dari localFilters
    const isSame =
      activeFilters.length === localFilters.length &&
      activeFilters.every(
        (af, i) =>
          af.id === localFilters[i]?.id &&
          af.key === localFilters[i]?.key &&
          af.operator === localFilters[i]?.operator &&
          JSON.stringify(af.value) === JSON.stringify(localFilters[i]?.value),
      )
    if (!isSame) {
      setLocalFilters(activeFilters)
    }
  }, [activeFilters, showFilterDialog])

  const handleSearchSubmit = useCallback(() => {
    if (onSearchChange) onSearchChange(pendingSearch)
    onActiveFiltersChange?.(localFilters)
  }, [onSearchChange, pendingSearch, onActiveFiltersChange, localFilters])
  const [showExportDropdown, setShowExportDropdown] = useState(false)
  const [colVisibility, setColVisibility] = useState<Record<string, boolean>>(
    Object.fromEntries(initialColumns.map((c) => [c.key, c.visible !== false])),
  )
  const [showColToggle, setShowColToggle] = useState(false)

  const columns = initialColumns.filter((c) => colVisibility[c.key] !== false)

  const getRowKey = useCallback(
    (row: T): string => {
      if (typeof rowKey === 'function') return rowKey(row)
      return String(row[rowKey])
    },
    [rowKey],
  )

  const handleSort = (col: ColumnDef<T>) => {
    if (!onSortChange) return
    const apiKey = col.sortKey ?? col.key
    if (sort?.key === apiKey) {
      if (sort.order === 'asc') onSortChange({ key: apiKey, order: 'desc' })
      else onSortChange(undefined)
    } else {
      onSortChange({ key: apiKey, order: 'asc' })
    }
  }

  const handleSelectAll = () => {
    if (!onSelectionChange) return
    const allIds = data.map(getRowKey)
    const allSelected = allIds.every((id) => selected.includes(id))
    onSelectionChange(allSelected ? [] : allIds)
  }

  const handleSelectRow = (id: string) => {
    if (!onSelectionChange) return
    onSelectionChange(
      selected.includes(id) ? selected.filter((s) => s !== id) : [...selected, id],
    )
  }

  const handleExport = (format: ExportFormat) => {
    const exportCols = initialColumns.map((c) => ({ key: c.key, header: c.header }))
    if (format === 'csv') exportToCSV(exportFilename, exportCols, data)
    else if (format === 'json') exportToJSON(exportFilename, exportCols, data)
    else if (format === 'pdf') {
      if (onExportPDF) {
        onExportPDF(exportFilename, data)
      } else {
        exportToPDF(exportFilename, exportCols, data)
      }
    }
    setShowExportDropdown(false)
  }

  const handleRemovePill = (id: string) => {
    const newFilters = localFilters.filter((f) => f.id !== id)
    setLocalFilters(newFilters)
    onActiveFiltersChange?.(newFilters)
  }

  const totalPages = pagination ? Math.ceil(pagination.total / pagination.pageSize) : 1
  const currentPage = pagination?.page ?? 0
  const dataLength = data?.length ?? 0
  const pageSize = pagination?.pageSize ?? dataLength
  const total = pagination?.total ?? dataLength
  const start = total === 0 ? 0 : currentPage * pageSize + 1
  const end = Math.min((currentPage + 1) * pageSize, total)

  const visibleNonLocked = initialColumns.filter((c) => !c.locked)
  const allNonLockedVisible = visibleNonLocked.every((c) => colVisibility[c.key] !== false)
  const hasFilterDefs = !!(filterDefs && filterDefs.length > 0)

  return (
    <div className={styles.root}>
      {/* Toolbar */}
      <div className={styles.toolbar}>
        <div className={styles.toolbarLeft}>
          {onSearchChange && (
            <div className={styles.searchWrap}>
              <Search size={14} className={styles.searchIcon} />
              <input
                className={styles.searchInput}
                type="text"
                placeholder={searchPlaceholder}
                value={pendingSearch}
                onChange={(e) => setPendingSearch(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' && searchMode === 'button') handleSearchSubmit()
                }}
              />
              {pendingSearch && (
                <button
                  className={styles.clearSearch}
                  onClick={() => {
                    setPendingSearch('')
                    if (debounceRef.current) clearTimeout(debounceRef.current)
                    onSearchChange('')
                  }}
                >
                  <X size={12} />
                </button>
              )}
            </div>
          )}

          {/* Active filter pills — appear before filter button */}
          {hasFilterDefs && localFilters.length > 0 && (
            <div className={styles.pillsWrap}>
              {localFilters.map((f) => (
                <span key={f.id} className={styles.pill}>
                  <span className={styles.pillLabel}>
                    {formatPillLabel(f, filterDefs!)}
                  </span>
                  <button
                    className={styles.pillRemove}
                    onClick={() => handleRemovePill(f.id)}
                    title="Hapus filter ini"
                  >
                    <X size={10} />
                  </button>
                </span>
              ))}
            </div>
          )}

          {/* Filter icon button */}
          {hasFilterDefs && (
            <button
              className={cn(styles.toolbarBtn, styles.toolbarBtnIcon, localFilters.length > 0 && styles.toolbarBtnActive)}
              onClick={() => setShowFilterDialog(true)}
              title="Filter data"
            >
              <SlidersHorizontal size={14} />
            </button>
          )}

          {/* Cari button */}
          {onSearchChange && searchMode === 'button' && (
            <button className={styles.searchBtn} onClick={handleSearchSubmit}>
              Cari
            </button>
          )}
        </div>
        <div className={styles.toolbarRight}>
          {showExport && (
            <div className={styles.exportWrap}>
              <button
                className={cn(styles.toolbarBtn, styles.toolbarBtnIcon)}
                onClick={() => setShowExportDropdown((v) => !v)}
                title="Ekspor data"
              >
                <Download size={14} />
              </button>
              {showExportDropdown && (
                <div className={styles.exportDropdown}>
                  <div className={styles.exportHeader}>Ekspor sebagai</div>
                  <button className={styles.exportItem} onClick={() => handleExport('csv')}>
                    CSV (.csv)
                  </button>
                  <button className={styles.exportItem} onClick={() => handleExport('json')}>
                    JSON (.json)
                  </button>
                  <button className={styles.exportItem} onClick={() => handleExport('pdf')}>
                    PDF (.pdf)
                  </button>
                </div>
              )}
            </div>
          )}
          <div className={styles.colToggleWrap}>
            <button
              className={cn(styles.toolbarBtn, styles.toolbarBtnIcon)}
              onClick={() => setShowColToggle(!showColToggle)}
              title="Tampilkan/sembunyikan kolom"
            >
              <Columns3 size={14} />
            </button>
            {showColToggle && (
              <div className={styles.colToggleDropdown}>
                <div className={styles.colToggleHeader}>
                  <span>Kolom</span>
                  <button
                    className={styles.colToggleReset}
                    onClick={() => {
                      const reset = Object.fromEntries(
                        initialColumns.map((c) => [c.key, true]),
                      )
                      setColVisibility(reset)
                    }}
                  >
                    Reset
                  </button>
                </div>
                <label className={styles.colToggleItem}>
                  <input
                    type="checkbox"
                    checked={allNonLockedVisible}
                    onChange={() => {
                      const newVis = { ...colVisibility }
                      visibleNonLocked.forEach((c) => {
                        newVis[c.key] = !allNonLockedVisible
                      })
                      setColVisibility(newVis)
                    }}
                  />
                  <span>Semua kolom</span>
                </label>
                <hr className={styles.colToggleDivider} />
                {initialColumns.map((col) => (
                  <label key={col.key} className={styles.colToggleItem}>
                    <input
                      type="checkbox"
                      checked={colVisibility[col.key] !== false}
                      disabled={col.locked}
                      onChange={() => {
                        setColVisibility((prev) => ({ ...prev, [col.key]: !prev[col.key] }))
                      }}
                    />
                    <span>{col.header}</span>
                  </label>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Inline quick filter */}
      {inlineFilter && filterDefs && filterDefs.length > 0 && onActiveFiltersChange && (
        <InlineFilter
          defs={filterDefs}
          onApply={(f) => onActiveFiltersChange([...activeFilters, f])}
        />
      )}

      {/* Table */}
      <div className={styles.tableWrapper}>
        <table className={styles.table}>
          <colgroup>
            {selectable && <col style={{ width: 44 }} />}
            {columns.map((col) => (
              <col key={col.key} style={{ width: col.width }} />
            ))}
          </colgroup>
          <thead className={styles.thead}>
            <tr>
              {selectable && (
                <th className={cn(styles.th, styles.checkboxCell)}>
                  <input
                    type="checkbox"
                    checked={data && data.length > 0 && data.every((r) => selected.includes(getRowKey(r)))}
                    ref={(el) => {
                      if (el && data) {
                        el.indeterminate = selected.length > 0 && !data.every((r) => selected.includes(getRowKey(r)))
                      }
                    }}
                    onChange={handleSelectAll}
                  />
                </th>
              )}
              {columns.map((col) => (
                <th
                  key={col.key}
                  className={cn(styles.th, col.sortable && styles.sortable, col.align === 'right' && styles.right, col.align === 'center' && styles.center)}
                  onClick={() => col.sortable && handleSort(col)}
                >
                  <span className={styles.thContent}>
                    {col.header}
                    {col.sortable && (
                      <span className={styles.sortIcon}>
                        {sort?.key === (col.sortKey ?? col.key) ? (
                          sort.order === 'asc' ? <ChevronUp size={12} /> : <ChevronDown size={12} />
                        ) : (
                          <ChevronsUpDown size={12} className={styles.sortIconInactive} />
                        )}
                      </span>
                    )}
                  </span>
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {isLoading ? (
              Array.from({ length: pageSize > 10 ? 10 : pageSize }).map((_, i) => (
                <SkeletonRow key={i} cols={columns.length} selectable={selectable} />
              ))
            ) : !data || data.length === 0 ? (
              <tr>
                <td colSpan={columns.length + (selectable ? 1 : 0)}>
                  <EmptyState title={emptyTitle} description={emptyDescription} />
                </td>
              </tr>
            ) : (
              data.map((row, rowIndex) => {
                const key = getRowKey(row)
                const isSelected = selected.includes(key)
                return (
                  <tr
                    key={key}
                    className={cn(
                      styles.tr,
                      rowIndex % 2 === 1 && styles.stripe,
                      isSelected && styles.selectedRow,
                      onRowClick && styles.clickable,
                    )}
                    onClick={() => onRowClick?.(row)}
                  >
                    {selectable && (
                      <td className={cn(styles.td, styles.checkboxCell)} onClick={(e) => e.stopPropagation()}>
                        <input
                          type="checkbox"
                          checked={isSelected}
                          onChange={() => handleSelectRow(key)}
                        />
                      </td>
                    )}
                    {columns.map((col) => (
                      <td
                        key={col.key}
                        className={cn(
                          styles.td,
                          col.align === 'right' && styles.right,
                          col.align === 'center' && styles.center,
                        )}
                      >
                        {col.render
                          ? col.render(row[col.key], row, rowIndex)
                          : String(row[col.key] ?? '')}
                      </td>
                    ))}
                  </tr>
                )
              })
            )}
          </tbody>
        </table>
      </div>

      {/* Footer */}
      {pagination && (
        <div className={styles.footer}>
          <div className={styles.footerInfo}>
            {total > 0 ? `Menampilkan ${start}–${end} dari ${total} data` : 'Tidak ada data'}
          </div>
          <div className={styles.footerPagination}>
            <button
              className={styles.pageBtn}
              onClick={() => onPageChange?.(0)}
              disabled={currentPage === 0}
              aria-label="Halaman pertama"
            >
              <ChevronsLeft size={14} />
            </button>
            <button
              className={styles.pageBtn}
              onClick={() => onPageChange?.(currentPage - 1)}
              disabled={currentPage === 0}
              aria-label="Halaman sebelumnya"
            >
              <ChevronLeft size={14} />
            </button>
            <span className={styles.pageInfo}>{currentPage + 1} / {totalPages}</span>
            <button
              className={styles.pageBtn}
              onClick={() => onPageChange?.(currentPage + 1)}
              disabled={currentPage >= totalPages - 1}
              aria-label="Halaman berikutnya"
            >
              <ChevronRight size={14} />
            </button>
            <button
              className={styles.pageBtn}
              onClick={() => onPageChange?.(totalPages - 1)}
              disabled={currentPage >= totalPages - 1}
              aria-label="Halaman terakhir"
            >
              <ChevronsRight size={14} />
            </button>
          </div>
          <div className={styles.pageSizeWrap}>
            <span className={styles.pageSizeLabel}>Per halaman:</span>
            <select
              className={styles.pageSizeSelect}
              value={pageSize}
              onChange={(e) => onPageSizeChange?.(Number(e.target.value))}
            >
              {PAGE_SIZE_OPTIONS.map((s) => (
                <option key={s} value={s}>{s}</option>
              ))}
            </select>
          </div>
        </div>
      )}

      {/* Filter dialog — portal ke document.body agar keluar dari overflow context */}
      {showFilterDialog && hasFilterDefs && createPortal(
        <FilterDialog
          defs={filterDefs!}
          activeFilters={localFilters}
          onSave={setLocalFilters}
          onClose={(saved) => {
            if (saved) {
              // User menyimpan perubahan, jangan sinkronisasi balik ke activeFilters
              setSkipSync(true)
            }
            setShowFilterDialog(false)
          }}
        />,
        document.body,
      )}
    </div>
  )
}
