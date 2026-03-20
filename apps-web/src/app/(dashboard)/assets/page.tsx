'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Plus, Package, Search } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import api from '@/lib/api';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { Pagination } from '@/components/ui/Pagination';
import { Modal } from '@/components/ui/Modal';
import { formatDate, formatCurrency } from '@/lib/utils';

type AssetCondition = 'GOOD' | 'DAMAGED' | 'LOST';

interface Asset {
  id: string;
  name: string;
  category: string;
  condition: AssetCondition;
  location: string;
  purchaseDate: string;
  purchasePrice: number;
}

const CONDITION_MAP: Record<AssetCondition, { label: string; variant: 'success' | 'warning' | 'danger' }> = {
  GOOD: { label: 'Baik', variant: 'success' },
  DAMAGED: { label: 'Rusak', variant: 'warning' },
  LOST: { label: 'Hilang', variant: 'danger' },
};

export default function AssetsPage() {
  const { user } = useAuth();
  const [assets, setAssets] = useState<Asset[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [conditionFilter, setConditionFilter] = useState('');
  const [page, setPage] = useState(1);
  const [meta, setMeta] = useState({ total: 0, totalPages: 1 });
  const [isAddOpen, setIsAddOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const [form, setForm] = useState({
    name: '',
    category: '',
    condition: 'GOOD' as AssetCondition,
    location: '',
    purchaseDate: '',
    purchasePrice: '',
  });

  const canManage = ['ADMIN_SEKOLAH', 'TATA_USAHA', 'BENDAHARA'].includes(user?.role || '');

  const fetchData = useCallback(async () => {
    setIsLoading(true);
    try {
      const params = new URLSearchParams({
        page: String(page),
        limit: '20',
        ...(search && { search }),
        ...(conditionFilter && { condition: conditionFilter }),
      });
      const res = await api.get(`/assets?${params}`);
      setAssets(res.data.data || []);
      setMeta(res.data.meta || { total: 0, totalPages: 1 });
    } catch {
      setAssets([]);
    } finally {
      setIsLoading(false);
    }
  }, [page, search, conditionFilter]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const handleSubmit = async () => {
    if (!form.name || !form.category) return;
    setIsSubmitting(true);
    try {
      await api.post('/assets', {
        ...form,
        purchasePrice: form.purchasePrice ? Number(form.purchasePrice) : undefined,
      });
      setIsAddOpen(false);
      setForm({ name: '', category: '', condition: 'GOOD', location: '', purchaseDate: '', purchasePrice: '' });
      fetchData();
    } catch {
      // handle
    } finally {
      setIsSubmitting(false);
    }
  };

  const CATEGORIES = [
    'Furnitur', 'Elektronik', 'Kendaraan', 'Peralatan Lab',
    'Peralatan Olahraga', 'Buku & Media', 'Lainnya',
  ];

  const columns = [
    {
      key: 'name',
      header: 'Nama Aset',
      render: (a: Asset) => (
        <div>
          <p className="font-medium text-gray-900">{a.name}</p>
          <p className="text-xs text-gray-400">{a.category}</p>
        </div>
      ),
    },
    {
      key: 'condition',
      header: 'Kondisi',
      render: (a: Asset) => {
        const c = CONDITION_MAP[a.condition] || { label: a.condition, variant: 'default' as const };
        return <Badge variant={c.variant}>{c.label}</Badge>;
      },
    },
    {
      key: 'location',
      header: 'Lokasi',
      render: (a: Asset) => (
        <span className="text-sm text-gray-700">{a.location || '—'}</span>
      ),
    },
    {
      key: 'purchaseDate',
      header: 'Tgl Beli',
      render: (a: Asset) => (
        <span className="text-sm text-gray-500">{a.purchaseDate ? formatDate(a.purchaseDate) : '—'}</span>
      ),
    },
    {
      key: 'purchasePrice',
      header: 'Harga Beli',
      render: (a: Asset) => (
        <span className="text-sm font-medium text-gray-900">
          {a.purchasePrice ? formatCurrency(a.purchasePrice) : '—'}
        </span>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900">Manajemen Aset</h1>
          <p className="text-sm text-gray-500 mt-0.5">Kelola inventaris dan aset sekolah</p>
        </div>
        {canManage && (
          <Button size="sm" leftIcon={<Plus className="w-3.5 h-3.5" />} onClick={() => setIsAddOpen(true)}>
            Tambah Aset
          </Button>
        )}
      </div>

      <Card padding="none">
        <div className="p-4 border-b border-gray-100 flex flex-wrap gap-3 items-center">
          <Input
            placeholder="Cari nama aset..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1); }}
            leftAddon={<Search className="w-4 h-4" />}
            className="max-w-xs"
          />
          <select
            value={conditionFilter}
            onChange={(e) => { setConditionFilter(e.target.value); setPage(1); }}
            className="text-sm border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500"
          >
            <option value="">Semua Kondisi</option>
            <option value="GOOD">Baik</option>
            <option value="DAMAGED">Rusak</option>
            <option value="LOST">Hilang</option>
          </select>
        </div>

        <Table
          columns={columns}
          data={assets}
          keyExtractor={(a) => a.id}
          isLoading={isLoading}
          emptyMessage="Belum ada data aset"
        />

        {meta.total > 0 && (
          <div className="px-4 py-3 border-t border-gray-100">
            <Pagination
              currentPage={page}
              totalPages={meta.totalPages}
              total={meta.total}
              limit={20}
              onPageChange={setPage}
            />
          </div>
        )}
      </Card>

      <Modal
        isOpen={isAddOpen}
        onClose={() => setIsAddOpen(false)}
        title="Tambah Aset Baru"
        description="Catat aset baru ke inventaris sekolah"
        size="md"
        footer={
          <>
            <Button variant="secondary" onClick={() => setIsAddOpen(false)}>Batal</Button>
            <Button
              onClick={handleSubmit}
              isLoading={isSubmitting}
              leftIcon={<Package className="w-4 h-4" />}
            >
              Simpan
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Nama Aset <span className="text-red-500">*</span></label>
            <input
              className="w-full text-sm border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500"
              placeholder="Contoh: Proyektor Epson EB-X41"
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
            />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Kategori <span className="text-red-500">*</span></label>
              <select
                className="w-full text-sm border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500"
                value={form.category}
                onChange={(e) => setForm({ ...form, category: e.target.value })}
              >
                <option value="">Pilih kategori</option>
                {CATEGORIES.map((c) => (
                  <option key={c} value={c}>{c}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Kondisi</label>
              <select
                className="w-full text-sm border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500"
                value={form.condition}
                onChange={(e) => setForm({ ...form, condition: e.target.value as AssetCondition })}
              >
                <option value="GOOD">Baik</option>
                <option value="DAMAGED">Rusak</option>
                <option value="LOST">Hilang</option>
              </select>
            </div>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Lokasi</label>
            <input
              className="w-full text-sm border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500"
              placeholder="Contoh: Ruang Kelas 9A, Lab Komputer"
              value={form.location}
              onChange={(e) => setForm({ ...form, location: e.target.value })}
            />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Tanggal Pembelian</label>
              <input
                type="date"
                className="w-full text-sm border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500"
                value={form.purchaseDate}
                onChange={(e) => setForm({ ...form, purchaseDate: e.target.value })}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Harga Beli (Rp)</label>
              <input
                type="number"
                className="w-full text-sm border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500"
                placeholder="0"
                min={0}
                value={form.purchasePrice}
                onChange={(e) => setForm({ ...form, purchasePrice: e.target.value })}
              />
            </div>
          </div>
        </div>
      </Modal>
    </div>
  );
}
