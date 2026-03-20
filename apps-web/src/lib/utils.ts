import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';
import { format, parseISO, formatDistanceToNow } from 'date-fns';
import { id } from 'date-fns/locale';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

// ─── Date Helpers ─────────────────────────────────────────────────────────────

export function formatDate(date: string | Date, fmt = 'dd MMM yyyy'): string {
  const d = typeof date === 'string' ? parseISO(date) : date;
  return format(d, fmt, { locale: id });
}

export function formatDateTime(date: string | Date): string {
  const d = typeof date === 'string' ? parseISO(date) : date;
  return format(d, 'dd MMM yyyy, HH:mm', { locale: id });
}

export function timeAgo(date: string | Date): string {
  const d = typeof date === 'string' ? parseISO(date) : date;
  return formatDistanceToNow(d, { addSuffix: true, locale: id });
}

// ─── Currency ─────────────────────────────────────────────────────────────────

export function formatCurrency(amount: number): string {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(amount);
}

export function formatNumber(n: number): string {
  return new Intl.NumberFormat('id-ID').format(n);
}

// ─── String Helpers ───────────────────────────────────────────────────────────

export function truncate(str: string, maxLength: number): string {
  if (str.length <= maxLength) return str;
  return str.slice(0, maxLength) + '...';
}

export function initials(name: string): string {
  return name
    .split(' ')
    .slice(0, 2)
    .map((n) => n[0])
    .join('')
    .toUpperCase();
}

export function genderLabel(gender: string): string {
  return gender === 'MALE' ? 'Laki-laki' : 'Perempuan';
}

// ─── Status Label Maps ────────────────────────────────────────────────────────

export const ROLE_LABELS: Record<string, string> = {
  SUPERADMIN: 'Super Admin',
  ADMIN_YAYASAN: 'Admin Yayasan',
  ADMIN_SEKOLAH: 'Admin Sekolah',
  KEPALA_SEKOLAH: 'Kepala Sekolah',
  KEPALA_KURIKULUM: 'Kepala Kurikulum',
  BENDAHARA: 'Bendahara',
  GURU: 'Guru',
  WALI_KELAS: 'Wali Kelas',
  GURU_BK: 'Guru BK',
  OPERATOR_SIMS: 'Operator SIMS',
  KASIR_KOPERASI: 'Kasir Koperasi',
  PUSTAKAWAN: 'Pustakawan',
  PETUGAS_UKS: 'Petugas UKS',
  TATA_USAHA: 'Tata Usaha',
  SISWA: 'Siswa',
  ORANG_TUA: 'Orang Tua',
};

export const INVOICE_STATUS_LABELS: Record<string, string> = {
  PENDING: 'Menunggu',
  PAID: 'Lunas',
  OVERDUE: 'Jatuh Tempo',
  CANCELLED: 'Dibatalkan',
};

export const ATTENDANCE_STATUS_LABELS: Record<string, string> = {
  PRESENT: 'Hadir',
  ABSENT: 'Tidak Hadir',
  LATE: 'Terlambat',
  SICK: 'Sakit',
  PERMITTED: 'Izin',
};

export const LOAN_STATUS_LABELS: Record<string, string> = {
  BORROWED: 'Dipinjam',
  RETURNED: 'Dikembalikan',
  OVERDUE: 'Terlambat',
};

export const SESSION_STATUS_LABELS: Record<string, string> = {
  SCHEDULED: 'Terjadwal',
  COMPLETED: 'Selesai',
  CANCELLED: 'Dibatalkan',
};

export const COUNSELING_TYPE_LABELS: Record<string, string> = {
  ACADEMIC: 'Akademik',
  PERSONAL: 'Personal',
  CAREER: 'Karir',
  SOCIAL: 'Sosial',
};

export const BULLYING_CATEGORY_LABELS: Record<string, string> = {
  PHYSICAL: 'Fisik',
  VERBAL: 'Verbal',
  CYBERBULLYING: 'Cyberbullying',
  SEXUAL_HARASSMENT: 'Pelecehan Seksual',
  OTHER: 'Lainnya',
};
