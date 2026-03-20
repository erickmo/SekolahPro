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
import { get, patch } from '../../src/lib/api';

// ─── Types ────────────────────────────────────────────────────────────────────

type NotificationType =
  | 'PAYMENT'
  | 'ACADEMIC'
  | 'ATTENDANCE'
  | 'HEALTH'
  | 'ANNOUNCEMENT'
  | 'SYSTEM';

interface AppNotification {
  id: string;
  title: string;
  body: string;
  type: NotificationType;
  isRead: boolean;
  createdAt: string;
}

const TYPE_CONFIG: Record<
  NotificationType,
  { label: string; icon: React.ComponentProps<typeof Ionicons>['name']; color: string; bg: string }
> = {
  PAYMENT: { label: 'Pembayaran', icon: 'card-outline', color: '#1D4ED8', bg: '#DBEAFE' },
  ACADEMIC: { label: 'Akademik', icon: 'school-outline', color: '#15803D', bg: '#DCFCE7' },
  ATTENDANCE: { label: 'Kehadiran', icon: 'calendar-outline', color: '#D97706', bg: '#FEF9C3' },
  HEALTH: { label: 'Kesehatan', icon: 'heart-outline', color: '#DC2626', bg: '#FEE2E2' },
  ANNOUNCEMENT: { label: 'Pengumuman', icon: 'megaphone-outline', color: '#7C3AED', bg: '#EDE9FE' },
  SYSTEM: { label: 'Sistem', icon: 'settings-outline', color: '#6B7280', bg: '#F3F4F6' },
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
  return `${Math.floor(diffDays / 30)} bulan lalu`;
}

// ─── Notification Card ────────────────────────────────────────────────────────

function NotificationCard({
  item,
  onRead,
}: {
  item: AppNotification;
  onRead: (id: string) => void;
}) {
  const cfg = TYPE_CONFIG[item.type] ?? TYPE_CONFIG.SYSTEM;

  return (
    <TouchableOpacity
      style={[styles.card, !item.isRead && styles.cardUnread]}
      activeOpacity={0.82}
      onPress={() => {
        if (!item.isRead) onRead(item.id);
      }}
    >
      {/* Unread indicator dot */}
      {!item.isRead && <View style={styles.unreadDot} />}

      <View style={[styles.iconWrapper, { backgroundColor: cfg.bg }]}>
        <Ionicons name={cfg.icon} size={20} color={cfg.color} />
      </View>

      <View style={styles.content}>
        <View style={styles.contentTop}>
          <Text
            style={[styles.title, !item.isRead && styles.titleUnread]}
            numberOfLines={1}
          >
            {item.title}
          </Text>
          <View style={[styles.typeBadge, { backgroundColor: cfg.bg }]}>
            <Text style={[styles.typeLabel, { color: cfg.color }]}>{cfg.label}</Text>
          </View>
        </View>
        <Text style={styles.body} numberOfLines={2}>
          {item.body}
        </Text>
        <Text style={styles.time}>{relativeTime(item.createdAt)}</Text>
      </View>
    </TouchableOpacity>
  );
}

// ─── Main Screen ──────────────────────────────────────────────────────────────

export default function NotificationsScreen() {
  const [notifications, setNotifications] = useState<AppNotification[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const [unreadCount, setUnreadCount] = useState(0);

  const fetchData = useCallback(async (pageNum: number, isLoadMore = false) => {
    if (!isLoadMore) setIsLoading(pageNum === 1);
    else setIsLoadingMore(true);

    try {
      const response = await get<{
        data: AppNotification[];
        meta: { hasNextPage: boolean; unreadCount: number };
      }>('/notifications', { page: pageNum, limit: PAGE_SIZE });

      const newItems = response.data ?? [];
      setHasMore(response.meta?.hasNextPage ?? false);
      if (pageNum === 1) setUnreadCount(response.meta?.unreadCount ?? 0);

      if (isLoadMore) {
        setNotifications((prev) => [...prev, ...newItems]);
      } else {
        setNotifications(newItems);
      }
    } catch {
      if (!isLoadMore) {
        setNotifications(MOCK_NOTIFICATIONS);
        setUnreadCount(MOCK_NOTIFICATIONS.filter((n) => !n.isRead).length);
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

  const handleRead = useCallback(async (id: string) => {
    // Optimistic update
    setNotifications((prev) =>
      prev.map((n) => (n.id === id ? { ...n, isRead: true } : n)),
    );
    setUnreadCount((prev) => Math.max(0, prev - 1));

    try {
      await patch(`/notifications/${id}/read`, {});
    } catch {
      // Revert on error
      setNotifications((prev) =>
        prev.map((n) => (n.id === id ? { ...n, isRead: false } : n)),
      );
      setUnreadCount((prev) => prev + 1);
    }
  }, []);

  const renderFooter = () =>
    isLoadingMore ? (
      <View style={styles.loadMoreContainer}>
        <ActivityIndicator size="small" color="#4F46E5" />
      </View>
    ) : null;

  const renderEmpty = () =>
    isLoading ? null : (
      <View style={styles.emptyContainer}>
        <Ionicons name="notifications-off-outline" size={48} color="#D1D5DB" />
        <Text style={styles.emptyTitle}>Tidak ada notifikasi</Text>
        <Text style={styles.emptySubtitle}>Semua notifikasi akan muncul di sini.</Text>
      </View>
    );

  return (
    <SafeAreaView style={styles.safeArea} edges={['top']}>
      {/* Header */}
      <View style={styles.header}>
        <View>
          <Text style={styles.headerTitle}>Notifikasi</Text>
          {unreadCount > 0 && (
            <Text style={styles.headerSubtitle}>
              {unreadCount} belum dibaca
            </Text>
          )}
        </View>
        <View style={styles.headerRight}>
          <View style={styles.headerIcon}>
            <Ionicons name="notifications" size={22} color="#4F46E5" />
            {unreadCount > 0 && (
              <View style={styles.badgeCount}>
                <Text style={styles.badgeCountText}>
                  {unreadCount > 99 ? '99+' : unreadCount}
                </Text>
              </View>
            )}
          </View>
        </View>
      </View>

      {/* List */}
      {isLoading ? (
        <View style={styles.loadingCenter}>
          <ActivityIndicator size="large" color="#4F46E5" />
          <Text style={styles.loadingText}>Memuat notifikasi...</Text>
        </View>
      ) : (
        <FlatList
          data={notifications}
          keyExtractor={(item) => item.id}
          renderItem={({ item }) => (
            <NotificationCard item={item} onRead={handleRead} />
          )}
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

const MOCK_TYPES: NotificationType[] = [
  'PAYMENT', 'ACADEMIC', 'ATTENDANCE', 'HEALTH', 'ANNOUNCEMENT', 'SYSTEM',
];

const MOCK_TITLES = [
  'Tagihan SPP Maret 2026 jatuh tempo', 'Nilai ujian tengah semester telah keluar',
  'Siswa tidak hadir hari ini', 'Kunjungan UKS: Ahmad Fauzi',
  'Pengumuman libur nasional', 'Pembaruan sistem terjadwal',
  'Pembayaran SPP berhasil diterima', 'Jadwal remedial telah diumumkan',
];

const MOCK_BODIES = [
  'Harap segera melakukan pembayaran sebelum tanggal 25 Maret 2026.',
  'Silakan cek portal untuk melihat detail nilai siswa Anda.',
  'Orang tua diharapkan mengkonfirmasi ketidakhadiran.',
  'Siswa mendapat penanganan untuk keluhan sakit kepala ringan.',
  'Sekolah libur pada 25 Maret 2026 memperingati Hari Raya Nyepi.',
  'Sistem akan offline pada 22 Maret 2026 pukul 00.00–02.00.',
  'Terima kasih atas pembayaran Anda. Kwitansi tersedia di portal.',
  'Remedial Matematika dilaksanakan Selasa, 24 Maret 2026.',
];

const MOCK_NOTIFICATIONS: AppNotification[] = Array.from({ length: 12 }, (_, i) => ({
  id: `notif-${i + 1}`,
  title: MOCK_TITLES[i % MOCK_TITLES.length],
  body: MOCK_BODIES[i % MOCK_BODIES.length],
  type: MOCK_TYPES[i % MOCK_TYPES.length],
  isRead: i % 3 !== 0,
  createdAt: new Date(2026, 2, 20 - i, 10 - i, 0).toISOString(),
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
    color: '#4F46E5',
    marginTop: 2,
    fontWeight: '500',
  },
  headerRight: {
    position: 'relative',
  },
  headerIcon: {
    width: 44,
    height: 44,
    borderRadius: 12,
    backgroundColor: '#EEF2FF',
    alignItems: 'center',
    justifyContent: 'center',
  },
  badgeCount: {
    position: 'absolute',
    top: -4,
    right: -4,
    backgroundColor: '#DC2626',
    borderRadius: 10,
    minWidth: 18,
    height: 18,
    alignItems: 'center',
    justifyContent: 'center',
    paddingHorizontal: 4,
    borderWidth: 1.5,
    borderColor: '#FFFFFF',
  },
  badgeCountText: {
    fontSize: 10,
    fontWeight: '700',
    color: '#FFFFFF',
  },
  listContent: {
    paddingHorizontal: 16,
    paddingTop: 4,
    paddingBottom: 24,
    gap: 8,
  },
  card: {
    backgroundColor: '#FFFFFF',
    borderRadius: 14,
    padding: 14,
    flexDirection: 'row',
    alignItems: 'flex-start',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 4,
    elevation: 2,
    position: 'relative',
  },
  cardUnread: {
    backgroundColor: '#F0F4FF',
    borderLeftWidth: 3,
    borderLeftColor: '#4F46E5',
  },
  unreadDot: {
    position: 'absolute',
    top: 14,
    left: 6,
    width: 7,
    height: 7,
    borderRadius: 4,
    backgroundColor: '#4F46E5',
  },
  iconWrapper: {
    width: 40,
    height: 40,
    borderRadius: 10,
    alignItems: 'center',
    justifyContent: 'center',
    marginRight: 12,
    flexShrink: 0,
  },
  content: {
    flex: 1,
  },
  contentTop: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    marginBottom: 4,
    gap: 6,
  },
  title: {
    flex: 1,
    fontSize: 14,
    fontWeight: '500',
    color: '#374151',
    lineHeight: 20,
  },
  titleUnread: {
    fontWeight: '700',
    color: '#111827',
  },
  typeBadge: {
    paddingHorizontal: 8,
    paddingVertical: 3,
    borderRadius: 6,
    flexShrink: 0,
  },
  typeLabel: {
    fontSize: 11,
    fontWeight: '600',
  },
  body: {
    fontSize: 13,
    color: '#6B7280',
    lineHeight: 18,
    marginBottom: 6,
  },
  time: {
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
