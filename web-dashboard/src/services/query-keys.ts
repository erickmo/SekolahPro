/**
 * QK — Centralized Query Key constants.
 *
 * Convention:
 *   - List  : QK.resourceSlug            → ['resource-slug']
 *   - Detail: QK.resourceSlugDetail      → ['resource-slug-detail']
 *
 * Usage:
 *   queryKey: [QK.users]
 *   queryKey: [QK.userDetail, id]
 *   invalidateQueries({ queryKey: [QK.users] })
 *
 * Add your own keys here as you build features.
 */
export const QK = {
  // ── Dashboard ────────────────────────────────────────────────────────────
  dashboardSummary: 'dashboard-summary',

  // ── Auth / Session ───────────────────────────────────────────────────────
  profile: 'profile',

  // ── Multi-tenant ─────────────────────────────────────────────────────────
  companyGroups: 'company-groups',
  companyGroupDetail: 'company-group-detail',

  // ── Chat ─────────────────────────────────────────────────────────────────
  chatChannels: 'chat-channels',
  chatMessages: 'chat-messages',
  chatMembers: 'chat-members',

  // ── Media ────────────────────────────────────────────────────────────────
  mediaFiles: 'media-files',

  // ── Phase 2: Master Data Sekolah ─────────────────────────────────────────
  academicYears: 'academic-years',
  academicYearDetail: 'academic-year-detail',
  teachers: 'teachers',
  teacherDetail: 'teacher-detail',
  classRooms: 'class-rooms',
  classRoomDetail: 'class-room-detail',

  // ── Users & Roles (ADR-013) ───────────────────────────────────────────────
  users: 'users',
  userDetail: 'user-detail',

  // ── Koperasi ─────────────────────────────────────────────────────────────
  produkAkad: 'produk-akad',
  produkAkadDetail: 'produk-akad-detail',
  nasabah: 'nasabah',
  nasabahDetail: 'nasabah-detail',
  rekening: 'rekening',
  rekeningDetail: 'rekening-detail',

  // ── Phase 4A: Student Core ──────────────────────────────────────────────────
  students: 'students',
  studentDetail: 'student-detail',
  studentGuardians: 'student-guardians',
  studentDocuments: 'student-documents',
  studentClassPlacements: 'student-class-placements',
  studentAdmissions: 'student-admissions',

  // ── Phase 4B: Simpanan Koperasi ──────────────────────────────────────────────
  simpananPokokWajib: 'simpanan-pokok-wajib',
  simpananPokokWajibDetail: 'simpanan-pokok-wajib-detail',
  tabungan: 'tabungan',
  tabunganDetail: 'tabungan-detail',
  deposito: 'deposito',
  depositoDetail: 'deposito-detail',

  // ── Phase 5: Academic & Koperasi ──────────────────────────────────────────────
  curricula: 'curricula',
  curriculumDetail: 'curriculum-detail',
  subjects: 'subjects',
  subjectDetail: 'subject-detail',
  academicCalendar: 'academic-calendar',
  academicCalendarDetail: 'academic-calendar-detail',
  teachingSchedule: 'teaching-schedule',
  teachingScheduleDetail: 'teaching-schedule-detail',
  lessonPlan: 'lesson-plan',
  lessonPlanDetail: 'lesson-plan-detail',
  teachingJournal: 'teaching-journal',
  teachingJournalDetail: 'teaching-journal-detail',
  denda: 'denda',
  dendaDetail: 'denda-detail',
  jaminan: 'jaminan',
  jaminanDetail: 'jaminan-detail',
} as const

export type QueryKeyValue = (typeof QK)[keyof typeof QK]
