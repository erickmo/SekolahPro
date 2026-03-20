'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { FileText, Plus, RefreshCw, Sparkles } from 'lucide-react';
import api from '@/lib/api';
import { Card, CardHeader } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { Modal } from '@/components/ui/Modal';
import { Pagination } from '@/components/ui/Pagination';
import { Input, Select } from '@/components/ui/Input';

const EXAM_TYPE_LABELS: Record<string, string> = {
  ULANGAN: 'Ulangan Harian',
  UTS: 'UTS',
  UAS: 'UAS',
};

const EXAM_TYPE_VARIANT: Record<string, 'gray' | 'info' | 'primary'> = {
  ULANGAN: 'gray',
  UTS: 'info',
  UAS: 'primary',
};

const STATUS_VARIANT: Record<string, 'gray' | 'warning' | 'success' | 'danger'> = {
  DRAFT: 'gray',
  PUBLISHED: 'success',
  ONGOING: 'warning',
  ENDED: 'danger',
};

const STATUS_LABEL: Record<string, string> = {
  DRAFT: 'Draf',
  PUBLISHED: 'Dipublikasi',
  ONGOING: 'Berlangsung',
  ENDED: 'Selesai',
};

interface Exam {
  id: string;
  title: string;
  subjectId: string;
  type: string;
  durationMins: number;
  totalQuestions: number;
  status: string;
  description?: string;
}

interface AddExamForm {
  title: string;
  subjectId: string;
  type: string;
  durationMins: number | '';
  description: string;
}

const EMPTY_FORM: AddExamForm = {
  title: '',
  subjectId: '',
  type: 'ULANGAN',
  durationMins: 60,
  description: '',
};

export default function ExamsPage() {
  const [exams, setExams] = useState<Exam[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [meta, setMeta] = useState({ total: 0, totalPages: 1 });
  const [isAddOpen, setIsAddOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [generatingId, setGeneratingId] = useState<string | null>(null);
  const [form, setForm] = useState<AddExamForm>(EMPTY_FORM);

  const fetchExams = useCallback(async () => {
    setIsLoading(true);
    try {
      const params = new URLSearchParams({ page: String(page), limit: '20' });
      const res = await api.get(`/exams?${params}`);
      setExams(res.data.data || []);
      setMeta(res.data.meta || { total: 0, totalPages: 1 });
    } catch {
      setExams([]);
    } finally {
      setIsLoading(false);
    }
  }, [page]);

  useEffect(() => {
    fetchExams();
  }, [fetchExams]);

  const handleAdd = async () => {
    setIsSubmitting(true);
    try {
      await api.post('/exams', {
        ...form,
        durationMins: Number(form.durationMins),
      });
      setIsAddOpen(false);
      setForm(EMPTY_FORM);
      fetchExams();
    } catch {
      // silent
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleGenerateQuestions = async (examId: string) => {
    setGeneratingId(examId);
    try {
      await api.post(`/exams/${examId}/generate-questions`);
      fetchExams();
    } catch {
      // silent
    } finally {
      setGeneratingId(null);
    }
  };

  const columns = [
    {
      key: 'title',
      header: 'Judul',
      render: (e: Exam) => (
        <div>
          <p className="font-medium text-gray-900">{e.title}</p>
          {e.description && (
            <p className="text-xs text-gray-400 mt-0.5 truncate max-w-xs">{e.description}</p>
          )}
        </div>
      ),
    },
    {
      key: 'subjectId',
      header: 'Mata Pelajaran',
      render: (e: Exam) => <span className="text-sm text-gray-600">{e.subjectId}</span>,
    },
    {
      key: 'type',
      header: 'Tipe',
      render: (e: Exam) => (
        <Badge variant={EXAM_TYPE_VARIANT[e.type] ?? 'default'} size="sm">
          {EXAM_TYPE_LABELS[e.type] ?? e.type}
        </Badge>
      ),
    },
    {
      key: 'durationMins',
      header: 'Durasi',
      render: (e: Exam) => <span className="text-sm text-gray-600">{e.durationMins} menit</span>,
    },
    {
      key: 'totalQuestions',
      header: 'Jumlah Soal',
      render: (e: Exam) => (
        <span className="text-sm font-medium text-gray-700">{e.totalQuestions ?? 0}</span>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (e: Exam) => (
        <Badge variant={STATUS_VARIANT[e.status] ?? 'default'} size="sm">
          {STATUS_LABEL[e.status] ?? e.status}
        </Badge>
      ),
    },
    {
      key: 'actions',
      header: '',
      render: (e: Exam) => (
        <Button
          variant="ghost"
          size="sm"
          leftIcon={<Sparkles className="w-3.5 h-3.5" />}
          isLoading={generatingId === e.id}
          onClick={(ev) => { ev.stopPropagation(); handleGenerateQuestions(e.id); }}
        >
          Generate AI
        </Button>
      ),
      className: 'w-32',
    },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="p-2 bg-purple-100 rounded-xl">
            <FileText className="w-5 h-5 text-purple-600" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-gray-900">Ujian & Soal</h1>
            <p className="text-sm text-gray-500 mt-0.5">Buat dan kelola ujian dengan bantuan AI</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="secondary" size="sm" leftIcon={<RefreshCw className="w-3.5 h-3.5" />} onClick={fetchExams}>
            Refresh
          </Button>
          <Button size="sm" leftIcon={<Plus className="w-3.5 h-3.5" />} onClick={() => setIsAddOpen(true)}>
            Buat Ujian
          </Button>
        </div>
      </div>

      {/* Table */}
      <Card padding="none">
        <div className="p-4 border-b border-gray-100">
          <CardHeader title="Daftar Ujian" description={`${meta.total} ujian tersimpan`} />
        </div>
        <Table
          columns={columns}
          data={exams}
          keyExtractor={(e) => e.id}
          isLoading={isLoading}
          emptyMessage="Belum ada ujian dibuat"
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

      {/* Add Exam Modal */}
      <Modal
        isOpen={isAddOpen}
        onClose={() => { setIsAddOpen(false); setForm(EMPTY_FORM); }}
        title="Buat Ujian Baru"
        description="Isi detail ujian. Soal dapat digenerate dengan AI setelah ujian dibuat."
        footer={
          <>
            <Button variant="secondary" onClick={() => { setIsAddOpen(false); setForm(EMPTY_FORM); }}>Batal</Button>
            <Button onClick={handleAdd} isLoading={isSubmitting}>Simpan</Button>
          </>
        }
      >
        <div className="space-y-4">
          <Input
            label="Judul Ujian"
            required
            placeholder="Contoh: UTS Matematika Kelas 10"
            value={form.title}
            onChange={(e) => setForm({ ...form, title: e.target.value })}
          />
          <div className="grid grid-cols-2 gap-4">
            <Input
              label="ID Mata Pelajaran"
              required
              placeholder="Contoh: MATEMATIKA"
              value={form.subjectId}
              onChange={(e) => setForm({ ...form, subjectId: e.target.value })}
            />
            <Select
              label="Tipe Ujian"
              options={Object.entries(EXAM_TYPE_LABELS).map(([value, label]) => ({ value, label }))}
              value={form.type}
              onChange={(e) => setForm({ ...form, type: e.target.value })}
            />
          </div>
          <Input
            label="Durasi (menit)"
            type="number"
            required
            placeholder="60"
            value={String(form.durationMins)}
            onChange={(e) => setForm({ ...form, durationMins: e.target.value === '' ? '' : Number(e.target.value) })}
          />
          <Input
            label="Deskripsi"
            placeholder="Keterangan tambahan (opsional)"
            value={form.description}
            onChange={(e) => setForm({ ...form, description: e.target.value })}
          />
        </div>
      </Modal>
    </div>
  );
}
