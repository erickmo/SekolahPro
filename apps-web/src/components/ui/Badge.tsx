import React from 'react';
import { cn } from '@/lib/utils';

type BadgeVariant =
  | 'default'
  | 'primary'
  | 'success'
  | 'warning'
  | 'danger'
  | 'info'
  | 'gray';

interface BadgeProps {
  children: React.ReactNode;
  variant?: BadgeVariant;
  size?: 'sm' | 'md';
  className?: string;
}

const variantClasses: Record<BadgeVariant, string> = {
  default: 'bg-gray-100 text-gray-700',
  primary: 'bg-primary-100 text-primary-700',
  success: 'bg-green-100 text-green-700',
  warning: 'bg-amber-100 text-amber-700',
  danger: 'bg-red-100 text-red-700',
  info: 'bg-blue-100 text-blue-700',
  gray: 'bg-gray-100 text-gray-600',
};

const sizeClasses = {
  sm: 'px-2 py-0.5 text-xs',
  md: 'px-2.5 py-1 text-xs',
};

export function Badge({
  children,
  variant = 'default',
  size = 'md',
  className,
}: BadgeProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center rounded-full font-medium',
        variantClasses[variant],
        sizeClasses[size],
        className
      )}
    >
      {children}
    </span>
  );
}

// ─── Status Badge helpers ─────────────────────────────────────────────────────

export function InvoiceStatusBadge({ status }: { status: string }) {
  const map: Record<string, { label: string; variant: BadgeVariant }> = {
    PENDING: { label: 'Menunggu', variant: 'warning' },
    PAID: { label: 'Lunas', variant: 'success' },
    OVERDUE: { label: 'Jatuh Tempo', variant: 'danger' },
    CANCELLED: { label: 'Dibatalkan', variant: 'gray' },
  };
  const { label, variant } = map[status] || { label: status, variant: 'default' };
  return <Badge variant={variant}>{label}</Badge>;
}

export function AttendanceBadge({ status }: { status: string }) {
  const map: Record<string, { label: string; variant: BadgeVariant }> = {
    PRESENT: { label: 'Hadir', variant: 'success' },
    ABSENT: { label: 'Tidak Hadir', variant: 'danger' },
    LATE: { label: 'Terlambat', variant: 'warning' },
    SICK: { label: 'Sakit', variant: 'info' },
    PERMITTED: { label: 'Izin', variant: 'primary' },
  };
  const { label, variant } = map[status] || { label: status, variant: 'default' };
  return <Badge variant={variant}>{label}</Badge>;
}

export function LoanStatusBadge({ status }: { status: string }) {
  const map: Record<string, { label: string; variant: BadgeVariant }> = {
    BORROWED: { label: 'Dipinjam', variant: 'warning' },
    RETURNED: { label: 'Dikembalikan', variant: 'success' },
    OVERDUE: { label: 'Terlambat', variant: 'danger' },
  };
  const { label, variant } = map[status] || { label: status, variant: 'default' };
  return <Badge variant={variant}>{label}</Badge>;
}

export function SessionStatusBadge({ status }: { status: string }) {
  const map: Record<string, { label: string; variant: BadgeVariant }> = {
    SCHEDULED: { label: 'Terjadwal', variant: 'primary' },
    COMPLETED: { label: 'Selesai', variant: 'success' },
    CANCELLED: { label: 'Dibatalkan', variant: 'gray' },
  };
  const { label, variant } = map[status] || { label: status, variant: 'default' };
  return <Badge variant={variant}>{label}</Badge>;
}

export function RiskScoreBadge({ score }: { score: number }) {
  let variant: BadgeVariant = 'success';
  let label = 'Rendah';
  if (score >= 70) { variant = 'danger'; label = 'Tinggi'; }
  else if (score >= 40) { variant = 'warning'; label = 'Sedang'; }
  return <Badge variant={variant}>{label} ({Math.round(score)})</Badge>;
}
