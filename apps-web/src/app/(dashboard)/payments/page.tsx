'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Plus, Search, Download, TrendingUp, CreditCard, Clock, AlertCircle } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import api from '@/lib/api';
import type { Invoice, FinancialSummary } from '@/types';
import { Card, CardHeader, StatCard } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Table } from '@/components/ui/Table';
import { InvoiceStatusBadge } from '@/components/ui/Badge';
import { Pagination } from '@/components/ui/Pagination';
import { Modal } from '@/components/ui/Modal';
import { formatCurrency, formatDate } from '@/lib/utils';

export default function PaymentsPage() {
  const { user } = useAuth();
  const [invoices, setInvoices] = useState<Invoice[]>([]);
  const [summary, setSummary] = useState<FinancialSummary | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [page, setPage] = useState(1);
  const [meta, setMeta] = useState({ total: 0, totalPages: 1 });
  const [generateOpen, setGenerateOpen] = useState(false);
  const [genMonth, setGenMonth] = useState('');
  const [genYear, setGenYear] = useState(String(new Date().getFullYear()));
  const [genAmount, setGenAmount] = useState('');
  const [isGenerating, setIsGenerating] = useState(false);

  const canManage = ['ADMIN_SEKOLAH', 'BENDAHARA'].includes(user?.role || '');
  const canViewSummary = ['ADMIN_SEKOLAH', 'BENDAHARA', 'KEPALA_SEKOLAH'].includes(user?.role || '');

  const fetchData = useCallback(async () => {
    setIsLoading(true);
    try {
      const params = new URLSearchParams({
        page: String(page),
        limit: '20',
        ...(search && { search }),
        ...(statusFilter && { status: statusFilter }),
      });
      const [invRes] = await Promise.all([
        api.get(`/payments/invoices?${params}`),
      ]);
      setInvoices(invRes.data.data || []);
      setMeta(invRes.data.meta || { total: 0, totalPages: 1 });

      if (canViewSummary) {
        try {
          const sumRes = await api.get('/payments/financial-summary');
          setSummary(sumRes.data.data);
        } catch {
          // summary optional
        }
      }
    } catch {
      setInvoices([]);
    } finally {
      setIsLoading(false);
    }
  }, [page, search, statusFilter, canViewSummary]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleGenerate = async () => {
    if (!genMonth || !genYear || !genAmount) return;
    setIsGenerating(true);
    try {
      await api.post('/payments/invoices/generate', {
        month: Number(genMonth),
        year: Number(genYear),
        amount: Number(genAmount),
        type: 'SPP',
        description: `SPP ${MONTHS[Number(genMonth) - 1]} ${genYear}`,
      });
      setGenerateOpen(false);
      fetchData();
    } catch {
      // handle
    } finally {
      setIsGenerating(false);
    }
  };

  const handlePayNow = async (invoice: Invoice) => {
    try {
      const res = await api.post(`/payments/invoices/${invoice.id}/pay`);
      const { paymentUrl } = res.data.data || {};
      if (paymentUrl) {
        window.open(paymentUrl, '_blank');
      }
    } catch {
      // handle
    }
  };

  const columns = [
    {
      key: 'studentName',
      header: 'Siswa',
      render: (inv: Invoice) => (
        <div>
          <p className="font-medium text-gray-900">{inv.studentName || '—'}</p>
          <p className="text-xs text-gray-400">NISN: {inv.studentNisn || '—'}</p>
        </div>
      ),
    },
    {
      key: 'description',
      header: 'Tagihan',
      render: (inv: Invoice) => (
        <div>
          <p className="font-medium">{inv.description}</p>
          <p className="text-xs text-gray-400">{inv.type}</p>
        </div>
      ),
    },
    {
      key: 'amount',
      header: 'Jumlah',
      render: (inv: Invoice) => (
        <span className="font-semibold text-gray-900">{formatCurrency(inv.amount)}</span>
      ),
    },
    {
      key: 'dueDate',
      header: 'Jatuh Tempo',
      render: (inv: Invoice) => formatDate(inv.dueDate),
    },
    {
      key: 'status',
      header: 'Status',
      render: (inv: Invoice) => <InvoiceStatusBadge status={inv.status} />,
    },
    {
      key: 'actions',
      header: '',
      render: (inv: Invoice) =>
        inv.status === 'PENDING' || inv.status === 'OVERDUE' ? (
          <Button variant="outline" size="sm" onClick={() => handlePayNow(inv)}>
            Bayar
          </Button>
        ) : null,
      className: 'w-24',
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900">Pembayaran SPP</h1>
          <p className="text-sm text-gray-500 mt-0.5">Manajemen tagihan dan pembayaran sekolah</p>
        </div>
        {canManage && (
          <Button size="sm" leftIcon={<Plus className="w-3.5 h-3.5" />} onClick={() => setGenerateOpen(true)}>
            Generate Tagihan
          </Button>
        )}
      </div>

      {/* Summary cards */}
      {canViewSummary && summary && (
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          <StatCard
            title="Total Tagihan"
            value={formatCurrency(summary.totalInvoiced)}
            icon={<CreditCard className="w-5 h-5 text-primary-600" />}
            iconBg="bg-primary-50"
          />
          <StatCard
            title="Sudah Dibayar"
            value={formatCurrency(summary.totalPaid)}
            icon={<TrendingUp className="w-5 h-5 text-green-600" />}
            iconBg="bg-green-50"
          />
          <StatCard
            title="Menunggu"
            value={formatCurrency(summary.totalPending)}
            icon={<Clock className="w-5 h-5 text-amber-600" />}
            iconBg="bg-amber-50"
          />
          <StatCard
            title="Jatuh Tempo"
            value={formatCurrency(summary.totalOverdue)}
            icon={<AlertCircle className="w-5 h-5 text-red-600" />}
            iconBg="bg-red-50"
            subtitle={`${summary.collectionRate?.toFixed(1)}% collection rate`}
          />
        </div>
      )}

      <Card padding="none">
        <div className="p-4 border-b border-gray-100 flex flex-wrap gap-3 items-center">
          <Input
            placeholder="Cari nama siswa..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1); }}
            leftAddon={<Search className="w-4 h-4" />}
            className="max-w-xs"
          />
          <select
            value={statusFilter}
            onChange={(e) => { setStatusFilter(e.target.value); setPage(1); }}
            className="text-sm border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500"
          >
            <option value="">Semua Status</option>
            <option value="PENDING">Menunggu</option>
            <option value="PAID">Lunas</option>
            <option value="OVERDUE">Jatuh Tempo</option>
          </select>
          <div className="ml-auto">
            <Button variant="secondary" size="sm" leftIcon={<Download className="w-3.5 h-3.5" />}>
              Export
            </Button>
          </div>
        </div>

        <Table
          columns={columns}
          data={invoices}
          keyExtractor={(inv) => inv.id}
          isLoading={isLoading}
          emptyMessage="Tidak ada tagihan"
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

      {/* Generate Modal */}
      <Modal
        isOpen={generateOpen}
        onClose={() => setGenerateOpen(false)}
        title="Generate Tagihan SPP"
        description="Buat tagihan SPP untuk semua siswa aktif"
        size="sm"
        footer={
          <>
            <Button variant="secondary" onClick={() => setGenerateOpen(false)}>Batal</Button>
            <Button onClick={handleGenerate} isLoading={isGenerating}>Generate</Button>
          </>
        }
      >
        <div className="space-y-4">
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Bulan</label>
              <select
                value={genMonth}
                onChange={(e) => setGenMonth(e.target.value)}
                className="w-full text-sm border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500"
              >
                <option value="">Pilih bulan</option>
                {MONTHS.map((m, i) => (
                  <option key={m} value={i + 1}>{m}</option>
                ))}
              </select>
            </div>
            <Input
              label="Tahun"
              type="number"
              value={genYear}
              onChange={(e) => setGenYear(e.target.value)}
              min={2020}
              max={2030}
            />
          </div>
          <Input
            label="Jumlah SPP (Rp)"
            type="number"
            placeholder="500000"
            value={genAmount}
            onChange={(e) => setGenAmount(e.target.value)}
            leftAddon="Rp"
          />
        </div>
      </Modal>
    </div>
  );
}

const MONTHS = [
  'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
  'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember',
];
