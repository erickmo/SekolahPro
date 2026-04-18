import type { BaseEntity } from './entity.types'

// ═══════════════════════════════════════════════════════════════════════════════
// ADR-S038: Library Management (Perpustakaan)
// ═══════════════════════════════════════════════════════════════════════════════

export type BookCategory =
  | 'fiction' | 'nonfiction' | 'textbook' | 'reference'
  | 'kitab_kuning' | 'al_quran' | 'hadits' | 'fiqh' | 'aqidah'
  | 'science' | 'literature' | 'other'

export type BookLanguage = 'id' | 'en' | 'ar' | 'jv' | 'other'

export interface LibraryBook extends BaseEntity {
  title: string
  author: string
  isbn: string
  publisher: string
  publicationYear: number | null
  category: BookCategory
  language: BookLanguage
  ddcClassification: string
  description: string
  coverUrl: string
}

export const BOOK_CATEGORY_LABELS: Record<BookCategory, string> = {
  fiction: 'Fiksi', nonfiction: 'Non-Fiksi', textbook: 'Buku Teks',
  reference: 'Referensi', kitab_kuning: 'Kitab Kuning', al_quran: 'Al-Quran',
  hadits: 'Hadits', fiqh: 'Fiqh', aqidah: 'Aqidah',
  science: 'Sains', literature: 'Sastra', other: 'Lainnya',
}

export const BOOK_LANGUAGE_LABELS: Record<BookLanguage, string> = {
  id: 'Indonesia', en: 'Inggris', ar: 'Arab', jv: 'Jawa', other: 'Lainnya',
}

export type CopyCondition = 'new' | 'good' | 'fair' | 'damaged' | 'lost'

export interface LibraryCopy extends BaseEntity {
  bookId: string
  bookTitle: string
  bookIsbn: string
  copyNumber: string
  barcode: string
  condition: CopyCondition
  locationShelf: string
  acquisitionDate: string
  acquisitionSource: string
}

export const COPY_CONDITION_LABELS: Record<CopyCondition, string> = {
  new: 'Baru', good: 'Baik', fair: 'Cukup', damaged: 'Rusak', lost: 'Hilang',
}

export type BorrowerType = 'student' | 'teacher'
export type FineStatus = 'none' | 'pending' | 'paid' | 'waived'

export interface LibraryBorrow extends BaseEntity {
  copyId: string
  copyBarcode: string
  copyCondition: string
  bookId: string
  bookTitle: string
  bookIsbn: string
  borrowerType: BorrowerType
  borrowerId: string
  academicYearId: string
  academicYearName: string
  borrowDate: string
  dueDate: string
  returnDate: string
  actualReturnDate: string
  fineAmount: number
  fineStatus: FineStatus
  note: string
}

export const FINE_STATUS_LABELS: Record<FineStatus, string> = {
  none: 'Tidak Ada', pending: 'Belum Dibayar', paid: 'Dibayar', waived: 'Dibebaskan',
}

// ═══════════════════════════════════════════════════════════════════════════════
// ADR-S039: Laboratory Management
// ═══════════════════════════════════════════════════════════════════════════════

export type LabType = 'ipa' | 'komputer' | 'bahasa' | 'multimedia'
export type SafetyRating = 'excellent' | 'good' | 'fair' | 'poor'
export type LabStatus = 'active' | 'inactive' | 'maintenance'

export interface Laboratory extends BaseEntity {
  name: string
  type: LabType
  capacity: number | null
  building: string
  floor: number | null
  roomNumber: string
  equipmentCount: number
  safetyRating: SafetyRating | null
  lastInspectionDate: string
  status: LabStatus
}

export const LAB_TYPE_LABELS: Record<LabType, string> = {
  ipa: 'IPA', komputer: 'Komputer', bahasa: 'Bahasa', multimedia: 'Multimedia',
}
export const SAFETY_RATING_LABELS: Record<SafetyRating, string> = {
  excellent: 'Sangat Baik', good: 'Baik', fair: 'Cukup', poor: 'Kurang',
}
export const LAB_STATUS_LABELS: Record<LabStatus, string> = {
  active: 'Aktif', inactive: 'Nonaktif', maintenance: 'Perbaikan',
}

export type EquipmentCategory =
  | 'optical' | 'electronic' | 'measuring' | 'chemical' | 'biological'
  | 'specimen' | 'tool' | 'safety' | 'furniture' | 'computer' | 'other'
export type EquipmentCondition = 'new' | 'good' | 'needs_repair' | 'damaged'

export interface LabEquipment extends BaseEntity {
  labId: string
  labName: string
  labType: string
  labRoomNumber: string
  name: string
  category: EquipmentCategory
  brand: string
  model: string
  serialNumber: string
  condition: EquipmentCondition
  quantity: number
  unit: string
  purchaseDate: string
  purchasePrice: number | null
  calibrationDate: string
  nextCalibration: string
}

export const EQUIPMENT_CATEGORY_LABELS: Record<EquipmentCategory, string> = {
  optical: 'Optik', electronic: 'Elektronik', measuring: 'Pengukuran',
  chemical: 'Kimia', biological: 'Biologi', specimen: 'Spesimen',
  tool: 'Alat', safety: 'Keselamatan', furniture: 'Furnitur',
  computer: 'Komputer', other: 'Lainnya',
}
export const EQUIPMENT_CONDITION_LABELS: Record<EquipmentCondition, string> = {
  new: 'Baru', good: 'Baik', needs_repair: 'Perlu Perbaikan', damaged: 'Rusak',
}

export interface LabUsageLog extends BaseEntity {
  labId: string
  labName: string
  labType: string
  teacherId: string
  teacherName: string
  teacherNip: string
  classRoomId: string
  classRoomName: string
  academicYearId: string
  academicYearName: string
  subjectId: string
  usageDate: string
  startTime: string
  endTime: string
  topic: string
  participantCount: number | null
  notes: string
}

// ═══════════════════════════════════════════════════════════════════════════════
// ADR-S040: Asset & Inventory Management
// ═══════════════════════════════════════════════════════════════════════════════

export type AssetCategory = 'tanah' | 'bangunan' | 'mesin' | 'kendaraan' | 'peralatan' | 'lainnya'
export type OwnershipType = 'owned' | 'leased' | 'donated'
export type AssetCondition = 'new' | 'good' | 'fair' | 'poor' | 'damaged'
export type LifecycleStatus = 'in_use' | 'idle' | 'maintenance' | 'disposed' | 'lost' | 'transferred'

export interface Asset extends BaseEntity {
  assetCode: string
  name: string
  category: AssetCategory
  description: string
  acquisitionDate: string
  acquisitionCost: number | null
  ownershipType: OwnershipType
  condition: AssetCondition
  locationRoomId: string
  locationRoomName: string
  locationBuilding: string
  responsiblePerson: string
  vendor: string
  warrantyExpiry: string
  depreciationMethod: string
  usefulLifeYears: number | null
  bookValue: number | null
  lifecycleStatus: LifecycleStatus
  disposalDate: string
}

export const ASSET_CATEGORY_LABELS: Record<AssetCategory, string> = {
  tanah: 'Tanah', bangunan: 'Bangunan', mesin: 'Mesin',
  kendaraan: 'Kendaraan', peralatan: 'Peralatan', lainnya: 'Lainnya',
}
export const OWNERSHIP_TYPE_LABELS: Record<OwnershipType, string> = {
  owned: 'Milik Sendiri', leased: 'Sewa', donated: 'Hibah',
}
export const ASSET_CONDITION_LABELS: Record<AssetCondition, string> = {
  new: 'Baru', good: 'Baik', fair: 'Cukup', poor: 'Kurang', damaged: 'Rusak',
}
export const LIFECYCLE_STATUS_LABELS: Record<LifecycleStatus, string> = {
  in_use: 'Digunakan', idle: 'Menganggur', maintenance: 'Perbaikan',
  disposed: 'Disposal', lost: 'Hilang', transferred: 'Dipindahkan',
}

export type MaintenanceType = 'preventive' | 'corrective' | 'emergency' | 'calibration' | 'inspection'
export type MaintenanceStatus = 'scheduled' | 'in_progress' | 'completed' | 'cancelled'

export interface AssetMaintenance extends BaseEntity {
  assetId: string
  assetCode: string
  assetName: string
  assetCondition: string
  type: MaintenanceType
  description: string
  startDate: string
  completionDate: string
  cost: number | null
  vendor: string
  technician: string
  status: MaintenanceStatus
  notes: string
}

export const MAINTENANCE_TYPE_LABELS: Record<MaintenanceType, string> = {
  preventive: 'Preventif', corrective: 'Korektif', emergency: 'Darurat',
  calibration: 'Kalibrasi', inspection: 'Inspeksi',
}
export const MAINTENANCE_STATUS_LABELS: Record<MaintenanceStatus, string> = {
  scheduled: 'Terjadwal', in_progress: 'Berlangsung', completed: 'Selesai', cancelled: 'Dibatalkan',
}

// ═══════════════════════════════════════════════════════════════════════════════
// ADR-S041: Room & Facility Booking
// ═══════════════════════════════════════════════════════════════════════════════

export type FacilityType = 'classroom' | 'lab' | 'library' | 'hall' | 'meeting_room' | 'sports_field' | 'mosque' | 'other'
export type FacilityStatus = 'active' | 'inactive' | 'maintenance'

export interface Facility extends BaseEntity {
  name: string
  type: FacilityType
  building: string
  floor: number | null
  roomNumber: string
  capacity: number | null
  isBookable: boolean
  requiresApproval: boolean
  bookingAdvanceMinDays: number
  bookingAdvanceMaxDays: number
  status: FacilityStatus
}

export const FACILITY_TYPE_LABELS: Record<FacilityType, string> = {
  classroom: 'Ruang Kelas', lab: 'Laboratorium', library: 'Perpustakaan',
  hall: 'Aula', meeting_room: 'Ruang Rapat', sports_field: 'Lapangan',
  mosque: 'Masjid', other: 'Lainnya',
}
export const FACILITY_STATUS_LABELS: Record<FacilityStatus, string> = {
  active: 'Aktif', inactive: 'Nonaktif', maintenance: 'Perbaikan',
}

export type RequesterType = 'teacher' | 'student' | 'staff' | 'external' | 'admin'
export type BookingStatus = 'pending' | 'approved' | 'rejected' | 'cancelled' | 'completed'
export type RecurrencePattern = 'daily' | 'weekly' | 'monthly' | 'none'

export interface FacilityBooking extends BaseEntity {
  facilityId: string
  facilityName: string
  facilityType: string
  facilityBuilding: string
  facilityRoomNumber: string
  requesterType: RequesterType
  requesterId: string
  academicYearId: string
  academicYearName: string
  bookingDate: string
  startTime: string
  endTime: string
  purpose: string
  participantCount: number | null
  status: BookingStatus
  approvedBy: string
  approvedAt: string
  rejectionReason: string
  recurrencePattern: RecurrencePattern | null
  recurrenceEndDate: string
  notes: string
}

export const REQUESTER_TYPE_LABELS: Record<RequesterType, string> = {
  teacher: 'Guru', student: 'Siswa', staff: 'Staff', external: 'Eksternal', admin: 'Admin',
}
export const BOOKING_STATUS_LABELS: Record<BookingStatus, string> = {
  pending: 'Menunggu', approved: 'Disetujui', rejected: 'Ditolak',
  cancelled: 'Dibatalkan', completed: 'Selesai',
}
export const RECURRENCE_LABELS: Record<RecurrencePattern, string> = {
  daily: 'Harian', weekly: 'Mingguan', monthly: 'Bulanan', none: 'Tidak Ada',
}

// ═══════════════════════════════════════════════════════════════════════════════
// ADR-S028: Teacher Performance Evaluation (PKG)
// ═══════════════════════════════════════════════════════════════════════════════

export type CompetencyArea = 'pedagogic' | 'personality' | 'social' | 'professional'

export interface EvaluationCompetency extends BaseEntity {
  name: string
  area: CompetencyArea
  indicatorCount: number
  weight: number
  description: string
  isActive: boolean
}

export const COMPETENCY_AREA_LABELS: Record<CompetencyArea, string> = {
  pedagogic: 'Pedagogik', personality: 'Kepribadian', social: 'Sosial', professional: 'Profesional',
}

export type Semester = 'ganjil' | 'genap'
export type EvaluationGrade = 'A' | 'B' | 'C' | 'D' | 'E'
export type WorkflowStatus = 'draft' | 'self_assessment' | 'peer_review' | 'supervisor_review' | 'final' | 'approved'

export interface TeacherEvaluation extends BaseEntity {
  teacherId: string
  teacherName: string
  teacherNip: string
  teacherEmployeeType: string
  academicYearId: string
  academicYearName: string
  semester: Semester
  evaluatorId: string
  totalScore: number | null
  grade: EvaluationGrade | null
  workflowStatus: WorkflowStatus
  evaluationDate: string
  notes: string
}

export const SEMESTER_LABELS: Record<Semester, string> = { ganjil: 'Ganjil', genap: 'Genap' }
export const EVAL_GRADE_LABELS: Record<EvaluationGrade, string> = {
  A: 'A (Amat Baik)', B: 'B (Baik)', C: 'C (Cukup)', D: 'D (Kurang)', E: 'E (Sangat Kurang)',
}
export const WORKFLOW_STATUS_LABELS: Record<WorkflowStatus, string> = {
  draft: 'Draft', self_assessment: 'Self Assessment', peer_review: 'Peer Review',
  supervisor_review: 'Review Atasan', final: 'Final', approved: 'Disetujui',
}

export type AssessorType = 'self' | 'peer' | 'supervisor'

export interface EvaluationScore extends BaseEntity {
  evaluationId: string
  evaluationTeacherId: string
  evaluationSemester: string
  evaluationWorkflowStatus: string
  competencyId: string
  competencyName: string
  competencyArea: string
  competencyWeight: number
  assessorType: AssessorType
  assessorId: string
  score: number
  evidence: string
  notes: string
}

export const ASSESSOR_TYPE_LABELS: Record<AssessorType, string> = {
  self: 'Diri Sendiri', peer: 'Teman Sejawat', supervisor: 'Atasan',
}

// ═══════════════════════════════════════════════════════════════════════════════
// ADR-S029: Professional Development (PKB)
// ═══════════════════════════════════════════════════════════════════════════════

export type CertificationType = 'profesi' | 'penilaian' | 'pengawas' | 'teknisi'
export type CertificationStatus = 'active' | 'expired' | 'revoked' | 'pending_renewal'

export interface TeacherCertification extends BaseEntity {
  teacherId: string
  teacherName: string
  teacherNip: string
  certificationType: CertificationType
  certificationNumber: string
  issueDate: string
  expiryDate: string
  issuingBody: string
  status: CertificationStatus
  notes: string
}

export const CERT_TYPE_LABELS: Record<CertificationType, string> = {
  profesi: 'Profesi', penilaian: 'Penilaian', pengawas: 'Pengawas', teknisi: 'Teknisi',
}
export const CERT_STATUS_LABELS: Record<CertificationStatus, string> = {
  active: 'Aktif', expired: 'Kedaluwarsa', revoked: 'Dicabut', pending_renewal: 'Perpanjangan',
}

export type ActivityStatus = 'registered' | 'attended' | 'completed' | 'cancelled'

export interface DevelopmentActivity extends BaseEntity {
  teacherId: string
  teacherName: string
  teacherNip: string
  activityType: string
  title: string
  organizer: string
  startDate: string
  endDate: string
  location: string
  creditPoints: number
  certificateNumber: string
  status: ActivityStatus
  description: string
}

export const ACTIVITY_STATUS_LABELS: Record<ActivityStatus, string> = {
  registered: 'Terdaftar', attended: 'Hadir', completed: 'Selesai', cancelled: 'Dibatalkan',
}

export type TeacherRank = 'I/a' | 'I/b' | 'II/a' | 'II/b' | 'III/a' | 'III/b' | 'IV/a' | 'IV/b' | 'IV/c'

export interface CreditSummary extends BaseEntity {
  teacherId: string
  teacherName: string
  teacherNip: string
  teacherEmployeeType: string
  currentRank: TeacherRank
  targetRank: string
  totalCredits: number
  requiredCredits: number
  creditGap: number | null
  lastPromotionDate: string
  nextEligibleDate: string
  skpScore: number | null
  notes: string
}

export const TEACHER_RANK_LABELS: Record<TeacherRank, string> = {
  'I/a': 'I/a', 'I/b': 'I/b', 'II/a': 'II/a', 'II/b': 'II/b',
  'III/a': 'III/a', 'III/b': 'III/b', 'IV/a': 'IV/a', 'IV/b': 'IV/b', 'IV/c': 'IV/c',
}

// ═══════════════════════════════════════════════════════════════════════════════
// ADR-S031: Staff Payroll (Penggajian)
// ═══════════════════════════════════════════════════════════════════════════════

export type EmployeeType = 'pns' | 'honorer' | 'yayasan'
export type PphStatus = 'non_pkp' | 'ptkp' | 'pkp'

export interface PayrollConfig extends BaseEntity {
  teacherId: string
  teacherName: string
  teacherNip: string
  teacherEmployeeType: string
  employeeType: EmployeeType
  baseSalary: number
  transportAllowance: number
  mealAllowance: number
  positionAllowance: number
  familyAllowance: number
  riceAllowance: number
  pphStatus: PphStatus
  bankName: string
  bankAccount: string
  isActive: boolean
}

export const EMPLOYEE_TYPE_LABELS: Record<EmployeeType, string> = {
  pns: 'PNS', honorer: 'Honorer', yayasan: 'Yayasan',
}
export const PPH_STATUS_LABELS: Record<PphStatus, string> = {
  non_pkp: 'Non-PKP', ptkp: 'PTKP', pkp: 'PKP',
}

export type PayrollWorkflowStatus = 'draft' | 'calculating' | 'calculated' | 'approved' | 'processing' | 'paid' | 'cancelled' | 'failed'

export interface PayrollPeriod extends BaseEntity {
  periodName: string
  academicYearId: string
  academicYearName: string
  month: number
  year: number
  startDate: string
  endDate: string
  workflowStatus: PayrollWorkflowStatus
  totalEmployees: number
  totalGross: number
  totalDeductions: number
  totalNet: number
  processedBy: string
  processedAt: string
}

export const PAYROLL_WORKFLOW_LABELS: Record<PayrollWorkflowStatus, string> = {
  draft: 'Draft', calculating: 'Menghitung', calculated: 'Terhitung',
  approved: 'Disetujui', processing: 'Diproses', paid: 'Dibayar',
  cancelled: 'Dibatalkan', failed: 'Gagal',
}
export const MONTH_LABELS: Record<number, string> = {
  1: 'Januari', 2: 'Februari', 3: 'Maret', 4: 'April', 5: 'Mei', 6: 'Juni',
  7: 'Juli', 8: 'Agustus', 9: 'September', 10: 'Oktober', 11: 'November', 12: 'Desember',
}

export interface PayrollEntry extends BaseEntity {
  periodId: string
  periodName: string
  periodMonth: number
  periodYear: number
  periodWorkflowStatus: string
  teacherId: string
  teacherName: string
  teacherNip: string
  employeeType: EmployeeType
  baseSalary: number
  totalAllowances: number
  grossSalary: number
  totalDeductions: number
  netSalary: number
  pph21Amount: number
  workingDays: number
  presentDays: number
  absentDays: number
  leaveDays: number
  payslipNumber: string
  notes: string
}

export type ComponentType = 'earning' | 'deduction'
export type CalculationMethod = 'fixed' | 'percentage' | 'formula' | 'attendance_based'

export interface PayrollComponent extends BaseEntity {
  entryId: string
  entryPayslipNumber: string
  entryTeacherId: string
  componentType: ComponentType
  category: string
  name: string
  amount: number
  calculationMethod: CalculationMethod
  isRecurring: boolean
  referenceId: string
  notes: string
}

export const COMPONENT_TYPE_LABELS: Record<ComponentType, string> = {
  earning: 'Pendapatan', deduction: 'Potongan',
}
export const CALC_METHOD_LABELS: Record<CalculationMethod, string> = {
  fixed: 'Tetap', percentage: 'Persentase', formula: 'Formula', attendance_based: 'Berdasarkan Kehadiran',
}
