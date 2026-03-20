import React, { useCallback, useEffect, useState } from 'react';
import {
  FlatList,
  RefreshControl,
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { Ionicons } from '@expo/vector-icons';
import { get } from '../../src/lib/api';
import InvoiceCard from '../../src/components/InvoiceCard';
import EmptyState from '../../src/components/EmptyState';
import LoadingSpinner from '../../src/components/LoadingSpinner';
import type { Invoice, InvoiceStatus, PaginatedResponse, PaymentSummary } from '../../src/types';

// ─── Types ────────────────────────────────────────────────────────────────────

type StatusFilter = 'ALL' | InvoiceStatus;

const STATUS_FILTERS: { key: StatusFilter; label: string; color: string }[] = [
  { key: 'ALL', label: 'Semua', color: '#6B7280' },
  { key: 'UNPAID', label: 'Belum Bayar', color: '#1E40AF' },
  { key: 'OVERDUE', label: 'Jatuh Tempo', color: '#DC2626' },
  { key: 'PAID', label: 'Lunas', color: '#059669' },
  { key: 'PARTIAL', label: 'Sebagian', color: '#D97706' },
];

function formatCurrency(amount: number): string {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
    notation: 'compact',
    compactDisplay: 'short',
  }).format(amount);
}

// ─── Summary Banner ───────────────────────────────────────────────────────────

function SummaryBanner({ summary }: { summary: PaymentSummary | null }) {
  if (!summary) return null;

  return (
    <View style={styles.summaryBanner}>
      <View style={styles.summaryItem}>
        <Text style={styles.summaryLabel}>Total Tagihan</Text>
        <Text style={styles.summaryValue}>
          {summary.totalInvoices.toLocaleString('id-ID')}
        </Text>
      </View>
      <View style={styles.summaryDivider} />
      <View style={styles.summaryItem}>
        <Text style={styles.summaryLabel}>Sudah Bayar</Text>
        <Text style={[styles.summaryValue, { color: '#059669' }]}>
          {summary.totalPaid.toLocaleString('id-ID')}
        </Text>
      </View>
      <View style={styles.summaryDivider} />
      <View style={styles.summaryItem}>
        <Text style={styles.summaryLabel}>Jatuh Tempo</Text>
        <Text style={[styles.summaryValue, { color: '#DC2626' }]}>
          {summary.totalOverdue.toLocaleString('id-ID')}
        </Text>
      </View>
      <View style={styles.summaryDivider} />
      <View style={styles.summaryItem}>
        <Text style={styles.summaryLabel}>Koleksi</Text>
        <Text style={[styles.summaryValue, { color: '#4F46E5' }]}>
          {summary.collectionRate.toFixed(0)}%
        </Text>
      </View>
    </View>
  );
}

// ─── Main Screen ──────────────────────────────────────────────────────────────

export default function PaymentsScreen() {
  const [invoices, setInvoices] = useState<Invoice[]>([]);
  const [summary, setSummary] = useState<PaymentSummary | null>(null);
  const [activeFilter, setActiveFilter] = useState<StatusFilter>('ALL');
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const [isLoadingMore, setIsLoadingMore] = useState(false);

  const fetchData = useCallback(
    async (pageNum: number, filter: StatusFilter, isLoadMore = false) => {
      if (!isLoadMore) setIsLoading(pageNum === 1);
      else setIsLoadingMore(true);

      try {
        const params: Record<string, unknown> = { page: pageNum, limit: 15 };
        if (filter !== 'ALL') params.status = filter;

        const [invoicesData, summaryData] = await Promise.all([
          get<PaginatedResponse<Invoice>>('/payments/invoices', params),
          pageNum === 1
            ? get<PaymentSummary>('/payments/summary')
            : Promise.resolve(null),
        ]);

        setHasMore(invoicesData.meta?.hasNextPage ?? false);
        if (isLoadMore) {
          setInvoices((prev) => [...prev, ...(invoicesData.items ?? [])]);
        } else {
          setInvoices(invoicesData.items ?? []);
        }
        if (summaryData) setSummary(summaryData);
      } catch {
        // Fallback mock data
        if (!isLoadMore) {
          setSummary({
            totalInvoices: 1248,
            totalPaid: 1025,
            totalUnpaid: 200,
            totalOverdue: 23,
            collectionRate: 82.1,
          });
          setInvoices(generateMockInvoices(15));
          setHasMore(true);
        }
      } finally {
        setIsLoading(false);
        setIsRefreshing(false);
        setIsLoadingMore(false);
      }
    },
    [],
  );

  useEffect(() => {
    setPage(1);
    fetchData(1, activeFilter);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [activeFilter]);

  const onRefresh = useCallback(() => {
    setIsRefreshing(true);
    setPage(1);
    fetchData(1, activeFilter);
  }, [fetchData, activeFilter]);

  const onLoadMore = useCallback(() => {
    if (!hasMore || isLoadingMore || isLoading) return;
    const next = page + 1;
    setPage(next);
    fetchData(next, activeFilter, true);
  }, [hasMore, isLoadingMore, isLoading, page, fetchData, activeFilter]);

  const renderFooter = () =>
    isLoadingMore ? (
      <View style={styles.loadMoreContainer}>
        <LoadingSpinner size="small" />
      </View>
    ) : null;

  const renderEmpty = () =>
    isLoading ? null : (
      <EmptyState
        icon="card-outline"
        title="Tidak ada tagihan"
        subtitle={
          activeFilter !== 'ALL'
            ? `Tidak ada tagihan dengan status "${STATUS_FILTERS.find((f) => f.key === activeFilter)?.label}".`
            : 'Belum ada data tagihan tersedia.'
        }
      />
    );

  return (
    <SafeAreaView style={styles.safeArea} edges={['top']}>
      {/* Header */}
      <View style={styles.header}>
        <View>
          <Text style={styles.headerTitle}>Pembayaran SPP</Text>
          <Text style={styles.headerSubtitle}>
            Kelola tagihan & status pembayaran siswa
          </Text>
        </View>
        <TouchableOpacity style={styles.exportButton} activeOpacity={0.7}>
          <Ionicons name="download-outline" size={20} color="#4F46E5" />
        </TouchableOpacity>
      </View>

      {/* Summary Banner */}
      <SummaryBanner summary={summary} />

      {/* Filter Tabs */}
      <View style={styles.filterTabsWrapper}>
        <ScrollView
          horizontal
          showsHorizontalScrollIndicator={false}
          contentContainerStyle={styles.filterTabsContent}
        >
          {STATUS_FILTERS.map((filter) => (
            <TouchableOpacity
              key={filter.key}
              style={[
                styles.filterTab,
                activeFilter === filter.key && {
                  backgroundColor: filter.color,
                  borderColor: filter.color,
                },
              ]}
              onPress={() => setActiveFilter(filter.key)}
              activeOpacity={0.75}
            >
              <Text
                style={[
                  styles.filterTabText,
                  activeFilter === filter.key && styles.filterTabTextActive,
                ]}
              >
                {filter.label}
              </Text>
            </TouchableOpacity>
          ))}
        </ScrollView>
      </View>

      {/* Invoice List */}
      {isLoading ? (
        <LoadingSpinner label="Memuat data tagihan..." style={styles.loadingCenter} />
      ) : (
        <FlatList
          data={invoices}
          keyExtractor={(item) => item.id}
          renderItem={({ item }) => <InvoiceCard invoice={item} />}
          contentContainerStyle={styles.listContent}
          showsVerticalScrollIndicator={false}
          refreshControl={
            <RefreshControl
              refreshing={isRefreshing}
              onRefresh={onRefresh}
              tintColor="#4F46E5"
              colors={['#4F46E5']}
            />
          }
          onEndReached={onLoadMore}
          onEndReachedThreshold={0.3}
          ListFooterComponent={renderFooter}
          ListEmptyComponent={renderEmpty}
        />
      )}
    </SafeAreaView>
  );
}

// ─── Mock Data ────────────────────────────────────────────────────────────────

const MOCK_STUDENT_NAMES = [
  'Ahmad Fauzi', 'Budi Santoso', 'Citra Dewi', 'Dini Rahayu', 'Eko Prasetyo',
  'Farah Azzahra', 'Gilang Ramadhan', 'Hana Putri', 'Ivan Setiawan', 'Julia Maharani',
  'Kevin Wijaya', 'Laila Nurdin', 'Muhammad Rizki', 'Nadia Sari', 'Oky Firmansyah',
];

const MOCK_STATUSES: Invoice['status'][] = ['PAID', 'UNPAID', 'OVERDUE', 'PARTIAL'];

const MOCK_MONTHS = [
  'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
  'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember',
];

function generateMockInvoices(count: number): Invoice[] {
  return Array.from({ length: count }, (_, i) => {
    const status = MOCK_STATUSES[i % MOCK_STATUSES.length];
    const amount = 350000 + (i % 5) * 50000;
    const paidAmount = status === 'PAID' ? amount : status === 'PARTIAL' ? amount * 0.5 : 0;
    const dueDate = new Date(2026, 2, 15 - (i % 20));

    return {
      id: `invoice-${i + 1}`,
      invoiceNumber: `INV/SPP/2026/${String(i + 1).padStart(4, '0')}`,
      studentId: `student-${i + 1}`,
      studentName: MOCK_STUDENT_NAMES[i % MOCK_STUDENT_NAMES.length],
      schoolId: 'school-1',
      type: 'SPP',
      title: `SPP ${MOCK_MONTHS[i % 12]} 2026`,
      amount,
      paidAmount,
      dueDate: dueDate.toISOString(),
      status,
      createdAt: new Date(2026, 2, 1).toISOString(),
      updatedAt: new Date().toISOString(),
    };
  });
}

// ─── Styles ───────────────────────────────────────────────────────────────────

const styles = StyleSheet.create({
  safeArea: {
    flex: 1,
    backgroundColor: '#F9FAFB',
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingTop: 20,
    paddingBottom: 12,
  },
  headerTitle: {
    fontSize: 24,
    fontWeight: '700',
    color: '#111827',
  },
  headerSubtitle: {
    fontSize: 13,
    color: '#6B7280',
    marginTop: 2,
  },
  exportButton: {
    width: 44,
    height: 44,
    borderRadius: 12,
    backgroundColor: '#EEF2FF',
    alignItems: 'center',
    justifyContent: 'center',
  },
  summaryBanner: {
    flexDirection: 'row',
    backgroundColor: '#FFFFFF',
    marginHorizontal: 16,
    borderRadius: 16,
    padding: 16,
    marginBottom: 12,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 4,
    elevation: 2,
  },
  summaryItem: {
    flex: 1,
    alignItems: 'center',
  },
  summaryLabel: {
    fontSize: 11,
    color: '#9CA3AF',
    marginBottom: 4,
    textAlign: 'center',
  },
  summaryValue: {
    fontSize: 16,
    fontWeight: '700',
    color: '#111827',
  },
  summaryDivider: {
    width: 1,
    backgroundColor: '#F3F4F6',
    marginVertical: 2,
  },
  filterTabsWrapper: {
    marginBottom: 8,
  },
  filterTabsContent: {
    paddingHorizontal: 16,
    gap: 8,
  },
  filterTab: {
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderRadius: 20,
    backgroundColor: '#FFFFFF',
    borderWidth: 1.5,
    borderColor: '#E5E7EB',
  },
  filterTabText: {
    fontSize: 13,
    fontWeight: '500',
    color: '#6B7280',
  },
  filterTabTextActive: {
    color: '#FFFFFF',
    fontWeight: '600',
  },
  listContent: {
    paddingHorizontal: 16,
    paddingTop: 4,
    paddingBottom: 20,
  },
  loadingCenter: {
    flex: 1,
  },
  loadMoreContainer: {
    paddingVertical: 16,
  },
});
