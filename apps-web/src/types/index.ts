// ─── Enums ────────────────────────────────────────────────────────────────────

export type Role =
  | 'SUPERADMIN'
  | 'ADMIN_YAYASAN'
  | 'ADMIN_SEKOLAH'
  | 'KEPALA_SEKOLAH'
  | 'KEPALA_KURIKULUM'
  | 'BENDAHARA'
  | 'GURU'
  | 'WALI_KELAS'
  | 'GURU_BK'
  | 'OPERATOR_SIMS'
  | 'KASIR_KOPERASI'
  | 'PUSTAKAWAN'
  | 'PETUGAS_UKS'
  | 'TATA_USAHA'
  | 'SISWA'
  | 'ORANG_TUA'
  | 'ALUMNI'
  | 'GURU_LES'
  | 'MITRA_DU_DI'
  | 'PUBLISHER'
  | 'DEVELOPER';

export type Gender = 'MALE' | 'FEMALE';
export type TenantPlan = 'BASIC' | 'PRO' | 'ENTERPRISE';
export type BloomLevel = 'C1' | 'C2' | 'C3' | 'C4' | 'C5' | 'C6';
export type Difficulty = 'EASY' | 'MEDIUM' | 'HARD';
export type QuestionType = 'MULTIPLE_CHOICE' | 'ESSAY' | 'TRUE_FALSE' | 'MATCHING';
export type ExamStatus = 'DRAFT' | 'PUBLISHED' | 'ONGOING' | 'COMPLETED';
export type AlertStatus = 'OPEN' | 'IN_PROGRESS' | 'RESOLVED' | 'DISMISSED';
export type NotificationType =
  | 'ATTENDANCE'
  | 'PAYMENT'
  | 'EXAM'
  | 'ANNOUNCEMENT'
  | 'ALERT'
  | 'HEALTH'
  | 'TRANSPORT';
export type NotificationChannel = 'WHATSAPP' | 'PUSH_NOTIFICATION' | 'EMAIL' | 'SMS';
export type NotificationStatus = 'PENDING' | 'SENT' | 'FAILED' | 'READ';
export type InvoiceStatus = 'PENDING' | 'PAID' | 'OVERDUE' | 'CANCELLED';
export type LoanStatus = 'BORROWED' | 'RETURNED' | 'OVERDUE';
export type AttendanceStatus = 'PRESENT' | 'ABSENT' | 'LATE' | 'SICK' | 'PERMITTED';

// ─── Shared ────────────────────────────────────────────────────────────────────

export interface PaginationMeta {
  total: number;
  page: number;
  limit: number;
  totalPages: number;
}

export interface ApiResponse<T = unknown> {
  success: boolean;
  data?: T;
  meta?: PaginationMeta;
  error?: {
    code: string;
    message: string;
  };
}

// ─── Auth ──────────────────────────────────────────────────────────────────────

export interface AuthUser {
  userId: string;
  email: string;
  name: string;
  role: Role;
  schoolId: string;
  schoolName: string;
  schoolSubdomain: string;
  photoUrl?: string;
}

export interface LoginCredentials {
  email: string;
  password: string;
  subdomain?: string;
}

export interface RegisterData {
  email: string;
  password: string;
  name: string;
  schoolId: string;
  role: Role;
}

export interface AuthTokens {
  accessToken: string;
  refreshToken: string;
}

export interface AuthResponse {
  user: AuthUser;
  accessToken: string;
  refreshToken: string;
}

// ─── School ────────────────────────────────────────────────────────────────────

export interface School {
  id: string;
  name: string;
  npsn: string;
  subdomain: string;
  address: string;
  logoUrl?: string;
  config: Record<string, unknown>;
  tenantPlan: TenantPlan;
  createdAt: string;
}

// ─── Student ───────────────────────────────────────────────────────────────────

export interface Guardian {
  id: string;
  name: string;
  relationship: string;
  phone: string;
  email?: string;
  address?: string;
}

export interface Student {
  id: string;
  nisn: string;
  nik?: string;
  name: string;
  photoUrl?: string;
  birthDate: string;
  birthPlace?: string;
  gender: Gender;
  address?: string;
  schoolId: string;
  riskScore?: number;
  guardians?: Guardian[];
  currentClass?: string;
  enrollmentStatus?: string;
  createdAt: string;
  updatedAt: string;
}

export interface StudentAttendance {
  id: string;
  studentId: string;
  date: string;
  status: AttendanceStatus;
  note?: string;
  recordedBy?: string;
}

// ─── Academic ─────────────────────────────────────────────────────────────────

export interface AcademicYear {
  id: string;
  schoolId: string;
  name: string;
  startDate: string;
  endDate: string;
  isActive: boolean;
  createdAt: string;
}

export interface Subject {
  id: string;
  schoolId: string;
  name: string;
  code: string;
  gradeLevel: string;
  curriculum: 'K13' | 'MERDEKA';
  weeklyHours: number;
  createdAt: string;
}

export interface SchoolClass {
  id: string;
  schoolId: string;
  name: string;
  gradeLevel: string;
  academicYearId: string;
  homeroomTeacherId?: string;
  capacity: number;
  studentCount?: number;
  createdAt: string;
}

export interface Grade {
  id: string;
  studentId: string;
  subjectId: string;
  classId: string;
  academicYearId: string;
  semester: number;
  dailyScore?: number;
  midtermScore?: number;
  finalScore?: number;
  finalGrade?: number;
  predicate?: string;
  createdAt: string;
}

export interface Teacher {
  id: string;
  schoolId: string;
  name: string;
  nip?: string;
  email: string;
  phone?: string;
  subjects?: string[];
  createdAt: string;
}

// ─── Payments ─────────────────────────────────────────────────────────────────

export interface Invoice {
  id: string;
  schoolId: string;
  studentId: string;
  studentName?: string;
  studentNisn?: string;
  className?: string;
  type: string;
  description: string;
  amount: number;
  dueDate: string;
  status: InvoiceStatus;
  paidAt?: string;
  paymentMethod?: string;
  createdAt: string;
}

export interface FinancialSummary {
  totalInvoiced: number;
  totalPaid: number;
  totalPending: number;
  totalOverdue: number;
  collectionRate: number;
  byMonth: Array<{ month: string; paid: number; pending: number }>;
}

// ─── Cooperative ──────────────────────────────────────────────────────────────

export interface SavingsAccount {
  id: string;
  studentId: string;
  schoolId: string;
  balance: number;
  totalDeposits: number;
  totalWithdrawals: number;
  createdAt: string;
  updatedAt: string;
}

export interface CoopTransaction {
  id: string;
  accountId: string;
  type: 'DEPOSIT' | 'WITHDRAWAL' | 'PURCHASE';
  amount: number;
  balanceAfter: number;
  description: string;
  cashierId?: string;
  createdAt: string;
}

export interface CoopProduct {
  id: string;
  schoolId: string;
  name: string;
  price: number;
  stock: number;
  category: string;
  imageUrl?: string;
  createdAt: string;
}

// ─── Library ──────────────────────────────────────────────────────────────────

export interface Book {
  id: string;
  schoolId: string;
  title: string;
  author: string;
  isbn?: string;
  publisher?: string;
  publishYear?: number;
  category: string;
  stock: number;
  available: number;
  coverUrl?: string;
  createdAt: string;
}

export interface BookLoan {
  id: string;
  schoolId: string;
  bookId: string;
  bookTitle?: string;
  studentId: string;
  studentName?: string;
  borrowedAt: string;
  dueDate: string;
  returnedAt?: string;
  status: LoanStatus;
  fine?: number;
  createdAt: string;
}

// ─── Health ───────────────────────────────────────────────────────────────────

export interface HealthRecord {
  id: string;
  schoolId: string;
  studentId: string;
  studentName?: string;
  bloodType?: string;
  allergies?: string[];
  conditions?: string[];
  height?: number;
  weight?: number;
  vision?: string;
  hearing?: string;
  notes?: string;
  updatedAt: string;
}

export interface UKSVisit {
  id: string;
  schoolId: string;
  studentId: string;
  studentName?: string;
  date: string;
  complaint: string;
  diagnosis?: string;
  treatment?: string;
  medicine?: string;
  referral?: boolean;
  staffId: string;
  staffName?: string;
  createdAt: string;
}

// ─── Counseling ───────────────────────────────────────────────────────────────

export interface CounselingSession {
  id: string;
  schoolId: string;
  studentId: string;
  studentName?: string;
  counselorId?: string;
  counselorName?: string;
  scheduledAt: string;
  status: 'SCHEDULED' | 'COMPLETED' | 'CANCELLED';
  type: 'ACADEMIC' | 'PERSONAL' | 'CAREER' | 'SOCIAL';
  notes?: string;
  followUp?: string;
  createdAt: string;
}

export interface BullyingReport {
  id: string;
  schoolId: string;
  reporterId?: string;
  isAnonymous: boolean;
  category: 'PHYSICAL' | 'VERBAL' | 'CYBERBULLYING' | 'SEXUAL_HARASSMENT' | 'OTHER';
  description: string;
  involvedStudents?: string[];
  status: 'OPEN' | 'INVESTIGATING' | 'RESOLVED' | 'CLOSED';
  assignedTo?: string;
  resolution?: string;
  createdAt: string;
}

// ─── Dashboard ────────────────────────────────────────────────────────────────

export interface DashboardStats {
  totalStudents: number;
  presentToday: number;
  absentToday: number;
  attendanceRate: number;
  totalTeachers: number;
  activeClasses: number;
  pendingPayments: number;
  totalRevenue: number;
  riskStudents: number;
  weeklyInsight?: string;
}

export interface AttendanceTrend {
  date: string;
  present: number;
  absent: number;
  late: number;
  rate: number;
}

// ─── EWS ──────────────────────────────────────────────────────────────────────

export interface RiskAlert {
  id: string;
  studentId: string;
  studentName?: string;
  riskScore: number;
  indicators: Record<string, unknown>;
  recommendation?: string;
  status: AlertStatus;
  assignedTo?: string;
  resolvedAt?: string;
  createdAt: string;
}
