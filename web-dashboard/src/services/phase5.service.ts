import { createVernonService } from './vernon.service'
import type {
  Curriculum,
  Subject,
  AcademicCalendar,
  TeachingSchedule,
  LessonPlan,
  TeachingJournal,
  Denda,
  Jaminan,
} from '@/types/phase5.types'

// ── Curriculum ─────────────────────────────────────────────────────────────────

function transformCurriculum(raw: Record<string, unknown>): Curriculum {
  return {
    id: raw.id as string,
    code: (raw.code ?? '') as string,
    name: (raw.name ?? '') as string,
    type: (raw.type ?? 'national') as Curriculum['type'],
    gradeLevel: (raw.grade_level ?? 'sd') as Curriculum['gradeLevel'],
    phase: (raw.phase ?? 'a') as Curriculum['phase'],
    description: (raw.description ?? '') as string,
    status: (raw.status ?? 'draft') as Curriculum['status'],
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const curriculaService = createVernonService<Curriculum, Record<string, unknown>>(
  '/curricula',
  transformCurriculum,
)

// ── Subject ────────────────────────────────────────────────────────────────────

function transformSubject(raw: Record<string, unknown>): Subject {
  return {
    id: raw.id as string,
    code: (raw.code ?? '') as string,
    name: (raw.name ?? '') as string,
    group: (raw.group ?? 'other') as Subject['group'],
    weightKnowledge: (raw.weight_knowledge ?? 0) as number,
    weightSkill: (raw.weight_skill ?? 0) as number,
    description: (raw.description ?? '') as string,
    status: (raw.status ?? 'draft') as Subject['status'],
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const subjectsService = createVernonService<Subject, Record<string, unknown>>(
  '/subjects',
  transformSubject,
)

// ── AcademicCalendar ───────────────────────────────────────────────────────────

function transformAcademicCalendar(raw: Record<string, unknown>): AcademicCalendar {
  return {
    id: raw.id as string,
    title: (raw.title ?? '') as string,
    eventType: (raw.event_type ?? 'other') as AcademicCalendar['eventType'],
    startDate: (raw.start_date ?? '') as string,
    endDate: (raw.end_date ?? '') as string,
    calendarType: (raw.calendar_type ?? 'school') as AcademicCalendar['calendarType'],
    description: (raw.description ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const academicCalendarService = createVernonService<AcademicCalendar, Record<string, unknown>>(
  '/academic_calendar_events',
  transformAcademicCalendar,
)

// ── TeachingSchedule ───────────────────────────────────────────────────────────

function transformTeachingSchedule(raw: Record<string, unknown>): TeachingSchedule {
  const data = (raw._data ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    dayOfWeek: (raw.day_of_week ?? 'monday') as TeachingSchedule['dayOfWeek'],
    slot: (raw.slot ?? '') as string,
    subjectId: (raw.subject_id ?? '') as string,
    subjectName: (data.subject_name ?? '') as string,
    teacherId: (raw.teacher_id ?? '') as string,
    teacherName: (data.teacher_name ?? '') as string,
    classRoomId: (raw.class_room_id ?? '') as string,
    classRoomName: (data.class_room_name ?? '') as string,
    semester: (raw.semester ?? '') as string,
    startTime: (raw.start_time ?? '') as string,
    endTime: (raw.end_time ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const teachingScheduleService = createVernonService<TeachingSchedule, Record<string, unknown>>(
  '/schedule_entries',
  transformTeachingSchedule,
)

// ── LessonPlan ─────────────────────────────────────────────────────────────────

function transformLessonPlan(raw: Record<string, unknown>): LessonPlan {
  const data = (raw._data ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    title: (raw.title ?? '') as string,
    planType: (raw.plan_type ?? 'daily') as LessonPlan['planType'],
    subjectId: (raw.subject_id ?? '') as string,
    subjectName: (data.subject_name ?? '') as string,
    teacherId: (raw.teacher_id ?? '') as string,
    teacherName: (data.teacher_name ?? '') as string,
    semester: (raw.semester ?? '') as string,
    description: (raw.description ?? '') as string,
    status: (raw.status ?? 'draft') as LessonPlan['status'],
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const lessonPlanService = createVernonService<LessonPlan, Record<string, unknown>>(
  '/lesson_plans',
  transformLessonPlan,
)

// ── TeachingJournal ────────────────────────────────────────────────────────────

function transformTeachingJournal(raw: Record<string, unknown>): TeachingJournal {
  const data = (raw._data ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    date: (raw.date ?? '') as string,
    subjectId: (raw.subject_id ?? '') as string,
    subjectName: (data.subject_name ?? '') as string,
    teacherId: (raw.teacher_id ?? '') as string,
    teacherName: (data.teacher_name ?? '') as string,
    classRoomId: (raw.class_room_id ?? '') as string,
    classRoomName: (data.class_room_name ?? '') as string,
    semester: (raw.semester ?? '') as string,
    topic: (raw.topic ?? '') as string,
    status: (raw.status ?? 'planned') as TeachingJournal['status'],
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const teachingJournalService = createVernonService<TeachingJournal, Record<string, unknown>>(
  '/teaching_journals',
  transformTeachingJournal,
)

// ── Denda ──────────────────────────────────────────────────────────────────────

function transformDenda(raw: Record<string, unknown>): Denda {
  const data = (raw._data ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    nasabahId: (raw.nasabah_id ?? '') as string,
    nasabahNama: (data.nasabah_nama ?? '') as string,
    pinjamanId: (raw.pinjaman_id ?? '') as string,
    penaltyType: (raw.penalty_type ?? 'other') as Denda['penaltyType'],
    amount: (raw.amount ?? 0) as number,
    period: (raw.period ?? '') as string,
    description: (raw.description ?? '') as string,
    status: (raw.status ?? 'pending') as Denda['status'],
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const dendaService = createVernonService<Denda, Record<string, unknown>>(
  '/denda',
  transformDenda,
)

// ── Jaminan ────────────────────────────────────────────────────────────────────

function transformJaminan(raw: Record<string, unknown>): Jaminan {
  const data = (raw._data ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    nasabahId: (raw.nasabah_id ?? '') as string,
    nasabahNama: (data.nasabah_nama ?? '') as string,
    collateralType: (raw.collateral_type ?? 'other') as Jaminan['collateralType'],
    description: (raw.description ?? '') as string,
    value: (raw.value ?? 0) as number,
    documentNumber: (raw.document_number ?? '') as string,
    status: (raw.status ?? 'pending') as Jaminan['status'],
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const jaminanService = createVernonService<Jaminan, Record<string, unknown>>(
  '/jaminan',
  transformJaminan,
)
