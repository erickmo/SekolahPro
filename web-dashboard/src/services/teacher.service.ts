import { createVernonService } from './vernon.service'
import type { Teacher } from '@/types/teacher.types'

function transformTeacher(raw: Record<string, unknown>): Teacher {
  return {
    id: raw.id as string,
    fullName: (raw.full_name ?? '') as string,
    nip: (raw.nip ?? '') as string,
    nuptk: (raw.nuptk ?? '') as string,
    gender: (raw.gender ?? 'L') as Teacher['gender'],
    birthPlace: (raw.birth_place ?? '') as string,
    birthDate: (raw.birth_date ?? '') as string,
    religion: (raw.religion ?? '') as Teacher['religion'],
    phone: (raw.phone ?? '') as string,
    email: (raw.email ?? '') as string,
    photoUrl: (raw.photo_url ?? '') as string,
    employeeType: (raw.employee_type ?? 'kontrak') as Teacher['employeeType'],
    role: (raw.role ?? 'guru_mapel') as Teacher['role'],
    joinDate: (raw.join_date ?? '') as string,
    resignDate: (raw.resign_date ?? '') as string,
    status: (raw.status ?? 'active') as Teacher['status'],
    signatureUrl: (raw.signature_url ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const teacherService = createVernonService<Teacher, Record<string, unknown>>(
  '/teachers',
  transformTeacher,
)
