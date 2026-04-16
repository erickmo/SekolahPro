import type { BaseEntity } from './entity.types'

// Student Core (S001)
export type StudentStatus = 'active' | 'inactive' | 'graduated' | 'transferred' | 'dropped_out'
export type StudentGender = 'L' | 'P'
export type StudentReligion = 'islam' | 'kristen' | 'katolik' | 'hindu' | 'buddha' | 'konghucu'

export interface Student extends BaseEntity {
  nis: string
  nisn: string
  fullName: string
  nickname: string
  gender: StudentGender
  birthPlace: string
  birthDate: string
  religion: StudentReligion
  bloodType: string
  phone: string
  email: string
  photoUrl: string
  status: StudentStatus
  admissionType: string
  notes: string
}

export const STUDENT_STATUS_LABELS: Record<StudentStatus, string> = {
  active: 'Aktif',
  inactive: 'Nonaktif',
  graduated: 'Lulus',
  transferred: 'Pindah',
  dropped_out: 'Drop Out',
}

export const STUDENT_GENDER_LABELS: Record<StudentGender, string> = {
  L: 'Laki-laki',
  P: 'Perempuan',
}

export const STUDENT_RELIGION_LABELS: Record<StudentReligion, string> = {
  islam: 'Islam',
  kristen: 'Kristen',
  katolik: 'Katolik',
  hindu: 'Hindu',
  buddha: 'Buddha',
  konghucu: 'Konghucu',
}

// Student Guardian (S003)
export type GuardianType = 'ayah' | 'ibu' | 'wali'

export interface StudentGuardian extends BaseEntity {
  studentId: string
  studentName: string
  guardianType: GuardianType
  fullName: string
  nik: string
  occupation: string
  phone: string
  email: string
  address: string
  isPrimary: boolean
  educationLevel: string
}

export const GUARDIAN_TYPE_LABELS: Record<GuardianType, string> = {
  ayah: 'Ayah',
  ibu: 'Ibu',
  wali: 'Wali',
}

// Student Document (S010)
export type DocumentType = 'akta_kelahiran' | 'kartu_keluarga' | 'ijazah' | 'skhun' | 'photo' | 'ktp_ortu' | 'surat_pindah' | 'sktm'
export type DocumentStatus = 'pending' | 'verified' | 'rejected'

export interface StudentDocument extends BaseEntity {
  studentId: string
  studentName: string
  documentType: DocumentType
  title: string
  fileName: string
  fileUrl: string
  fileSize: number
  mimeType: string
  status: DocumentStatus
  verifiedBy: string
  verifiedAt: string
}

export const DOCUMENT_TYPE_LABELS: Record<DocumentType, string> = {
  akta_kelahiran: 'Akta Kelahiran',
  kartu_keluarga: 'Kartu Keluarga',
  ijazah: 'Ijazah',
  skhun: 'SKHUN',
  photo: 'Pas Foto',
  ktp_ortu: 'KTP Orang Tua',
  surat_pindah: 'Surat Pindah',
  sktm: 'SKTM',
}

export const DOCUMENT_STATUS_LABELS: Record<DocumentStatus, string> = {
  pending: 'Menunggu',
  verified: 'Terverifikasi',
  rejected: 'Ditolak',
}

// Student Class Placement (S014)
export type PlacementStatus = 'active' | 'moved' | 'graduated'

export interface StudentClassPlacement extends BaseEntity {
  studentId: string
  studentName: string
  classRoomId: string
  classRoomName: string
  academicYearId: string
  academicYearName: string
  semester: string
  status: PlacementStatus
  enrollmentDate: string
  exitDate: string
  exitReason: string
}

export const PLACEMENT_STATUS_LABELS: Record<PlacementStatus, string> = {
  active: 'Aktif',
  moved: 'Pindah',
  graduated: 'Naik Kelas',
}

// Student Admission / PPDB (S016)
export type AdmissionType = 'regular' | 'pindahan' | 'afirmasi' | 'prestasi'
export type AdmissionStatus = 'pending' | 'document_review' | 'test' | 'written_test' | 'interview' | 'accepted' | 'rejected'

export interface StudentAdmission extends BaseEntity {
  registrationNumber: string
  studentId: string
  studentName: string
  academicYearId: string
  academicYearName: string
  admissionType: AdmissionType
  status: AdmissionStatus
  registrationDate: string
  testDate: string
  testScore: number
  interviewScore: number
  finalScore: number
  notes: string
}

export const ADMISSION_TYPE_LABELS: Record<AdmissionType, string> = {
  regular: 'Reguler',
  pindahan: 'Pindahan',
  afirmasi: 'Afirmasi',
  prestasi: 'Prestasi',
}

export const ADMISSION_STATUS_LABELS: Record<AdmissionStatus, string> = {
  pending: 'Menunggu',
  document_review: 'Verifikasi Dokumen',
  test: 'Tes',
  written_test: 'Tes Tulis',
  interview: 'Wawancara',
  accepted: 'Diterima',
  rejected: 'Ditolak',
}
