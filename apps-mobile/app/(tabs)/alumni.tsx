import React, { useCallback, useEffect, useRef, useState } from 'react';
import {
  ActivityIndicator,
  FlatList,
  RefreshControl,
  StyleSheet,
  Text,
  TextInput,
  TouchableOpacity,
  View,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { Ionicons } from '@expo/vector-icons';
import { get } from '../../src/lib/api';

// ─── Types ────────────────────────────────────────────────────────────────────

interface AlumniProfile {
  id: string;
  name: string;
  graduationYear: number;
  currentJob: string;
  currentCity: string;
}

const PAGE_SIZE = 20;

// ─── Helpers ──────────────────────────────────────────────────────────────────

function getInitials(name: string): string {
  return name
    .split(' ')
    .slice(0, 2)
    .map((n) => n[0])
    .join('')
    .toUpperCase();
}

const AVATAR_COLORS = [
  '#4F46E5', '#0891B2', '#059669', '#D97706', '#DC2626',
  '#7C3AED', '#DB2777', '#2563EB', '#16A34A', '#EA580C',
];

function avatarColor(id: string): string {
  let hash = 0;
  for (let i = 0; i < id.length; i++) hash = id.charCodeAt(i) + ((hash << 5) - hash);
  return AVATAR_COLORS[Math.abs(hash) % AVATAR_COLORS.length];
}

// ─── Alumni Card ──────────────────────────────────────────────────────────────

function AlumniCard({ item }: { item: AlumniProfile }) {
  const bg = avatarColor(item.id);

  return (
    <View style={styles.card}>
      {/* Gradient-like header strip */}
      <View style={[styles.cardStrip, { backgroundColor: bg + '20' }]}>
        <View style={[styles.avatar, { backgroundColor: bg }]}>
          <Text style={styles.avatarText}>{getInitials(item.name)}</Text>
        </View>
      </View>
      <View style={styles.cardBody}>
        <Text style={styles.name}>{item.name}</Text>
        <View style={styles.yearBadge}>
          <Ionicons name="school-outline" size={12} color="#4F46E5" />
          <Text style={styles.yearText}>Lulus {item.graduationYear}</Text>
        </View>
        <View style={styles.detailRow}>
          <Ionicons name="briefcase-outline" size={14} color="#9CA3AF" />
          <Text style={styles.detailText} numberOfLines={1}>
            {item.currentJob || 'Tidak diketahui'}
          </Text>
        </View>
        <View style={styles.detailRow}>
          <Ionicons name="location-outline" size={14} color="#9CA3AF" />
          <Text style={styles.detailText} numberOfLines={1}>
            {item.currentCity || 'Tidak diketahui'}
          </Text>
        </View>
      </View>
    </View>
  );
}

// ─── Main Screen ──────────────────────────────────────────────────────────────

export default function AlumniScreen() {
  const [alumni, setAlumni] = useState<AlumniProfile[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const [totalCount, setTotalCount] = useState(0);
  const searchTimeout = useRef<ReturnType<typeof setTimeout> | null>(null);

  const fetchData = useCallback(
    async (pageNum: number, query: string, isLoadMore = false) => {
      if (!isLoadMore) setIsLoading(pageNum === 1);
      else setIsLoadingMore(true);

      try {
        const params: Record<string, unknown> = { page: pageNum, limit: PAGE_SIZE };
        if (query.trim()) params.search = query.trim();

        const response = await get<{
          data: AlumniProfile[];
          meta: { total: number; hasNextPage: boolean };
        }>('/alumni/profiles', params);

        const newItems = response.data ?? [];
        setTotalCount(response.meta?.total ?? 0);
        setHasMore(response.meta?.hasNextPage ?? false);

        if (isLoadMore) {
          setAlumni((prev) => [...prev, ...newItems]);
        } else {
          setAlumni(newItems);
        }
      } catch {
        if (!isLoadMore) {
          setAlumni(MOCK_ALUMNI);
          setTotalCount(MOCK_ALUMNI.length);
          setHasMore(false);
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
    fetchData(1, '');
  }, [fetchData]);

  const handleSearch = (text: string) => {
    setSearchQuery(text);
    if (searchTimeout.current) clearTimeout(searchTimeout.current);
    searchTimeout.current = setTimeout(() => {
      setPage(1);
      fetchData(1, text);
    }, 400);
  };

  const onRefresh = useCallback(() => {
    setIsRefreshing(true);
    setPage(1);
    fetchData(1, searchQuery);
  }, [fetchData, searchQuery]);

  const onLoadMore = useCallback(() => {
    if (!hasMore || isLoadingMore || isLoading) return;
    const next = page + 1;
    setPage(next);
    fetchData(next, searchQuery, true);
  }, [hasMore, isLoadingMore, isLoading, page, fetchData, searchQuery]);

  const renderFooter = () =>
    isLoadingMore ? (
      <View style={styles.loadMoreContainer}>
        <ActivityIndicator size="small" color="#4F46E5" />
      </View>
    ) : null;

  const renderEmpty = () =>
    isLoading ? null : (
      <View style={styles.emptyContainer}>
        <Ionicons name="school-outline" size={48} color="#D1D5DB" />
        <Text style={styles.emptyTitle}>
          {searchQuery ? `Tidak ada hasil untuk "${searchQuery}"` : 'Belum ada data alumni'}
        </Text>
      </View>
    );

  return (
    <SafeAreaView style={styles.safeArea} edges={['top']}>
      {/* Header */}
      <View style={styles.header}>
        <View>
          <Text style={styles.headerTitle}>Alumni</Text>
          <Text style={styles.headerSubtitle}>
            {totalCount.toLocaleString('id-ID')} alumni terdaftar
          </Text>
        </View>
        <View style={styles.headerIcon}>
          <Ionicons name="school" size={22} color="#4F46E5" />
        </View>
      </View>

      {/* Search */}
      <View style={styles.searchContainer}>
        <View style={styles.searchBar}>
          <Ionicons name="search-outline" size={18} color="#9CA3AF" style={styles.searchIcon} />
          <TextInput
            style={styles.searchInput}
            placeholder="Cari nama atau tahun lulus..."
            placeholderTextColor="#D1D5DB"
            value={searchQuery}
            onChangeText={handleSearch}
            returnKeyType="search"
            autoCorrect={false}
            autoCapitalize="none"
          />
          {searchQuery.length > 0 && (
            <TouchableOpacity onPress={() => handleSearch('')} hitSlop={8}>
              <Ionicons name="close-circle" size={18} color="#9CA3AF" />
            </TouchableOpacity>
          )}
        </View>
      </View>

      {/* List */}
      {isLoading ? (
        <View style={styles.loadingCenter}>
          <ActivityIndicator size="large" color="#4F46E5" />
          <Text style={styles.loadingText}>Memuat data alumni...</Text>
        </View>
      ) : (
        <FlatList
          data={alumni}
          keyExtractor={(item) => item.id}
          renderItem={({ item }) => <AlumniCard item={item} />}
          numColumns={2}
          columnWrapperStyle={styles.columnWrapper}
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

const MOCK_JOBS = [
  'Software Engineer', 'Dokter Umum', 'Guru SD', 'Akuntan', 'Desainer Grafis',
  'Wirausahawan', 'Mahasiswa', 'Pegawai BUMN', 'Jurnalis', 'Arsitek',
];

const MOCK_CITIES = [
  'Jakarta', 'Bandung', 'Surabaya', 'Yogyakarta', 'Medan',
  'Semarang', 'Makassar', 'Palembang', 'Denpasar', 'Malang',
];

const MOCK_NAMES = [
  'Ahmad Fauzi', 'Budi Santoso', 'Citra Dewi', 'Dini Rahayu',
  'Eko Prasetyo', 'Farah Azzahra', 'Gilang Ramadhan', 'Hana Putri',
  'Ivan Setiawan', 'Julia Maharani', 'Kevin Wijaya', 'Laila Nurdin',
];

const MOCK_ALUMNI: AlumniProfile[] = Array.from({ length: 12 }, (_, i) => ({
  id: `alumni-${i + 1}`,
  name: MOCK_NAMES[i % MOCK_NAMES.length],
  graduationYear: 2020 - (i % 6),
  currentJob: MOCK_JOBS[i % MOCK_JOBS.length],
  currentCity: MOCK_CITIES[i % MOCK_CITIES.length],
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
  searchContainer: {
    paddingHorizontal: 16,
    paddingBottom: 12,
  },
  searchBar: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#FFFFFF',
    borderRadius: 12,
    paddingHorizontal: 14,
    height: 46,
    borderWidth: 1,
    borderColor: '#E5E7EB',
  },
  searchIcon: {
    marginRight: 10,
  },
  searchInput: {
    flex: 1,
    fontSize: 15,
    color: '#111827',
    paddingVertical: 0,
  },
  listContent: {
    paddingHorizontal: 12,
    paddingTop: 4,
    paddingBottom: 24,
  },
  columnWrapper: {
    gap: 10,
    marginBottom: 10,
    paddingHorizontal: 4,
  },
  card: {
    flex: 1,
    backgroundColor: '#FFFFFF',
    borderRadius: 14,
    overflow: 'hidden',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 4,
    elevation: 2,
  },
  cardStrip: {
    height: 56,
    alignItems: 'center',
    justifyContent: 'flex-end',
    paddingBottom: 0,
  },
  avatar: {
    width: 48,
    height: 48,
    borderRadius: 24,
    alignItems: 'center',
    justifyContent: 'center',
    marginBottom: -24,
    borderWidth: 3,
    borderColor: '#FFFFFF',
  },
  avatarText: {
    fontSize: 16,
    fontWeight: '700',
    color: '#FFFFFF',
  },
  cardBody: {
    paddingTop: 28,
    paddingHorizontal: 12,
    paddingBottom: 14,
    alignItems: 'center',
  },
  name: {
    fontSize: 14,
    fontWeight: '700',
    color: '#111827',
    textAlign: 'center',
    marginBottom: 6,
  },
  yearBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
    backgroundColor: '#EEF2FF',
    paddingHorizontal: 8,
    paddingVertical: 3,
    borderRadius: 20,
    marginBottom: 10,
  },
  yearText: {
    fontSize: 11,
    fontWeight: '600',
    color: '#4F46E5',
  },
  detailRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
    marginBottom: 3,
    width: '100%',
  },
  detailText: {
    flex: 1,
    fontSize: 12,
    color: '#6B7280',
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
    fontSize: 15,
    fontWeight: '600',
    color: '#374151',
    textAlign: 'center',
  },
});
