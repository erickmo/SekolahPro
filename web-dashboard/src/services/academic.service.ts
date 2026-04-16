import { createVernonService } from './vernon.service'
import type { AcademicYear } from '@/types/academic.types'

function transformAcademicYear(raw: Record<string, unknown>): AcademicYear {
  return {
    id: raw.id as string,
    name: raw.name as string,
    code: raw.code as string,
    startDate: (raw.start_date ?? '') as string,
    endDate: (raw.end_date ?? '') as string,
    semester1Start: (raw.semester1_start ?? '') as string,
    semester1End: (raw.semester1_end ?? '') as string,
    semester2Start: (raw.semester2_start ?? '') as string,
    semester2End: (raw.semester2_end ?? '') as string,
    isActive: (raw.is_active ?? false) as boolean,
    status: (raw.status ?? 'planning') as AcademicYear['status'],
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const academicYearService = createVernonService<AcademicYear, Record<string, unknown>>(
  '/academic_years',
  transformAcademicYear,
)
