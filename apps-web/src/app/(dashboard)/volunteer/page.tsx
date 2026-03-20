'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Plus, Users } from 'lucide-react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import api from '@/lib/api';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input, Select } from '@/components/ui/Input';
import { Badge } from '@/components/ui/Badge';
import { Modal } from '@/components/ui/Modal';
import { Table } from '@/components/ui/Table';
import { Pagination } from '@/components/ui/Pagination';

const schema = z.object({
  name: z.string().min(1, 'Nama wajib diisi'),
  email: z.string().email('Email tidak valid'),
  phone: z.string().min(1, 'Nomor telepon wajib diisi'),
  relation: z.enum(['PARENT', 'ALUMNI', 'EXTERNAL_PROFESSIONAL']),
  skills: z.string().min(1, 'Keahlian wajib diisi'),
});
type FormValues = z.infer<typeof schema>;

interface VolunteerProfile {
  id: string;
  name: string;
  relation: string;
  phone: string;
  skills: string[];
  totalHours?: number;
  status?: string;
}

const RELATION_LABELS: Record<string, string> = {
  PARENT: 'Orang Tua',
  ALUMNI: 'Alumni',
  EXTERNAL_PROFESSIONAL: 'Profesional Eksternal',
};

const RELATION_VARIANTS: Record<string, 'primary' | 'success' | 'info'> = {
  PARENT: 'primary',
  ALUMNI: 'success',
  EXTERNAL_PROFESSIONAL: 'info',
};

export default function VolunteerPage() {
  const [items, setItems] = useState<VolunteerProfile[]>([]);
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
    defaultValues: { relation: 'PARENT' },
  });

  const fetchData = useCallback(async () => {
    setIsLoading(true);
    try {
      const res = await api.get('/api/v1/volunteer/profiles', {
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
    const skills = values.skills
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean);
    await api.post('/api/v1/volunteer/profiles', { ...values, skills });
    reset();
    setIsAddOpen(false);
    fetchData();
  };

  const columns = [
    {
      key: 'name',
      header: 'Nama',
      render: (item: VolunteerProfile) => (
        <p className="font-medium text-gray-900">{item.name}</p>
      ),
    },
    {
      key: 'relation',
      header: 'Hubungan',
      render: (item: VolunteerProfile) => (
        <Badge variant={RELATION_VARIANTS[item.relation] ?? 'default'} size="sm">
          {RELATION_LABELS[item.relation] ?? item.relation}
        </Badge>
      ),
    },
    { key: 'phone', header: 'Telepon' },
    {
      key: 'skills',
      header: 'Keahlian',
      render: (item: VolunteerProfile) =>
        Array.isArray(item.skills) && item.skills.length > 0
          ? item.skills.join(', ')
          : '—',
    },
    {
      key: 'totalHours',
      header: 'Total Jam',
      render: (item: VolunteerProfile) =>
        item.totalHours !== undefined ? `${item.totalHours} jam` : '0 jam',
    },
    {
      key: 'status',
      header: 'Status',
      render: (item: VolunteerProfile) => {
        const s = item.status ?? 'ACTIVE';
        return (
          <Badge variant={s === 'ACTIVE' ? 'success' : 'gray'} size="sm">
            {s === 'ACTIVE' ? 'Aktif' : 'Tidak Aktif'}
          </Badge>
        );
      },
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="p-2 bg-primary-100 rounded-lg">
            <Users className="w-5 h-5 text-primary-600" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-gray-900">Volunteer & Komunitas</h1>
            <p className="text-sm text-gray-500 mt-0.5">Kelola relawan dan komunitas sekolah</p>
          </div>
        </div>
        <Button
          size="sm"
          leftIcon={<Plus className="w-3.5 h-3.5" />}
          onClick={() => setIsAddOpen(true)}
        >
          Tambah Volunteer
        </Button>
      </div>

      <Card padding="none">
        <Table
          columns={columns}
          data={items}
          keyExtractor={(item) => item.id}
          isLoading={isLoading}
          emptyMessage="Belum ada data volunteer"
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
        title="Tambah Volunteer"
        size="md"
        footer={
          <>
            <Button variant="secondary" onClick={() => { setIsAddOpen(false); reset(); }}>
              Batal
            </Button>
            <Button form="volunteer-form" type="submit" isLoading={isSubmitting}>
              Simpan
            </Button>
          </>
        }
      >
        <form id="volunteer-form" onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <Input
            label="Nama Lengkap"
            placeholder="Contoh: Budi Santoso"
            required
            {...register('name')}
            error={errors.name?.message}
          />
          <Input
            label="Email"
            type="email"
            placeholder="contoh@email.com"
            required
            {...register('email')}
            error={errors.email?.message}
          />
          <Input
            label="Nomor Telepon"
            placeholder="08xxxxxxxxxx"
            required
            {...register('phone')}
            error={errors.phone?.message}
          />
          <Select
            label="Hubungan dengan Sekolah"
            required
            options={[
              { value: 'PARENT', label: 'Orang Tua' },
              { value: 'ALUMNI', label: 'Alumni' },
              { value: 'EXTERNAL_PROFESSIONAL', label: 'Profesional Eksternal' },
            ]}
            {...register('relation')}
            error={errors.relation?.message}
          />
          <Input
            label="Keahlian"
            placeholder="Desain Grafis, Fotografi, Public Speaking (pisahkan koma)"
            required
            hint="Masukkan keahlian dipisahkan dengan koma"
            {...register('skills')}
            error={errors.skills?.message}
          />
        </form>
      </Modal>
    </div>
  );
}
