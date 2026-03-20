'use client';

import React, { useState, useEffect, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { Search, PiggyBank, TrendingUp, ShoppingBag } from 'lucide-react';
import api from '@/lib/api';
import { useAuth } from '@/context/AuthContext';
import type { Student, CoopProduct } from '@/types';
import { Card, CardHeader, StatCard } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Table } from '@/components/ui/Table';
import { formatCurrency } from '@/lib/utils';

export default function CooperativePage() {
  const { user } = useAuth();
  const router = useRouter();
  const [searchQuery, setSearchQuery] = useState('');
  const [students, setStudents] = useState<Student[]>([]);
  const [products, setProducts] = useState<CoopProduct[]>([]);
  const [dailySummary, setDailySummary] = useState<{
    totalDeposits: number;
    totalWithdrawals: number;
    totalTransactions: number;
  } | null>(null);
  const [isSearching, setIsSearching] = useState(false);
  const [isLoading, setIsLoading] = useState(false);

  const canManage = ['KASIR_KOPERASI', 'BENDAHARA', 'ADMIN_SEKOLAH'].includes(user?.role || '');

  const fetchProducts = useCallback(async () => {
    setIsLoading(true);
    try {
      const [pRes, dRes] = await Promise.all([
        api.get('/cooperative/products'),
        canManage ? api.get('/cooperative/reports/daily') : Promise.resolve(null),
      ]);
      setProducts(pRes.data.data || []);
      if (dRes) setDailySummary(dRes.data.data);
    } finally {
      setIsLoading(false);
    }
  }, [canManage]);

  useEffect(() => {
    fetchProducts();
  }, [fetchProducts]);

  const handleSearch = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!searchQuery.trim()) return;
    setIsSearching(true);
    try {
      const res = await api.get(`/students?search=${encodeURIComponent(searchQuery)}&limit=10`);
      setStudents(res.data.data || []);
    } finally {
      setIsSearching(false);
    }
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-bold text-gray-900">Koperasi & Tabungan</h1>
        <p className="text-sm text-gray-500 mt-0.5">Kelola tabungan siswa dan produk koperasi</p>
      </div>

      {/* Daily summary */}
      {canManage && dailySummary && (
        <div className="grid grid-cols-3 gap-4">
          <StatCard
            title="Setoran Hari Ini"
            value={formatCurrency(dailySummary.totalDeposits)}
            icon={<TrendingUp className="w-5 h-5 text-green-600" />}
            iconBg="bg-green-50"
          />
          <StatCard
            title="Penarikan Hari Ini"
            value={formatCurrency(dailySummary.totalWithdrawals)}
            icon={<PiggyBank className="w-5 h-5 text-red-600" />}
            iconBg="bg-red-50"
          />
          <StatCard
            title="Total Transaksi"
            value={dailySummary.totalTransactions}
            icon={<ShoppingBag className="w-5 h-5 text-primary-600" />}
            iconBg="bg-primary-50"
          />
        </div>
      )}

      {/* Search student */}
      <Card>
        <CardHeader title="Cari Tabungan Siswa" description="Masukkan nama atau NISN siswa untuk melihat tabungannya" />
        <form onSubmit={handleSearch} className="flex gap-3">
          <Input
            placeholder="Cari nama atau NISN siswa..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            leftAddon={<Search className="w-4 h-4" />}
            className="max-w-sm"
          />
          <Button type="submit" isLoading={isSearching}>Cari</Button>
        </form>

        {students.length > 0 && (
          <div className="mt-4">
            <Table
              columns={[
                { key: 'name', header: 'Nama Siswa', render: (s) => (
                  <div>
                    <p className="font-medium">{s.name}</p>
                    <p className="text-xs text-gray-400">NISN: {s.nisn}</p>
                  </div>
                )},
                { key: 'currentClass', header: 'Kelas', render: (s) => s.currentClass || '—' },
                { key: 'actions', header: '', render: (s) => (
                  <Button size="sm" variant="outline" onClick={() => router.push(`/dashboard/cooperative/${s.id}`)}>
                    Lihat Tabungan
                  </Button>
                ), className: 'w-36' },
              ]}
              data={students}
              keyExtractor={(s) => s.id}
              emptyMessage="Tidak ada siswa"
            />
          </div>
        )}
      </Card>

      {/* Products */}
      <Card padding="none">
        <div className="px-6 py-4 border-b border-gray-100">
          <h3 className="font-semibold text-sm text-gray-900">Produk Koperasi</h3>
        </div>
        {isLoading ? (
          <div className="p-8 text-center text-gray-400 text-sm">Memuat produk...</div>
        ) : products.length === 0 ? (
          <div className="p-8 text-center text-gray-400 text-sm">Belum ada produk koperasi</div>
        ) : (
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4 p-4">
            {products.map((p) => (
              <div key={p.id} className="border border-gray-100 rounded-xl p-4 hover:border-primary-200 transition-colors">
                {p.imageUrl ? (
                  <img src={p.imageUrl} alt={p.name} className="w-full h-28 object-cover rounded-lg mb-3" />
                ) : (
                  <div className="w-full h-28 bg-gray-100 rounded-lg mb-3 flex items-center justify-center">
                    <ShoppingBag className="w-8 h-8 text-gray-300" />
                  </div>
                )}
                <p className="font-medium text-sm text-gray-900">{p.name}</p>
                <p className="text-xs text-gray-400 mt-0.5">{p.category}</p>
                <p className="font-bold text-primary-700 mt-2">{formatCurrency(p.price)}</p>
                <p className="text-xs text-gray-400">Stok: {p.stock}</p>
              </div>
            ))}
          </div>
        )}
      </Card>
    </div>
  );
}
