'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Plus, Utensils, ShoppingCart, TrendingUp } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import api from '@/lib/api';
import { Card, CardHeader } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input, Select } from '@/components/ui/Input';
import { Table } from '@/components/ui/Table';
import { Pagination } from '@/components/ui/Pagination';
import { Modal } from '@/components/ui/Modal';
import { Badge } from '@/components/ui/Badge';
import { formatCurrency, formatDateTime } from '@/lib/utils';

interface MenuItem {
  id: string;
  name: string;
  category: string;
  price: number;
  stock: number;
  isAvailable: boolean;
}

interface CanteenOrder {
  id: string;
  studentName?: string;
  studentId: string;
  itemsCount: number;
  total: number;
  status: 'PENDING' | 'PROCESSING' | 'READY' | 'COMPLETED' | 'CANCELLED';
  createdAt: string;
}

type TabKey = 'menu' | 'orders';

const ORDER_STATUS_CONFIG: Record<string, { label: string; variant: 'warning' | 'info' | 'success' | 'danger' | 'gray' | 'primary' }> = {
  PENDING: { label: 'Menunggu', variant: 'warning' },
  PROCESSING: { label: 'Diproses', variant: 'info' },
  READY: { label: 'Siap Diambil', variant: 'primary' },
  COMPLETED: { label: 'Selesai', variant: 'success' },
  CANCELLED: { label: 'Dibatalkan', variant: 'danger' },
};

export default function CanteenPage() {
  const { user } = useAuth();
  const [activeTab, setActiveTab] = useState<TabKey>('menu');
  const [menuItems, setMenuItems] = useState<MenuItem[]>([]);
  const [orders, setOrders] = useState<CanteenOrder[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [menuPage, setMenuPage] = useState(1);
  const [orderPage, setOrderPage] = useState(1);
  const [menuMeta, setMenuMeta] = useState({ total: 0, totalPages: 1 });
  const [orderMeta, setOrderMeta] = useState({ total: 0, totalPages: 1 });
  const [addOpen, setAddOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Form state
  const [fName, setFName] = useState('');
  const [fCategory, setFCategory] = useState('');
  const [fPrice, setFPrice] = useState('');
  const [fStock, setFStock] = useState('');
  const [fAvailable, setFAvailable] = useState(true);

  const fetchMenu = useCallback(async () => {
    setIsLoading(true);
    try {
      const params = new URLSearchParams({ page: String(menuPage), limit: '20' });
      const res = await api.get(`/canteen/menu-items?${params}`);
      setMenuItems(res.data.data || []);
      setMenuMeta(res.data.meta || { total: 0, totalPages: 1 });
    } catch {
      setMenuItems([]);
    } finally {
      setIsLoading(false);
    }
  }, [menuPage]);

  const fetchOrders = useCallback(async () => {
    setIsLoading(true);
    try {
      const params = new URLSearchParams({ page: String(orderPage), limit: '20' });
      const res = await api.get(`/canteen/orders?${params}`);
      setOrders(res.data.data || []);
      setOrderMeta(res.data.meta || { total: 0, totalPages: 1 });
    } catch {
      setOrders([]);
    } finally {
      setIsLoading(false);
    }
  }, [orderPage]);

  useEffect(() => { fetchMenu(); }, [fetchMenu]);
  useEffect(() => { fetchOrders(); }, [fetchOrders]);

  const resetForm = () => {
    setFName(''); setFCategory(''); setFPrice(''); setFStock(''); setFAvailable(true);
  };

  const handleAddItem = async () => {
    if (!fName || !fPrice) return;
    setIsSubmitting(true);
    try {
      await api.post('/canteen/menu-items', {
        name: fName,
        category: fCategory,
        price: parseFloat(fPrice),
        stock: parseInt(fStock) || 0,
        isAvailable: fAvailable,
      });
      setAddOpen(false);
      resetForm();
      fetchMenu();
    } finally {
      setIsSubmitting(false);
    }
  };

  // Summary stats
  const todayOrders = orders.filter((o) => {
    const today = new Date().toDateString();
    return new Date(o.createdAt).toDateString() === today;
  });
  const todayRevenue = todayOrders
    .filter((o) => o.status === 'COMPLETED')
    .reduce((sum, o) => sum + o.total, 0);

  const menuColumns = [
    {
      key: 'name',
      header: 'Nama',
      render: (m: MenuItem) => <span className="font-medium text-gray-900">{m.name}</span>,
    },
    {
      key: 'category',
      header: 'Kategori',
      render: (m: MenuItem) => m.category || <span className="text-gray-300">—</span>,
    },
    {
      key: 'price',
      header: 'Harga',
      render: (m: MenuItem) => formatCurrency(m.price),
    },
    {
      key: 'stock',
      header: 'Stok',
      render: (m: MenuItem) => (
        <span className={m.stock <= 5 ? 'text-red-600 font-medium' : 'text-gray-700'}>{m.stock}</span>
      ),
    },
    {
      key: 'isAvailable',
      header: 'Status',
      render: (m: MenuItem) => (
        <Badge variant={m.isAvailable ? 'success' : 'gray'} size="sm">
          {m.isAvailable ? 'Tersedia' : 'Tidak Tersedia'}
        </Badge>
      ),
    },
  ];

  const orderColumns = [
    {
      key: 'student',
      header: 'Siswa',
      render: (o: CanteenOrder) => o.studentName || o.studentId,
    },
    {
      key: 'itemsCount',
      header: 'Jumlah Item',
      render: (o: CanteenOrder) => `${o.itemsCount} item`,
    },
    {
      key: 'total',
      header: 'Total',
      render: (o: CanteenOrder) => <span className="font-medium">{formatCurrency(o.total)}</span>,
    },
    {
      key: 'status',
      header: 'Status',
      render: (o: CanteenOrder) => {
        const cfg = ORDER_STATUS_CONFIG[o.status] || { label: o.status, variant: 'gray' as const };
        return <Badge variant={cfg.variant} size="sm">{cfg.label}</Badge>;
      },
    },
    {
      key: 'createdAt',
      header: 'Waktu',
      render: (o: CanteenOrder) => formatDateTime(o.createdAt),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900">Kantin Digital</h1>
          <p className="text-sm text-gray-500 mt-0.5">Manajemen menu dan pesanan kantin sekolah</p>
        </div>
        {activeTab === 'menu' && (
          <Button
            size="sm"
            leftIcon={<Plus className="w-3.5 h-3.5" />}
            onClick={() => setAddOpen(true)}
          >
            Tambah Menu
          </Button>
        )}
      </div>

      {/* Stats */}
      <div className="grid grid-cols-3 gap-4">
        <Card padding="sm">
          <div className="flex items-center gap-3">
            <div className="p-2.5 bg-primary-50 rounded-lg">
              <Utensils className="w-5 h-5 text-primary-600" />
            </div>
            <div>
              <p className="text-xl font-bold text-gray-900">{menuMeta.total}</p>
              <p className="text-xs text-gray-500">Total Menu</p>
            </div>
          </div>
        </Card>
        <Card padding="sm">
          <div className="flex items-center gap-3">
            <div className="p-2.5 bg-blue-50 rounded-lg">
              <ShoppingCart className="w-5 h-5 text-blue-600" />
            </div>
            <div>
              <p className="text-xl font-bold text-gray-900">{todayOrders.length}</p>
              <p className="text-xs text-gray-500">Pesanan Hari Ini</p>
            </div>
          </div>
        </Card>
        <Card padding="sm">
          <div className="flex items-center gap-3">
            <div className="p-2.5 bg-green-50 rounded-lg">
              <TrendingUp className="w-5 h-5 text-green-600" />
            </div>
            <div>
              <p className="text-xl font-bold text-gray-900">{formatCurrency(todayRevenue)}</p>
              <p className="text-xs text-gray-500">Pendapatan Hari Ini</p>
            </div>
          </div>
        </Card>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 bg-gray-100 p-1 rounded-xl w-fit">
        {([
          { key: 'menu', label: 'Daftar Menu' },
          { key: 'orders', label: 'Pesanan' },
        ] as { key: TabKey; label: string }[]).map((t) => (
          <button
            key={t.key}
            onClick={() => setActiveTab(t.key)}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
              activeTab === t.key ? 'bg-white text-primary-700 shadow-sm' : 'text-gray-500'
            }`}
          >
            {t.label}
          </button>
        ))}
      </div>

      {activeTab === 'menu' && (
        <Card padding="none">
          <Table
            columns={menuColumns}
            data={menuItems}
            keyExtractor={(m) => m.id}
            isLoading={isLoading}
            emptyMessage="Belum ada item menu kantin"
          />
          {menuMeta.total > 0 && (
            <div className="px-4 py-3 border-t border-gray-100">
              <Pagination
                currentPage={menuPage}
                totalPages={menuMeta.totalPages}
                total={menuMeta.total}
                limit={20}
                onPageChange={setMenuPage}
              />
            </div>
          )}
        </Card>
      )}

      {activeTab === 'orders' && (
        <Card padding="none">
          <Table
            columns={orderColumns}
            data={orders}
            keyExtractor={(o) => o.id}
            isLoading={isLoading}
            emptyMessage="Belum ada pesanan kantin"
          />
          {orderMeta.total > 0 && (
            <div className="px-4 py-3 border-t border-gray-100">
              <Pagination
                currentPage={orderPage}
                totalPages={orderMeta.totalPages}
                total={orderMeta.total}
                limit={20}
                onPageChange={setOrderPage}
              />
            </div>
          )}
        </Card>
      )}

      {/* Add Menu Item Modal */}
      <Modal
        isOpen={addOpen}
        onClose={() => { setAddOpen(false); resetForm(); }}
        title="Tambah Item Menu Kantin"
        size="sm"
        footer={
          <>
            <Button variant="secondary" onClick={() => { setAddOpen(false); resetForm(); }}>Batal</Button>
            <Button onClick={handleAddItem} isLoading={isSubmitting}>Simpan</Button>
          </>
        }
      >
        <div className="space-y-4">
          <Input
            label="Nama Menu"
            placeholder="Contoh: Nasi Goreng Spesial"
            required
            value={fName}
            onChange={(e) => setFName(e.target.value)}
          />
          <Input
            label="Kategori"
            placeholder="Contoh: Makanan Berat, Minuman, Snack..."
            value={fCategory}
            onChange={(e) => setFCategory(e.target.value)}
          />
          <div className="grid grid-cols-2 gap-4">
            <Input
              label="Harga (Rp)"
              type="number"
              placeholder="Contoh: 15000"
              required
              value={fPrice}
              onChange={(e) => setFPrice(e.target.value)}
            />
            <Input
              label="Stok"
              type="number"
              placeholder="Contoh: 50"
              value={fStock}
              onChange={(e) => setFStock(e.target.value)}
            />
          </div>
          <div className="flex items-center gap-3 p-3 bg-gray-50 rounded-lg">
            <input
              type="checkbox"
              id="isAvailable"
              checked={fAvailable}
              onChange={(e) => setFAvailable(e.target.checked)}
              className="w-4 h-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
            />
            <label htmlFor="isAvailable" className="text-sm font-medium text-gray-700">
              Tersedia untuk dijual
            </label>
          </div>
        </div>
      </Modal>
    </div>
  );
}
