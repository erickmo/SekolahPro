import type { BaseEntity } from './entity.types'

// ── Curriculum ─────────────────────────────────────────────────────────────────

export type CurriculumType = 'national' | 'school' | 'international' | 'hybrid'
export type CurriculumGradeLevel = 'paud' | 'sd' | 'smp' | 'sma' | 'smk'
export type CurriculumPhase = 'a' | 'b' | 'c' | 'd'
export type CurriculumStatus = 'draft' | 'active' | 'archived'

export interface Curriculum extends BaseEntity {
  code: string
  name: string
  type: CurriculumType
  gradeLevel: CurriculumGradeLevel
  phase: CurriculumPhase
  description: string
  status: CurriculumStatus
}

export const CURRICULUM_TYPE_LABELS: Record<CurriculumType, string> = {
  national: 'Nasional',
  school: 'Sekolah',
  international: 'Internasional',
  hybrid: 'Hibrida',
}

export const CURRICULUM_GRADE_LEVEL_LABELS: Record<CurriculumGradeLevel, string> = {
  paud: 'PAUD',
  sd: 'SD',
  smp: 'SMP',
  sma: 'SMA',
  smk: 'SMK',
}

export const CURRICULUM_PHASE_LABELS: Record<CurriculumPhase, string> = {
  a: 'Fase A',
  b: 'Fase B',
  c: 'Fase C',
  d: 'Fase D',
}

export const CURRICULUM_STATUS_LABELS: Record<CurriculumStatus, string> = {
  draft: 'Draft',
  active: 'Aktif',
  archived: 'Diarsipkan',
}

// ── Subject ────────────────────────────────────────────────────────────────────

export type SubjectGroup = 'science' | 'social' | 'language' | 'mathematics' | 'religion' | 'arts' | 'physical' | 'technology' | 'other'
export type SubjectStatus = 'draft' | 'active' | 'inactive' | 'archived'

export interface Subject extends BaseEntity {
  code: string
  name: string
  group: SubjectGroup
  weightKnowledge: number
  weightSkill: number
  description: string
  status: SubjectStatus
}

export const SUBJECT_GROUP_LABELS: Record<SubjectGroup, string> = {
  science: 'IPA',
  social: 'IPS',
  language: 'Bahasa',
  mathematics: 'Matematika',
  religion: 'Pendidikan Agama',
  arts: 'Seni',
  physical: 'PJOK',
  technology: 'Teknologi',
  other: 'Lainnya',
}

export const SUBJECT_STATUS_LABELS: Record<SubjectStatus, string> = {
  draft: 'Draft',
  active: 'Aktif',
  inactive: 'Nonaktif',
  archived: 'Diarsipkan',
}

// ── AcademicCalendar ───────────────────────────────────────────────────────────

export type AcademicCalendarEventType = 'semester_start' | 'semester_end' | 'holiday' | 'exam' | 'event' | 'break' | 'other'
export type AcademicCalendarType = 'national' | 'school' | 'academic'

export interface AcademicCalendar extends BaseEntity {
  title: string
  eventType: AcademicCalendarEventType
  startDate: string
  endDate: string
  calendarType: AcademicCalendarType
  description: string
}

export const ACADEMIC_CALENDAR_EVENT_TYPE_LABELS: Record<AcademicCalendarEventType, string> = {
  semester_start: 'Mulai Semester',
  semester_end: 'Akhir Semester',
  holiday: 'Libur',
  exam: 'Ujian',
  event: 'Acara',
  break: 'Jeda',
  other: 'Lainnya',
}

export const ACADEMIC_CALENDAR_TYPE_LABELS: Record<AcademicCalendarType, string> = {
  national: 'Nasional',
  school: 'Sekolah',
  academic: 'Akademik',
}

// ── TeachingSchedule ───────────────────────────────────────────────────────────

export type TeachingScheduleDay = 'monday' | 'tuesday' | 'wednesday' | 'thursday' | 'friday' | 'saturday'

export interface TeachingSchedule extends BaseEntity {
  dayOfWeek: TeachingScheduleDay
  slot: string
  subjectId: string
  subjectName: string
  teacherId: string
  teacherName: string
  classRoomId: string
  classRoomName: string
  semester: string
  startTime: string
  endTime: string
}

export const TEACHING_SCHEDULE_DAY_LABELS: Record<TeachingScheduleDay, string> = {
  monday: 'Senin',
  tuesday: 'Selasa',
  wednesday: 'Rabu',
  thursday: 'Kamis',
  friday: 'Jumat',
  saturday: 'Sabtu',
}

// ── LessonPlan ─────────────────────────────────────────────────────────────────

export type LessonPlanType = 'daily' | 'weekly' | 'unit' | 'semester'
export type LessonPlanStatus = 'draft' | 'submitted' | 'approved' | 'revision' | 'archived'

export interface LessonPlan extends BaseEntity {
  title: string
  planType: LessonPlanType
  subjectId: string
  subjectName: string
  teacherId: string
  teacherName: string
  semester: string
  description: string
  status: LessonPlanStatus
}

export const LESSON_PLAN_TYPE_LABELS: Record<LessonPlanType, string> = {
  daily: 'Harian',
  weekly: 'Mingguan',
  unit: 'Unit',
  semester: 'Semester',
}

export const LESSON_PLAN_STATUS_LABELS: Record<LessonPlanStatus, string> = {
  draft: 'Draft',
  submitted: 'Dikirim',
  approved: 'Disetujui',
  revision: 'Revisi',
  archived: 'Diarsipkan',
}

// ── TeachingJournal ────────────────────────────────────────────────────────────

export type TeachingJournalStatus = 'planned' | 'completed' | 'cancelled' | 'rescheduled'

export interface TeachingJournal extends BaseEntity {
  date: string
  subjectId: string
  subjectName: string
  teacherId: string
  teacherName: string
  classRoomId: string
  classRoomName: string
  semester: string
  topic: string
  status: TeachingJournalStatus
}

export const TEACHING_JOURNAL_STATUS_LABELS: Record<TeachingJournalStatus, string> = {
  planned: 'Direncanakan',
  completed: 'Selesai',
  cancelled: 'Dibatalkan',
  rescheduled: 'Dijadwalkan Ulang',
}

// ── Denda ──────────────────────────────────────────────────────────────────────

export type DendaPenaltyType = 'late_payment' | 'early_withdrawal' | 'overdue' | 'administrative' | 'other'
export type DendaStatus = 'pending' | 'paid' | 'waived' | 'cancelled'

export interface Denda extends BaseEntity {
  nasabahId: string
  nasabahNama: string
  pinjamanId: string
  penaltyType: DendaPenaltyType
  amount: number
  period: string
  description: string
  status: DendaStatus
}

export const DENDA_PENALTY_TYPE_LABELS: Record<DendaPenaltyType, string> = {
  late_payment: 'Keterlambatan',
  early_withdrawal: 'Pencairan Awal',
  overdue: 'Tunggakan',
  administrative: 'Administratif',
  other: 'Lainnya',
}

export const DENDA_STATUS_LABELS: Record<DendaStatus, string> = {
  pending: 'Pending',
  paid: 'Lunas',
  waived: 'Dibebaskan',
  cancelled: 'Dibatalkan',
}

// ── Jaminan ────────────────────────────────────────────────────────────────────

export type JaminanCollateralType = 'property' | 'vehicle' | 'gold' | 'land' | 'savings' | 'guarantor' | 'other'
export type JaminanStatus = 'verified' | 'pending' | 'rejected' | 'released' | 'foreclosed'

export interface Jaminan extends BaseEntity {
  nasabahId: string
  nasabahNama: string
  collateralType: JaminanCollateralType
  description: string
  value: number
  documentNumber: string
  status: JaminanStatus
}

export const JAMINAN_COLLATERAL_TYPE_LABELS: Record<JaminanCollateralType, string> = {
  property: 'Properti',
  vehicle: 'Kendaraan',
  gold: 'Emas',
  land: 'Tanah',
  savings: 'Simpanan',
  guarantor: 'Penjamin',
  other: 'Lainnya',
}

export const JAMINAN_STATUS_LABELS: Record<JaminanStatus, string> = {
  verified: 'Terverifikasi',
  pending: 'Pending',
  rejected: 'Ditolak',
  released: 'Dilepas',
  foreclosed: 'Dieksekusi',
}
