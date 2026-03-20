'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { GraduationCap, Plus, RefreshCw, Users } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import api from '@/lib/api';
import { Card, CardHeader } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { Modal } from '@/components/ui/Modal';
import { Pagination } from '@/components/ui/Pagination';
import { Input, Select } from '@/components/ui/Input';
import { formatDate } from '@/lib/utils';

type ProgramType = 'WEBINAR' | 'WORKSHOP' | 'SEMINAR' | 'COURSE';
type ProgramStatus = 'DRAFT' | 'PUBLISHED' | 'ONGOING' | 'COMPLETED' | 'CANCELLED';

interface Program {
  id: string;
  title: string;
  type: ProgramType;
  startDate: string;
  endDate: string;
  provider: string;
  description: string;
  maxParticipants: number;
  enrollmentCount: number;
  status: ProgramStatus;
}

const TYPE_MAP: Record<ProgramType, { label: string; variant: 'primary' | 'info' | 'success' | 'warning' }> = {
  WEBINAR: { label: 'Webinar', variant: 'info' },
  WORKSHOP: { label: 'Workshop', variant: 'primary' },
  SEMINAR: { label: 'Seminar', variant: 'success' },
  COURSE: { label: 'Kursus', variant: 'warning' },
};

const STATUS_MAP: Record<ProgramStatus, { label: string; variant: 'gray' | 'primary' | 'warning' | 'success' | 'danger' }> = {
  DRAFT: { label: 'Draf', variant: 'gray' },
  PUBLISHED: { label: 'Dipublikasi', variant: 'primary' },
  ONGOING: { label: 'Berlangsung', variant: 'warning' },
  COMPLETED: { label: 'Selesai', variant: 'success' },
  CANCELLED: { label: 'Dibatalkan', variant: 'danger' },
};

const EMPTY_FORM = {
  title: '',
  type: 'WEBINAR' as ProgramType,
  startDate: '',
  endDate: '',
  provider: '',
  description: '',
  maxParticipants: '',
};

export default function TeacherTrainingPage() {
  const { user } = useAuth();
  const [programs, setPrograms] = useState<Program[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [meta, setMeta] = useState({ total: 0, totalPages: 1 });
  const [isAddOpen, setIsAddOpen] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [form, setForm] = useState(EMPTY_FORM);
  const [errors, setErrors] = useState<Partial<Record<keyof typeof EMPTY_FORM, string>>>({});

  const canManage = ['ADMIN_SEKOLAH', 'KEPALA_SEKOLAH', 'KEPALA_KURIKULUM'].includes(user?.role || '');

  const fetchData = useCallback(async () => {
    setIsLoading(true);
    try {
      const res = await api.get(`/teacher-training/programs?page=${page}&limit=20`);
      setPrograms(res.data.data || []);
      setMeta(res.data.meta || { total: 0, totalPages: 1 });
    } catch {
      setPrograms([]);
    } finally {
      setIsLoading(false);
    }
  }, [page]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const validate = () => {
    const e: Partial<Record<keyof typeof EMPTY_FORM, string>> = {};
    if (!form.title.trim()) e.title = 'Judul program wajib diisi';
    if (!form.startDate) e.startDate = 'Tanggal mulai wajib diisi';
    if (!form.endDate) e.endDate = 'Tanggal selesai wajib diisi';
    if (!form.provider.trim()) e.provider = 'Penyelenggara wajib diisi';
    if (!form.maxParticipants || Number(form.maxParticipants) < 1)
      e.maxParticipants = 'Kapasitas peserta wajib diisi';
    setErrors(e);
    return Object.keys(e).length === 0;
  };

  const handleSave = async () => {
    if (!validate()) return;
    setIsSaving(true);
    try {
      await api.post('/teacher-training/programs', {
        ...form,
        maxParticipants: Number(form.maxParticipants),
      });
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

  const set = <K extends keyof typeof EMPTY_FORM>(key: K) =>
    (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) =>
      setForm((f) => ({ ...f, [key]: e.target.value }));

  const columns = [
    {
      key: 'title',
      header: 'Judul Program',
      render: (p: Program) => (
        <div>
          <p className="font-medium text-gray-900">{p.title}</p>
          <p className="text-xs text-gray-400">{p.provider}</p>
        </div>
      ),
    },
    {
      key: 'type',
      header: 'Tipe',
      render: (p: Program) => {
        const t = TYPE_MAP[p.type] || { label: p.type, variant: 'default' as const };
        return <Badge variant={t.variant}>{t.label}</Badge>;
      },
    },
    {
      key: 'dates',
      header: 'Tanggal',
      render: (p: Program) => (
        <span className="text-sm text-gray-700">
          {formatDate(p.startDate)} — {formatDate(p.endDate)}
        </span>
      ),
    },
    {
      key: 'participants',
      header: 'Peserta',
      render: (p: Program) => (
        <div className="flex items-center gap-1.5">
          <Users className="w-3.5 h-3.5 text-gray-400" />
          <span className="text-sm text-gray-700">
            {p.enrollmentCount} / {p.maxParticipants}
          </span>
        </div>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (p: Program) => {
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
          <h1 className="text-xl font-bold text-gray-900">Pelatihan & CPD Guru</h1>
          <p className="text-sm text-gray-500 mt-0.5">
            Program pengembangan profesional dan pelatihan kompetensi guru — {meta.total} program
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
              Tambah Program
            </Button>
          )}
        </div>
      </div>

      {/* Table */}
      <Card padding="none">
        <Table
          columns={columns}
          data={programs}
          keyExtractor={(p) => p.id}
          isLoading={isLoading}
          emptyMessage="Belum ada program pelatihan"
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
        title="Tambah Program Pelatihan"
        description="Buat program pelatihan atau pengembangan profesional guru"
        size="lg"
        footer={
          <>
            <Button variant="secondary" onClick={handleClose}>Batal</Button>
            <Button
              leftIcon={<GraduationCap className="w-4 h-4" />}
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
            label="Judul Program"
            placeholder="Contoh: Workshop Kurikulum Merdeka"
            required
            value={form.title}
            onChange={set('title')}
            error={errors.title}
          />
          <div className="grid grid-cols-2 gap-4">
            <Select
              label="Tipe Program"
              required
              options={[
                { value: 'WEBINAR', label: 'Webinar' },
                { value: 'WORKSHOP', label: 'Workshop' },
                { value: 'SEMINAR', label: 'Seminar' },
                { value: 'COURSE', label: 'Kursus' },
              ]}
              value={form.type}
              onChange={set('type')}
            />
            <Input
              label="Kapasitas Peserta"
              type="number"
              placeholder="Jumlah maks peserta"
              required
              min="1"
              value={form.maxParticipants}
              onChange={set('maxParticipants')}
              error={errors.maxParticipants}
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
          <Input
            label="Penyelenggara"
            placeholder="Nama lembaga / instansi penyelenggara"
            required
            value={form.provider}
            onChange={set('provider')}
            error={errors.provider}
          />
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Deskripsi</label>
            <textarea
              className="w-full text-sm border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500 resize-none"
              rows={3}
              placeholder="Deskripsi singkat program pelatihan..."
              value={form.description}
              onChange={set('description')}
            />
          </div>
        </div>
      </Modal>
    </div>
  );
}
