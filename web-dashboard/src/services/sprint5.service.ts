import { createVernonService } from './vernon.service'
import type {
  DailyAttendance,
  AcademicRecord,
  SubjectGrade,
  ExamAssessment,
  HealthRecord,
  Discipline,
  Achievement,
  Extracurricular,
} from '@/types/sprint5.types'

// ── DailyAttendance ──────────────────────────────────────────────────────────────

function transformDailyAttendance(raw: Record<string, unknown>): DailyAttendance {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const studentData = (data.student ?? {}) as Record<string, unknown>
  const classRoomData = (data.class_room ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    studentId: (raw.student_id ?? '') as string,
    studentName: (studentData.full_name ?? '') as string,
    studentNis: (studentData.nis ?? '') as string,
    classRoomId: (raw.class_room_id ?? '') as string,
    classRoomName: (classRoomData.name ?? '') as string,
    classRoomGradeLevel: (classRoomData.grade_level ?? '') as string,
    attendanceDate: (raw.attendance_date ?? '') as string,
    status: (raw.status ?? 'present') as DailyAttendance['status'],
    notes: (raw.notes ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const dailyAttendanceService = createVernonService<DailyAttendance, Record<string, unknown>>(
  '/daily_attendances',
  transformDailyAttendance,
)

// ── AcademicRecord ───────────────────────────────────────────────────────────────

function transformAcademicRecord(raw: Record<string, unknown>): AcademicRecord {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const studentData = (data.student ?? {}) as Record<string, unknown>
  const ayData = (data.academic_year ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    studentId: (raw.student_id ?? '') as string,
    studentName: (studentData.full_name ?? '') as string,
    studentNis: (studentData.nis ?? '') as string,
    academicYearId: (raw.academic_year_id ?? '') as string,
    academicYearName: (ayData.name ?? '') as string,
    academicYearCode: (ayData.code ?? '') as string,
    semester: (raw.semester ?? '1') as string,
    gpa: (raw.gpa ?? 0) as number,
    rank: (raw.rank ?? 0) as number,
    totalScore: (raw.total_score ?? 0) as number,
    status: (raw.status ?? 'pass') as AcademicRecord['status'],
    notes: (raw.notes ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const academicRecordService = createVernonService<AcademicRecord, Record<string, unknown>>(
  '/academic_records',
  transformAcademicRecord,
)

// ── SubjectGrade ─────────────────────────────────────────────────────────────────

function transformSubjectGrade(raw: Record<string, unknown>): SubjectGrade {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const studentData = (data.student ?? {}) as Record<string, unknown>
  const subjectData = (data.subject ?? {}) as Record<string, unknown>
  const teacherData = (data.teacher ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    studentId: (raw.student_id ?? '') as string,
    studentName: (studentData.full_name ?? '') as string,
    studentNis: (studentData.nis ?? '') as string,
    subjectId: (raw.subject_id ?? '') as string,
    subjectName: (subjectData.name ?? '') as string,
    subjectCode: (subjectData.code ?? '') as string,
    teacherId: (raw.teacher_id ?? '') as string,
    teacherName: (teacherData.full_name ?? '') as string,
    teacherNip: (teacherData.nip ?? '') as string,
    grade: (raw.grade ?? 0) as number,
    gradeType: (raw.grade_type ?? 'assignment') as SubjectGrade['gradeType'],
    examType: (raw.exam_type ?? '') as string,
    weight: (raw.weight ?? 0) as number,
    description: (raw.description ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const subjectGradeService = createVernonService<SubjectGrade, Record<string, unknown>>(
  '/subject_grades',
  transformSubjectGrade,
)

// ── ExamAssessment ───────────────────────────────────────────────────────────────

function transformExamAssessment(raw: Record<string, unknown>): ExamAssessment {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const subjectData = (data.subject ?? {}) as Record<string, unknown>
  const teacherData = (data.teacher ?? {}) as Record<string, unknown>
  const classRoomData = (data.class_room ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    subjectId: (raw.subject_id ?? '') as string,
    subjectName: (subjectData.name ?? '') as string,
    subjectCode: (subjectData.code ?? '') as string,
    teacherId: (raw.teacher_id ?? '') as string,
    teacherName: (teacherData.full_name ?? '') as string,
    teacherNip: (teacherData.nip ?? '') as string,
    classRoomId: (raw.class_room_id ?? '') as string,
    classRoomName: (classRoomData.name ?? '') as string,
    classRoomGradeLevel: (classRoomData.grade_level ?? '') as string,
    title: (raw.title ?? '') as string,
    examType: (raw.exam_type ?? 'quiz') as ExamAssessment['examType'],
    date: (raw.date ?? '') as string,
    maxScore: (raw.max_score ?? 100) as number,
    description: (raw.description ?? '') as string,
    status: (raw.status ?? 'draft') as ExamAssessment['status'],
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const examAssessmentService = createVernonService<ExamAssessment, Record<string, unknown>>(
  '/exam_assessments',
  transformExamAssessment,
)

// ── HealthRecord ─────────────────────────────────────────────────────────────────

function transformHealthRecord(raw: Record<string, unknown>): HealthRecord {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const studentData = (data.student ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    studentId: (raw.student_id ?? '') as string,
    studentName: (studentData.full_name ?? '') as string,
    studentNis: (studentData.nis ?? '') as string,
    recordType: (raw.record_type ?? 'general') as HealthRecord['recordType'],
    recordDate: (raw.record_date ?? '') as string,
    description: (raw.description ?? '') as string,
    diagnosis: (raw.diagnosis ?? '') as string,
    treatment: (raw.treatment ?? '') as string,
    doctorName: (raw.doctor_name ?? '') as string,
    followUpDate: (raw.follow_up_date ?? '') as string,
    status: (raw.status ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const healthRecordService = createVernonService<HealthRecord, Record<string, unknown>>(
  '/health_records',
  transformHealthRecord,
)

// ── Discipline ───────────────────────────────────────────────────────────────────

function transformDiscipline(raw: Record<string, unknown>): Discipline {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const studentData = (data.student ?? {}) as Record<string, unknown>
  const teacherData = (data.teacher ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    studentId: (raw.student_id ?? '') as string,
    studentName: (studentData.full_name ?? '') as string,
    studentNis: (studentData.nis ?? '') as string,
    teacherId: (raw.teacher_id ?? '') as string,
    teacherName: (teacherData.full_name ?? '') as string,
    teacherNip: (teacherData.nip ?? '') as string,
    violationType: (raw.violation_type ?? 'minor') as Discipline['violationType'],
    incidentDate: (raw.incident_date ?? '') as string,
    description: (raw.description ?? '') as string,
    sanction: (raw.sanction ?? '') as string,
    sanctionDate: (raw.sanction_date ?? '') as string,
    points: (raw.points ?? 0) as number,
    status: (raw.status ?? 'reported') as Discipline['status'],
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const disciplineService = createVernonService<Discipline, Record<string, unknown>>(
  '/disciplines',
  transformDiscipline,
)

// ── Achievement ──────────────────────────────────────────────────────────────────

function transformAchievement(raw: Record<string, unknown>): Achievement {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const studentData = (data.student ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    studentId: (raw.student_id ?? '') as string,
    studentName: (studentData.full_name ?? '') as string,
    studentNis: (studentData.nis ?? '') as string,
    title: (raw.title ?? '') as string,
    achievementType: (raw.achievement_type ?? 'academic') as Achievement['achievementType'],
    level: (raw.level ?? 'school') as Achievement['level'],
    date: (raw.date ?? '') as string,
    organizer: (raw.organizer ?? '') as string,
    description: (raw.description ?? '') as string,
    certificateUrl: (raw.certificate_url ?? '') as string,
    status: (raw.status ?? 'draft') as Achievement['status'],
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const achievementService = createVernonService<Achievement, Record<string, unknown>>(
  '/achievements',
  transformAchievement,
)

// ── Extracurricular ──────────────────────────────────────────────────────────────

function transformExtracurricular(raw: Record<string, unknown>): Extracurricular {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const studentData = (data.student ?? {}) as Record<string, unknown>
  const teacherData = (data.teacher ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    studentId: (raw.student_id ?? '') as string,
    studentName: (studentData.full_name ?? '') as string,
    studentNis: (studentData.nis ?? '') as string,
    teacherId: (raw.teacher_id ?? '') as string,
    teacherName: (teacherData.full_name ?? '') as string,
    teacherNip: (teacherData.nip ?? '') as string,
    activityName: (raw.activity_name ?? '') as string,
    activityType: (raw.activity_type ?? 'sport') as Extracurricular['activityType'],
    joinDate: (raw.join_date ?? '') as string,
    exitDate: (raw.exit_date ?? '') as string,
    role: (raw.role ?? '') as string,
    achievement: (raw.achievement ?? '') as string,
    schedule: (raw.schedule ?? '') as string,
    status: (raw.status ?? 'active') as Extracurricular['status'],
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const extracurricularService = createVernonService<Extracurricular, Record<string, unknown>>(
  '/extracurriculars',
  transformExtracurricular,
)
