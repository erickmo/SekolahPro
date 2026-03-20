import React, { useCallback, useEffect, useState } from 'react';
import {
  ActivityIndicator,
  FlatList,
  RefreshControl,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { Ionicons } from '@expo/vector-icons';
import { get } from '../../src/lib/api';

// ─── Types ────────────────────────────────────────────────────────────────────

type TicketStatus = 'OPEN' | 'IN_PROGRESS' | 'RESOLVED' | 'CLOSED';
type TicketPriority = 'LOW' | 'MEDIUM' | 'HIGH' | 'URGENT';

interface Ticket {
  id: string;
  subject: string;
  category: string;
  priority: TicketPriority;
  status: TicketStatus;
  description: string;
  createdAt: string;
}

const STATUS_CONFIG: Record<TicketStatus, { label: string; bg: string; text: string }> = {
  OPEN: { label: 'Terbuka', bg: '#DBEAFE', text: '#1D4ED8' },
  IN_PROGRESS: { label: 'Diproses', bg: '#FEF9C3', text: '#A16207' },
  RESOLVED: { label: 'Selesai', bg: '#DCFCE7', text: '#15803D' },
  CLOSED: { label: 'Ditutup', bg: '#F3F4F6', text: '#6B7280' },
};

const PRIORITY_CONFIG: Record<TicketPriority, { label: string; color: string }> = {
  LOW: { label: 'Rendah', color: '#6B7280' },
  MEDIUM: { label: 'Sedang', color: '#D97706' },
  HIGH: { label: 'Tinggi', color: '#DC2626' },
  URGENT: { label: 'Mendesak', color: '#7C3AED' },
};

const PAGE_SIZE = 20;

// ─── Helpers ──────────────────────────────────────────────────────────────────

function relativeTime(dateStr: string): string {
  const now = new Date();
  const date = new Date(dateStr);
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffMins < 1) return 'Baru saja';
  if (diffMins < 60) return `${diffMins} menit lalu`;
  if (diffHours < 24) return `${diffHours} jam lalu`;
  if (diffDays === 1) return 'Kemarin';
  if (diffDays < 30) return `${diffDays} hari lalu`;
  if (diffDays < 365) return `${Math.floor(diffDays / 30)} bulan lalu`;
  return `${Math.floor(diffDays / 365)} tahun lalu`;
}

// ─── Ticket Card ──────────────────────────────────────────────────────────────

function TicketCard({ item }: { item: Ticket }) {
  const [expanded, setExpanded] = useState(false);
  const statusCfg = STATUS_CONFIG[item.status];
  const priorityCfg = PRIORITY_CONFIG[item.priority];

  return (
    <TouchableOpacity
      style={styles.card}
      activeOpacity={0.85}
      onPress={() => setExpanded((prev) => !prev)}
    >
      <View style={styles.cardTop}>
        <Text style={styles.subject} numberOfLines={expanded ? undefined : 2}>
          {item.subject}
        </Text>
        <Ionicons
          name={expanded ? 'chevron-up' : 'chevron-down'}
          size={16}
          color="#9CA3AF"
          style={styles.chevron}
        />
      </View>

      <View style={styles.badgeRow}>
        <View style={styles.categoryBadge}>
          <Text style={styles.categoryText}>{item.category}</Text>
        </View>
        <View style={[styles.statusBadge, { backgroundColor: statusCfg.bg }]}>
          <Text style={[styles.statusText, { color: statusCfg.text }]}>{statusCfg.label}</Text>
        </View>
        <View style={styles.priorityBadge}>
          <View style={[styles.priorityDot, { backgroundColor: priorityCfg.color }]} />
          <Text style={[styles.priorityText, { color: priorityCfg.color }]}>
            {priorityCfg.label}
          </Text>
        </View>
      </View>

      {expanded && item.description ? (
        <Text style={styles.description}>{item.description}</Text>
      ) : null}

      <Text style={styles.timeText}>{relativeTime(item.createdAt)}</Text>
    </TouchableOpacity>
  );
}

// ─── Main Screen ──────────────────────────────────────────────────────────────

export default function HelpdeskScreen() {
  const [tickets, setTickets] = useState<Ticket[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const [totalCount, setTotalCount] = useState(0);

  const fetchData = useCallback(async (pageNum: number, isLoadMore = false) => {
    if (!isLoadMore) setIsLoading(pageNum === 1);
    else setIsLoadingMore(true);

    try {
      const response = await get<{
        data: Ticket[];
        meta: { total: number; hasNextPage: boolean };
      }>('/helpdesk/tickets', { page: pageNum, limit: PAGE_SIZE });

      const newItems = response.data ?? [];
      setTotalCount(response.meta?.total ?? 0);
      setHasMore(response.meta?.hasNextPage ?? false);

      if (isLoadMore) {
        setTickets((prev) => [...prev, ...newItems]);
      } else {
        setTickets(newItems);
      }
    } catch {
      if (!isLoadMore) {
        setTickets(MOCK_TICKETS);
        setTotalCount(MOCK_TICKETS.length);
        setHasMore(false);
      }
    } finally {
      setIsLoading(false);
      setIsRefreshing(false);
      setIsLoadingMore(false);
    }
  }, []);

  useEffect(() => {
    fetchData(1);
  }, [fetchData]);

  const onRefresh = useCallback(() => {
    setIsRefreshing(true);
    setPage(1);
    fetchData(1);
  }, [fetchData]);

  const onLoadMore = useCallback(() => {
    if (!hasMore || isLoadingMore || isLoading) return;
    const next = page + 1;
    setPage(next);
    fetchData(next, true);
  }, [hasMore, isLoadingMore, isLoading, page, fetchData]);

  const renderFooter = () =>
    isLoadingMore ? (
      <View style={styles.loadMoreContainer}>
        <ActivityIndicator size="small" color="#4F46E5" />
      </View>
    ) : null;

  const renderEmpty = () =>
    isLoading ? null : (
      <View style={styles.emptyContainer}>
        <Ionicons name="help-circle-outline" size={48} color="#D1D5DB" />
        <Text style={styles.emptyTitle}>Tidak ada tiket</Text>
        <Text style={styles.emptySubtitle}>Belum ada tiket helpdesk yang masuk.</Text>
      </View>
    );

  return (
    <SafeAreaView style={styles.safeArea} edges={['top']}>
      {/* Header */}
      <View style={styles.header}>
        <View>
          <Text style={styles.headerTitle}>Helpdesk</Text>
          <Text style={styles.headerSubtitle}>
            {totalCount.toLocaleString('id-ID')} tiket
          </Text>
        </View>
        <View style={styles.headerIcon}>
          <Ionicons name="help-circle" size={22} color="#4F46E5" />
        </View>
      </View>

      {/* List */}
      {isLoading ? (
        <View style={styles.loadingCenter}>
          <ActivityIndicator size="large" color="#4F46E5" />
          <Text style={styles.loadingText}>Memuat tiket helpdesk...</Text>
        </View>
      ) : (
        <FlatList
          data={tickets}
          keyExtractor={(item) => item.id}
          renderItem={({ item }) => <TicketCard item={item} />}
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

const MOCK_CATEGORIES = ['Akademik', 'Administrasi', 'Fasilitas', 'Keuangan', 'Teknis'];
const MOCK_SUBJECTS = [
  'Nilai rapor belum keluar', 'Pembayaran SPP tidak tercatat',
  'AC di kelas X IPA 2 rusak', 'Akses portal tidak bisa login',
  'Jadwal ujian bentrok', 'Kartu pelajar hilang',
  'Laporan absensi tidak sinkron', 'Permohonan surat keterangan',
];
const MOCK_PRIORITIES: TicketPriority[] = ['LOW', 'MEDIUM', 'HIGH', 'URGENT'];
const MOCK_STATUSES: TicketStatus[] = ['OPEN', 'IN_PROGRESS', 'RESOLVED', 'CLOSED'];

const MOCK_TICKETS: Ticket[] = Array.from({ length: 12 }, (_, i) => ({
  id: `ticket-${i + 1}`,
  subject: MOCK_SUBJECTS[i % MOCK_SUBJECTS.length],
  category: MOCK_CATEGORIES[i % MOCK_CATEGORIES.length],
  priority: MOCK_PRIORITIES[i % MOCK_PRIORITIES.length],
  status: MOCK_STATUSES[i % MOCK_STATUSES.length],
  description: 'Mohon segera ditindaklanjuti. Masalah ini sudah berlangsung beberapa hari dan mengganggu kegiatan belajar mengajar.',
  createdAt: new Date(2026, 2, 20 - i, 9, 0).toISOString(),
}));

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
  listContent: {
    paddingHorizontal: 16,
    paddingTop: 8,
    paddingBottom: 24,
    gap: 10,
  },
  card: {
    backgroundColor: '#FFFFFF',
    borderRadius: 14,
    padding: 14,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 4,
    elevation: 2,
  },
  cardTop: {
    flexDirection: 'row',
    alignItems: 'flex-start',
    marginBottom: 10,
  },
  subject: {
    flex: 1,
    fontSize: 15,
    fontWeight: '600',
    color: '#111827',
    lineHeight: 22,
    marginRight: 8,
  },
  chevron: {
    marginTop: 3,
  },
  badgeRow: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 6,
    marginBottom: 8,
  },
  categoryBadge: {
    backgroundColor: '#F3F4F6',
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 6,
  },
  categoryText: {
    fontSize: 12,
    color: '#374151',
    fontWeight: '500',
  },
  statusBadge: {
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 6,
  },
  statusText: {
    fontSize: 12,
    fontWeight: '600',
  },
  priorityBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 6,
    backgroundColor: '#F9FAFB',
  },
  priorityDot: {
    width: 6,
    height: 6,
    borderRadius: 3,
  },
  priorityText: {
    fontSize: 12,
    fontWeight: '600',
  },
  description: {
    fontSize: 13,
    color: '#6B7280',
    lineHeight: 19,
    marginBottom: 8,
    paddingTop: 4,
    borderTopWidth: 1,
    borderTopColor: '#F3F4F6',
  },
  timeText: {
    fontSize: 12,
    color: '#9CA3AF',
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
  },
});
