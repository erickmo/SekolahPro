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
import { useRouter } from 'expo-router';
import { useAuth } from '../../src/lib/auth';
import StatsCard from '../../src/components/StatsCard';
import LoadingSpinner from '../../src/components/LoadingSpinner';
import { get } from '../../src/lib/api';
import type { DashboardStats, Notification } from '../../src/types';

// ─── Helpers ──────────────────────────────────────────────────────────────────

function getGreeting(): string {
  const hour = new Date().getHours();
  if (hour < 12) return 'Selamat Pagi';
  if (hour < 15) return 'Selamat Siang';
  if (hour < 18) return 'Selamat Sore';
  return 'Selamat Malam';
}

function getNotificationIcon(type: Notification['type']): React.ComponentProps<typeof Ionicons>['name'] {
  switch (type) {
    case 'ATTENDANCE': return 'checkmark-circle-outline';
    case 'PAYMENT': return 'card-outline';
    case 'EXAM': return 'document-text-outline';
    case 'ANNOUNCEMENT': return 'megaphone-outline';
    case 'ALERT': return 'warning-outline';
    case 'HEALTH': return 'medical-outline';
    case 'TRANSPORT': return 'bus-outline';
    default: return 'notifications-outline';
  }
}

function getNotificationColor(type: Notification['type']): string {
  switch (type) {
    case 'ATTENDANCE': return '#059669';
    case 'PAYMENT': return '#4F46E5';
    case 'EXAM': return '#7C3AED';
    case 'ANNOUNCEMENT': return '#D97706';
    case 'ALERT': return '#DC2626';
    case 'HEALTH': return '#0891B2';
    case 'TRANSPORT': return '#0284C7';
    default: return '#6B7280';
  }
}

function formatNotificationTime(dateString: string): string {
  const date = new Date(dateString);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);

  if (diffMins < 1) return 'Baru saja';
  if (diffMins < 60) return `${diffMins} menit lalu`;
  if (diffHours < 24) return `${diffHours} jam lalu`;
  if (diffDays < 7) return `${diffDays} hari lalu`;
  return date.toLocaleDateString('id-ID', { day: 'numeric', month: 'short' });
}

// ─── Notification Item ────────────────────────────────────────────────────────

function NotificationItem({ item }: { item: Notification }) {
  const iconName = getNotificationIcon(item.type);
  const iconColor = getNotificationColor(item.type);
  const bgColor = iconColor + '15'; // 15 = ~8% opacity in hex

  return (
    <TouchableOpacity style={styles.notifItem} activeOpacity={0.7}>
      <View style={[styles.notifIconContainer, { backgroundColor: bgColor }]}>
        <Ionicons name={iconName} size={18} color={iconColor} />
      </View>
      <View style={styles.notifContent}>
        <Text style={styles.notifTitle} numberOfLines={1}>
          {item.title}
        </Text>
        <Text style={styles.notifBody} numberOfLines={2}>
          {item.body}
        </Text>
        <Text style={styles.notifTime}>{formatNotificationTime(item.createdAt)}</Text>
      </View>
      {!item.isRead && <View style={styles.unreadDot} />}
    </TouchableOpacity>
  );
}

// ─── Main Screen ──────────────────────────────────────────────────────────────

export default function DashboardScreen() {
  const { user, school } = useAuth();
  const router = useRouter();

  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);

  const fetchData = useCallback(async () => {
    try {
      const [statsData, notifData] = await Promise.all([
        get<DashboardStats>('/dashboard/stats'),
        get<{ items: Notification[] }>('/notifications?limit=5'),
      ]);
      setStats(statsData);
      setNotifications(notifData.items ?? []);
    } catch {
      // Use mock data if API not available yet
      setStats({
        totalStudents: 1248,
        activeStudents: 1205,
        attendanceToday: {
          date: new Date().toISOString(),
          totalStudents: 1205,
          present: 1083,
          absent: 42,
          late: 56,
          excused: 18,
          sick: 6,
          attendanceRate: 89.8,
        },
        pendingPayments: 87,
        overduePayments: 23,
        collectionRate: 82.4,
        riskStudents: 14,
      });
      setNotifications([
        {
          id: '1',
          type: 'ATTENDANCE',
          title: 'Rekap Kehadiran Hari Ini',
          body: '1.083 dari 1.205 siswa hadir (89,8%). Ada 42 siswa tidak hadir.',
          isRead: false,
          createdAt: new Date(Date.now() - 3600000).toISOString(),
        },
        {
          id: '2',
          type: 'PAYMENT',
          title: 'Tagihan SPP Jatuh Tempo',
          body: '23 siswa memiliki tagihan SPP yang telah melewati jatuh tempo bulan ini.',
          isRead: false,
          createdAt: new Date(Date.now() - 7200000).toISOString(),
        },
        {
          id: '3',
          type: 'ALERT',
          title: 'Early Warning: 14 Siswa Berisiko',
          body: 'Sistem mendeteksi 14 siswa dengan skor risiko tinggi. Segera lakukan intervensi.',
          isRead: true,
          createdAt: new Date(Date.now() - 86400000).toISOString(),
        },
        {
          id: '4',
          type: 'ANNOUNCEMENT',
          title: 'Rapat Komite Sekolah',
          body: 'Rapat komite sekolah akan diadakan pada Sabtu, 22 Maret 2026 pukul 08.00 WIB.',
          isRead: true,
          createdAt: new Date(Date.now() - 172800000).toISOString(),
        },
        {
          id: '5',
          type: 'EXAM',
          title: 'Ujian Tengah Semester Dimulai',
          body: 'UTS Semester Genap 2025/2026 dimulai hari ini hingga 28 Maret 2026.',
          isRead: true,
          createdAt: new Date(Date.now() - 259200000).toISOString(),
        },
      ]);
    } finally {
      setIsLoading(false);
      setIsRefreshing(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const onRefresh = useCallback(() => {
    setIsRefreshing(true);
    fetchData();
  }, [fetchData]);

  if (isLoading) {
    return <LoadingSpinner fullScreen label="Memuat data..." />;
  }

  const attendanceRate = stats?.attendanceToday.attendanceRate ?? 0;

  return (
    <SafeAreaView style={styles.safeArea} edges={['top']}>
      <ScrollView
        style={styles.scroll}
        contentContainerStyle={styles.scrollContent}
        showsVerticalScrollIndicator={false}
        refreshControl={
          <RefreshControl
            refreshing={isRefreshing}
            onRefresh={onRefresh}
            tintColor="#4F46E5"
            colors={['#4F46E5']}
          />
        }
      >
        {/* Header */}
        <View style={styles.pageHeader}>
          <View style={styles.greetingContainer}>
            <Text style={styles.greeting}>
              {getGreeting()}, {user?.name?.split(' ')[0] ?? 'Pengguna'} 👋
            </Text>
            <Text style={styles.schoolName} numberOfLines={1}>
              {school?.name ?? 'Sekolah'}
            </Text>
          </View>
          <TouchableOpacity style={styles.notifButton} activeOpacity={0.7}>
            <Ionicons name="notifications-outline" size={22} color="#374151" />
            {notifications.some((n) => !n.isRead) && (
              <View style={styles.notifBadge} />
            )}
          </TouchableOpacity>
        </View>

        {/* Attendance Banner */}
        <View style={styles.attendanceBanner}>
          <View style={styles.attendanceBannerLeft}>
            <Text style={styles.attendanceBannerTitle}>Kehadiran Hari Ini</Text>
            <Text style={styles.attendanceBannerRate}>
              {attendanceRate.toFixed(1)}%
            </Text>
            <Text style={styles.attendanceBannerSub}>
              {stats?.attendanceToday.present.toLocaleString('id-ID')} dari{' '}
              {stats?.attendanceToday.totalStudents.toLocaleString('id-ID')} siswa hadir
            </Text>
          </View>
          <View style={styles.attendanceBannerRight}>
            <View style={styles.attendanceRing}>
              <Text style={styles.attendanceRingText}>
                {Math.round(attendanceRate)}%
              </Text>
            </View>
          </View>
        </View>

        {/* Stats Grid */}
        <Text style={styles.sectionTitle}>Ringkasan</Text>
        <View style={styles.statsGrid}>
          <View style={styles.statsRow}>
            <StatsCard
              title="Total Siswa"
              value={stats?.totalStudents.toLocaleString('id-ID') ?? '0'}
              subtitle="Siswa aktif"
              icon="people"
              iconColor="#4F46E5"
              iconBgColor="#EEF2FF"
              style={styles.statsCardHalf}
              onPress={() => router.push('/(tabs)/students')}
            />
            <View style={styles.statsGap} />
            <StatsCard
              title="Tidak Hadir"
              value={stats?.attendanceToday.absent ?? 0}
              subtitle="Hari ini"
              icon="close-circle"
              iconColor="#EF4444"
              iconBgColor="#FEE2E2"
              style={styles.statsCardHalf}
            />
          </View>

          <View style={styles.statsRow}>
            <StatsCard
              title="Tagihan Pending"
              value={stats?.pendingPayments ?? 0}
              subtitle="Belum terbayar"
              icon="card"
              iconColor="#D97706"
              iconBgColor="#FEF3C7"
              style={styles.statsCardHalf}
              onPress={() => router.push('/(tabs)/payments')}
            />
            <View style={styles.statsGap} />
            <StatsCard
              title="Siswa Berisiko"
              value={stats?.riskStudents ?? 0}
              subtitle="Perlu intervensi"
              icon="warning"
              iconColor="#DC2626"
              iconBgColor="#FEE2E2"
              style={styles.statsCardHalf}
            />
          </View>
        </View>

        {/* Quick Actions */}
        <Text style={styles.sectionTitle}>Aksi Cepat</Text>
        <View style={styles.quickActions}>
          {QUICK_ACTIONS.map((action) => (
            <TouchableOpacity
              key={action.label}
              style={styles.quickActionItem}
              activeOpacity={0.7}
              onPress={() => router.push(action.route as never)}
            >
              <View
                style={[
                  styles.quickActionIcon,
                  { backgroundColor: action.bgColor },
                ]}
              >
                <Ionicons name={action.icon} size={22} color={action.color} />
              </View>
              <Text style={styles.quickActionLabel}>{action.label}</Text>
            </TouchableOpacity>
          ))}
        </View>

        {/* Recent Notifications */}
        <View style={styles.notifSection}>
          <View style={styles.sectionHeader}>
            <Text style={styles.sectionTitle}>Notifikasi Terbaru</Text>
            <TouchableOpacity activeOpacity={0.7}>
              <Text style={styles.seeAll}>Lihat Semua</Text>
            </TouchableOpacity>
          </View>

          {notifications.length > 0 ? (
            <View style={styles.notifList}>
              {notifications.map((item) => (
                <NotificationItem key={item.id} item={item} />
              ))}
            </View>
          ) : (
            <View style={styles.notifEmpty}>
              <Ionicons name="notifications-off-outline" size={32} color="#D1D5DB" />
              <Text style={styles.notifEmptyText}>Belum ada notifikasi</Text>
            </View>
          )}
        </View>

        <View style={styles.bottomPadding} />
      </ScrollView>
    </SafeAreaView>
  );
}

// ─── Quick Actions Config ─────────────────────────────────────────────────────

const QUICK_ACTIONS = [
  {
    label: 'Data Siswa',
    icon: 'people' as const,
    color: '#4F46E5',
    bgColor: '#EEF2FF',
    route: '/(tabs)/students',
  },
  {
    label: 'SPP',
    icon: 'card' as const,
    color: '#059669',
    bgColor: '#D1FAE5',
    route: '/(tabs)/payments',
  },
  {
    label: 'Absensi',
    icon: 'checkmark-circle' as const,
    color: '#0284C7',
    bgColor: '#E0F2FE',
    route: '/(tabs)',
  },
  {
    label: 'Laporan',
    icon: 'bar-chart' as const,
    color: '#7C3AED',
    bgColor: '#EDE9FE',
    route: '/(tabs)',
  },
];

// ─── Styles ───────────────────────────────────────────────────────────────────

const styles = StyleSheet.create({
  safeArea: {
    flex: 1,
    backgroundColor: '#F9FAFB',
  },
  scroll: {
    flex: 1,
  },
  scrollContent: {
    paddingHorizontal: 16,
  },
  pageHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingTop: 20,
    paddingBottom: 16,
  },
  greetingContainer: {
    flex: 1,
    marginRight: 16,
  },
  greeting: {
    fontSize: 22,
    fontWeight: '700',
    color: '#111827',
    marginBottom: 2,
  },
  schoolName: {
    fontSize: 14,
    color: '#6B7280',
  },
  notifButton: {
    width: 44,
    height: 44,
    borderRadius: 12,
    backgroundColor: '#FFFFFF',
    alignItems: 'center',
    justifyContent: 'center',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.06,
    shadowRadius: 4,
    elevation: 2,
    position: 'relative',
  },
  notifBadge: {
    position: 'absolute',
    top: 10,
    right: 10,
    width: 8,
    height: 8,
    borderRadius: 4,
    backgroundColor: '#EF4444',
    borderWidth: 1.5,
    borderColor: '#FFFFFF',
  },
  attendanceBanner: {
    backgroundColor: '#4F46E5',
    borderRadius: 20,
    padding: 20,
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 24,
    shadowColor: '#4F46E5',
    shadowOffset: { width: 0, height: 6 },
    shadowOpacity: 0.3,
    shadowRadius: 12,
    elevation: 6,
  },
  attendanceBannerLeft: {
    flex: 1,
  },
  attendanceBannerTitle: {
    fontSize: 13,
    color: '#C7D2FE',
    fontWeight: '500',
    marginBottom: 4,
  },
  attendanceBannerRate: {
    fontSize: 36,
    fontWeight: '800',
    color: '#FFFFFF',
    letterSpacing: -1,
    marginBottom: 4,
  },
  attendanceBannerSub: {
    fontSize: 12,
    color: '#A5B4FC',
    lineHeight: 16,
  },
  attendanceBannerRight: {
    marginLeft: 16,
  },
  attendanceRing: {
    width: 72,
    height: 72,
    borderRadius: 36,
    borderWidth: 4,
    borderColor: '#818CF8',
    backgroundColor: '#6366F1',
    alignItems: 'center',
    justifyContent: 'center',
  },
  attendanceRingText: {
    fontSize: 16,
    fontWeight: '700',
    color: '#FFFFFF',
  },
  sectionTitle: {
    fontSize: 17,
    fontWeight: '700',
    color: '#111827',
    marginBottom: 12,
  },
  sectionHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 12,
  },
  seeAll: {
    fontSize: 14,
    color: '#4F46E5',
    fontWeight: '500',
  },
  statsGrid: {
    marginBottom: 24,
    gap: 10,
  },
  statsRow: {
    flexDirection: 'row',
  },
  statsCardHalf: {
    flex: 1,
  },
  statsGap: {
    width: 10,
  },
  quickActions: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginBottom: 24,
  },
  quickActionItem: {
    alignItems: 'center',
    flex: 1,
  },
  quickActionIcon: {
    width: 56,
    height: 56,
    borderRadius: 16,
    alignItems: 'center',
    justifyContent: 'center',
    marginBottom: 8,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.06,
    shadowRadius: 6,
    elevation: 2,
  },
  quickActionLabel: {
    fontSize: 12,
    fontWeight: '500',
    color: '#374151',
    textAlign: 'center',
  },
  notifSection: {
    marginBottom: 8,
  },
  notifList: {
    backgroundColor: '#FFFFFF',
    borderRadius: 16,
    overflow: 'hidden',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 4,
    elevation: 2,
  },
  notifItem: {
    flexDirection: 'row',
    alignItems: 'flex-start',
    padding: 14,
    borderBottomWidth: 1,
    borderBottomColor: '#F9FAFB',
  },
  notifIconContainer: {
    width: 40,
    height: 40,
    borderRadius: 12,
    alignItems: 'center',
    justifyContent: 'center',
    marginRight: 12,
    marginTop: 1,
  },
  notifContent: {
    flex: 1,
  },
  notifTitle: {
    fontSize: 14,
    fontWeight: '600',
    color: '#111827',
    marginBottom: 2,
  },
  notifBody: {
    fontSize: 13,
    color: '#6B7280',
    lineHeight: 18,
    marginBottom: 4,
  },
  notifTime: {
    fontSize: 11,
    color: '#9CA3AF',
  },
  unreadDot: {
    width: 8,
    height: 8,
    borderRadius: 4,
    backgroundColor: '#4F46E5',
    marginLeft: 8,
    marginTop: 6,
  },
  notifEmpty: {
    backgroundColor: '#FFFFFF',
    borderRadius: 16,
    padding: 32,
    alignItems: 'center',
    gap: 8,
  },
  notifEmptyText: {
    fontSize: 14,
    color: '#9CA3AF',
  },
  bottomPadding: {
    height: 20,
  },
});
