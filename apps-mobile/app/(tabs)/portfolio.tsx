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
import { useRouter } from 'expo-router';
import { get } from '../../src/lib/api';

// ─── Types ────────────────────────────────────────────────────────────────────

type PortfolioItemType = 'DOCUMENT' | 'IMAGE' | 'VIDEO' | 'LINK' | 'CERTIFICATE';

interface Portfolio {
  id: string;
  studentName: string;
  itemCount: number;
  itemTypes: PortfolioItemType[];
  lastUpdated: string;
}

const TYPE_CONFIG: Record<PortfolioItemType, { label: string; icon: React.ComponentProps<typeof Ionicons>['name']; color: string }> = {
  DOCUMENT: { label: 'Dokumen', icon: 'document-text-outline', color: '#2563EB' },
  IMAGE: { label: 'Foto', icon: 'image-outline', color: '#059669' },
  VIDEO: { label: 'Video', icon: 'videocam-outline', color: '#DC2626' },
  LINK: { label: 'Tautan', icon: 'link-outline', color: '#D97706' },
  CERTIFICATE: { label: 'Sertifikat', icon: 'ribbon-outline', color: '#7C3AED' },
};

const PAGE_SIZE = 20;

// ─── Helpers ──────────────────────────────────────────────────────────────────

function formatLastUpdated(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  });
}

function getInitials(name: string): string {
  return name
    .split(' ')
    .slice(0, 2)
    .map((n) => n[0])
    .join('')
    .toUpperCase();
}

// ─── Portfolio Card ───────────────────────────────────────────────────────────

function PortfolioCard({ item, onPress }: { item: Portfolio; onPress: (id: string) => void }) {
  return (
    <TouchableOpacity
      style={styles.card}
      activeOpacity={0.85}
      onPress={() => onPress(item.id)}
    >
      <View style={styles.cardLeft}>
        <View style={styles.avatar}>
          <Text style={styles.avatarText}>{getInitials(item.studentName)}</Text>
        </View>
      </View>
      <View style={styles.cardBody}>
        <View style={styles.cardTop}>
          <Text style={styles.studentName} numberOfLines={1}>
            {item.studentName}
          </Text>
          <Ionicons name="chevron-forward" size={18} color="#9CA3AF" />
        </View>
        <View style={styles.itemTypesRow}>
          {item.itemTypes.slice(0, 4).map((type) => {
            const cfg = TYPE_CONFIG[type];
            return (
              <View key={type} style={[styles.typeBadge, { borderColor: cfg.color + '40', backgroundColor: cfg.color + '12' }]}>
                <Ionicons name={cfg.icon} size={11} color={cfg.color} />
                <Text style={[styles.typeLabel, { color: cfg.color }]}>{cfg.label}</Text>
              </View>
            );
          })}
        </View>
        <View style={styles.cardFooter}>
          <View style={styles.itemCount}>
            <Ionicons name="folder-open-outline" size={13} color="#6B7280" />
            <Text style={styles.itemCountText}>{item.itemCount} item</Text>
          </View>
          <Text style={styles.lastUpdated}>
            Diperbarui {formatLastUpdated(item.lastUpdated)}
          </Text>
        </View>
      </View>
    </TouchableOpacity>
  );
}

// ─── Main Screen ──────────────────────────────────────────────────────────────

export default function PortfolioScreen() {
  const router = useRouter();
  const [portfolios, setPortfolios] = useState<Portfolio[]>([]);
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
        data: Portfolio[];
        meta: { total: number; hasNextPage: boolean };
      }>('/portfolio', { page: pageNum, limit: PAGE_SIZE });

      const newItems = response.data ?? [];
      setTotalCount(response.meta?.total ?? 0);
      setHasMore(response.meta?.hasNextPage ?? false);

      if (isLoadMore) {
        setPortfolios((prev) => [...prev, ...newItems]);
      } else {
        setPortfolios(newItems);
      }
    } catch {
      if (!isLoadMore) {
        setPortfolios(MOCK_PORTFOLIOS);
        setTotalCount(MOCK_PORTFOLIOS.length);
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

  const onPortfolioPress = useCallback(
    (id: string) => {
      router.push(`/portfolio/${id}` as any);
    },
    [router],
  );

  const renderFooter = () =>
    isLoadingMore ? (
      <View style={styles.loadMoreContainer}>
        <ActivityIndicator size="small" color="#4F46E5" />
      </View>
    ) : null;

  const renderEmpty = () =>
    isLoading ? null : (
      <View style={styles.emptyContainer}>
        <Ionicons name="folder-open-outline" size={48} color="#D1D5DB" />
        <Text style={styles.emptyTitle}>Belum ada portofolio</Text>
        <Text style={styles.emptySubtitle}>
          Portofolio digital siswa akan muncul di sini.
        </Text>
      </View>
    );

  return (
    <SafeAreaView style={styles.safeArea} edges={['top']}>
      {/* Header */}
      <View style={styles.header}>
        <View>
          <Text style={styles.headerTitle}>Portofolio Siswa</Text>
          <Text style={styles.headerSubtitle}>
            {totalCount.toLocaleString('id-ID')} portofolio
          </Text>
        </View>
        <View style={styles.headerIcon}>
          <Ionicons name="folder" size={22} color="#4F46E5" />
        </View>
      </View>

      {/* List */}
      {isLoading ? (
        <View style={styles.loadingCenter}>
          <ActivityIndicator size="large" color="#4F46E5" />
          <Text style={styles.loadingText}>Memuat portofolio...</Text>
        </View>
      ) : (
        <FlatList
          data={portfolios}
          keyExtractor={(item) => item.id}
          renderItem={({ item }) => (
            <PortfolioCard item={item} onPress={onPortfolioPress} />
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

const ALL_TYPES: PortfolioItemType[] = ['DOCUMENT', 'IMAGE', 'VIDEO', 'LINK', 'CERTIFICATE'];
const MOCK_NAMES = [
  'Ahmad Fauzi', 'Budi Santoso', 'Citra Dewi', 'Dini Rahayu', 'Eko Prasetyo',
  'Farah Azzahra', 'Gilang Ramadhan', 'Hana Putri', 'Ivan Setiawan', 'Julia Maharani',
];

const MOCK_PORTFOLIOS: Portfolio[] = Array.from({ length: 10 }, (_, i) => ({
  id: `portfolio-${i + 1}`,
  studentName: MOCK_NAMES[i % MOCK_NAMES.length],
  itemCount: 3 + (i % 10),
  itemTypes: ALL_TYPES.slice(0, 2 + (i % 4)) as PortfolioItemType[],
  lastUpdated: new Date(2026, 2, 15 - i).toISOString(),
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
    flexDirection: 'row',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 4,
    elevation: 2,
  },
  cardLeft: {
    marginRight: 12,
  },
  avatar: {
    width: 44,
    height: 44,
    borderRadius: 22,
    backgroundColor: '#4F46E5',
    alignItems: 'center',
    justifyContent: 'center',
  },
  avatarText: {
    fontSize: 15,
    fontWeight: '700',
    color: '#FFFFFF',
  },
  cardBody: {
    flex: 1,
  },
  cardTop: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    marginBottom: 8,
  },
  studentName: {
    flex: 1,
    fontSize: 15,
    fontWeight: '700',
    color: '#111827',
    marginRight: 4,
  },
  itemTypesRow: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 5,
    marginBottom: 10,
  },
  typeBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 3,
    paddingHorizontal: 7,
    paddingVertical: 3,
    borderRadius: 6,
    borderWidth: 1,
  },
  typeLabel: {
    fontSize: 11,
    fontWeight: '500',
  },
  cardFooter: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  itemCount: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
  },
  itemCountText: {
    fontSize: 12,
    color: '#6B7280',
    fontWeight: '500',
  },
  lastUpdated: {
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
