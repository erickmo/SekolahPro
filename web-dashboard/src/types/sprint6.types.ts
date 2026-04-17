import type { BaseEntity } from './entity.types'

// ── FeeType (Master Jenis Tagihan) ──────────────────────────────────────────

export type FeeCategory = 'monthly' | 'annual' | 'one_time' | 'incidental'

export interface FeeType extends BaseEntity {
  name: string
  code: string
  feeCategory: FeeCategory
  defaultAmount: number
  description: string
  isActive: boolean
}

export const FEE_CATEGORY_LABELS: Record<FeeCategory, string> = {
  monthly: 'Bulanan',
  annual: 'Tahunan',
  one_time: 'Sekali Bayar',
  incidental: 'Insidental',
}

// ── Invoice (Tagihan Siswa) ─────────────────────────────────────────────────

export type InvoiceStatus = 'unpaid' | 'partial' | 'paid' | 'overdue' | 'waived'

export interface StudentInvoice extends BaseEntity {
  studentId: string
  studentName: string
  studentNis: string
  academicYearId: string
  academicYearName: string
  feeTypeId: string
  feeTypeName: string
  feeTypeCode: string
  feeCategory: FeeCategory
  invoiceNo: string
  periodMonth: number | null
  periodYear: number
  amount: number
  totalAmount: number
  paidAmount: number
  dueDate: string
  status: InvoiceStatus
}

export const INVOICE_STATUS_LABELS: Record<InvoiceStatus, string> = {
  unpaid: 'Belum Bayar',
  partial: 'Cicilan',
  paid: 'Lunas',
  overdue: 'Terlambat',
  waived: 'Dibebaskan',
}

// ── Payment (Pembayaran) ────────────────────────────────────────────────────

export type PaymentMethod = 'cash' | 'transfer' | 'debit' | 'qris' | 'va'

export interface StudentPayment extends BaseEntity {
  invoiceId: string
  invoiceNo: string
  invoiceTotalAmount: number
  studentId: string
  studentName: string
  studentNis: string
  receiptNo: string
  amount: number
  paymentMethod: PaymentMethod
  paymentDate: string
  receivedBy: string
  notes: string
}

export const PAYMENT_METHOD_LABELS: Record<PaymentMethod, string> = {
  cash: 'Tunai',
  transfer: 'Transfer',
  debit: 'Kartu Debit',
  qris: 'QRIS',
  va: 'Virtual Account',
}

// ── RaporTemplate ───────────────────────────────────────────────────────────

export type CurriculumType = 'merdeka' | 'k13' | 'diniyah' | 'custom'

export interface RaporTemplate extends BaseEntity {
  name: string
  curriculumType: CurriculumType
  gradeLevels: string[]
  description: string
  layoutConfig: Record<string, unknown>
  headerConfig: Record<string, unknown>
  isActive: boolean
}

export const CURRICULUM_TYPE_LABELS: Record<CurriculumType, string> = {
  merdeka: 'Merdeka',
  k13: 'Kurikulum 2013',
  diniyah: 'Diniyah',
  custom: 'Kustom',
}

// ── RaporRecord ─────────────────────────────────────────────────────────────

export type Semester = 'ganjil' | 'genap'
export type RecordStatus = 'draft' | 'generated' | 'reviewed' | 'finalized'

export interface RaporRecord extends BaseEntity {
  studentId: string
  studentName: string
  studentNis: string
  academicYearId: string
  academicYearName: string
  classRoomId: string
  classRoomName: string
  templateId: string
  templateName: string
  curriculumType: CurriculumType
  semester: Semester
  status: RecordStatus
  notes: string
  pdfUrl: string
  generatedAt: string
}

export const SEMESTER_LABELS: Record<Semester, string> = {
  ganjil: 'Ganjil',
  genap: 'Genap',
}

export const RECORD_STATUS_LABELS: Record<RecordStatus, string> = {
  draft: 'Draft',
  generated: 'Dihasilkan',
  reviewed: 'Direview',
  finalized: 'Difinalisasi',
}

// ── Leave Type (ADR-S030) ──────────────────────────────────────────────────────

export type LeaveTypeCode = 'cuti_tahunan' | 'cuti_sakit' | 'cuti_melahirkan' | 'cuti_besar' | 'izin' | 'tugas_belajar'
export type LeaveTypeApplicableTo = 'all' | 'pns' | 'honorer' | 'yayasan'

export interface LeaveType extends BaseEntity {
  code: LeaveTypeCode
  name: string
  description: string
  maxDaysPerYear: number | null
  isPaid: boolean
  applicableTo: LeaveTypeApplicableTo
  approvalLevels: number
  requiresDocument: boolean
  isActive: boolean
}

export const LEAVE_TYPE_CODE_LABELS: Record<LeaveTypeCode, string> = {
  cuti_tahunan: 'Cuti Tahunan',
  cuti_sakit: 'Cuti Sakit',
  cuti_melahirkan: 'Cuti Melahirkan',
  cuti_besar: 'Cuti Besar',
  izin: 'Izin',
  tugas_belajar: 'Tugas Belajar',
}

export const APPLICABLE_TO_LABELS: Record<LeaveTypeApplicableTo, string> = {
  all: 'Semua',
  pns: 'PNS',
  honorer: 'Honorer',
  yayasan: 'Yayasan',
}

// ── Leave Balance (ADR-S030) ───────────────────────────────────────────────────

export interface LeaveBalance extends BaseEntity {
  teacherId: string
  teacherName: string
  teacherNip: string
  teacherEmployeeType: string
  leaveTypeId: string
  leaveTypeCode: string
  leaveTypeName: string
  academicYearId: string
  academicYearName: string
  year: number
  initialBalance: number
  used: number
  remaining: number
  carryOver: number
}

// ── Leave Request (ADR-S030) ───────────────────────────────────────────────────

export type LeaveRequestStatus = 'draft' | 'submitted' | 'pending_approval' | 'approved' | 'rejected' | 'cancelled' | 'completed'

export interface LeaveRequest extends BaseEntity {
  teacherId: string
  teacherName: string
  teacherNip: string
  teacherEmployeeType: string
  leaveTypeId: string
  leaveTypeName: string
  leaveTypeCode: string
  leaveBalanceId: string
  academicYearId: string
  academicYearName: string
  startDate: string
  endDate: string
  totalDays: number
  reason: string
  status: LeaveRequestStatus
  documentUrl: string
  approvalNotes: string
  currentApprovalLevel: number
  notes: string
}

export const LEAVE_REQUEST_STATUS_LABELS: Record<LeaveRequestStatus, string> = {
  draft: 'Draft',
  submitted: 'Diajukan',
  pending_approval: 'Menunggu Persetujuan',
  approved: 'Disetujui',
  rejected: 'Ditolak',
  cancelled: 'Dibatalkan',
  completed: 'Selesai',
}

// ── Leave Approval Log (ADR-S030) ──────────────────────────────────────────────

export type ApprovalAction = 'approved' | 'rejected' | 'returned'
export type ApproverRole = 'wakil_kepsek' | 'kepala_sekolah' | 'yayasan' | 'dinas_pendidikan'

export interface LeaveApprovalLog extends BaseEntity {
  leaveRequestId: string
  approverId: string
  approverName: string
  approverRole: string
  approvalLevel: number
  action: ApprovalAction
  notes: string
}

export const APPROVAL_ACTION_LABELS: Record<ApprovalAction, string> = {
  approved: 'Disetujui',
  rejected: 'Ditolak',
  returned: 'Dikembalikan',
}

export const APPROVER_ROLE_LABELS: Record<ApproverRole, string> = {
  wakil_kepsek: 'Wakil Kepala Sekolah',
  kepala_sekolah: 'Kepala Sekolah',
  yayasan: 'Yayasan',
  dinas_pendidikan: 'Dinas Pendidikan',
}

// ── Duty Schedule (ADR-S032) ────────────────────────────────────────────────────

export type DutyType = 'piket_pagi' | 'piket_kelas' | 'piket_gerbang' | 'piket_upacara' | 'piket_siang' | 'piket_asrama' | 'piket_malam'

export interface DutySchedule extends BaseEntity {
  teacherId: string
  teacherName: string
  teacherNip: string
  teacherEmployeeType: string
  academicYearId: string
  academicYearName: string
  semester: Semester
  dayOfWeek: number
  dutyType: DutyType
  startTime: string
  endTime: string
  notes: string
}

export const DUTY_TYPE_LABELS: Record<DutyType, string> = {
  piket_pagi: 'Piket Pagi',
  piket_kelas: 'Piket Kelas',
  piket_gerbang: 'Piket Gerbang',
  piket_upacara: 'Piket Upacara',
  piket_siang: 'Piket Siang',
  piket_asrama: 'Piket Asrama',
  piket_malam: 'Piket Malam',
}

export const DAY_OF_WEEK_LABELS: Record<number, string> = {
  1: 'Senin',
  2: 'Selasa',
  3: 'Rabu',
  4: 'Kamis',
  5: 'Jumat',
  6: 'Sabtu',
  7: 'Minggu',
}

// ── Teacher Substitution (ADR-S032) ─────────────────────────────────────────────

export type SubstitutionReasonType = 'sakit' | 'cuti' | 'izin' | 'dinas_luar' | 'tugas_belajar' | 'terlambat' | 'lain_lain'
export type SubstitutionStatus = 'pending' | 'notified' | 'accepted' | 'declined' | 'in_progress' | 'completed' | 'cancelled'

export interface TeacherSubstitution extends BaseEntity {
  originalTeacherId: string
  originalTeacherName: string
  substituteTeacherId: string
  substituteTeacherName: string
  academicYearId: string
  academicYearName: string
  substitutionDate: string
  reasonType: SubstitutionReasonType
  classRoomId: string
  classRoomName: string
  status: SubstitutionStatus
  scheduleEntryId: string
  timeSlotId: string
  originalSubjectId: string
  notes: string
}

export const SUBSTITUTION_REASON_LABELS: Record<SubstitutionReasonType, string> = {
  sakit: 'Sakit',
  cuti: 'Cuti',
  izin: 'Izin',
  dinas_luar: 'Dinas Luar',
  tugas_belajar: 'Tugas Belajar',
  terlambat: 'Terlambat',
  lain_lain: 'Lain-lain',
}

export const SUBSTITUTION_STATUS_LABELS: Record<SubstitutionStatus, string> = {
  pending: 'Menunggu',
  notified: 'Dinotifikasi',
  accepted: 'Diterima',
  declined: 'Ditolak',
  in_progress: 'Berlangsung',
  completed: 'Selesai',
  cancelled: 'Dibatalkan',
}

// ── Substitution Log (ADR-S032) ─────────────────────────────────────────────────

export type SubstitutionLogAction = 'created' | 'notified' | 'accepted' | 'declined' | 'reassigned' | 'started' | 'completed' | 'cancelled'

export interface SubstitutionLog extends BaseEntity {
  substitutionId: string
  actorId: string
  actorName: string
  action: SubstitutionLogAction
  notes: string
}

export const SUBSTITUTION_LOG_ACTION_LABELS: Record<SubstitutionLogAction, string> = {
  created: 'Dibuat',
  notified: 'Dinotifikasi',
  accepted: 'Diterima',
  declined: 'Ditolak',
  reassigned: 'Dialihkan',
  started: 'Dimulai',
  completed: 'Selesai',
  cancelled: 'Dibatalkan',
}

// ── Student Counseling (ADR-S017) ─────────────────────────────────────────────────

export type CaseCategory = 'academic' | 'social' | 'personal' | 'career' | 'behavioral' | 'family' | 'other'
export type CaseSeverity = 'low' | 'medium' | 'high' | 'critical'
export type CaseStatus = 'open' | 'in_progress' | 'referred' | 'resolved' | 'closed'
export type SessionType = 'individual' | 'group' | 'home_visit' | 'parent_conference' | 'referral'

export interface CounselingCase extends BaseEntity {
  studentId: string
  studentName: string
  studentNis: string
  academicYearId: string
  academicYearName: string
  caseNo: string
  category: CaseCategory
  title: string
  description: string
  severity: CaseSeverity
  status: CaseStatus
  openedDate: string
  closedDate: string
  resolution: string
  counselorId: string
  referredTo: string
}

export interface CounselingSession extends BaseEntity {
  caseId: string
  caseNo: string
  caseTitle: string
  studentId: string
  studentName: string
  studentNis: string
  sessionDate: string
  sessionType: SessionType
  durationMinutes: number
  notes: string
  recommendation: string
  counselorId: string
  parentPresent: boolean
  followUpDate: string
  followUpNote: string
}

export const CASE_CATEGORY_LABELS: Record<CaseCategory, string> = {
  academic: 'Akademik',
  social: 'Sosial',
  personal: 'Pribadi',
  career: 'Karir',
  behavioral: 'Perilaku',
  family: 'Keluarga',
  other: 'Lainnya',
}

export const CASE_SEVERITY_LABELS: Record<CaseSeverity, string> = {
  low: 'Rendah',
  medium: 'Sedang',
  high: 'Tinggi',
  critical: 'Kritis',
}

export const CASE_STATUS_LABELS: Record<CaseStatus, string> = {
  open: 'Terbuka',
  in_progress: 'Dalam Proses',
  referred: 'Dirujuk',
  resolved: 'Terselesaikan',
  closed: 'Ditutup',
}

export const SESSION_TYPE_LABELS: Record<SessionType, string> = {
  individual: 'Individual',
  group: 'Kelompok',
  home_visit: 'Kunjungan Rumah',
  parent_conference: 'Konferensi Orang Tua',
  referral: 'Rujukan',
}

// ── Teacher Attendance (ADR-S026) ──────────────────────────────────────────────────

export type TeacherAttendanceStatus = 'present' | 'sick' | 'permitted' | 'absent' | 'dinas_luar' | 'cuti' | 'libur'
export type ClockMethod = 'fingerprint' | 'face_recognition' | 'gps' | 'manual' | 'qr_code'

export interface TeacherAttendance extends BaseEntity {
  teacherId: string
  teacherName: string
  teacherNip: string
  academicYearId: string
  academicYearName: string
  attendanceDate: string
  status: TeacherAttendanceStatus
  clockIn: string
  clockOut: string
  lateMinutes: number
  earlyLeaveMinutes: number
  clockInMethod: string
  clockOutMethod: string
  note: string
}

export const TEACHER_ATT_STATUS_LABELS: Record<TeacherAttendanceStatus, string> = {
  present: 'Hadir',
  sick: 'Sakit',
  permitted: 'Izin',
  absent: 'Tidak Hadir',
  dinas_luar: 'Dinas Luar',
  cuti: 'Cuti',
  libur: 'Libur',
}

// ── Teacher Workload (ADR-S027) ────────────────────────────────────────────────────

export type FulfillmentStatus = 'kurang' | 'terpenuhi' | 'lebih'
export type WorkloadItemType = 'mengajar' | 'wali_kelas' | 'pembina_ekskul' | 'guru_bk' | 'kepala_sekolah' | 'wakil_kepsek' | 'koordinator' | 'panitia' | 'tugas_tambahan'

export interface TeacherWorkload extends BaseEntity {
  teacherId: string
  teacherName: string
  teacherNip: string
  academicYearId: string
  academicYearName: string
  semester: Semester
  teachingHours: number
  additionalHours: number
  totalHours: number
  minimumRequired: number
  isFulfilled: boolean
  fulfillmentStatus: FulfillmentStatus
}

export interface TeacherWorkloadItem extends BaseEntity {
  workloadId: string
  teacherId: string
  teacherName: string
  academicYearId: string
  itemType: WorkloadItemType
  description: string
  hoursPerWeek: number
  referenceType: string
  referenceId: string
}

export const FULFILLMENT_STATUS_LABELS: Record<FulfillmentStatus, string> = {
  kurang: 'Kurang',
  terpenuhi: 'Terpenuhi',
  lebih: 'Lebih',
}

export const WORKLOAD_ITEM_TYPE_LABELS: Record<WorkloadItemType, string> = {
  mengajar: 'Mengajar',
  wali_kelas: 'Wali Kelas',
  pembina_ekskul: 'Pembina Ekskul',
  guru_bk: 'Guru BK',
  kepala_sekolah: 'Kepala Sekolah',
  wakil_kepsek: 'Wakil Kepala Sekolah',
  koordinator: 'Koordinator',
  panitia: 'Panitia',
  tugas_tambahan: 'Tugas Tambahan',
}
