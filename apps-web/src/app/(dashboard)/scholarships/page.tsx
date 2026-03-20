'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Plus, Award, Search } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import api from '@/lib/api';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { Pagination } from '@/components/ui/Pagination';
import { Modal } from '@/components/ui/Modal';
import { formatCurrency, formatDate } from '@/lib/utils';

type TabKey = 'programs' | 'applications';

interface ScholarshipProgram {
  id: string;
  name: string;
  type: string;
  amount: number;
  quota: number;
  requirements: string;
  isActive: boolean;
}

interface ScholarshipApplication {
  id: string;
  studentName: string;
  programName: string;
  status: string;
  appliedAt: string;
}

const APP_STATUS_MAP: Record<string, { label: string; variant: 'warning' | 'success' | 'danger' | 'gray' }> = {
  PENDING: { label: 'Menunggu', variant: 'warning' },
  APPROVED: { label: 'Diterima', variant: 'success' },
  REJECTED: { label: 'Ditolak', variant: 'danger' },
  CANCELLED: { label: 'Dibatalkan', variant: 'gray' },
};

export default function ScholarshipsPage() {
  const { user } = useAuth();
  const [activeTab, setActiveTab] = useState<TabKey>('programs');
  const [programs, setPrograms] = useState<ScholarshipProgram[]>([]);
  const [applications, setApplications] = useState<ScholarshipApplication[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [page, setPage] = useState(1);
  const [meta, setMeta] = useState({ total: 0, totalPages: 1 });
  const [isAddOpen, setIsAddOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const [form, setForm] = useState({
    name: '',
    type: '',
    amount: '',
    quota: '',
    requirements: '',
  });

  const canManage = ['ADMIN_SEKOLAH', 'BENDAHARA', 'KEPALA_SEKOLAH'].includes(user?.role || '');

  const fetchPrograms = useCallback(async () => {
    setIsLoading(true);
    try {
      const params = new URLSearchParams({
        page: String(page),
        limit: '20',
        ...(search && { search }),
      });
      const res = await api.get(`/scholarships/programs?${params}`);
      setPrograms(res.data.data || []);
      setMeta(res.data.meta || { total: 0, totalPages: 1 });
    } catch {
      setPrograms([]);
    } finally {
      setIsLoading(false);
    }
  }, [page, search]);

  const fetchApplications = useCallback(async () => {
    setIsLoading(true);
    try {
      const params = new URLSearchParams({ page: String(page), limit: '20' });
      const res = await api.get(`/scholarships/applications?${params}`);
      setApplications(res.data.data || []);
      setMeta(res.data.meta || { total: 0, totalPages: 1 });
    } catch {
      setApplications([]);
    } finally {
      setIsLoading(false);
    }
  }, [page]);

  useEffect(() => {
    if (activeTab === 'programs') fetchPrograms();
    else fetchApplications();
  }, [activeTab, fetchPrograms, fetchApplications]);

  const handleSubmit = async () => {
    if (!form.name || !form.type || !form.amount) return;
    setIsSubmitting(true);
    try {
      await api.post('/scholarships/programs', {
        ...form,
        amount: Number(form.amount),
        quota: form.quota ? Number(form.quota) : undefined,
      });
      setIsAddOpen(false);
      setForm({ name: '', type: '', amount: '', quota: '', requirements: '' });
      fetchPrograms();
    } catch {
      // handle
    } finally {
      setIsSubmitting(false);
    }
  };

  const programColumns = [
    {
      key: 'name',
      header: 'Program',
      render: (p: ScholarshipProgram) => (
        <div>
          <p className="font-medium text-gray-900">{p.name}</p>
          <p className="text-xs text-gray-400 line-clamp-1">{p.requirements}</p>
        </div>
      ),
    },
    {
      key: 'type',
      header: 'Tipe',
      render: (p: ScholarshipProgram) => (
        <Badge variant="primary">{p.type}</Badge>
      ),
    },
    {
      key: 'amount',
      header: 'Nilai',
      render: (p: ScholarshipProgram) => (
        <span className="text-sm font-semibold text-gray-900">{formatCurrency(p.amount)}</span>
      ),
    },
    {
      key: 'quota',
      header: 'Kuota',
      render: (p: ScholarshipProgram) => (
        <span className="text-sm text-gray-700">{p.quota ?? '—'} siswa</span>
      ),
    },
    {
      key: 'isActive',
      header: 'Status',
      render: (p: ScholarshipProgram) => (
        <Badge variant={p.isActive ? 'success' : 'gray'}>{p.isActive ? 'Aktif' : 'Nonaktif'}</Badge>
      ),
    },
  ];

  const applicationColumns = [
    {
      key: 'student',
      header: 'Siswa',
      render: (a: ScholarshipApplication) => (
        <span className="font-medium text-gray-900">{a.studentName}</span>
      ),
    },
    {
      key: 'program',
      header: 'Program',
      render: (a: ScholarshipApplication) => (
        <span className="text-sm text-gray-700">{a.programName}</span>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (a: ScholarshipApplication) => {
        const s = APP_STATUS_MAP[a.status] || { label: a.status, variant: 'default' as const };
        return <Badge variant={s.variant}>{s.label}</Badge>;
      },
    },
    {
      key: 'appliedAt',
      header: 'Tanggal',
      render: (a: ScholarshipApplication) => (
        <span className="text-sm text-gray-500">{formatDate(a.appliedAt)}</span>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900">Beasiswa & Bantuan Siswa</h1>
          <p className="text-sm text-gray-500 mt-0.5">Kelola program beasiswa dan pengajuan bantuan</p>
        </div>
        {canManage && activeTab === 'programs' && (
          <Button size="sm" leftIcon={<Plus className="w-3.5 h-3.5" />} onClick={() => setIsAddOpen(true)}>
            Tambah Program
          </Button>
        )}
      </div>

      {/* Tabs */}
      <div className="flex gap-1 bg-gray-100 p-1 rounded-xl w-fit">
        {(['programs', 'applications'] as TabKey[]).map((t) => (
          <button
            key={t}
            onClick={() => { setActiveTab(t); setPage(1); }}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
              activeTab === t ? 'bg-white text-primary-700 shadow-sm' : 'text-gray-500'
            }`}
          >
            {t === 'programs' ? 'Program Beasiswa' : 'Pengajuan'}
          </button>
        ))}
      </div>

      <Card padding="none">
        {activeTab === 'programs' && (
          <div className="p-4 border-b border-gray-100">
            <Input
              placeholder="Cari program beasiswa..."
              value={search}
              onChange={(e) => { setSearch(e.target.value); setPage(1); }}
              leftAddon={<Search className="w-4 h-4" />}
              className="max-w-xs"
            />
          </div>
        )}

        <Table
          columns={activeTab === 'programs' ? programColumns : applicationColumns}
          data={activeTab === 'programs' ? programs : applications}
          keyExtractor={(item) => item.id}
          isLoading={isLoading}
          emptyMessage={activeTab === 'programs' ? 'Belum ada program beasiswa' : 'Belum ada pengajuan'}
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
        title="Tambah Program Beasiswa"
        description="Buat program beasiswa atau bantuan siswa baru"
        size="md"
        footer={
          <>
            <Button variant="secondary" onClick={() => setIsAddOpen(false)}>Batal</Button>
            <Button
              onClick={handleSubmit}
              isLoading={isSubmitting}
              leftIcon={<Award className="w-4 h-4" />}
            >
              Simpan
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Nama Program <span className="text-red-500">*</span></label>
            <input
              className="w-full text-sm border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500"
              placeholder="Contoh: Beasiswa Prestasi 2026"
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
            />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Tipe <span className="text-red-500">*</span></label>
              <select
                className="w-full text-sm border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500"
                value={form.type}
                onChange={(e) => setForm({ ...form, type: e.target.value })}
              >
                <option value="">Pilih tipe</option>
                <option value="PRESTASI">Prestasi</option>
                <option value="EKONOMI">Ekonomi</option>
                <option value="KHUSUS">Khusus</option>
                <option value="PEMERINTAH">Pemerintah</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Kuota (siswa)</label>
              <input
                type="number"
                className="w-full text-sm border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500"
                placeholder="0"
                min={1}
                value={form.quota}
                onChange={(e) => setForm({ ...form, quota: e.target.value })}
              />
            </div>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Nilai Beasiswa (Rp) <span className="text-red-500">*</span></label>
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
            <label className="block text-sm font-medium text-gray-700 mb-1">Persyaratan</label>
            <textarea
              className="w-full text-sm border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500 resize-none"
              rows={3}
              placeholder="Tuliskan persyaratan penerima beasiswa..."
              value={form.requirements}
              onChange={(e) => setForm({ ...form, requirements: e.target.value })}
            />
          </div>
        </div>
      </Modal>
    </div>
  );
}
