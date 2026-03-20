'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Briefcase, Plus, RefreshCw } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import api from '@/lib/api';
import { Card, CardHeader } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { Modal } from '@/components/ui/Modal';
import { Pagination } from '@/components/ui/Pagination';
import { Input } from '@/components/ui/Input';
import { formatDate } from '@/lib/utils';

interface Placement {
  id: string;
  studentName: string;
  studentId: string;
  companyName: string;
  position: string;
  startDate: string;
  endDate: string;
  supervisorName: string;
  supervisorPhone: string;
  status: 'PENDING' | 'ACTIVE' | 'COMPLETED' | 'CANCELLED';
}

const STATUS_MAP: Record<string, { label: string; variant: 'warning' | 'success' | 'info' | 'gray' }> = {
  PENDING: { label: 'Menunggu', variant: 'warning' },
  ACTIVE: { label: 'Aktif', variant: 'success' },
  COMPLETED: { label: 'Selesai', variant: 'info' },
  CANCELLED: { label: 'Dibatalkan', variant: 'gray' },
};

const EMPTY_FORM = {
  studentId: '',
  companyName: '',
  position: '',
  startDate: '',
  endDate: '',
  supervisorName: '',
  supervisorPhone: '',
};

export default function InternshipPage() {
  const { user } = useAuth();
  const [placements, setPlacements] = useState<Placement[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [meta, setMeta] = useState({ total: 0, totalPages: 1 });
  const [isAddOpen, setIsAddOpen] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [form, setForm] = useState(EMPTY_FORM);
  const [errors, setErrors] = useState<Partial<typeof EMPTY_FORM>>({});

  const canManage = ['ADMIN_SEKOLAH', 'KEPALA_SEKOLAH', 'GURU', 'TATA_USAHA'].includes(user?.role || '');

  const fetchData = useCallback(async () => {
    setIsLoading(true);
    try {
      const res = await api.get(`/internship/placements?page=${page}&limit=20`);
      setPlacements(res.data.data || []);
      setMeta(res.data.meta || { total: 0, totalPages: 1 });
    } catch {
      setPlacements([]);
    } finally {
      setIsLoading(false);
    }
  }, [page]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const validate = () => {
    const e: Partial<typeof EMPTY_FORM> = {};
    if (!form.studentId.trim()) e.studentId = 'ID Siswa wajib diisi';
    if (!form.companyName.trim()) e.companyName = 'Nama perusahaan wajib diisi';
    if (!form.position.trim()) e.position = 'Posisi wajib diisi';
    if (!form.startDate) e.startDate = 'Tanggal mulai wajib diisi';
    if (!form.endDate) e.endDate = 'Tanggal selesai wajib diisi';
    if (!form.supervisorName.trim()) e.supervisorName = 'Nama pembimbing wajib diisi';
    if (!form.supervisorPhone.trim()) e.supervisorPhone = 'No. telepon pembimbing wajib diisi';
    setErrors(e);
    return Object.keys(e).length === 0;
  };

  const handleSave = async () => {
    if (!validate()) return;
    setIsSaving(true);
    try {
      await api.post('/internship/placements', form);
      setIsAddOpen(false);
      setForm(EMPTY_FORM);
      setErrors({});
      fetchData();
    } catch {
      // handle
    } finally {
      setIsSaving(false);
    }
  };

  const handleClose = () => {
    setIsAddOpen(false);
    setForm(EMPTY_FORM);
    setErrors({});
  };

  const set = (key: keyof typeof EMPTY_FORM) => (e: React.ChangeEvent<HTMLInputElement>) =>
    setForm((f) => ({ ...f, [key]: e.target.value }));

  const columns = [
    {
      key: 'student',
      header: 'Nama Siswa',
      render: (p: Placement) => (
        <div>
          <p className="font-medium text-gray-900">{p.studentName}</p>
          <p className="text-xs text-gray-400">{p.studentId}</p>
        </div>
      ),
    },
    {
      key: 'company',
      header: 'Perusahaan',
      render: (p: Placement) => (
        <div>
          <p className="text-sm font-medium text-gray-900">{p.companyName}</p>
          <p className="text-xs text-gray-400">{p.position}</p>
        </div>
      ),
    },
    {
      key: 'startDate',
      header: 'Tanggal Mulai',
      render: (p: Placement) => <span className="text-sm text-gray-700">{formatDate(p.startDate)}</span>,
    },
    {
      key: 'endDate',
      header: 'Tanggal Selesai',
      render: (p: Placement) => <span className="text-sm text-gray-700">{formatDate(p.endDate)}</span>,
    },
    {
      key: 'status',
      header: 'Status',
      render: (p: Placement) => {
        const s = STATUS_MAP[p.status] || { label: p.status, variant: 'default' as const };
        return <Badge variant={s.variant}>{s.label}</Badge>;
      },
    },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900">Prakerin / Magang</h1>
          <p className="text-sm text-gray-500 mt-0.5">
            Kelola penempatan praktek kerja industri siswa SMK — {meta.total} penempatan
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="secondary"
            size="sm"
            leftIcon={<RefreshCw className="w-3.5 h-3.5" />}
            onClick={fetchData}
          >
            Refresh
          </Button>
          {canManage && (
            <Button
              size="sm"
              leftIcon={<Plus className="w-3.5 h-3.5" />}
              onClick={() => setIsAddOpen(true)}
            >
              Tambah Penempatan
            </Button>
          )}
        </div>
      </div>

      {/* Table */}
      <Card padding="none">
        <Table
          columns={columns}
          data={placements}
          keyExtractor={(p) => p.id}
          isLoading={isLoading}
          emptyMessage="Belum ada data penempatan prakerin"
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

      {/* Add Modal */}
      <Modal
        isOpen={isAddOpen}
        onClose={handleClose}
        title="Tambah Penempatan Prakerin"
        description="Isi data penempatan praktek kerja industri siswa"
        size="lg"
        footer={
          <>
            <Button variant="secondary" onClick={handleClose}>Batal</Button>
            <Button
              leftIcon={<Briefcase className="w-4 h-4" />}
              isLoading={isSaving}
              onClick={handleSave}
            >
              Simpan
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          <Input
            label="ID Siswa"
            placeholder="Masukkan ID siswa"
            required
            value={form.studentId}
            onChange={set('studentId')}
            error={errors.studentId}
          />
          <div className="grid grid-cols-2 gap-4">
            <Input
              label="Nama Perusahaan"
              placeholder="PT / CV / UD..."
              required
              value={form.companyName}
              onChange={set('companyName')}
              error={errors.companyName}
            />
            <Input
              label="Posisi / Divisi"
              placeholder="Contoh: IT Support"
              required
              value={form.position}
              onChange={set('position')}
              error={errors.position}
            />
            <Input
              label="Tanggal Mulai"
              type="date"
              required
              value={form.startDate}
              onChange={set('startDate')}
              error={errors.startDate}
            />
            <Input
              label="Tanggal Selesai"
              type="date"
              required
              value={form.endDate}
              onChange={set('endDate')}
              error={errors.endDate}
            />
          </div>
          <div className="pt-2 border-t border-gray-100">
            <p className="text-sm font-medium text-gray-700 mb-3">Data Pembimbing Lapangan</p>
            <div className="grid grid-cols-2 gap-4">
              <Input
                label="Nama Pembimbing"
                placeholder="Nama lengkap pembimbing"
                required
                value={form.supervisorName}
                onChange={set('supervisorName')}
                error={errors.supervisorName}
              />
              <Input
                label="No. Telepon Pembimbing"
                placeholder="08xxxxxxxxxx"
                required
                value={form.supervisorPhone}
                onChange={set('supervisorPhone')}
                error={errors.supervisorPhone}
              />
            </div>
          </div>
        </div>
      </Modal>
    </div>
  );
}
