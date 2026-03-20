'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Plus, RefreshCw } from 'lucide-react';
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

const schema = z.object({
  subject: z.string().min(1, 'Mata pelajaran wajib diisi'),
  gradeLevel: z.string().min(1, 'Kelas wajib diisi'),
  type: z.enum(['REMEDIAL', 'ENRICHMENT']),
  title: z.string().min(1, 'Judul wajib diisi'),
  targetKDs: z.string().min(1, 'KD target wajib diisi'),
});
type FormValues = z.infer<typeof schema>;

interface Program {
  id: string;
  title: string;
  subject: string;
  gradeLevel: string;
  type: string;
  targetKDs: string[];
  createdAt: string;
}

export default function RemediationPage() {
  const { user } = useAuth();
  const [items, setItems] = useState<Program[]>([]);
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
    defaultValues: { type: 'REMEDIAL' },
  });

  const fetchData = useCallback(async () => {
    setIsLoading(true);
    try {
      const res = await api.get('/api/v1/remediation/programs', {
        params: { page, limit: 20 },
      });
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
    const targetKDs = values.targetKDs
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean);
    await api.post('/api/v1/remediation/programs', {
      ...values,
      teacherId: user?.userId,
      targetKDs,
    });
    reset();
    setIsAddOpen(false);
    fetchData();
  };

  const typeBadge = (type: string) => {
    if (type === 'REMEDIAL') return <Badge variant="danger" size="sm">REMEDIAL</Badge>;
    return <Badge variant="info" size="sm">ENRICHMENT</Badge>;
  };

  const columns = [
    {
      key: 'title',
      header: 'Judul',
      render: (item: Program) => (
        <p className="font-medium text-gray-900">{item.title}</p>
      ),
    },
    { key: 'subject', header: 'Mata Pelajaran' },
    { key: 'gradeLevel', header: 'Kelas' },
    {
      key: 'type',
      header: 'Tipe',
      render: (item: Program) => typeBadge(item.type),
    },
    {
      key: 'targetKDs',
      header: 'KD Target',
      render: (item: Program) =>
        Array.isArray(item.targetKDs) && item.targetKDs.length > 0
          ? item.targetKDs.join(', ')
          : '—',
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="p-2 bg-primary-100 rounded-lg">
            <RefreshCw className="w-5 h-5 text-primary-600" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-gray-900">Remedial & Pengayaan</h1>
            <p className="text-sm text-gray-500 mt-0.5">Program remedial dan pengayaan siswa</p>
          </div>
        </div>
        <Button
          size="sm"
          leftIcon={<Plus className="w-3.5 h-3.5" />}
          onClick={() => setIsAddOpen(true)}
        >
          Tambah Program
        </Button>
      </div>

      <Card padding="none">
        <Table
          columns={columns}
          data={items}
          keyExtractor={(item) => item.id}
          isLoading={isLoading}
          emptyMessage="Belum ada program remedial atau pengayaan"
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
        title="Tambah Program"
        size="md"
        footer={
          <>
            <Button variant="secondary" onClick={() => { setIsAddOpen(false); reset(); }}>
              Batal
            </Button>
            <Button form="remediation-form" type="submit" isLoading={isSubmitting}>
              Simpan
            </Button>
          </>
        }
      >
        <form id="remediation-form" onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <Input
            label="Judul Program"
            placeholder="Contoh: Remedial Persamaan Kuadrat"
            required
            {...register('title')}
            error={errors.title?.message}
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
            placeholder="Contoh: X IPA 1"
            required
            {...register('gradeLevel')}
            error={errors.gradeLevel?.message}
          />
          <Select
            label="Tipe"
            required
            options={[
              { value: 'REMEDIAL', label: 'Remedial' },
              { value: 'ENRICHMENT', label: 'Pengayaan (Enrichment)' },
            ]}
            {...register('type')}
            error={errors.type?.message}
          />
          <Input
            label="KD Target"
            placeholder="3.1, 3.2, 4.1 (pisahkan dengan koma)"
            required
            hint="Masukkan kode Kompetensi Dasar dipisahkan koma"
            {...register('targetKDs')}
            error={errors.targetKDs?.message}
          />
        </form>
      </Modal>
    </div>
  );
}
