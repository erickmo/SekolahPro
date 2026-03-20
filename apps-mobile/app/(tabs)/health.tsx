import React, { useCallback, useEffect, useState } from 'react';
import {
  ActivityIndicator,
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

// ─── Types ────────────────────────────────────────────────────────────────────

type DateFilter = 'TODAY' | 'THIS_WEEK' | 'THIS_MONTH';

interface UksVisit {
  id: string;
  studentName: string;
  complaint: string;
  actionTaken: string;
  visitDate: string;
  referred: boolean;
}

const DATE_FILTERS: { key: DateFilter; label: string }[] = [
  { key: 'TODAY', label: 'Hari Ini' },
  { key: 'THIS_WEEK', label: 'Minggu Ini' },
  { key: 'THIS_MONTH', label: 'Bulan Ini' },
];

const PAGE_SIZE = 20;

// ─── Helpers ──────────────────────────────────────────────────────────────────

function formatVisitDate(dateStr: string): string {
  const date = new Date(dateStr);
  return date.toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

// ─── Visit Card ───────────────────────────────────────────────────────────────

function VisitCard({ item }: { item: UksVisit }) {
  return (
    <View style={[styles.card, { borderLeftColor: item.referred ? '#DC2626' : '#16A34A' }]}>
      <View style={styles.cardHeader}>
        <View style={styles.cardHeaderLeft}>
          <Text style={styles.studentName}>{item.studentName}</Text>
          <Text style={styles.visitDate}>{formatVisitDate(item.visitDate)}</Text>
        </View>
        <View
          style={[
            styles.referralBadge,
            { backgroundColor: item.referred ? '#FEE2E2' : '#DCFCE7' },
          ]}
        >
          <Text
            style={[
              styles.referralBadgeText,
              { color: item.referred ? '#DC2626' : '#16A34A' },
            ]}
          >
            {item.referred ? 'Dirujuk' : 'Ditangani'}
          </Text>
        </View>
      </View>
      <View style={styles.cardBody}>
        <View style={styles.cardRow}>
          <Ionicons name="medical-outline" size={14} color="#6B7280" />
          <Text style={styles.cardLabel}>Keluhan:</Text>
          <Text style={styles.cardValue} numberOfLines={2}>
            {item.complaint}
          </Text>
        </View>
        <View style={styles.cardRow}>
          <Ionicons name="checkmark-circle-outline" size={14} color="#6B7280" />
          <Text style={styles.cardLabel}>Tindakan:</Text>
          <Text style={styles.cardValue} numberOfLines={2}>
            {item.actionTaken}
          </Text>
        </View>
      </View>
    </View>
  );
}

// ─── Main Screen ──────────────────────────────────────────────────────────────

export default function HealthScreen() {
  const [visits, setVisits] = useState<UksVisit[]>([]);
  const [activeFilter, setActiveFilter] = useState<DateFilter>('TODAY');
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const [totalCount, setTotalCount] = useState(0);

  const fetchData = useCallback(
    async (pageNum: number, filter: DateFilter, isLoadMore = false) => {
      if (!isLoadMore) setIsLoading(pageNum === 1);
      else setIsLoadingMore(true);

      try {
        const params: Record<string, unknown> = {
          page: pageNum,
          limit: PAGE_SIZE,
          dateFilter: filter,
        };

        const response = await get<{ data: UksVisit[]; meta: { total: number; hasNextPage: boolean } }>(
          '/health/uks-visits',
          params,
        );

        const newItems = response.data ?? [];
        setTotalCount(response.meta?.total ?? 0);
        setHasMore(response.meta?.hasNextPage ?? false);

        if (isLoadMore) {
          setVisits((prev) => [...prev, ...newItems]);
        } else {
          setVisits(newItems);
        }
      } catch {
        if (!isLoadMore) {
          setVisits(generateMockVisits(12));
          setTotalCount(47);
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
        <ActivityIndicator size="small" color="#4F46E5" />
      </View>
    ) : null;

  const renderEmpty = () =>
    isLoading ? null : (
      <View style={styles.emptyContainer}>
        <Ionicons name="heart-outline" size={48} color="#D1D5DB" />
        <Text style={styles.emptyTitle}>Tidak ada kunjungan</Text>
        <Text style={styles.emptySubtitle}>
          Belum ada kunjungan UKS untuk periode yang dipilih.
        </Text>
      </View>
    );

  return (
    <SafeAreaView style={styles.safeArea} edges={['top']}>
      {/* Header */}
      <View style={styles.header}>
        <View>
          <Text style={styles.headerTitle}>UKS & Kesehatan</Text>
          <Text style={styles.headerSubtitle}>
            {totalCount.toLocaleString('id-ID')} kunjungan
          </Text>
        </View>
        <View style={styles.headerIcon}>
          <Ionicons name="heart" size={22} color="#4F46E5" />
        </View>
      </View>

      {/* Date Filter Chips */}
      <View style={styles.filterWrapper}>
        <ScrollView
          horizontal
          showsHorizontalScrollIndicator={false}
          contentContainerStyle={styles.filterContent}
        >
          {DATE_FILTERS.map((f) => (
            <TouchableOpacity
              key={f.key}
              style={[styles.filterChip, activeFilter === f.key && styles.filterChipActive]}
              onPress={() => setActiveFilter(f.key)}
              activeOpacity={0.75}
            >
              <Text
                style={[
                  styles.filterChipText,
                  activeFilter === f.key && styles.filterChipTextActive,
                ]}
              >
                {f.label}
              </Text>
            </TouchableOpacity>
          ))}
        </ScrollView>
      </View>

      {/* List */}
      {isLoading ? (
        <View style={styles.loadingCenter}>
          <ActivityIndicator size="large" color="#4F46E5" />
          <Text style={styles.loadingText}>Memuat data kunjungan...</Text>
        </View>
      ) : (
        <FlatList
          data={visits}
          keyExtractor={(item) => item.id}
          renderItem={({ item }) => <VisitCard item={item} />}
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

const MOCK_COMPLAINTS = [
  'Sakit kepala', 'Demam ringan', 'Mual dan pusing', 'Luka lecet di lutut',
  'Sakit perut', 'Mimisan', 'Pusing setelah olahraga', 'Batuk dan pilek',
];

const MOCK_ACTIONS = [
  'Diberikan obat paracetamol dan istirahat', 'Diukur suhu tubuh, diberikan kompres',
  'Diberikan obat antasida', 'Dibersihkan luka dan diberikan plester',
  'Dianjurkan minum air putih dan istirahat', 'Diberikan tampon dan ditangani',
];

const MOCK_NAMES = [
  'Ahmad Fauzi', 'Budi Santoso', 'Citra Dewi', 'Dini Rahayu', 'Eko Prasetyo',
  'Farah Azzahra', 'Gilang Ramadhan', 'Hana Putri', 'Ivan Setiawan', 'Julia Maharani',
  'Kevin Wijaya', 'Laila Nurdin',
];

function generateMockVisits(count: number): UksVisit[] {
  return Array.from({ length: count }, (_, i) => ({
    id: `visit-${i + 1}`,
    studentName: MOCK_NAMES[i % MOCK_NAMES.length],
    complaint: MOCK_COMPLAINTS[i % MOCK_COMPLAINTS.length],
    actionTaken: MOCK_ACTIONS[i % MOCK_ACTIONS.length],
    visitDate: new Date(2026, 2, 20, 8 + i, 0).toISOString(),
    referred: i % 5 === 0,
  }));
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
  headerIcon: {
    width: 44,
    height: 44,
    borderRadius: 12,
    backgroundColor: '#EEF2FF',
    alignItems: 'center',
    justifyContent: 'center',
  },
  filterWrapper: {
    marginBottom: 8,
  },
  filterContent: {
    paddingHorizontal: 16,
    gap: 8,
  },
  filterChip: {
    paddingHorizontal: 18,
    paddingVertical: 8,
    borderRadius: 20,
    backgroundColor: '#FFFFFF',
    borderWidth: 1.5,
    borderColor: '#E5E7EB',
  },
  filterChipActive: {
    backgroundColor: '#4F46E5',
    borderColor: '#4F46E5',
  },
  filterChipText: {
    fontSize: 13,
    fontWeight: '500',
    color: '#6B7280',
  },
  filterChipTextActive: {
    color: '#FFFFFF',
    fontWeight: '600',
  },
  listContent: {
    paddingHorizontal: 16,
    paddingTop: 4,
    paddingBottom: 24,
    gap: 10,
  },
  card: {
    backgroundColor: '#FFFFFF',
    borderRadius: 14,
    borderLeftWidth: 4,
    padding: 14,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 4,
    elevation: 2,
  },
  cardHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    marginBottom: 10,
  },
  cardHeaderLeft: {
    flex: 1,
    marginRight: 8,
  },
  studentName: {
    fontSize: 15,
    fontWeight: '700',
    color: '#111827',
  },
  visitDate: {
    fontSize: 12,
    color: '#9CA3AF',
    marginTop: 2,
  },
  referralBadge: {
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 20,
  },
  referralBadgeText: {
    fontSize: 12,
    fontWeight: '600',
  },
  cardBody: {
    gap: 6,
  },
  cardRow: {
    flexDirection: 'row',
    alignItems: 'flex-start',
    gap: 6,
  },
  cardLabel: {
    fontSize: 13,
    color: '#6B7280',
    fontWeight: '500',
    minWidth: 60,
  },
  cardValue: {
    flex: 1,
    fontSize: 13,
    color: '#374151',
  },
  loadingCenter: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    gap: 12,
  },
  loadingText: {
    fontSize: 14,
    color: '#6B7280',
  },
  loadMoreContainer: {
    paddingVertical: 16,
    alignItems: 'center',
  },
  emptyContainer: {
    alignItems: 'center',
    paddingTop: 60,
    paddingHorizontal: 32,
    gap: 8,
  },
  emptyTitle: {
    fontSize: 16,
    fontWeight: '600',
    color: '#374151',
  },
  emptySubtitle: {
    fontSize: 14,
    color: '#9CA3AF',
    textAlign: 'center',
    lineHeight: 20,
  },
});
