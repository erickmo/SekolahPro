import type { BaseEntity } from './entity.types'

// ── DailyAttendance ─────────────────────────────────────────────────────────────

export type AttendanceStatus = 'present' | 'absent' | 'late' | 'excused' | 'sick' | 'permit'

export interface DailyAttendance extends BaseEntity {
  studentId: string
  studentName: string
  studentNis: string
  classRoomId: string
  classRoomName: string
  classRoomGradeLevel: string
  attendanceDate: string
  status: AttendanceStatus
  notes: string
}

export const ATTENDANCE_STATUS_LABELS: Record<AttendanceStatus, string> = {
  present: 'Hadir',
  absent: 'Tidak Hadir',
  late: 'Terlambat',
  excused: 'Izin',
  sick: 'Sakit',
  permit: 'Dispensasi',
}

// ── AcademicRecord ───────────────────────────────────────────────────────────────

export type AcademicRecordStatus = 'pass' | 'fail' | 'remedial'

export interface AcademicRecord extends BaseEntity {
  studentId: string
  studentName: string
  studentNis: string
  academicYearId: string
  academicYearName: string
  academicYearCode: string
  semester: string
  gpa: number
  rank: number
  totalScore: number
  status: AcademicRecordStatus
  notes: string
}

export const ACADEMIC_RECORD_STATUS_LABELS: Record<AcademicRecordStatus, string> = {
  pass: 'Lulus',
  fail: 'Tidak Lulus',
  remedial: 'Remedial',
}

// ── SubjectGrade ─────────────────────────────────────────────────────────────────

export type GradeType = 'assignment' | 'midterm' | 'final' | 'quiz' | 'project'

export interface SubjectGrade extends BaseEntity {
  studentId: string
  studentName: string
  studentNis: string
  subjectId: string
  subjectName: string
  subjectCode: string
  teacherId: string
  teacherName: string
  teacherNip: string
  grade: number
  gradeType: GradeType
  examType: string
  weight: number
  description: string
}

export const GRADE_TYPE_LABELS: Record<GradeType, string> = {
  assignment: 'Tugas',
  midterm: 'UTS',
  final: 'UAS',
  quiz: 'Kuis',
  project: 'Proyek',
}

// ── ExamAssessment ───────────────────────────────────────────────────────────────

export type ExamType = 'uts' | 'uas' | 'quiz' | 'assignment' | 'project' | 'practice'
export type ExamStatus = 'draft' | 'published' | 'completed'

export interface ExamAssessment extends BaseEntity {
  subjectId: string
  subjectName: string
  subjectCode: string
  teacherId: string
  teacherName: string
  teacherNip: string
  classRoomId: string
  classRoomName: string
  classRoomGradeLevel: string
  title: string
  examType: ExamType
  date: string
  maxScore: number
  description: string
  status: ExamStatus
}

export const EXAM_TYPE_LABELS: Record<ExamType, string> = {
  uts: 'UTS',
  uas: 'UAS',
  quiz: 'Kuis',
  assignment: 'Tugas',
  project: 'Proyek',
  practice: 'Praktik',
}

export const EXAM_STATUS_LABELS: Record<ExamStatus, string> = {
  draft: 'Draft',
  published: 'Dipublikasi',
  completed: 'Selesai',
}

// ── HealthRecord ─────────────────────────────────────────────────────────────────

export type HealthRecordType = 'general' | 'vaccination' | 'allergy' | 'injury' | 'chronic' | 'vision' | 'dental'

export interface HealthRecord extends BaseEntity {
  studentId: string
  studentName: string
  studentNis: string
  recordType: HealthRecordType
  recordDate: string
  description: string
  diagnosis: string
  treatment: string
  doctorName: string
  followUpDate: string
  status: string
}

export const HEALTH_RECORD_TYPE_LABELS: Record<HealthRecordType, string> = {
  general: 'Umum',
  vaccination: 'Vaksinasi',
  allergy: 'Alergi',
  injury: 'Cedera',
  chronic: 'Kronis',
  vision: 'Penglihatan',
  dental: 'Gigi',
}

// ── Discipline ───────────────────────────────────────────────────────────────────

export type ViolationType = 'minor' | 'moderate' | 'major'
export type DisciplineStatus = 'reported' | 'reviewed' | 'sanctioned' | 'resolved'

export interface Discipline extends BaseEntity {
  studentId: string
  studentName: string
  studentNis: string
  teacherId: string
  teacherName: string
  teacherNip: string
  violationType: ViolationType
  incidentDate: string
  description: string
  sanction: string
  sanctionDate: string
  points: number
  status: DisciplineStatus
}

export const VIOLATION_TYPE_LABELS: Record<ViolationType, string> = {
  minor: 'Ringan',
  moderate: 'Sedang',
  major: 'Berat',
}

export const DISCIPLINE_STATUS_LABELS: Record<DisciplineStatus, string> = {
  reported: 'Dilaporkan',
  reviewed: 'Direview',
  sanctioned: 'Dikenakan Sanksi',
  resolved: 'Terselesaikan',
}

// ── Achievement ──────────────────────────────────────────────────────────────────

export type AchievementType = 'academic' | 'sport' | 'art' | 'technology' | 'social'
export type AchievementLevel = 'school' | 'district' | 'province' | 'national' | 'international'
export type AchievementStatus = 'draft' | 'verified' | 'rejected'

export interface Achievement extends BaseEntity {
  studentId: string
  studentName: string
  studentNis: string
  title: string
  achievementType: AchievementType
  level: AchievementLevel
  date: string
  organizer: string
  description: string
  certificateUrl: string
  status: AchievementStatus
}

export const ACHIEVEMENT_TYPE_LABELS: Record<AchievementType, string> = {
  academic: 'Akademik',
  sport: 'Olahraga',
  art: 'Seni',
  technology: 'Teknologi',
  social: 'Sosial',
}

export const ACHIEVEMENT_LEVEL_LABELS: Record<AchievementLevel, string> = {
  school: 'Sekolah',
  district: 'Kabupaten/Kota',
  province: 'Provinsi',
  national: 'Nasional',
  international: 'Internasional',
}

export const ACHIEVEMENT_STATUS_LABELS: Record<AchievementStatus, string> = {
  draft: 'Draft',
  verified: 'Terverifikasi',
  rejected: 'Ditolak',
}

// ── Extracurricular ──────────────────────────────────────────────────────────────

export type ActivityType = 'sport' | 'art' | 'academic' | 'social' | 'religion' | 'technology'
export type ExtracurricularStatus = 'active' | 'inactive' | 'completed'

export interface Extracurricular extends BaseEntity {
  studentId: string
  studentName: string
  studentNis: string
  teacherId: string
  teacherName: string
  teacherNip: string
  activityName: string
  activityType: ActivityType
  joinDate: string
  exitDate: string
  role: string
  achievement: string
  schedule: string
  status: ExtracurricularStatus
}

export const ACTIVITY_TYPE_LABELS: Record<ActivityType, string> = {
  sport: 'Olahraga',
  art: 'Seni',
  academic: 'Akademik',
  social: 'Sosial',
  religion: 'Keagamaan',
  technology: 'Teknologi',
}

export const EXTRACURRICULAR_STATUS_LABELS: Record<ExtracurricularStatus, string> = {
  active: 'Aktif',
  inactive: 'Nonaktif',
  completed: 'Selesai',
}
