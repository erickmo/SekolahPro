import type { BaseEntity } from './entity.types'

export type AcademicYearStatus = 'planning' | 'active' | 'closed'

export interface AcademicYear extends BaseEntity {
  name: string
  code: string
  startDate: string
  endDate: string
  semester1Start: string
  semester1End: string
  semester2Start: string
  semester2End: string
  isActive: boolean
  status: AcademicYearStatus
}

export const STATUS_LABELS: Record<AcademicYearStatus, string> = {
  planning: 'Perencanaan',
  active: 'Aktif',
  closed: 'Ditutup',
}
