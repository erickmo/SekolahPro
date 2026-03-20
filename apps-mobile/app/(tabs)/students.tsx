import React, { useCallback, useEffect, useRef, useState } from 'react';
import {
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
import { useRouter } from 'expo-router';
import { get } from '../../src/lib/api';
import StudentCard from '../../src/components/StudentCard';
import EmptyState from '../../src/components/EmptyState';
import LoadingSpinner from '../../src/components/LoadingSpinner';
import type { PaginatedResponse, StudentListItem } from '../../src/types';

// ─── Filter Types ─────────────────────────────────────────────────────────────

type StatusFilter = 'ALL' | 'ACTIVE' | 'INACTIVE' | 'TRANSFERRED' | 'GRADUATED';

const STATUS_FILTERS: { key: StatusFilter; label: string }[] = [
  { key: 'ALL', label: 'Semua' },
  { key: 'ACTIVE', label: 'Aktif' },
  { key: 'INACTIVE', label: 'Tidak Aktif' },
  { key: 'TRANSFERRED', label: 'Pindah' },
  { key: 'GRADUATED', label: 'Lulus' },
];

const PAGE_SIZE = 20;

// ─── Component ────────────────────────────────────────────────────────────────

export default function StudentsScreen() {
  const router = useRouter();
  const searchTimeout = useRef<ReturnType<typeof setTimeout> | null>(null);

  const [students, setStudents] = useState<StudentListItem[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [activeFilter, setActiveFilter] = useState<StatusFilter>('ALL');
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const [totalCount, setTotalCount] = useState(0);

  const fetchStudents = useCallback(
    async (
      pageNum: number,
      query: string,
      filter: StatusFilter,
      isLoadMore = false,
    ) => {
      if (!isLoadMore) {
        if (pageNum === 1) setIsLoading(true);
      } else {
        setIsLoadingMore(true);
      }

      try {
        const params: Record<string, unknown> = {
          page: pageNum,
          limit: PAGE_SIZE,
        };
        if (query.trim()) params.search = query.trim();
        if (filter !== 'ALL') params.status = filter;

        const response = await get<PaginatedResponse<StudentListItem>>(
          '/students',
          params,
        );

        const newItems = response.items ?? [];
        setTotalCount(response.meta?.total ?? 0);
        setHasMore(response.meta?.hasNextPage ?? false);

        if (isLoadMore) {
          setStudents((prev) => [...prev, ...newItems]);
        } else {
          setStudents(newItems);
        }
      } catch {
        // Fallback mock data
        if (!isLoadMore) {
          const mockStudents: StudentListItem[] = Array.from(
            { length: 20 },
            (_, i) => ({
              id: `student-${i + 1}`,
              nisn: `${3010000001 + i}`,
              name: MOCK_NAMES[i % MOCK_NAMES.length],
              gender: i % 3 === 0 ? 'FEMALE' : 'MALE',
              status: i % 10 === 0 ? 'INACTIVE' : 'ACTIVE',
              grade: `${10 + (i % 3)}`,
              className: `IPA ${(i % 4) + 1}`,
              riskScore: i % 7 === 0 ? Math.floor(Math.random() * 60) + 30 : undefined,
            }),
          );
          setStudents(mockStudents);
          setTotalCount(247);
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

  // Initial load + filter/search changes
  useEffect(() => {
    setPage(1);
    fetchStudents(1, searchQuery, activeFilter);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [activeFilter]);

  // Debounced search
  const handleSearch = (text: string) => {
    setSearchQuery(text);
    if (searchTimeout.current) clearTimeout(searchTimeout.current);
    searchTimeout.current = setTimeout(() => {
      setPage(1);
      fetchStudents(1, text, activeFilter);
    }, 400);
  };

  const onRefresh = useCallback(() => {
    setIsRefreshing(true);
    setPage(1);
    fetchStudents(1, searchQuery, activeFilter, false);
  }, [fetchStudents, searchQuery, activeFilter]);

  const onLoadMore = useCallback(() => {
    if (!hasMore || isLoadingMore || isLoading) return;
    const nextPage = page + 1;
    setPage(nextPage);
    fetchStudents(nextPage, searchQuery, activeFilter, true);
  }, [hasMore, isLoadingMore, isLoading, page, fetchStudents, searchQuery, activeFilter]);

  const onStudentPress = useCallback(
    (id: string) => {
      router.push(`/students/${id}`);
    },
    [router],
  );

  const renderFooter = () => {
    if (!isLoadingMore) return null;
    return (
      <View style={styles.loadMoreContainer}>
        <LoadingSpinner size="small" label="Memuat lebih banyak..." />
      </View>
    );
  };

  const renderEmpty = () => {
    if (isLoading) return null;
    return (
      <EmptyState
        icon="people-outline"
        title="Tidak ada siswa ditemukan"
        subtitle={
          searchQuery
            ? `Tidak ada hasil untuk "${searchQuery}"`
            : 'Belum ada data siswa yang terdaftar.'
        }
        actionLabel={searchQuery ? 'Hapus Pencarian' : undefined}
        onAction={searchQuery ? () => handleSearch('') : undefined}
      />
    );
  };

  return (
    <SafeAreaView style={styles.safeArea} edges={['top']}>
      {/* Header */}
      <View style={styles.header}>
        <View>
          <Text style={styles.headerTitle}>Data Siswa</Text>
          <Text style={styles.headerSubtitle}>
            {totalCount.toLocaleString('id-ID')} siswa terdaftar
          </Text>
        </View>
        <TouchableOpacity style={styles.filterButton} activeOpacity={0.7}>
          <Ionicons name="options-outline" size={20} color="#4F46E5" />
        </TouchableOpacity>
      </View>

      {/* Search Bar */}
      <View style={styles.searchContainer}>
        <View style={styles.searchBar}>
          <Ionicons name="search-outline" size={18} color="#9CA3AF" style={styles.searchIcon} />
          <TextInput
            style={styles.searchInput}
            placeholder="Cari nama, NISN, kelas..."
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

      {/* Status Filter Tabs */}
      <View style={styles.filterTabs}>
        <FlatList
          data={STATUS_FILTERS}
          horizontal
          showsHorizontalScrollIndicator={false}
          contentContainerStyle={styles.filterTabsContent}
          keyExtractor={(item) => item.key}
          renderItem={({ item }) => (
            <TouchableOpacity
              style={[
                styles.filterTab,
                activeFilter === item.key && styles.filterTabActive,
              ]}
              onPress={() => setActiveFilter(item.key)}
              activeOpacity={0.75}
            >
              <Text
                style={[
                  styles.filterTabText,
                  activeFilter === item.key && styles.filterTabTextActive,
                ]}
              >
                {item.label}
              </Text>
            </TouchableOpacity>
          )}
        />
      </View>

      {/* List */}
      {isLoading ? (
        <LoadingSpinner label="Memuat data siswa..." style={styles.loadingCenter} />
      ) : (
        <FlatList
          data={students}
          keyExtractor={(item) => item.id}
          renderItem={({ item }) => (
            <StudentCard student={item} onPress={onStudentPress} />
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

const MOCK_NAMES = [
  'Ahmad Fauzi', 'Budi Santoso', 'Citra Dewi', 'Dini Rahayu', 'Eko Prasetyo',
  'Farah Azzahra', 'Gilang Ramadhan', 'Hana Putri', 'Ivan Setiawan', 'Julia Maharani',
  'Kevin Wijaya', 'Laila Nurdin', 'Muhammad Rizki', 'Nadia Sari', 'Oky Firmansyah',
  'Putri Ayu', 'Qori Fadillah', 'Reza Mahendra', 'Siti Aminah', 'Taufik Hidayat',
];

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
  filterButton: {
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
  filterTabs: {
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
    borderWidth: 1,
    borderColor: '#E5E7EB',
  },
  filterTabActive: {
    backgroundColor: '#4F46E5',
    borderColor: '#4F46E5',
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
    paddingTop: 8,
    paddingBottom: 20,
  },
  loadingCenter: {
    flex: 1,
  },
  loadMoreContainer: {
    paddingVertical: 16,
  },
});
