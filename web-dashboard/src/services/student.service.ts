import { createVernonService } from './vernon.service'
import type {
  Student,
  StudentGuardian,
  StudentDocument,
  StudentClassPlacement,
  StudentAdmission,
} from '@/types/student.types'

// ── Students ──────────────────────────────────────────────────────────────────

function transformStudent(raw: Record<string, unknown>): Student {
  return {
    id: raw.id as string,
    nis: (raw.nis ?? '') as string,
    nisn: (raw.nisn ?? '') as string,
    fullName: (raw.full_name ?? '') as string,
    nickname: (raw.nickname ?? '') as string,
    gender: (raw.gender ?? 'L') as Student['gender'],
    birthPlace: (raw.birth_place ?? '') as string,
    birthDate: (raw.birth_date ?? '') as string,
    religion: (raw.religion ?? 'islam') as Student['religion'],
    bloodType: (raw.blood_type ?? '') as string,
    phone: (raw.phone ?? '') as string,
    email: (raw.email ?? '') as string,
    photoUrl: (raw.photo_url ?? '') as string,
    status: (raw.status ?? 'active') as Student['status'],
    admissionType: (raw.admission_type ?? '') as string,
    notes: (raw.notes ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const studentService = createVernonService<Student, Record<string, unknown>>(
  '/students',
  transformStudent,
)

// ── Student Guardians ─────────────────────────────────────────────────────────

function transformStudentGuardian(raw: Record<string, unknown>): StudentGuardian {
  const data = (raw._data ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    studentId: (raw.student_id ?? '') as string,
    studentName: (data.student_full_name ?? '') as string,
    guardianType: (raw.guardian_type ?? 'ayah') as StudentGuardian['guardianType'],
    fullName: (raw.full_name ?? '') as string,
    nik: (raw.nik ?? '') as string,
    occupation: (raw.occupation ?? '') as string,
    phone: (raw.phone ?? '') as string,
    email: (raw.email ?? '') as string,
    address: (raw.address ?? '') as string,
    isPrimary: (raw.is_primary ?? false) as boolean,
    educationLevel: (raw.education_level ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const studentGuardianService = createVernonService<StudentGuardian, Record<string, unknown>>(
  '/student_guardians',
  transformStudentGuardian,
)

// ── Student Documents ─────────────────────────────────────────────────────────

function transformStudentDocument(raw: Record<string, unknown>): StudentDocument {
  const data = (raw._data ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    studentId: (raw.student_id ?? '') as string,
    studentName: (data.student_full_name ?? '') as string,
    documentType: (raw.document_type ?? 'photo') as StudentDocument['documentType'],
    title: (raw.title ?? '') as string,
    fileName: (raw.file_name ?? '') as string,
    fileUrl: (raw.file_url ?? '') as string,
    fileSize: (raw.file_size ?? 0) as number,
    mimeType: (raw.mime_type ?? '') as string,
    status: (raw.status ?? 'pending') as StudentDocument['status'],
    verifiedBy: (raw.verified_by ?? '') as string,
    verifiedAt: (raw.verified_at ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const studentDocumentService = createVernonService<StudentDocument, Record<string, unknown>>(
  '/student_documents',
  transformStudentDocument,
)

// ── Student Class Placements ──────────────────────────────────────────────────

function transformStudentClassPlacement(raw: Record<string, unknown>): StudentClassPlacement {
  const data = (raw._data ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    studentId: (raw.student_id ?? '') as string,
    studentName: (data.student_full_name ?? '') as string,
    classRoomId: (raw.class_room_id ?? '') as string,
    classRoomName: (data.class_room_name ?? '') as string,
    academicYearId: (raw.academic_year_id ?? '') as string,
    academicYearName: (data.academic_year_name ?? '') as string,
    semester: (raw.semester ?? '1') as string,
    status: (raw.status ?? 'active') as StudentClassPlacement['status'],
    enrollmentDate: (raw.enrollment_date ?? '') as string,
    exitDate: (raw.exit_date ?? '') as string,
    exitReason: (raw.exit_reason ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const studentClassPlacementService = createVernonService<StudentClassPlacement, Record<string, unknown>>(
  '/student_class_placements',
  transformStudentClassPlacement,
)

// ── Student Admissions (PPDB) ────────────────────────────────────────────────

function transformStudentAdmission(raw: Record<string, unknown>): StudentAdmission {
  const data = (raw._data ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    registrationNumber: (raw.registration_number ?? '') as string,
    studentId: (raw.student_id ?? '') as string,
    studentName: (data.student_full_name ?? '') as string,
    academicYearId: (raw.academic_year_id ?? '') as string,
    academicYearName: (data.academic_year_name ?? '') as string,
    admissionType: (raw.admission_type ?? 'regular') as StudentAdmission['admissionType'],
    status: (raw.status ?? 'pending') as StudentAdmission['status'],
    registrationDate: (raw.registration_date ?? '') as string,
    testDate: (raw.test_date ?? '') as string,
    testScore: (raw.test_score ?? 0) as number,
    interviewScore: (raw.interview_score ?? 0) as number,
    finalScore: (raw.final_score ?? 0) as number,
    notes: (raw.notes ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const studentAdmissionService = createVernonService<StudentAdmission, Record<string, unknown>>(
  '/student_admissions',
  transformStudentAdmission,
)
