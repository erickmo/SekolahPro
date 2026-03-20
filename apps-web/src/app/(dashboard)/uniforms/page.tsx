'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Plus, Shirt } from 'lucide-react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import api from '@/lib/api';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Badge } from '@/components/ui/Badge';
import { Modal } from '@/components/ui/Modal';
import { Table } from '@/components/ui/Table';
import { Pagination } from '@/components/ui/Pagination';
import { formatCurrency } from '@/lib/utils';

type TabKey = 'items' | 'orders';

const ALL_SIZES = ['S', 'M', 'L', 'XL', 'XXL'];

const itemSchema = z.object({
  name: z.string().min(1, 'Nama item wajib diisi'),
  description: z.string().optional(),
  price: z.number().min(0, 'Harga tidak boleh negatif'),
  sizes: z.array(z.string()).min(1, 'Pilih minimal satu ukuran'),
});
type ItemForm = z.infer<typeof itemSchema>;

interface UniformItem {
  id: string;
  name: string;
  sizes: string[];
  price: number;
  isActive: boolean;
}

interface UniformOrder {
  id: string;
  studentId?: string;
  itemName?: string;
  size: string;
  qty: number;
  totalPrice: number;
  status: string;
}

const ORDER_STATUS_MAP: Record<string, { label: string; variant: 'warning' | 'info' | 'success' | 'danger' | 'gray' }> = {
  PENDING: { label: 'Menunggu', variant: 'warning' },
  PROCESSING: { label: 'Diproses', variant: 'info' },
  DELIVERED: { label: 'Dikirim', variant: 'success' },
  CANCELLED: { label: 'Dibatalkan', variant: 'danger' },
};

export default function UniformsPage() {
  const [activeTab, setActiveTab] = useState<TabKey>('items');

  const [items, setItems] = useState<UniformItem[]>([]);
  const [itemLoading, setItemLoading] = useState(true);
  const [itemPage, setItemPage] = useState(1);
  const [itemMeta, setItemMeta] = useState({ total: 0, totalPages: 1 });
  const [isAddOpen, setIsAddOpen] = useState(false);

  const [orders, setOrders] = useState<UniformOrder[]>([]);
  const [orderLoading, setOrderLoading] = useState(true);
  const [orderPage, setOrderPage] = useState(1);
  const [orderMeta, setOrderMeta] = useState({ total: 0, totalPages: 1 });

  const {
    register,
    handleSubmit,
    reset,
    watch,
    setValue,
    formState: { errors, isSubmitting },
  } = useForm<ItemForm>({
    resolver: zodResolver(itemSchema),
    defaultValues: { price: 0, sizes: [] },
  });

  const selectedSizes = watch('sizes') ?? [];

  const toggleSize = (size: string) => {
    const current = selectedSizes ?? [];
    if (current.includes(size)) {
      setValue('sizes', current.filter((s) => s !== size), { shouldValidate: true });
    } else {
      setValue('sizes', [...current, size], { shouldValidate: true });
    }
  };

  const fetchItems = useCallback(async () => {
    setItemLoading(true);
    try {
      const res = await api.get('/api/v1/uniforms/items', {
        params: { page: itemPage, limit: 20 },
      });
      setItems(res.data.data || []);
      if (res.data.meta) setItemMeta(res.data.meta);
    } catch {
      // handle silently
    } finally {
      setItemLoading(false);
    }
  }, [itemPage]);

  const fetchOrders = useCallback(async () => {
    setOrderLoading(true);
    try {
      const res = await api.get('/api/v1/uniforms/orders', {
        params: { page: orderPage, limit: 20 },
      });
      setOrders(res.data.data || []);
      if (res.data.meta) setOrderMeta(res.data.meta);
    } catch {
      // handle silently
    } finally {
      setOrderLoading(false);
    }
  }, [orderPage]);

  useEffect(() => {
    fetchItems();
  }, [fetchItems]);

  useEffect(() => {
    fetchOrders();
  }, [fetchOrders]);

  const onAddItem = async (values: ItemForm) => {
    await api.post('/api/v1/uniforms/items', values);
    reset();
    setIsAddOpen(false);
    fetchItems();
  };

  const tabs = [
    { key: 'items' as TabKey, label: 'Daftar Item', count: items.length },
    { key: 'orders' as TabKey, label: 'Pesanan', count: orders.length },
  ];

  const itemColumns = [
    {
      key: 'name',
      header: 'Nama',
      render: (item: UniformItem) => (
        <p className="font-medium text-gray-900">{item.name}</p>
      ),
    },
    {
      key: 'sizes',
      header: 'Ukuran',
      render: (item: UniformItem) =>
        Array.isArray(item.sizes) ? item.sizes.join(', ') : '—',
    },
    {
      key: 'price',
      header: 'Harga',
      render: (item: UniformItem) => (
        <span className="font-semibold text-gray-900">{formatCurrency(item.price)}</span>
      ),
    },
    {
      key: 'isActive',
      header: 'Status',
      render: (item: UniformItem) => (
        <Badge variant={item.isActive ? 'success' : 'gray'} size="sm">
          {item.isActive ? 'Aktif' : 'Tidak Aktif'}
        </Badge>
      ),
    },
  ];

  const orderColumns = [
    {
      key: 'studentId',
      header: 'Siswa ID',
      render: (order: UniformOrder) => (
        <span className="font-mono text-sm text-gray-700">{order.studentId || '—'}</span>
      ),
    },
    {
      key: 'itemName',
      header: 'Item',
      render: (order: UniformOrder) => order.itemName || '—',
    },
    { key: 'size', header: 'Ukuran' },
    { key: 'qty', header: 'Qty' },
    {
      key: 'totalPrice',
      header: 'Total',
      render: (order: UniformOrder) => (
        <span className="font-semibold text-gray-900">{formatCurrency(order.totalPrice)}</span>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (order: UniformOrder) => {
        const map = ORDER_STATUS_MAP[order.status] ?? { label: order.status, variant: 'gray' as const };
        return <Badge variant={map.variant} size="sm">{map.label}</Badge>;
      },
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="p-2 bg-primary-100 rounded-lg">
            <Shirt className="w-5 h-5 text-primary-600" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-gray-900">Seragam & Atribut</h1>
            <p className="text-sm text-gray-500 mt-0.5">Kelola item seragam dan pesanan siswa</p>
          </div>
        </div>
        {activeTab === 'items' && (
          <Button
            size="sm"
            leftIcon={<Plus className="w-3.5 h-3.5" />}
            onClick={() => setIsAddOpen(true)}
          >
            Tambah Item
          </Button>
        )}
      </div>

      {/* Tabs */}
      <div className="flex gap-2 border-b border-gray-200">
        {tabs.map((tab) => (
          <button
            key={tab.key}
            onClick={() => setActiveTab(tab.key)}
            className={`flex items-center gap-2 px-4 py-3 text-sm font-medium border-b-2 transition-colors -mb-px ${
              activeTab === tab.key
                ? 'border-primary-600 text-primary-600'
                : 'border-transparent text-gray-500 hover:text-gray-700'
            }`}
          >
            {tab.label}
            <span
              className={`text-xs px-1.5 py-0.5 rounded-full ${
                activeTab === tab.key
                  ? 'bg-primary-100 text-primary-700'
                  : 'bg-gray-100 text-gray-500'
              }`}
            >
              {tab.count}
            </span>
          </button>
        ))}
      </div>

      {/* Items */}
      {activeTab === 'items' && (
        <Card padding="none">
          <Table
            columns={itemColumns}
            data={items}
            keyExtractor={(item) => item.id}
            isLoading={itemLoading}
            emptyMessage="Belum ada item seragam"
          />
          {itemMeta.total > 0 && (
            <div className="px-4 py-3 border-t border-gray-100">
              <Pagination
                currentPage={itemPage}
                totalPages={itemMeta.totalPages}
                total={itemMeta.total}
                limit={20}
                onPageChange={setItemPage}
              />
            </div>
          )}
        </Card>
      )}

      {/* Orders */}
      {activeTab === 'orders' && (
        <Card padding="none">
          <Table
            columns={orderColumns}
            data={orders}
            keyExtractor={(order) => order.id}
            isLoading={orderLoading}
            emptyMessage="Belum ada pesanan seragam"
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

      {/* Add Item Modal */}
      <Modal
        isOpen={isAddOpen}
        onClose={() => { setIsAddOpen(false); reset(); }}
        title="Tambah Item Seragam"
        size="md"
        footer={
          <>
            <Button variant="secondary" onClick={() => { setIsAddOpen(false); reset(); }}>
              Batal
            </Button>
            <Button form="uniform-item-form" type="submit" isLoading={isSubmitting}>
              Simpan
            </Button>
          </>
        }
      >
        <form id="uniform-item-form" onSubmit={handleSubmit(onAddItem)} className="space-y-4">
          <Input
            label="Nama Item"
            placeholder="Contoh: Kemeja Putih Lengan Panjang"
            required
            {...register('name')}
            error={errors.name?.message}
          />
          <div className="w-full">
            <label className="block text-sm font-medium text-gray-700 mb-1">Deskripsi</label>
            <textarea
              rows={2}
              placeholder="Deskripsi item seragam (opsional)"
              className="block w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent resize-none"
              {...register('description')}
            />
          </div>
          <Input
            label="Harga (Rp)"
            type="number"
            min={0}
            required
            leftAddon="Rp"
            {...register('price', { valueAsNumber: true })}
            error={errors.price?.message}
          />
          <div className="w-full">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Ukuran Tersedia <span className="text-red-500 ml-1">*</span>
            </label>
            <div className="flex flex-wrap gap-2">
              {ALL_SIZES.map((size) => (
                <button
                  key={size}
                  type="button"
                  onClick={() => toggleSize(size)}
                  className={`px-4 py-2 rounded-lg border text-sm font-medium transition-colors ${
                    selectedSizes.includes(size)
                      ? 'bg-primary-600 border-primary-600 text-white'
                      : 'border-gray-300 text-gray-700 hover:bg-gray-50'
                  }`}
                >
                  {size}
                </button>
              ))}
            </div>
            {errors.sizes && (
              <p className="mt-1 text-xs text-red-600">{errors.sizes.message}</p>
            )}
          </div>
        </form>
      </Modal>
    </div>
  );
}
