import type { BaseEntity } from './entity.types'

export type Major = 'ipa' | 'ips' | 'bahasa' | 'agama' | 'umum'

export interface ClassRoomEmbeddedAcademicYear {
  name: string
  code: string
  isActive: boolean
}

export interface ClassRoomEmbeddedTeacher {
  fullName: string
  nip: string
  nuptk: string
}

export interface ClassRoom extends BaseEntity {
  name: string
  gradeLevel: string
  parallelId: string
  capacity: number
  currentCount: number
  major: Major
  isActive: boolean
  academicYearId: string
  homeroomTeacherId: string
  academicYear?: ClassRoomEmbeddedAcademicYear
  homeroomTeacher?: ClassRoomEmbeddedTeacher
}

export const MAJOR_LABELS: Record<Major, string> = {
  ipa: 'IPA',
  ips: 'IPS',
  bahasa: 'Bahasa',
  agama: 'Agama',
  umum: 'Umum',
}
