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

type ActiveTab = 'BOOKS' | 'LOANS';
type LoanStatus = 'BORROWED' | 'RETURNED' | 'OVERDUE';

interface Book {
  id: string;
  title: string;
  author: string;
  category: string;
  availableCopies: number;
  totalCopies: number;
}

interface Loan {
  id: string;
  bookTitle: string;
  borrowerName: string;
  dueDate: string;
  status: LoanStatus;
}

const LOAN_STATUS_CONFIG: Record<LoanStatus, { label: string; bg: string; text: string }> = {
  BORROWED: { label: 'Dipinjam', bg: '#DBEAFE', text: '#1D4ED8' },
  RETURNED: { label: 'Dikembalikan', bg: '#DCFCE7', text: '#15803D' },
  OVERDUE: { label: 'Terlambat', bg: '#FEE2E2', text: '#DC2626' },
};

const PAGE_SIZE = 20;

// ─── Helpers ──────────────────────────────────────────────────────────────────

function formatDueDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  });
}

// ─── Book Card ────────────────────────────────────────────────────────────────

function BookCard({ item }: { item: Book }) {
  const isLow = item.availableCopies === 0;
  return (
    <View style={styles.card}>
      <View style={styles.bookIconWrapper}>
        <Ionicons name="book" size={22} color="#4F46E5" />
      </View>
      <View style={styles.bookInfo}>
        <Text style={styles.bookTitle} numberOfLines={2}>
          {item.title}
        </Text>
        <Text style={styles.bookAuthor}>{item.author}</Text>
        <View style={styles.bookMeta}>
          <View style={styles.categoryBadge}>
            <Text style={styles.categoryText}>{item.category}</Text>
          </View>
          <Text style={[styles.copyCount, { color: isLow ? '#DC2626' : '#16A34A' }]}>
            {isLow
              ? 'Habis'
              : `${item.availableCopies}/${item.totalCopies} tersedia`}
          </Text>
        </View>
      </View>
    </View>
  );
}

// ─── Loan Card ────────────────────────────────────────────────────────────────

function LoanCard({ item }: { item: Loan }) {
  const cfg = LOAN_STATUS_CONFIG[item.status];
  return (
    <View style={styles.card}>
      <View style={styles.loanContent}>
        <View style={styles.loanTop}>
          <Text style={styles.bookTitle} numberOfLines={2}>
            {item.bookTitle}
          </Text>
          <View style={[styles.statusBadge, { backgroundColor: cfg.bg }]}>
            <Text style={[styles.statusText, { color: cfg.text }]}>{cfg.label}</Text>
          </View>
        </View>
        <View style={styles.loanBottom}>
          <View style={styles.loanDetail}>
            <Ionicons name="person-outline" size={13} color="#9CA3AF" />
            <Text style={styles.loanDetailText}>{item.borrowerName}</Text>
          </View>
          <View style={styles.loanDetail}>
            <Ionicons name="calendar-outline" size={13} color="#9CA3AF" />
            <Text
              style={[
                styles.loanDetailText,
                item.status === 'OVERDUE' && { color: '#DC2626', fontWeight: '600' },
              ]}
            >
              Kembali: {formatDueDate(item.dueDate)}
            </Text>
          </View>
        </View>
      </View>
    </View>
  );
}

// ─── Main Screen ──────────────────────────────────────────────────────────────

export default function LibraryScreen() {
  const [activeTab, setActiveTab] = useState<ActiveTab>('BOOKS');
  const [books, setBooks] = useState<Book[]>([]);
  const [loans, setLoans] = useState<Loan[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [isLoadingBooks, setIsLoadingBooks] = useState(true);
  const [isLoadingLoans, setIsLoadingLoans] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [bookPage, setBookPage] = useState(1);
  const [hasMoreBooks, setHasMoreBooks] = useState(true);
  const searchTimeout = useRef<ReturnType<typeof setTimeout> | null>(null);

  const fetchBooks = useCallback(
    async (pageNum: number, query: string, isLoadMore = false) => {
      if (!isLoadMore) setIsLoadingBooks(pageNum === 1);
      else setIsLoadingMore(true);

      try {
        const params: Record<string, unknown> = { page: pageNum, limit: PAGE_SIZE };
        if (query.trim()) params.search = query.trim();

        const response = await get<{ data: Book[]; meta: { hasNextPage: boolean } }>(
          '/library/books',
          params,
        );
        const newItems = response.data ?? [];
        setHasMoreBooks(response.meta?.hasNextPage ?? false);

        if (isLoadMore) {
          setBooks((prev) => [...prev, ...newItems]);
        } else {
          setBooks(newItems);
        }
      } catch {
        if (!isLoadMore) {
          setBooks(MOCK_BOOKS);
          setHasMoreBooks(false);
        }
      } finally {
        setIsLoadingBooks(false);
        setIsLoadingMore(false);
        setIsRefreshing(false);
      }
    },
    [],
  );

  const fetchLoans = useCallback(async () => {
    setIsLoadingLoans(true);
    try {
      const response = await get<{ data: Loan[] }>('/library/loans');
      setLoans(response.data ?? []);
    } catch {
      setLoans(MOCK_LOANS);
    } finally {
      setIsLoadingLoans(false);
      setIsRefreshing(false);
    }
  }, []);

  useEffect(() => {
    fetchBooks(1, '');
    fetchLoans();
  }, [fetchBooks, fetchLoans]);

  const handleSearch = (text: string) => {
    setSearchQuery(text);
    if (searchTimeout.current) clearTimeout(searchTimeout.current);
    searchTimeout.current = setTimeout(() => {
      setBookPage(1);
      fetchBooks(1, text);
    }, 400);
  };

  const onRefresh = useCallback(() => {
    setIsRefreshing(true);
    if (activeTab === 'BOOKS') {
      setBookPage(1);
      fetchBooks(1, searchQuery);
    } else {
      fetchLoans();
    }
  }, [activeTab, fetchBooks, fetchLoans, searchQuery]);

  const onLoadMoreBooks = useCallback(() => {
    if (!hasMoreBooks || isLoadingMore || isLoadingBooks) return;
    const next = bookPage + 1;
    setBookPage(next);
    fetchBooks(next, searchQuery, true);
  }, [hasMoreBooks, isLoadingMore, isLoadingBooks, bookPage, fetchBooks, searchQuery]);

  const isLoading = activeTab === 'BOOKS' ? isLoadingBooks : isLoadingLoans;

  const renderFooter = () =>
    isLoadingMore && activeTab === 'BOOKS' ? (
      <View style={styles.loadMoreContainer}>
        <ActivityIndicator size="small" color="#4F46E5" />
      </View>
    ) : null;

  const renderEmpty = () =>
    isLoading ? null : (
      <View style={styles.emptyContainer}>
        <Ionicons name="book-outline" size={48} color="#D1D5DB" />
        <Text style={styles.emptyTitle}>
          {activeTab === 'BOOKS' ? 'Tidak ada buku ditemukan' : 'Tidak ada peminjaman aktif'}
        </Text>
      </View>
    );

  return (
    <SafeAreaView style={styles.safeArea} edges={['top']}>
      {/* Header */}
      <View style={styles.header}>
        <View>
          <Text style={styles.headerTitle}>Perpustakaan</Text>
          <Text style={styles.headerSubtitle}>Koleksi & peminjaman buku</Text>
        </View>
        <View style={styles.headerIcon}>
          <Ionicons name="book" size={22} color="#4F46E5" />
        </View>
      </View>

      {/* Tabs */}
      <View style={styles.tabRow}>
        {(['BOOKS', 'LOANS'] as ActiveTab[]).map((tab) => (
          <TouchableOpacity
            key={tab}
            style={[styles.tabButton, activeTab === tab && styles.tabButtonActive]}
            onPress={() => setActiveTab(tab)}
            activeOpacity={0.75}
          >
            <Text style={[styles.tabText, activeTab === tab && styles.tabTextActive]}>
              {tab === 'BOOKS' ? 'Buku' : 'Peminjaman'}
            </Text>
          </TouchableOpacity>
        ))}
      </View>

      {/* Search (books only) */}
      {activeTab === 'BOOKS' && (
        <View style={styles.searchContainer}>
          <View style={styles.searchBar}>
            <Ionicons name="search-outline" size={18} color="#9CA3AF" style={styles.searchIcon} />
            <TextInput
              style={styles.searchInput}
              placeholder="Cari judul atau penulis..."
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
      )}

      {/* List */}
      {isLoading ? (
        <View style={styles.loadingCenter}>
          <ActivityIndicator size="large" color="#4F46E5" />
          <Text style={styles.loadingText}>
            {activeTab === 'BOOKS' ? 'Memuat koleksi buku...' : 'Memuat data peminjaman...'}
          </Text>
        </View>
      ) : (
        <FlatList
          key={activeTab}
          data={activeTab === 'BOOKS' ? (books as any[]) : (loans as any[])}
          keyExtractor={(item) => item.id}
          renderItem={({ item }) =>
            activeTab === 'BOOKS' ? <BookCard item={item} /> : <LoanCard item={item} />
          }
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
          onEndReached={activeTab === 'BOOKS' ? onLoadMoreBooks : undefined}
          onEndReachedThreshold={0.3}
          ListFooterComponent={renderFooter}
          ListEmptyComponent={renderEmpty}
        />
      )}
    </SafeAreaView>
  );
}

// ─── Mock Data ────────────────────────────────────────────────────────────────

const MOCK_BOOKS: Book[] = [
  { id: 'b1', title: 'Laskar Pelangi', author: 'Andrea Hirata', category: 'Fiksi', availableCopies: 3, totalCopies: 5 },
  { id: 'b2', title: 'Bumi Manusia', author: 'Pramoedya Ananta Toer', category: 'Sejarah', availableCopies: 1, totalCopies: 3 },
  { id: 'b3', title: 'Matematika Kelas X', author: 'Kemendikbud', category: 'Pelajaran', availableCopies: 0, totalCopies: 10 },
  { id: 'b4', title: 'Kimia Dasar', author: 'Raymond Chang', category: 'Sains', availableCopies: 2, totalCopies: 4 },
  { id: 'b5', title: 'Fisika Terapan', author: 'Halliday & Resnick', category: 'Sains', availableCopies: 5, totalCopies: 6 },
  { id: 'b6', title: 'Atlas Dunia', author: 'National Geographic', category: 'Referensi', availableCopies: 1, totalCopies: 2 },
];

const MOCK_LOAN_NAMES = [
  'Ahmad Fauzi', 'Budi Santoso', 'Citra Dewi', 'Dini Rahayu', 'Eko Prasetyo',
];

const MOCK_LOAN_STATUSES: LoanStatus[] = ['BORROWED', 'RETURNED', 'OVERDUE'];

const MOCK_LOANS: Loan[] = Array.from({ length: 10 }, (_, i) => ({
  id: `loan-${i + 1}`,
  bookTitle: MOCK_BOOKS[i % MOCK_BOOKS.length].title,
  borrowerName: MOCK_LOAN_NAMES[i % MOCK_LOAN_NAMES.length],
  dueDate: new Date(2026, 2, 10 + i).toISOString(),
  status: MOCK_LOAN_STATUSES[i % MOCK_LOAN_STATUSES.length],
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
  tabRow: {
    flexDirection: 'row',
    marginHorizontal: 16,
    marginBottom: 12,
    backgroundColor: '#F3F4F6',
    borderRadius: 12,
    padding: 4,
  },
  tabButton: {
    flex: 1,
    paddingVertical: 10,
    alignItems: 'center',
    borderRadius: 10,
  },
  tabButtonActive: {
    backgroundColor: '#FFFFFF',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.08,
    shadowRadius: 3,
    elevation: 2,
  },
  tabText: {
    fontSize: 14,
    fontWeight: '500',
    color: '#6B7280',
  },
  tabTextActive: {
    color: '#111827',
    fontWeight: '700',
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
    paddingHorizontal: 16,
    paddingTop: 4,
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
  bookIconWrapper: {
    width: 44,
    height: 44,
    borderRadius: 12,
    backgroundColor: '#EEF2FF',
    alignItems: 'center',
    justifyContent: 'center',
    marginRight: 12,
    flexShrink: 0,
  },
  bookInfo: {
    flex: 1,
  },
  bookTitle: {
    fontSize: 15,
    fontWeight: '700',
    color: '#111827',
    marginBottom: 2,
  },
  bookAuthor: {
    fontSize: 13,
    color: '#6B7280',
    marginBottom: 8,
  },
  bookMeta: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  categoryBadge: {
    backgroundColor: '#F3F4F6',
    paddingHorizontal: 8,
    paddingVertical: 3,
    borderRadius: 6,
  },
  categoryText: {
    fontSize: 12,
    color: '#374151',
    fontWeight: '500',
  },
  copyCount: {
    fontSize: 12,
    fontWeight: '600',
  },
  loanContent: {
    flex: 1,
  },
  loanTop: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    marginBottom: 10,
    gap: 8,
  },
  statusBadge: {
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 20,
    flexShrink: 0,
  },
  statusText: {
    fontSize: 12,
    fontWeight: '600',
  },
  loanBottom: {
    gap: 4,
  },
  loanDetail: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
  },
  loanDetailText: {
    fontSize: 13,
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
    gap: 8,
  },
  emptyTitle: {
    fontSize: 16,
    fontWeight: '600',
    color: '#374151',
  },
});
