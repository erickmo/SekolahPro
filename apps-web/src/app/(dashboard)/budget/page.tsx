'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Plus, DollarSign, TrendingUp, TrendingDown } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import api from '@/lib/api';
import { Card, StatCard } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { Pagination } from '@/components/ui/Pagination';
import { Modal } from '@/components/ui/Modal';
import { formatCurrency } from '@/lib/utils';

type FundSource = 'BOS' | 'SPP' | 'KOMITE' | 'OTHER';
type BudgetStatus = 'DRAFT' | 'APPROVED' | 'ACTIVE' | 'CLOSED';

interface BudgetPlan {
  id: string;
  fiscalYear: number;
  fundSource: FundSource;
  totalAmount: number;
  totalRealized: number;
  status: BudgetStatus;
}

interface BudgetSummary {
  totalPlanned: number;
  totalRealized: number;
  activePlans: number;
}

const STATUS_MAP: Record<BudgetStatus, { label: string; variant: 'gray' | 'primary' | 'success' | 'warning' }> = {
  DRAFT: { label: 'Draft', variant: 'gray' },
  APPROVED: { label: 'Disetujui', variant: 'primary' },
  ACTIVE: { label: 'Aktif', variant: 'success' },
  CLOSED: { label: 'Ditutup', variant: 'warning' },
};

const FUND_SOURCE_LABELS: Record<FundSource, string> = {
  BOS: 'Dana BOS',
  SPP: 'SPP',
  KOMITE: 'Dana Komite',
  OTHER: 'Lainnya',
};

export default function BudgetPage() {
  const { user } = useAuth();
  const [plans, setPlans] = useState<BudgetPlan[]>([]);
  const [summary, setSummary] = useState<BudgetSummary | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [meta, setMeta] = useState({ total: 0, totalPages: 1 });
  const [isAddOpen, setIsAddOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const [form, setForm] = useState({
    fiscalYear: String(new Date().getFullYear()),
    fundSource: 'BOS' as FundSource,
    totalAmount: '',
  });

  const canManage = ['ADMIN_SEKOLAH', 'BENDAHARA', 'KEPALA_SEKOLAH'].includes(user?.role || '');

  const fetchData = useCallback(async () => {
    setIsLoading(true);
    try {
      const params = new URLSearchParams({ page: String(page), limit: '20' });
      const [plansRes] = await Promise.all([
        api.get(`/budget/plans?${params}`),
      ]);
      setPlans(plansRes.data.data || []);
      setMeta(plansRes.data.meta || { total: 0, totalPages: 1 });

      try {
        const sumRes = await api.get('/budget/summary');
        setSummary(sumRes.data.data);
      } catch {
        // summary optional
      }
    } catch {
      setPlans([]);
    } finally {
      setIsLoading(false);
    }
  }, [page]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const handleSubmit = async () => {
    if (!form.fiscalYear || !form.totalAmount) return;
    setIsSubmitting(true);
    try {
      await api.post('/budget/plans', {
        fiscalYear: Number(form.fiscalYear),
        fundSource: form.fundSource,
        totalAmount: Number(form.totalAmount),
      });
      setIsAddOpen(false);
      setForm({ fiscalYear: String(new Date().getFullYear()), fundSource: 'BOS', totalAmount: '' });
      fetchData();
    } catch {
      // handle
    } finally {
      setIsSubmitting(false);
    }
  };

  const realizationPercent = (plan: BudgetPlan) => {
    if (!plan.totalAmount) return 0;
    return Math.min(100, Math.round((plan.totalRealized / plan.totalAmount) * 100));
  };

  const columns = [
    {
      key: 'fiscalYear',
      header: 'Tahun Anggaran',
      render: (p: BudgetPlan) => (
        <span className="font-semibold text-gray-900">{p.fiscalYear}</span>
      ),
    },
    {
      key: 'fundSource',
      header: 'Sumber Dana',
      render: (p: BudgetPlan) => (
        <Badge variant="info">{FUND_SOURCE_LABELS[p.fundSource] || p.fundSource}</Badge>
      ),
    },
    {
      key: 'totalAmount',
      header: 'Total Direncanakan',
      render: (p: BudgetPlan) => (
        <span className="text-sm font-semibold text-gray-900">{formatCurrency(p.totalAmount)}</span>
      ),
    },
    {
      key: 'realization',
      header: 'Realisasi',
      render: (p: BudgetPlan) => {
        const pct = realizationPercent(p);
        return (
          <div className="space-y-1">
            <div className="flex justify-between text-xs">
              <span className="text-gray-500">{formatCurrency(p.totalRealized || 0)}</span>
              <span className="font-medium text-gray-700">{pct}%</span>
            </div>
            <div className="w-full bg-gray-100 rounded-full h-1.5">
              <div
                className={`h-1.5 rounded-full ${pct >= 80 ? 'bg-green-500' : pct >= 50 ? 'bg-amber-500' : 'bg-primary-500'}`}
                style={{ width: `${pct}%` }}
              />
            </div>
          </div>
        );
      },
    },
    {
      key: 'status',
      header: 'Status',
      render: (p: BudgetPlan) => {
        const s = STATUS_MAP[p.status] || { label: p.status, variant: 'default' as const };
        return <Badge variant={s.variant}>{s.label}</Badge>;
      },
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900">Anggaran & RKAS</h1>
          <p className="text-sm text-gray-500 mt-0.5">Rencana Kegiatan dan Anggaran Sekolah</p>
        </div>
        {canManage && (
          <Button size="sm" leftIcon={<Plus className="w-3.5 h-3.5" />} onClick={() => setIsAddOpen(true)}>
            Tambah Rencana
          </Button>
        )}
      </div>

      {/* Summary Cards */}
      {summary && (
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <StatCard
            title="Total Direncanakan"
            value={formatCurrency(summary.totalPlanned)}
            icon={<DollarSign className="w-5 h-5 text-primary-600" />}
            iconBg="bg-primary-50"
          />
          <StatCard
            title="Total Terealisasi"
            value={formatCurrency(summary.totalRealized)}
            icon={<TrendingUp className="w-5 h-5 text-green-600" />}
            iconBg="bg-green-50"
            subtitle={
              summary.totalPlanned
                ? `${Math.round((summary.totalRealized / summary.totalPlanned) * 100)}% dari rencana`
                : undefined
            }
          />
          <StatCard
            title="Sisa Anggaran"
            value={formatCurrency(Math.max(0, summary.totalPlanned - summary.totalRealized))}
            icon={<TrendingDown className="w-5 h-5 text-amber-600" />}
            iconBg="bg-amber-50"
            subtitle={`${summary.activePlans ?? 0} rencana aktif`}
          />
        </div>
      )}

      <Card padding="none">
        <Table
          columns={columns}
          data={plans}
          keyExtractor={(p) => p.id}
          isLoading={isLoading}
          emptyMessage="Belum ada rencana anggaran"
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
        title="Tambah Rencana Anggaran"
        description="Buat rencana anggaran untuk tahun fiskal tertentu"
        size="sm"
        footer={
          <>
            <Button variant="secondary" onClick={() => setIsAddOpen(false)}>Batal</Button>
            <Button
              onClick={handleSubmit}
              isLoading={isSubmitting}
              leftIcon={<DollarSign className="w-4 h-4" />}
            >
              Simpan
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Tahun Anggaran <span className="text-red-500">*</span></label>
            <input
              type="number"
              className="w-full text-sm border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500"
              placeholder={String(new Date().getFullYear())}
              min={2020}
              max={2099}
              value={form.fiscalYear}
              onChange={(e) => setForm({ ...form, fiscalYear: e.target.value })}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Sumber Dana</label>
            <select
              className="w-full text-sm border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500"
              value={form.fundSource}
              onChange={(e) => setForm({ ...form, fundSource: e.target.value as FundSource })}
            >
              <option value="BOS">Dana BOS</option>
              <option value="SPP">SPP</option>
              <option value="KOMITE">Dana Komite</option>
              <option value="OTHER">Lainnya</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Total Anggaran (Rp) <span className="text-red-500">*</span></label>
            <input
              type="number"
              className="w-full text-sm border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500"
              placeholder="0"
              min={0}
              value={form.totalAmount}
              onChange={(e) => setForm({ ...form, totalAmount: e.target.value })}
            />
          </div>
        </div>
      </Modal>
    </div>
  );
}
