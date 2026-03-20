'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Plus, Heart, TrendingUp } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import api from '@/lib/api';
import { Card, StatCard } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { Pagination } from '@/components/ui/Pagination';
import { Modal } from '@/components/ui/Modal';
import { formatCurrency, formatDate } from '@/lib/utils';

interface Donation {
  id: string;
  donorName: string;
  amount: number;
  purpose: string;
  isAnonymous: boolean;
  createdAt: string;
}

interface DonationSummary {
  totalAmount: number;
  totalCount: number;
  monthlyAmount: number;
}

export default function DonationsPage() {
  const { user } = useAuth();
  const [donations, setDonations] = useState<Donation[]>([]);
  const [summary, setSummary] = useState<DonationSummary | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [meta, setMeta] = useState({ total: 0, totalPages: 1 });
  const [isAddOpen, setIsAddOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const [form, setForm] = useState({
    donorName: '',
    amount: '',
    purpose: '',
    isAnonymous: false,
  });

  const canManage = ['ADMIN_SEKOLAH', 'BENDAHARA', 'KEPALA_SEKOLAH'].includes(user?.role || '');

  const fetchData = useCallback(async () => {
    setIsLoading(true);
    try {
      const params = new URLSearchParams({ page: String(page), limit: '20' });
      const res = await api.get(`/donations?${params}`);
      setDonations(res.data.data || []);
      setMeta(res.data.meta || { total: 0, totalPages: 1 });

      try {
        const sumRes = await api.get('/donations/summary');
        setSummary(sumRes.data.data);
      } catch {
        // summary optional
      }
    } catch {
      setDonations([]);
    } finally {
      setIsLoading(false);
    }
  }, [page]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const handleSubmit = async () => {
    if (!form.amount || (!form.donorName && !form.isAnonymous)) return;
    setIsSubmitting(true);
    try {
      await api.post('/donations', {
        ...form,
        donorName: form.isAnonymous ? 'Anonim' : form.donorName,
        amount: Number(form.amount),
      });
      setIsAddOpen(false);
      setForm({ donorName: '', amount: '', purpose: '', isAnonymous: false });
      fetchData();
    } catch {
      // handle
    } finally {
      setIsSubmitting(false);
    }
  };

  const columns = [
    {
      key: 'donor',
      header: 'Donatur',
      render: (d: Donation) => (
        <div className="flex items-center gap-2">
          <span className="font-medium text-gray-900">
            {d.isAnonymous ? 'Anonim' : d.donorName}
          </span>
          {d.isAnonymous && (
            <Badge variant="gray" size="sm">Anonim</Badge>
          )}
        </div>
      ),
    },
    {
      key: 'amount',
      header: 'Jumlah',
      render: (d: Donation) => (
        <span className="text-sm font-semibold text-gray-900">{formatCurrency(d.amount)}</span>
      ),
    },
    {
      key: 'purpose',
      header: 'Tujuan / Keterangan',
      render: (d: Donation) => (
        <span className="text-sm text-gray-700">{d.purpose || '—'}</span>
      ),
    },
    {
      key: 'createdAt',
      header: 'Tanggal',
      render: (d: Donation) => (
        <span className="text-sm text-gray-500">{formatDate(d.createdAt)}</span>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900">Dana Komite & Donasi</h1>
          <p className="text-sm text-gray-500 mt-0.5">Catatan donasi dan sumbangan komite sekolah</p>
        </div>
        {canManage && (
          <Button size="sm" leftIcon={<Plus className="w-3.5 h-3.5" />} onClick={() => setIsAddOpen(true)}>
            Catat Donasi
          </Button>
        )}
      </div>

      {/* Summary Cards */}
      {summary && (
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <StatCard
            title="Total Donasi"
            value={formatCurrency(summary.totalAmount)}
            icon={<Heart className="w-5 h-5 text-red-500" />}
            iconBg="bg-red-50"
          />
          <StatCard
            title="Donasi Bulan Ini"
            value={formatCurrency(summary.monthlyAmount)}
            icon={<TrendingUp className="w-5 h-5 text-green-600" />}
            iconBg="bg-green-50"
          />
          <StatCard
            title="Jumlah Transaksi"
            value={`${summary.totalCount} donasi`}
            icon={<Heart className="w-5 h-5 text-primary-600" />}
            iconBg="bg-primary-50"
          />
        </div>
      )}

      <Card padding="none">
        <Table
          columns={columns}
          data={donations}
          keyExtractor={(d) => d.id}
          isLoading={isLoading}
          emptyMessage="Belum ada data donasi"
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
        title="Catat Donasi"
        description="Tambahkan catatan donasi atau sumbangan"
        size="sm"
        footer={
          <>
            <Button variant="secondary" onClick={() => setIsAddOpen(false)}>Batal</Button>
            <Button
              onClick={handleSubmit}
              isLoading={isSubmitting}
              leftIcon={<Heart className="w-4 h-4" />}
            >
              Simpan
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          <div className="flex items-center gap-2">
            <input
              id="anonymous"
              type="checkbox"
              className="w-4 h-4 text-primary-600 border-gray-300 rounded focus:ring-primary-500"
              checked={form.isAnonymous}
              onChange={(e) => setForm({ ...form, isAnonymous: e.target.checked, donorName: e.target.checked ? '' : form.donorName })}
            />
            <label htmlFor="anonymous" className="text-sm font-medium text-gray-700">Donasi Anonim</label>
          </div>
          {!form.isAnonymous && (
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Nama Donatur <span className="text-red-500">*</span></label>
              <input
                className="w-full text-sm border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500"
                placeholder="Nama lengkap donatur"
                value={form.donorName}
                onChange={(e) => setForm({ ...form, donorName: e.target.value })}
              />
            </div>
          )}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Jumlah (Rp) <span className="text-red-500">*</span></label>
            <input
              type="number"
              className="w-full text-sm border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500"
              placeholder="0"
              min={0}
              value={form.amount}
              onChange={(e) => setForm({ ...form, amount: e.target.value })}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Tujuan / Keterangan</label>
            <input
              className="w-full text-sm border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500"
              placeholder="Contoh: Pembangunan musholla, Dana kegiatan"
              value={form.purpose}
              onChange={(e) => setForm({ ...form, purpose: e.target.value })}
            />
          </div>
        </div>
      </Modal>
    </div>
  );
}
