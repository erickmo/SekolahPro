import { useState, useEffect } from 'react'
import { Plus, Trash2, HelpCircle, X } from 'lucide-react'
import { useQueryClient } from '@tanstack/react-query'
import { useSearchParams } from 'react-router-dom'
import { PageHeader } from '@/layouts/PageHeader/PageHeader'
import { PageWrapper } from '@/widgets/PageWrapper/PageWrapper'
import { DataTable } from '@/widgets/DataTable/DataTable'
import { useDataSource } from '@/hooks/useDataSource'
import { useModuleAccess } from '@/hooks/useModuleAccess'
import { toast } from '@/widgets/Toast/Toast'
import { StatusPills } from '@/widgets/StatusPills/StatusPills'
import type { ColumnDef, RowActionDef, FilterDef, ActiveFilter } from '@/widgets/DataTable/DataTable'
import { serializeFilters, parseFiltersFromURL } from '@/widgets/DataTable/filter.utils'
import type { ListParams } from '@/services/createEntityService'
import type { PaginatedResponse } from '@/types/api.types'
import styles from './ListPageTemplate.module.css'

// ─── Types ────────────────────────────────────────────────────────────────────

export interface DeleteConfig<T> {
  /** The API call. Should throw on error. */
  onDelete: (row: T) => Promise<void>
  /** Dialog title, e.g. "Hapus Item?" */
  dialogTitle: string
  /** Dialog body content */
  dialogBody: (row: T) => React.ReactNode
  /** Toast success message */
  successMessage: (row: T) => string
  /** Toast error message fallback */
  errorMessage?: string
}

interface ListPageTemplateProps<T extends { id: string }> {
  /** Page title shown in PageHeader */
  title: string

  /** When provided, renders an "Add" button in the header */
  addLabel?: string
  onAdd?: () => void

  /** Data source — passed to useDataSource internally */
  queryKey: string
  fetcher: (params: ListParams) => Promise<PaginatedResponse<T>>
  defaultPageSize?: number

  /** Table configuration */
  columns: ColumnDef<T>[]
  rowActions?: RowActionDef<T>[]
  onRowClick?: (row: T) => void
  searchPlaceholder?: string
  exportFilename?: string
  /** Custom PDF export handler */
  onExportPDF?: (filename: string, data: T[]) => void
  emptyTitle?: string
  emptyDescription?: string

  /** When provided, a (?) icon appears in the header that opens a help modal */
  helpTitle?: string
  helpText?: string

  /** When provided, a "Hapus" row action is injected and delete dialog is managed */
  deleteConfig?: DeleteConfig<T>

  /** When provided, a Filter button appears in the DataTable toolbar */
  filterDefs?: FilterDef[]

  /** When provided, sets the initial sort for the data source */
  defaultSort?: { key: string; order: 'asc' | 'desc' }

  /** When true: hides pagination footer and loads all data (uses pageSize 9999) */
  hidePagination?: boolean

  /** When true: hides Add button, suppresses delete row action, shows "Hanya Baca" pill */
  readonly?: boolean
  /** When true: shows "Dikelola HQ" pill */
  managedByHQ?: boolean

  /** When provided, custom action buttons are rendered in the header alongside Add/Help buttons */
  actions?: React.ReactNode
}

// ─── Component ────────────────────────────────────────────────────────────────

export function ListPageTemplate<T extends { id: string }>({
  title,
  addLabel,
  onAdd,
  queryKey,
  fetcher,
  defaultPageSize = 100,
  columns,
  rowActions = [],
  onRowClick,
  searchPlaceholder,
  exportFilename,
  onExportPDF,
  emptyTitle,
  emptyDescription,
  helpTitle,
  helpText,
  deleteConfig,
  filterDefs,
  defaultSort,
  hidePagination = false,
  readonly: isReadonly = false,
  managedByHQ = false,
  actions,
}: ListPageTemplateProps<T>) {
  const moduleAccess = useModuleAccess()
  const effectiveReadonly = isReadonly || moduleAccess.readonly
  const effectiveManagedByHQ = managedByHQ || moduleAccess.managedByHQ

  const queryClient = useQueryClient()
  const [searchParams] = useSearchParams()
  const [deletingRow, setDeletingRow] = useState<T | null>(null)
  const [isDeleting, setIsDeleting] = useState(false)
  const [showHelp, setShowHelp] = useState(false)
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([])
  const [urlFiltersApplied, setUrlFiltersApplied] = useState(false)
  const hasHelp = !!(helpTitle || helpText)

  const {
    data, total, isLoading, error, refetch,
    pagination, sort, search,
    setPage, setPageSize, setSort, setSearch, setFilters,
  } = useDataSource<T>({
    queryKey,
    fetcher,
    defaultPageSize: hidePagination ? 9999 : defaultPageSize,
    defaultSort,
  })

  // Parse filter dari URL query string (?filters=[...]) saat pertama kali dimuat
  useEffect(() => {
    if (urlFiltersApplied) return
    if (!filterDefs) {
      setUrlFiltersApplied(true)
      return
    }
    const { matched, adhoc } = parseFiltersFromURL(searchParams, filterDefs)
    const hasFilters = matched.length > 0 || adhoc.length > 0
    if (hasFilters) {
      // Sertakan SEMUA filter (matched + adhoc) ke UI agar user bisa melihatnya
      setActiveFilters([...matched, ...adhoc])
      // Gabung matched filters + adhoc filters (semua dalam format baru)
      const allFilters = [...matched, ...adhoc]
      setFilters(serializeFilters(allFilters))
    }
    setUrlFiltersApplied(true)
  }, [searchParams, filterDefs, urlFiltersApplied, setFilters])

  function handleActiveFiltersChange(filters: ActiveFilter[]) {
    setActiveFilters(filters)
    setFilters(serializeFilters(filters))
  }

  const allRowActions: RowActionDef<T>[] = deleteConfig && !effectiveReadonly && !effectiveManagedByHQ
    ? [
        ...rowActions,
        {
          key: '_delete',
          label: 'Hapus',
          icon: <Trash2 size={14} />,
          variant: 'danger' as const,
          onClick: (row) => setDeletingRow(row),
        },
      ]
    : rowActions

  async function handleConfirmDelete() {
    if (!deletingRow || !deleteConfig) return
    setIsDeleting(true)
    try {
      await deleteConfig.onDelete(deletingRow)
      await queryClient.invalidateQueries({ queryKey: [queryKey] })
      toast.success(deleteConfig.successMessage(deletingRow))
      setDeletingRow(null)
      void refetch()
    } catch (err) {
      toast.error(
        err instanceof Error
          ? err.message
          : (deleteConfig.errorMessage ?? 'Gagal menghapus data'),
      )
    } finally {
      setIsDeleting(false)
    }
  }

  return (
    <>
      <PageHeader
        title={title}
        pills={<StatusPills readonly={effectiveReadonly} managedByHQ={effectiveManagedByHQ} />}
        actions={
          (actions || (onAdd && addLabel && !effectiveReadonly && !effectiveManagedByHQ) || hasHelp) ? (
            <>
              {actions}
              {hasHelp && (
                <button
                  className={styles.helpBtn}
                  onClick={() => setShowHelp(true)}
                  title="Tentang halaman ini"
                >
                  <HelpCircle size={18} />
                </button>
              )}
              {onAdd && addLabel && !effectiveReadonly && !effectiveManagedByHQ && (
                <button className={styles.addBtn} onClick={onAdd}>
                  <Plus size={16} />
                  {addLabel}
                </button>
              )}
            </>
          ) : undefined
        }
      />

      <PageWrapper error={error as Error | null} onRetry={refetch}>
        <DataTable<T>
          columns={columns}
          data={data}
          isLoading={isLoading}
          pagination={hidePagination ? undefined : { page: pagination.page, pageSize: pagination.pageSize, total }}
          onPageChange={hidePagination ? undefined : setPage}
          onPageSizeChange={hidePagination ? undefined : setPageSize}
          sort={sort}
          onSortChange={setSort}
          search={search}
          onSearchChange={setSearch}
          searchPlaceholder={searchPlaceholder}
          rowActions={allRowActions}
          onRowClick={onRowClick}
          exportFilename={exportFilename}
          onExportPDF={onExportPDF}
          emptyTitle={emptyTitle}
          emptyDescription={emptyDescription}
          filterDefs={filterDefs}
          activeFilters={activeFilters}
          onActiveFiltersChange={handleActiveFiltersChange}
        />
      </PageWrapper>

      {showHelp && hasHelp && (
        <div className={styles.overlay} onClick={() => setShowHelp(false)}>
          <div className={styles.helpModal} onClick={(e) => e.stopPropagation()}>
            <div className={styles.helpModalHeader}>
              <div className={styles.helpModalIcon}>
                <HelpCircle size={20} />
              </div>
              <h3 className={styles.helpModalTitle}>{helpTitle ?? title}</h3>
              <button className={styles.helpModalClose} onClick={() => setShowHelp(false)}>
                <X size={16} />
              </button>
            </div>
            <p className={styles.helpModalBody}>{helpText}</p>
          </div>
        </div>
      )}

      {deleteConfig && deletingRow && (
        <div className={styles.overlay} onClick={() => !isDeleting && setDeletingRow(null)}>
          <div className={styles.dialog} onClick={(e) => e.stopPropagation()}>
            <div className={styles.dialogIcon}>
              <Trash2 size={24} />
            </div>
            <h3 className={styles.dialogTitle}>{deleteConfig.dialogTitle}</h3>
            <p className={styles.dialogBody}>{deleteConfig.dialogBody(deletingRow)}</p>
            <div className={styles.dialogActions}>
              <button
                className={styles.btnSecondary}
                onClick={() => setDeletingRow(null)}
                disabled={isDeleting}
              >
                Batal
              </button>
              <button
                className={styles.btnDanger}
                onClick={handleConfirmDelete}
                disabled={isDeleting}
              >
                {isDeleting ? 'Menghapus...' : 'Ya, Hapus'}
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  )
}
