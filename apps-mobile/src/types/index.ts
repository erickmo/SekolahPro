// ─── Auth & User ─────────────────────────────────────────────────────────────

export type UserRole =
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
  | 'ALUMNI';

export interface User {
  id: string;
  name: string;
  email: string;
  role: UserRole;
  schoolId: string;
  photoUrl?: string;
  createdAt: string;
  updatedAt: string;
}

export interface School {
  id: string;
  name: string;
  npsn: string;
  subdomain: string;
  address: string;
  logoUrl?: string;
  tenantPlan: 'BASIC' | 'PRO' | 'ENTERPRISE';
}

export interface AuthTokens {
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
}

export interface LoginCredentials {
  subdomain: string;
  email: string;
  password: string;
}

export interface AuthState {
  user: User | null;
  school: School | null;
  tokens: AuthTokens | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}

// ─── Student ──────────────────────────────────────────────────────────────────

export type Gender = 'MALE' | 'FEMALE';
export type StudentStatus = 'ACTIVE' | 'INACTIVE' | 'TRANSFERRED' | 'GRADUATED';

export interface Guardian {
  id: string;
  name: string;
  relationship: string;
  phone: string;
  email?: string;
  occupation?: string;
}

export interface Enrollment {
  id: string;
  academicYear: string;
  grade: string;
  className: string;
  semester: number;
  status: 'ACTIVE' | 'COMPLETED' | 'DROPPED';
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
  religion?: string;
  address?: string;
  phone?: string;
  email?: string;
  schoolId: string;
  status: StudentStatus;
  riskScore?: number;
  guardians: Guardian[];
  currentEnrollment?: Enrollment;
  createdAt: string;
  updatedAt: string;
}

export interface StudentListItem {
  id: string;
  nisn: string;
  name: string;
  photoUrl?: string;
  gender: Gender;
  status: StudentStatus;
  grade?: string;
  className?: string;
  riskScore?: number;
}

// ─── Payments / Invoices ──────────────────────────────────────────────────────

export type InvoiceStatus = 'UNPAID' | 'PAID' | 'OVERDUE' | 'PARTIAL' | 'CANCELLED';
export type PaymentChannel = 'QRIS' | 'VIRTUAL_ACCOUNT' | 'GOPAY' | 'OVO' | 'CASH';

export interface Invoice {
  id: string;
  invoiceNumber: string;
  studentId: string;
  studentName: string;
  schoolId: string;
  type: 'SPP' | 'PPDB' | 'ACTIVITY' | 'OTHER';
  title: string;
  amount: number;
  paidAmount: number;
  dueDate: string;
  status: InvoiceStatus;
  paymentChannel?: PaymentChannel;
  paidAt?: string;
  createdAt: string;
  updatedAt: string;
}

export interface PaymentSummary {
  totalInvoices: number;
  totalPaid: number;
  totalUnpaid: number;
  totalOverdue: number;
  collectionRate: number;
}

// ─── Attendance ───────────────────────────────────────────────────────────────

export type AttendanceStatus = 'PRESENT' | 'ABSENT' | 'LATE' | 'EXCUSED' | 'SICK';

export interface AttendanceRecord {
  id: string;
  studentId: string;
  date: string;
  status: AttendanceStatus;
  checkInTime?: string;
  checkOutTime?: string;
  notes?: string;
}

export interface AttendanceSummary {
  date: string;
  totalStudents: number;
  present: number;
  absent: number;
  late: number;
  excused: number;
  sick: number;
  attendanceRate: number;
}

// ─── Notifications ────────────────────────────────────────────────────────────

export type NotificationType =
  | 'ATTENDANCE'
  | 'PAYMENT'
  | 'EXAM'
  | 'ANNOUNCEMENT'
  | 'ALERT'
  | 'HEALTH'
  | 'TRANSPORT';

export interface Notification {
  id: string;
  type: NotificationType;
  title: string;
  body: string;
  isRead: boolean;
  createdAt: string;
  data?: Record<string, unknown>;
}

// ─── Dashboard ────────────────────────────────────────────────────────────────

export interface DashboardStats {
  totalStudents: number;
  activeStudents: number;
  attendanceToday: AttendanceSummary;
  pendingPayments: number;
  overduePayments: number;
  collectionRate: number;
  riskStudents: number;
}

// ─── API Response ─────────────────────────────────────────────────────────────

export interface ApiResponse<T> {
  success: boolean;
  data: T;
  meta?: PaginationMeta;
}

export interface ApiError {
  success: false;
  error: {
    code: string;
    message: string;
  };
}

export interface PaginationMeta {
  total: number;
  page: number;
  limit: number;
  totalPages: number;
  hasNextPage: boolean;
  hasPrevPage: boolean;
}

export interface PaginatedResponse<T> {
  items: T[];
  meta: PaginationMeta;
}

// ─── Form Types ───────────────────────────────────────────────────────────────

export interface LoginFormData {
  subdomain: string;
  email: string;
  password: string;
}

// ─── Navigation ───────────────────────────────────────────────────────────────

export type RootStackParamList = {
  '(auth)/login': undefined;
  '(tabs)': undefined;
  'students/[id]': { id: string };
};
