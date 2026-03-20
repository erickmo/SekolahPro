'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Plus, BookOpen } from 'lucide-react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import api from '@/lib/api';
import { useAuth } from '@/context/AuthContext';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input, Select } from '@/components/ui/Input';
import { Badge } from '@/components/ui/Badge';
import { Modal } from '@/components/ui/Modal';
import { Table } from '@/components/ui/Table';
import { Pagination } from '@/components/ui/Pagination';
import { formatDate } from '@/lib/utils';

const schema = z.object({
  subject: z.string().min(1, 'Mata pelajaran wajib diisi'),
  gradeLevel: z.string().min(1, 'Kelas wajib diisi'),
  topic: z.string().min(1, 'Topik wajib diisi'),
  curriculum: z.enum(['MERDEKA', 'K13']),
  durationMins: z.number().min(1, 'Durasi wajib diisi'),
  content: z.string().optional(),
});
type FormValues = z.infer<typeof schema>;

interface LessonPlan {
  id: string;
  topic: string;
  subject: string;
  gradeLevel: string;
  curriculum: string;
  durationMins: number;
  status: string;
  createdAt: string;
}

export default function LessonPlansPage() {
  const { user } = useAuth();
  const [items, setItems] = useState<LessonPlan[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [meta, setMeta] = useState({ total: 0, totalPages: 1 });
  const [isAddOpen, setIsAddOpen] = useState(false);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { curriculum: 'MERDEKA', durationMins: 45 },
  });

  const fetchData = useCallback(async () => {
    setIsLoading(true);
    try {
      const res = await api.get('/api/v1/lesson-plans', { params: { page, limit: 20 } });
      setItems(res.data.data || []);
      if (res.data.meta) setMeta(res.data.meta);
    } catch {
      // handle silently
    } finally {
      setIsLoading(false);
    }
  }, [page]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const onSubmit = async (values: FormValues) => {
    let parsedContent: Record<string, unknown> | undefined;
    if (values.content) {
      try {
        parsedContent = JSON.parse(values.content);
      } catch {
        parsedContent = { raw: values.content };
      }
    }
    await api.post('/api/v1/lesson-plans', {
      ...values,
      teacherId: user?.userId,
      content: parsedContent,
    });
    reset();
    setIsAddOpen(false);
    fetchData();
  };

  const statusBadge = (status: string) => {
    if (status === 'PUBLISHED') return <Badge variant="success" size="sm">PUBLISHED</Badge>;
    return <Badge variant="gray" size="sm">DRAFT</Badge>;
  };

  const columns = [
    {
      key: 'topic',
      header: 'Topik',
      render: (item: LessonPlan) => (
        <p className="font-medium text-gray-900">{item.topic}</p>
      ),
    },
    { key: 'subject', header: 'Mata Pelajaran' },
    { key: 'gradeLevel', header: 'Kelas' },
    {
      key: 'curriculum',
      header: 'Kurikulum',
      render: (item: LessonPlan) => (
        <Badge variant={item.curriculum === 'MERDEKA' ? 'success' : 'primary'} size="sm">
          {item.curriculum}
        </Badge>
      ),
    },
    {
      key: 'durationMins',
      header: 'Durasi (menit)',
      render: (item: LessonPlan) => `${item.durationMins} menit`,
    },
    {
      key: 'status',
      header: 'Status',
      render: (item: LessonPlan) => statusBadge(item.status),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="p-2 bg-primary-100 rounded-lg">
            <BookOpen className="w-5 h-5 text-primary-600" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-gray-900">RPP Digital</h1>
            <p className="text-sm text-gray-500 mt-0.5">Kelola rencana pelaksanaan pembelajaran</p>
          </div>
        </div>
        <Button size="sm" leftIcon={<Plus className="w-3.5 h-3.5" />} onClick={() => setIsAddOpen(true)}>
          Tambah RPP
        </Button>
      </div>

      <Card padding="none">
        <Table
          columns={columns}
          data={items}
          keyExtractor={(item) => item.id}
          isLoading={isLoading}
          emptyMessage="Belum ada RPP yang dibuat"
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
        onClose={() => { setIsAddOpen(false); reset(); }}
        title="Tambah RPP"
        size="md"
        footer={
          <>
            <Button variant="secondary" onClick={() => { setIsAddOpen(false); reset(); }}>
              Batal
            </Button>
            <Button form="lesson-plan-form" type="submit" isLoading={isSubmitting}>
              Simpan
            </Button>
          </>
        }
      >
        <form id="lesson-plan-form" onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <Input
            label="Topik"
            placeholder="Contoh: Persamaan Linear Satu Variabel"
            required
            {...register('topic')}
            error={errors.topic?.message}
          />
          <Input
            label="Mata Pelajaran"
            placeholder="Contoh: Matematika"
            required
            {...register('subject')}
            error={errors.subject?.message}
          />
          <Input
            label="Kelas"
            placeholder="Contoh: VII A"
            required
            {...register('gradeLevel')}
            error={errors.gradeLevel?.message}
          />
          <Select
            label="Kurikulum"
            required
            options={[
              { value: 'MERDEKA', label: 'Merdeka Belajar' },
              { value: 'K13', label: 'Kurikulum 2013' },
            ]}
            {...register('curriculum')}
            error={errors.curriculum?.message}
          />
          <Input
            label="Durasi (menit)"
            type="number"
            min={1}
            required
            {...register('durationMins', { valueAsNumber: true })}
            error={errors.durationMins?.message}
          />
          <div className="w-full">
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Konten <span className="text-xs text-gray-400">(JSON opsional)</span>
            </label>
            <textarea
              rows={4}
              placeholder='{"tujuan": "...", "langkah": []}'
              className="block w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent resize-none"
              {...register('content')}
            />
          </div>
        </form>
      </Modal>
    </div>
  );
}
