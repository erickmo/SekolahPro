import React, { useCallback, useEffect, useState } from 'react';
import {
  ActivityIndicator,
  FlatList,
  RefreshControl,
  SectionList,
  StyleSheet,
  Text,
  View,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { Ionicons } from '@expo/vector-icons';
import { get } from '../../src/lib/api';

// ─── Types ────────────────────────────────────────────────────────────────────

type MenuCategory = 'Makanan Berat' | 'Minuman' | 'Snack';

interface MenuItem {
  id: string;
  name: string;
  category: MenuCategory;
  price: number;
  stock: number;
}

interface MenuSection {
  title: MenuCategory;
  data: MenuItem[];
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

function formatRupiah(amount: number): string {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(amount);
}

function groupByCategory(items: MenuItem[]): MenuSection[] {
  const order: MenuCategory[] = ['Makanan Berat', 'Minuman', 'Snack'];
  const grouped: Record<MenuCategory, MenuItem[]> = {
    'Makanan Berat': [],
    'Minuman': [],
    'Snack': [],
  };

  items.forEach((item) => {
    if (grouped[item.category]) {
      grouped[item.category].push(item);
    }
  });

  return order
    .filter((cat) => grouped[cat].length > 0)
    .map((cat) => ({ title: cat, data: grouped[cat] }));
}

const CATEGORY_ICON: Record<MenuCategory, React.ComponentProps<typeof Ionicons>['name']> = {
  'Makanan Berat': 'restaurant',
  'Minuman': 'cafe',
  'Snack': 'nutrition',
};

const CATEGORY_COLOR: Record<MenuCategory, string> = {
  'Makanan Berat': '#EEF2FF',
  'Minuman': '#ECFDF5',
  'Snack': '#FFF7ED',
};

const CATEGORY_ICON_COLOR: Record<MenuCategory, string> = {
  'Makanan Berat': '#4F46E5',
  'Minuman': '#059669',
  'Snack': '#EA580C',
};

// ─── Menu Item Card ───────────────────────────────────────────────────────────

function MenuItemCard({ item }: { item: MenuItem }) {
  const outOfStock = item.stock === 0;
  const iconBg = CATEGORY_COLOR[item.category] ?? '#F3F4F6';
  const iconColor = CATEGORY_ICON_COLOR[item.category] ?? '#6B7280';
  const icon = CATEGORY_ICON[item.category] ?? 'fast-food';

  return (
    <View style={[styles.card, outOfStock && styles.cardUnavailable]}>
      <View style={[styles.itemIconWrapper, { backgroundColor: iconBg }]}>
        <Ionicons name={icon} size={22} color={outOfStock ? '#D1D5DB' : iconColor} />
      </View>
      <View style={styles.itemInfo}>
        <View style={styles.itemTop}>
          <Text
            style={[styles.itemName, outOfStock && styles.itemNameUnavailable]}
            numberOfLines={1}
          >
            {item.name}
          </Text>
          {outOfStock && (
            <View style={styles.outOfStockBadge}>
              <Text style={styles.outOfStockText}>Habis</Text>
            </View>
          )}
        </View>
        <View style={styles.itemBottom}>
          <Text style={[styles.itemPrice, outOfStock && styles.itemPriceUnavailable]}>
            {formatRupiah(item.price)}
          </Text>
          {!outOfStock && (
            <Text style={styles.stockText}>Stok: {item.stock}</Text>
          )}
        </View>
      </View>
    </View>
  );
}

// ─── Section Header ───────────────────────────────────────────────────────────

function SectionHeader({ title }: { title: MenuCategory }) {
  return (
    <View style={styles.sectionHeader}>
      <Ionicons name={CATEGORY_ICON[title]} size={16} color={CATEGORY_ICON_COLOR[title] ?? '#6B7280'} />
      <Text style={styles.sectionTitle}>{title}</Text>
    </View>
  );
}

// ─── Main Screen ──────────────────────────────────────────────────────────────

export default function CanteenScreen() {
  const [sections, setSections] = useState<MenuSection[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [totalItems, setTotalItems] = useState(0);

  const fetchData = useCallback(async (refresh = false) => {
    if (refresh) setIsRefreshing(true);
    else setIsLoading(true);

    try {
      const response = await get<{ data: MenuItem[] }>('/canteen/menu-items');
      const items = response.data ?? [];
      setTotalItems(items.length);
      setSections(groupByCategory(items));
    } catch {
      const mockItems = MOCK_MENU_ITEMS;
      setTotalItems(mockItems.length);
      setSections(groupByCategory(mockItems));
    } finally {
      setIsLoading(false);
      setIsRefreshing(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const onRefresh = useCallback(() => {
    fetchData(true);
  }, [fetchData]);

  return (
    <SafeAreaView style={styles.safeArea} edges={['top']}>
      {/* Header */}
      <View style={styles.header}>
        <View>
          <Text style={styles.headerTitle}>Kantin Digital</Text>
          <Text style={styles.headerSubtitle}>
            {totalItems} menu tersedia
          </Text>
        </View>
        <View style={styles.headerIcon}>
          <Ionicons name="restaurant" size={22} color="#4F46E5" />
        </View>
      </View>

      {/* Content */}
      {isLoading ? (
        <View style={styles.loadingCenter}>
          <ActivityIndicator size="large" color="#4F46E5" />
          <Text style={styles.loadingText}>Memuat menu kantin...</Text>
        </View>
      ) : sections.length === 0 ? (
        <View style={styles.emptyContainer}>
          <Ionicons name="restaurant-outline" size={48} color="#D1D5DB" />
          <Text style={styles.emptyTitle}>Menu tidak tersedia</Text>
          <Text style={styles.emptySubtitle}>Belum ada menu yang ditambahkan.</Text>
        </View>
      ) : (
        <SectionList
          sections={sections}
          keyExtractor={(item) => item.id}
          renderItem={({ item }) => <MenuItemCard item={item} />}
          renderSectionHeader={({ section }) => (
            <SectionHeader title={section.title} />
          )}
          contentContainerStyle={styles.listContent}
          showsVerticalScrollIndicator={false}
          stickySectionHeadersEnabled={false}
          refreshControl={
            <RefreshControl
              refreshing={isRefreshing}
              onRefresh={onRefresh}
              tintColor="#4F46E5"
              colors={['#4F46E5']}
            />
          }
        />
      )}
    </SafeAreaView>
  );
}

// ─── Mock Data ────────────────────────────────────────────────────────────────

const MOCK_MENU_ITEMS: MenuItem[] = [
  { id: 'm1', name: 'Nasi Goreng Spesial', category: 'Makanan Berat', price: 12000, stock: 15 },
  { id: 'm2', name: 'Nasi Ayam Geprek', category: 'Makanan Berat', price: 13000, stock: 10 },
  { id: 'm3', name: 'Mie Goreng', category: 'Makanan Berat', price: 10000, stock: 0 },
  { id: 'm4', name: 'Bakso Kuah', category: 'Makanan Berat', price: 11000, stock: 20 },
  { id: 'm5', name: 'Air Mineral', category: 'Minuman', price: 3000, stock: 50 },
  { id: 'm6', name: 'Teh Manis', category: 'Minuman', price: 4000, stock: 30 },
  { id: 'm7', name: 'Jus Jeruk', category: 'Minuman', price: 7000, stock: 12 },
  { id: 'm8', name: 'Susu Kotak', category: 'Minuman', price: 5000, stock: 0 },
  { id: 'm9', name: 'Keripik Singkong', category: 'Snack', price: 3000, stock: 25 },
  { id: 'm10', name: 'Biskuit Coklat', category: 'Snack', price: 4000, stock: 18 },
  { id: 'm11', name: 'Gorengan (3 pcs)', category: 'Snack', price: 5000, stock: 40 },
  { id: 'm12', name: 'Risoles', category: 'Snack', price: 3500, stock: 0 },
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
    paddingBottom: 16,
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
    paddingBottom: 24,
  },
  sectionHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    paddingTop: 16,
    paddingBottom: 10,
  },
  sectionTitle: {
    fontSize: 16,
    fontWeight: '700',
    color: '#111827',
  },
  card: {
    backgroundColor: '#FFFFFF',
    borderRadius: 14,
    padding: 14,
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 8,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 4,
    elevation: 2,
  },
  cardUnavailable: {
    opacity: 0.65,
  },
  itemIconWrapper: {
    width: 46,
    height: 46,
    borderRadius: 12,
    alignItems: 'center',
    justifyContent: 'center',
    marginRight: 12,
    flexShrink: 0,
  },
  itemInfo: {
    flex: 1,
  },
  itemTop: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    marginBottom: 4,
  },
  itemName: {
    flex: 1,
    fontSize: 15,
    fontWeight: '600',
    color: '#111827',
    marginRight: 8,
  },
  itemNameUnavailable: {
    color: '#9CA3AF',
  },
  outOfStockBadge: {
    backgroundColor: '#FEE2E2',
    paddingHorizontal: 8,
    paddingVertical: 3,
    borderRadius: 6,
  },
  outOfStockText: {
    fontSize: 11,
    fontWeight: '700',
    color: '#DC2626',
  },
  itemBottom: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  itemPrice: {
    fontSize: 15,
    fontWeight: '700',
    color: '#4F46E5',
  },
  itemPriceUnavailable: {
    color: '#9CA3AF',
  },
  stockText: {
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
  emptyContainer: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
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
  },
});
