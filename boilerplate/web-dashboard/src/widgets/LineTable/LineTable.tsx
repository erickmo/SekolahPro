import { Ellipsis, Trash2, Plus } from 'lucide-react'
import styles from './LineTable.module.css'

// ─── Column definition ────────────────────────────────────────────────────────

export interface LineTableColumn<T> {
  key: string
  label: string
  /** CSS width, e.g. '80px', '1fr', '20%'. Default: auto */
  width?: string
  align?: 'left' | 'right' | 'center'
  render: (
    row: T,
    index: number,
    onChange: (updates: Partial<T>) => void,
  ) => React.ReactNode
}

// ─── Props ────────────────────────────────────────────────────────────────────

interface LineTableProps<T> {
  rows: T[]
  columns: LineTableColumn<T>[]
  onAdd: () => void
  onDelete: (index: number) => void
  onRowChange: (index: number, updates: Partial<T>) => void
  /** Opens extended-fields dialog for the row. Caller controls dialog content. */
  onMore?: (index: number) => void
  addLabel?: string
  moreTooltip?: string
  emptyMessage?: string
  /** Rendered inside <tfoot> — useful for totals row */
  footer?: React.ReactNode
}

// ─── LineTable ────────────────────────────────────────────────────────────────

export function LineTable<T>({
  rows,
  columns,
  onAdd,
  onDelete,
  onRowChange,
  onMore,
  addLabel = 'Tambah Baris',
  moreTooltip = 'Detail lainnya',
  emptyMessage = 'Belum ada baris. Klik "+ Tambah Baris" untuk mulai.',
  footer,
}: LineTableProps<T>) {
  return (
    <div className={styles.wrapper}>
      <table className={styles.table}>
        <colgroup>
          <col style={{ width: '36px' }} />
          {columns.map((col) => (
            <col key={col.key} style={{ width: col.width ?? 'auto' }} />
          ))}
          <col style={{ width: onMore ? '64px' : '40px' }} />
        </colgroup>

        <thead>
          <tr>
            <th className={`${styles.th} ${styles.thNum}`}>#</th>
            {columns.map((col) => (
              <th
                key={col.key}
                className={styles.th}
                style={{ textAlign: col.align ?? 'left' }}
              >
                {col.label}
              </th>
            ))}
            <th className={`${styles.th} ${styles.thActions}`} />
          </tr>
        </thead>

        <tbody>
          {rows.length === 0 && (
            <tr>
              <td colSpan={columns.length + 2} className={styles.emptyCell}>
                {emptyMessage}
              </td>
            </tr>
          )}
          {rows.map((row, index) => (
            <tr key={index} className={styles.tr}>
              <td className={`${styles.td} ${styles.tdNum}`}>{index + 1}</td>
              {columns.map((col) => (
                <td
                  key={col.key}
                  className={styles.td}
                  style={{ textAlign: col.align ?? 'left' }}
                >
                  {col.render(row, index, (updates) => onRowChange(index, updates))}
                </td>
              ))}
              <td className={`${styles.td} ${styles.tdActions}`}>
                {onMore && (
                  <button
                    type="button"
                    className={styles.actionBtn}
                    onClick={() => onMore(index)}
                    title={moreTooltip}
                  >
                    <Ellipsis size={14} />
                  </button>
                )}
                <button
                  type="button"
                  className={`${styles.actionBtn} ${styles.actionBtnDelete}`}
                  onClick={() => onDelete(index)}
                  title="Hapus baris"
                >
                  <Trash2 size={14} />
                </button>
              </td>
            </tr>
          ))}
        </tbody>

        {footer && (
          <tfoot>
            <tr>
              <td colSpan={columns.length + 2} className={styles.footer}>
                {footer}
              </td>
            </tr>
          </tfoot>
        )}
      </table>

      <div className={styles.addRow}>
        <button type="button" className={styles.addBtn} onClick={onAdd}>
          <Plus size={14} />
          {addLabel}
        </button>
      </div>
    </div>
  )
}

// ─── CellInput ────────────────────────────────────────────────────────────────

interface CellInputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  mono?: boolean
  align?: 'left' | 'right' | 'center'
  error?: boolean
}

export function CellInput({ mono, align, error, className, ...props }: CellInputProps) {
  return (
    <input
      className={[
        styles.cellInput,
        mono ? styles.cellInputMono : '',
        align === 'right' ? styles.cellInputRight : '',
        align === 'center' ? styles.cellInputCenter : '',
        error ? styles.cellInputError : '',
        className ?? '',
      ].join(' ')}
      {...props}
    />
  )
}

// ─── CellSelect ───────────────────────────────────────────────────────────────

interface CellSelectProps extends React.SelectHTMLAttributes<HTMLSelectElement> {
  error?: boolean
}

export function CellSelect({ error, className, children, ...props }: CellSelectProps) {
  return (
    <select
      className={[
        styles.cellSelect,
        error ? styles.cellInputError : '',
        className ?? '',
      ].join(' ')}
      {...props}
    >
      {children}
    </select>
  )
}

// ─── CellValue (read-only display) ───────────────────────────────────────────

interface CellValueProps {
  children: React.ReactNode
  mono?: boolean
  align?: 'left' | 'right' | 'center'
  secondary?: boolean
}

export function CellValue({ children, mono, align, secondary }: CellValueProps) {
  return (
    <span
      className={[
        styles.cellValue,
        mono ? styles.cellValueMono : '',
        align === 'right' ? styles.cellValueRight : '',
        secondary ? styles.cellValueSecondary : '',
      ].join(' ')}
    >
      {children}
    </span>
  )
}
