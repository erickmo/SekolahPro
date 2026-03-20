'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Plus, Heart, Check, X } from 'lucide-react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import api from '@/lib/api';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Modal } from '@/components/ui/Modal';
import { Table } from '@/components/ui/Table';
import { Pagination } from '@/components/ui/Pagination';
import { formatDate } from '@/lib/utils';

const schema = z.object({
  studentId: z.string().min(1, 'ID Siswa wajib diisi'),
  encryptedData: z.string().min(1, 'Data medis wajib diisi'),
  requiresExtraTime: z.boolean().optional(),
  requiresAssistant: z.boolean().optional(),
  requiresLargeFont: z.boolean().optional(),
});
type FormValues = z.infer<typeof schema>;

interface AbkProfile {
  id: string;
  studentNisn?: string;
  studentName?: string;
  requiresExtraTime: boolean;
  requiresAssistant: boolean;
  requiresLargeFont?: boolean;
  gpkName?: string;
  createdAt: string;
}

function BoolCell({ value }: { value: boolean }) {
  return value ? (
    <Check className="w-4 h-4 text-green-600" />
  ) : (
    <X className="w-4 h-4 text-gray-300" />
  );
}

export default function SpecialNeedsPage() {
  const [items, setItems] = useState<AbkProfile[]>([]);
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
    defaultValues: {
      requiresExtraTime: false,
      requiresAssistant: false,
      requiresLargeFont: false,
    },
  });

  const fetchData = useCallback(async () => {
    setIsLoading(true);
    try {
      const res = await api.get('/api/v1/special-needs/profiles', {
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
    await api.post('/api/v1/special-needs/profiles', values);
    reset();
    setIsAddOpen(false);
    fetchData();
  };

  const columns = [
    {
      key: 'studentNisn',
      header: 'Siswa NISN',
      render: (item: AbkProfile) => (
        <div>
          {item.studentName && (
            <p className="font-medium text-gray-900">{item.studentName}</p>
          )}
          <p className="text-xs text-gray-500">{item.studentNisn || '—'}</p>
        </div>
      ),
    },
    {
      key: 'encryptedData',
      header: 'Data Medis',
      render: () => (
        <span className="text-xs text-gray-400 italic">[TERENKRIPSI]</span>
      ),
    },
    {
      key: 'requiresExtraTime',
      header: 'Waktu Extra',
      render: (item: AbkProfile) => <BoolCell value={item.requiresExtraTime} />,
    },
    {
      key: 'requiresAssistant',
      header: 'Butuh Asisten',
      render: (item: AbkProfile) => <BoolCell value={item.requiresAssistant} />,
    },
    {
      key: 'gpkName',
      header: 'GPK',
      render: (item: AbkProfile) => item.gpkName || '—',
    },
    {
      key: 'createdAt',
      header: 'Dibuat',
      render: (item: AbkProfile) => formatDate(item.createdAt),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="p-2 bg-red-100 rounded-lg">
            <Heart className="w-5 h-5 text-red-600" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-gray-900">Manajemen ABK</h1>
            <p className="text-sm text-gray-500 mt-0.5">Anak Berkebutuhan Khusus — data dienkripsi</p>
          </div>
        </div>
        <Button
          size="sm"
          leftIcon={<Plus className="w-3.5 h-3.5" />}
          onClick={() => setIsAddOpen(true)}
        >
          Tambah Profil ABK
        </Button>
      </div>

      <div className="flex items-start gap-3 p-4 bg-amber-50 border border-amber-200 rounded-xl text-sm text-amber-800">
        <span>⚠️</span>
        <p>
          Data medis siswa ABK dienkripsi AES-256 dan hanya dapat diakses oleh pihak yang berwenang.
          Kolom Data Medis selalu ditampilkan sebagai <strong>[TERENKRIPSI]</strong> di halaman ini.
        </p>
      </div>

      <Card padding="none">
        <Table
          columns={columns}
          data={items}
          keyExtractor={(item) => item.id}
          isLoading={isLoading}
          emptyMessage="Belum ada profil ABK yang terdaftar"
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
        title="Tambah Profil ABK"
        description="Data medis akan dienkripsi sebelum disimpan"
        size="md"
        footer={
          <>
            <Button variant="secondary" onClick={() => { setIsAddOpen(false); reset(); }}>
              Batal
            </Button>
            <Button form="abk-form" type="submit" isLoading={isSubmitting}>
              Simpan
            </Button>
          </>
        }
      >
        <form id="abk-form" onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <Input
            label="ID Siswa"
            placeholder="Masukkan NISN atau ID siswa"
            required
            {...register('studentId')}
            error={errors.studentId?.message}
          />
          <div className="w-full">
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Data Medis <span className="text-red-500 ml-1">*</span>
              <span className="ml-2 text-xs text-gray-400 font-normal">(akan dienkripsi)</span>
            </label>
            <textarea
              rows={4}
              placeholder="Masukkan data medis, diagnosis, atau catatan kebutuhan khusus siswa..."
              className="block w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent resize-none"
              {...register('encryptedData')}
            />
            {errors.encryptedData && (
              <p className="mt-1 text-xs text-red-600">{errors.encryptedData.message}</p>
            )}
          </div>
          <div className="space-y-3">
            <p className="text-sm font-medium text-gray-700">Kebutuhan Akomodasi</p>
            <label className="flex items-center gap-3 cursor-pointer">
              <input
                type="checkbox"
                className="w-4 h-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                {...register('requiresExtraTime')}
              />
              <span className="text-sm text-gray-700">Membutuhkan waktu extra saat ujian</span>
            </label>
            <label className="flex items-center gap-3 cursor-pointer">
              <input
                type="checkbox"
                className="w-4 h-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                {...register('requiresAssistant')}
              />
              <span className="text-sm text-gray-700">Membutuhkan pendamping (GPK)</span>
            </label>
            <label className="flex items-center gap-3 cursor-pointer">
              <input
                type="checkbox"
                className="w-4 h-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                {...register('requiresLargeFont')}
              />
              <span className="text-sm text-gray-700">Membutuhkan huruf besar (large font)</span>
            </label>
          </div>
        </form>
      </Modal>
    </div>
  );
}
