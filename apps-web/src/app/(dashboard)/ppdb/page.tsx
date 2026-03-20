'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { UserPlus, Plus, RefreshCw } from 'lucide-react';
import api from '@/lib/api';
import { Card, CardHeader } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { Modal } from '@/components/ui/Modal';
import { Pagination } from '@/components/ui/Pagination';
import { Input } from '@/components/ui/Input';
import { formatDate } from '@/lib/utils';

type TabKey = 'periods' | 'applications';

const PERIOD_STATUS_VARIANT: Record<string, 'gray' | 'warning' | 'success' | 'danger' | 'info'> = {
  DRAFT: 'gray',
  OPEN: 'success',
  CLOSED: 'danger',
  ANNOUNCED: 'info',
};

const PERIOD_STATUS_LABEL: Record<string, string> = {
  DRAFT: 'Draf',
  OPEN: 'Buka',
  CLOSED: 'Tutup',
  ANNOUNCED: 'Diumumkan',
};

const APP_STATUS_VARIANT: Record<string, 'warning' | 'success' | 'danger' | 'info' | 'gray'> = {
  PENDING: 'warning',
  VERIFIED: 'info',
  ACCEPTED: 'success',
  REJECTED: 'danger',
  CANCELLED: 'gray',
};

const APP_STATUS_LABEL: Record<string, string> = {
  PENDING: 'Menunggu',
  VERIFIED: 'Terverifikasi',
  ACCEPTED: 'Diterima',
  REJECTED: 'Ditolak',
  CANCELLED: 'Dibatalkan',
};

interface PpdbPeriod {
  id: string;
  name: string;
  academicYear: string;
  quota: number;
  registrationStart: string;
  registrationEnd: string;
  announcementDate: string;
  status: string;
}

interface PpdbApplication {
  id: string;
  applicantName: string;
  nisn: string;
  status: string;
  createdAt: string;
  periodId: string;
}

interface AddPeriodForm {
  name: string;
  academicYear: string;
  quota: number | '';
  registrationStart: string;
  registrationEnd: string;
  announcementDate: string;
}

const EMPTY_FORM: AddPeriodForm = {
  name: '',
  academicYear: '',
  quota: '',
  registrationStart: '',
  registrationEnd: '',
  announcementDate: '',
};

export default function PpdbPage() {
  const [activeTab, setActiveTab] = useState<TabKey>('periods');
  const [periods, setPeriods] = useState<PpdbPeriod[]>([]);
  const [applications, setApplications] = useState<PpdbApplication[]>([]);
  const [isLoadingPeriods, setIsLoadingPeriods] = useState(true);
  const [isLoadingApps, setIsLoadingApps] = useState(false);
  const [periodsMeta, setPeriodsMeta] = useState({ total: 0, totalPages: 1 });
  const [appsMeta, setAppsMeta] = useState({ total: 0, totalPages: 1 });
  const [periodsPage, setPeriodsPage] = useState(1);
  const [appsPage, setAppsPage] = useState(1);
  const [isAddOpen, setIsAddOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [form, setForm] = useState<AddPeriodForm>(EMPTY_FORM);

  const fetchPeriods = useCallback(async () => {
    setIsLoadingPeriods(true);
    try {
      const params = new URLSearchParams({ page: String(periodsPage), limit: '20' });
      const res = await api.get(`/ppdb/periods?${params}`);
      setPeriods(res.data.data || []);
      setPeriodsMeta(res.data.meta || { total: 0, totalPages: 1 });
    } catch {
      setPeriods([]);
    } finally {
      setIsLoadingPeriods(false);
    }
  }, [periodsPage]);

  const fetchApplications = useCallback(async () => {
    setIsLoadingApps(true);
    try {
      const params = new URLSearchParams({ page: String(appsPage), limit: '20' });
      const res = await api.get(`/ppdb/applications?${params}`);
      setApplications(res.data.data || []);
      setAppsMeta(res.data.meta || { total: 0, totalPages: 1 });
    } catch {
      setApplications([]);
    } finally {
      setIsLoadingApps(false);
    }
  }, [appsPage]);

  useEffect(() => {
    fetchPeriods();
  }, [fetchPeriods]);

  useEffect(() => {
    if (activeTab === 'applications') {
      fetchApplications();
    }
  }, [activeTab, fetchApplications]);

  const handleAdd = async () => {
    setIsSubmitting(true);
    try {
      await api.post('/ppdb/periods', {
        ...form,
        quota: Number(form.quota),
      });
      setIsAddOpen(false);
      setForm(EMPTY_FORM);
      fetchPeriods();
    } catch {
      // silent
    } finally {
      setIsSubmitting(false);
    }
  };

  const periodColumns = [
    {
      key: 'name',
      header: 'Nama Periode',
      render: (p: PpdbPeriod) => (
        <div>
          <p className="font-medium text-gray-900">{p.name}</p>
          <p className="text-xs text-gray-400">TA: {p.academicYear}</p>
        </div>
      ),
    },
    {
      key: 'quota',
      header: 'Kuota',
      render: (p: PpdbPeriod) => <span className="text-sm text-gray-700">{p.quota}</span>,
    },
    {
      key: 'registration',
      header: 'Pendaftaran',
      render: (p: PpdbPeriod) => (
        <span className="text-sm text-gray-600">
          {formatDate(p.registrationStart)} — {formatDate(p.registrationEnd)}
        </span>
      ),
    },
    {
      key: 'announcementDate',
      header: 'Pengumuman',
      render: (p: PpdbPeriod) => (
        <span className="text-sm text-gray-500">{formatDate(p.announcementDate)}</span>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (p: PpdbPeriod) => (
        <Badge variant={PERIOD_STATUS_VARIANT[p.status] ?? 'default'} size="sm">
          {PERIOD_STATUS_LABEL[p.status] ?? p.status}
        </Badge>
      ),
    },
  ];

  const appColumns = [
    {
      key: 'applicantName',
      header: 'Nama Calon Siswa',
      render: (a: PpdbApplication) => (
        <div>
          <p className="font-medium text-gray-900">{a.applicantName}</p>
          <p className="text-xs text-gray-400">NISN: {a.nisn}</p>
        </div>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (a: PpdbApplication) => (
        <Badge variant={APP_STATUS_VARIANT[a.status] ?? 'default'} size="sm">
          {APP_STATUS_LABEL[a.status] ?? a.status}
        </Badge>
      ),
    },
    {
      key: 'createdAt',
      header: 'Tanggal Daftar',
      render: (a: PpdbApplication) => (
        <span className="text-sm text-gray-500">{formatDate(a.createdAt)}</span>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="p-2 bg-green-100 rounded-xl">
            <UserPlus className="w-5 h-5 text-green-600" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-gray-900">PPDB Online</h1>
            <p className="text-sm text-gray-500 mt-0.5">Penerimaan peserta didik baru</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="secondary"
            size="sm"
            leftIcon={<RefreshCw className="w-3.5 h-3.5" />}
            onClick={() => { fetchPeriods(); if (activeTab === 'applications') fetchApplications(); }}
          >
            Refresh
          </Button>
          {activeTab === 'periods' && (
            <Button size="sm" leftIcon={<Plus className="w-3.5 h-3.5" />} onClick={() => setIsAddOpen(true)}>
              Buat Periode
            </Button>
          )}
        </div>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 p-1 bg-gray-100 rounded-lg w-fit">
        {(['periods', 'applications'] as TabKey[]).map((tab) => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab)}
            className={`px-4 py-2 text-sm font-medium rounded-md transition-colors ${
              activeTab === tab
                ? 'bg-white text-gray-900 shadow-sm'
                : 'text-gray-500 hover:text-gray-700'
            }`}
          >
            {tab === 'periods' ? 'Periode PPDB' : 'Pendaftar'}
          </button>
        ))}
      </div>

      {/* Periods Tab */}
      {activeTab === 'periods' && (
        <Card padding="none">
          <div className="p-4 border-b border-gray-100">
            <CardHeader title="Periode Penerimaan" description={`${periodsMeta.total} periode tercatat`} />
          </div>
          <Table
            columns={periodColumns}
            data={periods}
            keyExtractor={(p) => p.id}
            isLoading={isLoadingPeriods}
            emptyMessage="Belum ada periode PPDB dibuat"
          />
          {periodsMeta.total > 0 && (
            <div className="px-4 py-3 border-t border-gray-100">
              <Pagination
                currentPage={periodsPage}
                totalPages={periodsMeta.totalPages}
                total={periodsMeta.total}
                limit={20}
                onPageChange={setPeriodsPage}
              />
            </div>
          )}
        </Card>
      )}

      {/* Applications Tab */}
      {activeTab === 'applications' && (
        <Card padding="none">
          <div className="p-4 border-b border-gray-100">
            <CardHeader title="Data Pendaftar" description={`${appsMeta.total} pendaftar tercatat`} />
          </div>
          <Table
            columns={appColumns}
            data={applications}
            keyExtractor={(a) => a.id}
            isLoading={isLoadingApps}
            emptyMessage="Belum ada pendaftar"
          />
          {appsMeta.total > 0 && (
            <div className="px-4 py-3 border-t border-gray-100">
              <Pagination
                currentPage={appsPage}
                totalPages={appsMeta.totalPages}
                total={appsMeta.total}
                limit={20}
                onPageChange={setAppsPage}
              />
            </div>
          )}
        </Card>
      )}

      {/* Add Period Modal */}
      <Modal
        isOpen={isAddOpen}
        onClose={() => { setIsAddOpen(false); setForm(EMPTY_FORM); }}
        title="Buat Periode PPDB"
        description="Buat periode penerimaan peserta didik baru"
        size="lg"
        footer={
          <>
            <Button variant="secondary" onClick={() => { setIsAddOpen(false); setForm(EMPTY_FORM); }}>Batal</Button>
            <Button onClick={handleAdd} isLoading={isSubmitting}>Simpan</Button>
          </>
        }
      >
        <div className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="col-span-2">
              <Input
                label="Nama Periode"
                required
                placeholder="Contoh: PPDB Tahun Ajaran 2026/2027"
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
              />
            </div>
            <Input
              label="Tahun Ajaran"
              required
              placeholder="2026/2027"
              value={form.academicYear}
              onChange={(e) => setForm({ ...form, academicYear: e.target.value })}
            />
            <Input
              label="Kuota"
              type="number"
              required
              placeholder="200"
              value={String(form.quota)}
              onChange={(e) => setForm({ ...form, quota: e.target.value === '' ? '' : Number(e.target.value) })}
            />
            <Input
              label="Mulai Pendaftaran"
              type="date"
              required
              value={form.registrationStart}
              onChange={(e) => setForm({ ...form, registrationStart: e.target.value })}
            />
            <Input
              label="Tutup Pendaftaran"
              type="date"
              required
              value={form.registrationEnd}
              onChange={(e) => setForm({ ...form, registrationEnd: e.target.value })}
            />
            <div className="col-span-2">
              <Input
                label="Tanggal Pengumuman"
                type="date"
                required
                value={form.announcementDate}
                onChange={(e) => setForm({ ...form, announcementDate: e.target.value })}
              />
            </div>
          </div>
        </div>
      </Modal>
    </div>
  );
}
