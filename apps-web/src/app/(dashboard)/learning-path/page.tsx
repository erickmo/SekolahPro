'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { TrendingUp, Plus, RefreshCw, Sparkles } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import api from '@/lib/api';
import { Card, CardHeader } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { Modal } from '@/components/ui/Modal';
import { Pagination } from '@/components/ui/Pagination';
import { Input, Select } from '@/components/ui/Input';

interface LearningPath {
  id: string;
  studentName: string;
  studentId: string;
  subject: string;
  currentLevel: string;
  progressPercent: number;
  recommendedActivities: string[];
  updatedAt: string;
}

const LEVEL_MAP: Record<string, { label: string; variant: 'danger' | 'warning' | 'success' | 'info' }> = {
  BEGINNER: { label: 'Pemula', variant: 'danger' },
  INTERMEDIATE: { label: 'Menengah', variant: 'warning' },
  ADVANCED: { label: 'Mahir', variant: 'success' },
  EXPERT: { label: 'Expert', variant: 'info' },
};

const SUBJECTS = [
  'Matematika',
  'Bahasa Indonesia',
  'Bahasa Inggris',
  'IPA',
  'IPS',
  'PKn',
  'Agama',
  'Seni Budaya',
  'PJOK',
  'Informatika',
];

const EMPTY_FORM = {
  studentId: '',
  subject: '',
};

export default function LearningPathPage() {
  const { user } = useAuth();
  const [paths, setPaths] = useState<LearningPath[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [meta, setMeta] = useState({ total: 0, totalPages: 1 });
  const [subjectFilter, setSubjectFilter] = useState('');
  const [isAddOpen, setIsAddOpen] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [isGenerating, setIsGenerating] = useState(false);
  const [form, setForm] = useState(EMPTY_FORM);
  const [errors, setErrors] = useState<Partial<typeof EMPTY_FORM>>({});

  const canManage = ['ADMIN_SEKOLAH', 'KEPALA_SEKOLAH', 'KEPALA_KURIKULUM', 'GURU'].includes(user?.role || '');

  const fetchData = useCallback(async () => {
    setIsLoading(true);
    try {
      const params = new URLSearchParams({
        page: String(page),
        limit: '20',
        ...(subjectFilter && { subject: subjectFilter }),
      });
      const res = await api.get(`/learning-path/paths?${params}`);
      setPaths(res.data.data || []);
      setMeta(res.data.meta || { total: 0, totalPages: 1 });
    } catch {
      setPaths([]);
    } finally {
      setIsLoading(false);
    }
  }, [page, subjectFilter]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const validate = () => {
    const e: Partial<typeof EMPTY_FORM> = {};
    if (!form.studentId.trim()) e.studentId = 'ID Siswa wajib diisi';
    if (!form.subject.trim()) e.subject = 'Mata pelajaran wajib diisi';
    setErrors(e);
    return Object.keys(e).length === 0;
  };

  const handleSave = async () => {
    if (!validate()) return;
    setIsSaving(true);
    try {
      await api.post('/learning-path/paths', form);
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

  const handleGenerate = async () => {
    setIsGenerating(true);
    try {
      await api.post('/learning-path/generate', {});
      fetchData();
    } catch {
      // handle
    } finally {
      setIsGenerating(false);
    }
  };

  const handleClose = () => {
    setIsAddOpen(false);
    setForm(EMPTY_FORM);
    setErrors({});
  };

  const columns = [
    {
      key: 'student',
      header: 'Nama Siswa',
      render: (p: LearningPath) => (
        <div>
          <p className="font-medium text-gray-900">{p.studentName}</p>
          <p className="text-xs text-gray-400">{p.studentId}</p>
        </div>
      ),
    },
    {
      key: 'subject',
      header: 'Mata Pelajaran',
      render: (p: LearningPath) => (
        <span className="text-sm font-medium text-gray-700">{p.subject}</span>
      ),
    },
    {
      key: 'currentLevel',
      header: 'Level Saat Ini',
      render: (p: LearningPath) => {
        const l = LEVEL_MAP[p.currentLevel] || { label: p.currentLevel, variant: 'default' as const };
        return <Badge variant={l.variant}>{l.label}</Badge>;
      },
    },
    {
      key: 'progress',
      header: 'Progress',
      render: (p: LearningPath) => (
        <div className="flex items-center gap-2 min-w-[100px]">
          <div className="flex-1 h-1.5 bg-gray-100 rounded-full overflow-hidden">
            <div
              className="h-full bg-primary-500 rounded-full transition-all"
              style={{ width: `${Math.min(p.progressPercent, 100)}%` }}
            />
          </div>
          <span className="text-xs text-gray-500 tabular-nums w-8">{p.progressPercent}%</span>
        </div>
      ),
    },
    {
      key: 'activities',
      header: 'Rekomendasi Aktivitas',
      render: (p: LearningPath) => (
        <div className="flex flex-wrap gap-1 max-w-xs">
          {(p.recommendedActivities || []).slice(0, 2).map((act, i) => (
            <span
              key={i}
              className="inline-block text-xs bg-gray-50 border border-gray-200 text-gray-600 rounded px-1.5 py-0.5"
            >
              {act}
            </span>
          ))}
          {(p.recommendedActivities || []).length > 2 && (
            <span className="text-xs text-gray-400">+{p.recommendedActivities.length - 2}</span>
          )}
        </div>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900">Personalized Learning Path</h1>
          <p className="text-sm text-gray-500 mt-0.5">
            Jalur belajar personal berbasis AI untuk setiap siswa — {meta.total} jalur aktif
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
            <>
              <Button
                variant="secondary"
                size="sm"
                leftIcon={<Sparkles className={`w-3.5 h-3.5 ${isGenerating ? 'animate-pulse' : ''}`} />}
                isLoading={isGenerating}
                onClick={handleGenerate}
              >
                Generate AI
              </Button>
              <Button
                size="sm"
                leftIcon={<Plus className="w-3.5 h-3.5" />}
                onClick={() => setIsAddOpen(true)}
              >
                Tambah Jalur
              </Button>
            </>
          )}
        </div>
      </div>

      {/* Filter */}
      <Card padding="none">
        <div className="p-4 border-b border-gray-100 flex items-center gap-4 flex-wrap">
          <CardHeader
            title="Daftar Learning Path"
            description="Jalur belajar personal yang disesuaikan dengan kemampuan siswa"
          />
          <Select
            options={[
              { value: '', label: 'Semua Mata Pelajaran' },
              ...SUBJECTS.map((s) => ({ value: s, label: s })),
            ]}
            value={subjectFilter}
            onChange={(e) => { setSubjectFilter(e.target.value); setPage(1); }}
            className="w-48"
          />
        </div>

        <Table
          columns={columns}
          data={paths}
          keyExtractor={(p) => p.id}
          isLoading={isLoading}
          emptyMessage="Belum ada jalur belajar"
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
        title="Tambah Learning Path"
        description="Buat jalur belajar baru untuk siswa"
        size="md"
        footer={
          <>
            <Button variant="secondary" onClick={handleClose}>Batal</Button>
            <Button
              leftIcon={<TrendingUp className="w-4 h-4" />}
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
            onChange={(e) => setForm((f) => ({ ...f, studentId: e.target.value }))}
            error={errors.studentId}
          />
          <Select
            label="Mata Pelajaran"
            required
            options={[
              { value: '', label: 'Pilih mata pelajaran' },
              ...SUBJECTS.map((s) => ({ value: s, label: s })),
            ]}
            value={form.subject}
            onChange={(e) => setForm((f) => ({ ...f, subject: e.target.value }))}
            error={errors.subject}
          />
        </div>
      </Modal>
    </div>
  );
}
