import jsPDF from 'jspdf'
import autoTable from 'jspdf-autotable'

export type ExportFormat = 'csv' | 'json' | 'pdf'

export interface ExportColumn {
  key: string
  header: string
}

export function exportToCSV<T extends Record<string, unknown>>(
  filename: string,
  columns: ExportColumn[],
  data: T[],
): void {
  const header = columns.map((c) => `"${c.header}"`).join(',')
  const rows = data.map((row) =>
    columns
      .map((c) => {
        const val = row[c.key]
        const str = val === null || val === undefined ? '' : String(val)
        return `"${str.replace(/"/g, '""')}"`
      })
      .join(','),
  )

  const csv = '\uFEFF' + [header, ...rows].join('\n')
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${filename}.csv`
  a.click()
  URL.revokeObjectURL(url)
}

export function exportToJSON<T extends Record<string, unknown>>(
  filename: string,
  columns: ExportColumn[],
  data: T[],
): void {
  const rows = data.map((row) =>
    Object.fromEntries(columns.map((c) => [c.header, row[c.key] ?? null])),
  )
  const json = JSON.stringify(rows, null, 2)
  const blob = new Blob([json], { type: 'application/json;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${filename}.json`
  a.click()
  URL.revokeObjectURL(url)
}

export function exportToPDF<T extends Record<string, unknown>>(
  filename: string,
  columns: ExportColumn[],
  data: T[],
): void {
  const doc = new jsPDF()

  // Title
  doc.setFontSize(16)
  doc.text(filename.replace(/-/g, ' ').replace(/\b\w/g, (l) => l.toUpperCase()), 14, 15)

  // Table
  autoTable(doc, {
    startY: 25,
    head: [columns.map((c) => c.header)],
    body: data.map((row) => columns.map((c) => String(row[c.key] ?? ''))),
    styles: {
      fontSize: 9,
      cellPadding: 3,
    },
    headStyles: {
      fillColor: [77, 41, 117],
      textColor: 255,
      fontStyle: 'bold',
    },
    alternateRowStyles: {
      fillColor: [245, 245, 245],
    },
  })

  // Footer with timestamp
  const pageCount = doc.getNumberOfPages()
  doc.setFontSize(8)
  doc.setTextColor(128)

  for (let i = 1; i <= pageCount; i++) {
    doc.setPage(i)
    doc.text(
      `Generated: ${new Date().toLocaleString('id-ID')}`,
      14,
      doc.internal.pageSize.height - 10,
    )
    doc.text(
      `Page ${i} of ${pageCount}`,
      doc.internal.pageSize.width - 25,
      doc.internal.pageSize.height - 10,
    )
  }

  doc.save(`${filename}.pdf`)
}
