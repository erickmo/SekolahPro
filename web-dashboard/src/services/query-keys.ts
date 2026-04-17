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

  // ── Phase 5B: Koperasi Operations ──────────────────────────────────────────────
  transaksi: 'transaksi',
  transaksiDetail: 'transaksi-detail',
  tellerSession: 'teller-session',
  tellerSessionDetail: 'teller-session-detail',
  moneyDenomination: 'money-denomination',
  moneyDenominationDetail: 'money-denomination-detail',
  kas: 'kas',
  kasDetail: 'kas-detail',

  // ── Sprint 5: Student Academic & Activity ──────────────────────────────────────
  dailyAttendance: 'daily-attendance',
  dailyAttendanceDetail: 'daily-attendance-detail',
  academicRecord: 'academic-record',
  academicRecordDetail: 'academic-record-detail',
  subjectGrade: 'subject-grade',
  subjectGradeDetail: 'subject-grade-detail',
  examAssessment: 'exam-assessment',
  examAssessmentDetail: 'exam-assessment-detail',
  healthRecord: 'health-record',
  healthRecordDetail: 'health-record-detail',
  discipline: 'discipline',
  disciplineDetail: 'discipline-detail',
  achievement: 'achievement',
  achievementDetail: 'achievement-detail',
  extracurricular: 'extracurricular',
  extracurricularDetail: 'extracurricular-detail',

  // ── Sprint 6: Koperasi Accounting & SHU ───────────────────────────────────────
  coa: 'coa',
  coaDetail: 'coa-detail',
  jurnal: 'jurnal',
  jurnalDetail: 'jurnal-detail',
  accountingPeriod: 'accounting-period',
  accountingPeriodDetail: 'accounting-period-detail',
  journalMapping: 'journal-mapping',
  journalMappingDetail: 'journal-mapping-detail',
  shuPeriode: 'shu-periode',
  shuPeriodeDetail: 'shu-periode-detail',
  shuAnggota: 'shu-anggota',
  shuAnggotaDetail: 'shu-anggota-detail',

  // ── Sprint 6: Student Finance SPP ──────────────────────────────────────────
  feeTypes: 'fee-types',
  feeTypeDetail: 'fee-type-detail',
  studentInvoices: 'student-invoices',
  studentInvoiceDetail: 'student-invoice-detail',
  studentPayments: 'student-payments',
  studentPaymentDetail: 'student-payment-detail',

  // ── Sprint 6: Rapor Generation ─────────────────────────────────────────────
  raporTemplates: 'rapor-templates',
  raporTemplateDetail: 'rapor-template-detail',
  raporRecords: 'rapor-records',
  raporRecordDetail: 'rapor-record-detail',

  // ── Sprint 6: Leave Management (ADR-S030) ──────────────────────────────────────
  leaveType: 'leave-type',
  leaveTypeDetail: 'leave-type-detail',
  leaveBalance: 'leave-balance',
  leaveBalanceDetail: 'leave-balance-detail',
  leaveRequest: 'leave-request',
  leaveRequestDetail: 'leave-request-detail',
  leaveApprovalLog: 'leave-approval-log',
  leaveApprovalLogDetail: 'leave-approval-log-detail',

  // ── Sprint 6: Teacher Substitution (ADR-S032) ─────────────────────────────────
  dutySchedule: 'duty-schedule',
  dutyScheduleDetail: 'duty-schedule-detail',
  teacherSubstitution: 'teacher-substitution',
  teacherSubstitutionDetail: 'teacher-substitution-detail',
  substitutionLog: 'substitution-log',
  substitutionLogDetail: 'substitution-log-detail',

  // ── Sprint 6: Student Counseling (ADR-S017) ─────────────────────────────────────
  counselingCase: 'counseling-case',
  counselingCaseDetail: 'counseling-case-detail',
  counselingSession: 'counseling-session',
  counselingSessionDetail: 'counseling-session-detail',

  // ── Sprint 6: Teacher Attendance (ADR-S026) ──────────────────────────────────────
  teacherAttendance: 'teacher-attendance',
  teacherAttendanceDetail: 'teacher-attendance-detail',

  // ── Sprint 6: Teacher Workload (ADR-S027) ────────────────────────────────────────
  teacherWorkload: 'teacher-workload',
  teacherWorkloadDetail: 'teacher-workload-detail',
  teacherWorkloadItem: 'teacher-workload-item',
  teacherWorkloadItemDetail: 'teacher-workload-item-detail',

  // ── Sprint 7: Laporan Regulasi (ADR-K017) ────────────────────────────────────────
  laporanRegulasi: 'laporan-regulasi',
  laporanRegulasiDetail: 'laporan-regulasi-detail',
  laporanConfig: 'laporan-config',
  laporanConfigDetail: 'laporan-config-detail',
  laporanVersi: 'laporan-versi',
  laporanVersiDetail: 'laporan-versi-detail',

  // ── Sprint 7: Zakat & Infaq (ADR-K018) ───────────────────────────────────────────
  zakatCollection: 'zakat-collection',
  zakatCollectionDetail: 'zakat-collection-detail',
  zakatDistribution: 'zakat-distribution',
  zakatDistributionDetail: 'zakat-distribution-detail',
  infaq: 'infaq',
  infaqDetail: 'infaq-detail',
  mustahik: 'mustahik',
  mustahikDetail: 'mustahik-detail',
  tazirFund: 'tazir-fund',
  tazirFundDetail: 'tazir-fund-detail',
} as const

export type QueryKeyValue = (typeof QK)[keyof typeof QK]
