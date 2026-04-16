import type { BaseEntity } from './entity.types'

export type Gender = 'L' | 'P'
export type EmployeeType = 'pns' | 'p3k' | 'honorer' | 'yayasan' | 'kontrak'
export type TeacherRole =
  | 'kepala_sekolah'
  | 'wakasek_kurikulum'
  | 'wakasek_kesiswaan'
  | 'wakasek_sarana'
  | 'wakasek_humas'
  | 'guru_mapel'
  | 'guru_bk'
  | 'guru_piket'
  | 'admin_tu'
  | 'bendahara'
  | 'pustakawan'
  | 'staff_umum'

export type TeacherStatus = 'active' | 'on_leave' | 'resigned' | 'retired'
export type Religion = 'islam' | 'kristen' | 'katolik' | 'hindu' | 'buddha' | 'konghucu'

export interface Teacher extends BaseEntity {
  fullName: string
  nip: string
  nuptk: string
  gender: Gender
  birthPlace: string
  birthDate: string
  religion: Religion
  phone: string
  email: string
  photoUrl: string
  employeeType: EmployeeType
  role: TeacherRole
  joinDate: string
  resignDate: string
  status: TeacherStatus
  signatureUrl: string
}

export const EMPLOYEE_TYPE_LABELS: Record<EmployeeType, string> = {
  pns: 'PNS',
  p3k: 'P3K',
  honorer: 'Honorer',
  yayasan: 'Yayasan',
  kontrak: 'Kontrak',
}

export const ROLE_LABELS: Record<TeacherRole, string> = {
  kepala_sekolah: 'Kepala Sekolah',
  wakasek_kurikulum: 'Waka Kurikulum',
  wakasek_kesiswaan: 'Waka Kesiswaan',
  wakasek_sarana: 'Waka Sarana',
  wakasek_humas: 'Waka Humas',
  guru_mapel: 'Guru Mapel',
  guru_bk: 'Guru BK',
  guru_piket: 'Guru Piket',
  admin_tu: 'Admin TU',
  bendahara: 'Bendahara',
  pustakawan: 'Pustakawan',
  staff_umum: 'Staff Umum',
}

export const STATUS_LABELS: Record<TeacherStatus, string> = {
  active: 'Aktif',
  on_leave: 'Cuti',
  resigned: 'Berhenti',
  retired: 'Pensiun',
}
