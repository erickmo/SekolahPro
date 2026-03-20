import React from 'react';
import {
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import type { Invoice, InvoiceStatus } from '../types';

// ─── Helpers ──────────────────────────────────────────────────────────────────

function getStatusConfig(status: InvoiceStatus) {
  switch (status) {
    case 'PAID':
      return {
        label: 'Lunas',
        color: '#065F46',
        bgColor: '#D1FAE5',
        borderColor: '#6EE7B7',
        icon: 'checkmark-circle' as const,
      };
    case 'UNPAID':
      return {
        label: 'Belum Bayar',
        color: '#1E40AF',
        bgColor: '#DBEAFE',
        borderColor: '#93C5FD',
        icon: 'time-outline' as const,
      };
    case 'OVERDUE':
      return {
        label: 'Jatuh Tempo',
        color: '#991B1B',
        bgColor: '#FEE2E2',
        borderColor: '#FCA5A5',
        icon: 'alert-circle' as const,
      };
    case 'PARTIAL':
      return {
        label: 'Sebagian',
        color: '#92400E',
        bgColor: '#FEF3C7',
        borderColor: '#FCD34D',
        icon: 'ellipsis-horizontal-circle' as const,
      };
    case 'CANCELLED':
      return {
        label: 'Dibatalkan',
        color: '#374151',
        bgColor: '#F3F4F6',
        borderColor: '#D1D5DB',
        icon: 'close-circle-outline' as const,
      };
    default:
      return {
        label: status,
        color: '#374151',
        bgColor: '#F3F4F6',
        borderColor: '#D1D5DB',
        icon: 'help-circle-outline' as const,
      };
  }
}

function formatCurrency(amount: number): string {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(amount);
}

function formatDate(dateString: string): string {
  return new Date(dateString).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  });
}

function getInvoiceTypeLabel(type: Invoice['type']): string {
  switch (type) {
    case 'SPP':
      return 'SPP Bulanan';
    case 'PPDB':
      return 'PPDB';
    case 'ACTIVITY':
      return 'Kegiatan';
    case 'OTHER':
      return 'Lainnya';
    default:
      return type;
  }
}

// ─── Types ────────────────────────────────────────────────────────────────────

interface InvoiceCardProps {
  invoice: Invoice;
  onPress?: (invoice: Invoice) => void;
}

// ─── Component ────────────────────────────────────────────────────────────────

export default function InvoiceCard({ invoice, onPress }: InvoiceCardProps) {
  const statusConfig = getStatusConfig(invoice.status);
  const isOverdue = invoice.status === 'OVERDUE';
  const remainingAmount = invoice.amount - invoice.paidAmount;

  return (
    <TouchableOpacity
      style={[
        styles.card,
        isOverdue && styles.cardOverdue,
      ]}
      onPress={() => onPress?.(invoice)}
      activeOpacity={0.75}
      disabled={!onPress}
    >
      {/* Left accent bar */}
      <View
        style={[styles.accentBar, { backgroundColor: statusConfig.borderColor }]}
      />

      <View style={styles.content}>
        {/* Top row */}
        <View style={styles.topRow}>
          <View style={styles.titleContainer}>
            <Text style={styles.invoiceTitle} numberOfLines={1}>
              {invoice.title}
            </Text>
            <Text style={styles.invoiceNumber} numberOfLines={1}>
              {invoice.invoiceNumber} · {getInvoiceTypeLabel(invoice.type)}
            </Text>
          </View>

          <View
            style={[
              styles.statusBadge,
              { backgroundColor: statusConfig.bgColor },
            ]}
          >
            <Ionicons
              name={statusConfig.icon}
              size={11}
              color={statusConfig.color}
            />
            <Text style={[styles.statusText, { color: statusConfig.color }]}>
              {statusConfig.label}
            </Text>
          </View>
        </View>

        {/* Student name */}
        <Text style={styles.studentName} numberOfLines={1}>
          {invoice.studentName}
        </Text>

        {/* Bottom row */}
        <View style={styles.bottomRow}>
          <View>
            <Text style={styles.amountLabel}>
              {invoice.status === 'PARTIAL' ? 'Sisa tagihan' : 'Total tagihan'}
            </Text>
            <Text
              style={[
                styles.amount,
                isOverdue && styles.amountOverdue,
              ]}
            >
              {formatCurrency(
                invoice.status === 'PARTIAL' ? remainingAmount : invoice.amount,
              )}
            </Text>
          </View>

          <View style={styles.dueDateContainer}>
            <Ionicons
              name="calendar-outline"
              size={12}
              color={isOverdue ? '#EF4444' : '#9CA3AF'}
            />
            <Text
              style={[
                styles.dueDate,
                isOverdue && styles.dueDateOverdue,
              ]}
            >
              {formatDate(invoice.dueDate)}
            </Text>
          </View>
        </View>

        {/* Progress bar for partial payments */}
        {invoice.status === 'PARTIAL' && invoice.amount > 0 && (
          <View style={styles.progressContainer}>
            <View style={styles.progressBg}>
              <View
                style={[
                  styles.progressFill,
                  {
                    width: `${Math.min(
                      (invoice.paidAmount / invoice.amount) * 100,
                      100,
                    )}%`,
                  },
                ]}
              />
            </View>
            <Text style={styles.progressText}>
              {Math.round((invoice.paidAmount / invoice.amount) * 100)}% terbayar
            </Text>
          </View>
        )}
      </View>
    </TouchableOpacity>
  );
}

// ─── Styles ───────────────────────────────────────────────────────────────────

const styles = StyleSheet.create({
  card: {
    backgroundColor: '#FFFFFF',
    borderRadius: 14,
    marginBottom: 10,
    flexDirection: 'row',
    overflow: 'hidden',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 4,
    elevation: 2,
  },
  cardOverdue: {
    shadowColor: '#EF4444',
    shadowOpacity: 0.1,
  },
  accentBar: {
    width: 4,
    borderRadius: 0,
  },
  content: {
    flex: 1,
    padding: 14,
    paddingLeft: 12,
  },
  topRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    marginBottom: 4,
  },
  titleContainer: {
    flex: 1,
    marginRight: 8,
  },
  invoiceTitle: {
    fontSize: 15,
    fontWeight: '600',
    color: '#111827',
  },
  invoiceNumber: {
    fontSize: 11,
    color: '#9CA3AF',
    marginTop: 1,
  },
  statusBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 3,
    paddingHorizontal: 8,
    paddingVertical: 3,
    borderRadius: 20,
  },
  statusText: {
    fontSize: 11,
    fontWeight: '600',
  },
  studentName: {
    fontSize: 13,
    color: '#6B7280',
    marginBottom: 10,
  },
  bottomRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-end',
  },
  amountLabel: {
    fontSize: 11,
    color: '#9CA3AF',
    marginBottom: 2,
  },
  amount: {
    fontSize: 17,
    fontWeight: '700',
    color: '#111827',
    letterSpacing: -0.3,
  },
  amountOverdue: {
    color: '#DC2626',
  },
  dueDateContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
  },
  dueDate: {
    fontSize: 12,
    color: '#9CA3AF',
  },
  dueDateOverdue: {
    color: '#EF4444',
    fontWeight: '500',
  },
  progressContainer: {
    marginTop: 10,
    gap: 4,
  },
  progressBg: {
    height: 4,
    backgroundColor: '#F3F4F6',
    borderRadius: 2,
    overflow: 'hidden',
  },
  progressFill: {
    height: '100%',
    backgroundColor: '#F59E0B',
    borderRadius: 2,
  },
  progressText: {
    fontSize: 10,
    color: '#9CA3AF',
  },
});
