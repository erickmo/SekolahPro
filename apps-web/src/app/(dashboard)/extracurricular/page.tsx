'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Star, Plus, RefreshCw } from 'lucide-react';
import api from '@/lib/api';
import { Card, CardHeader } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { Modal } from '@/components/ui/Modal';
import { Pagination } from '@/components/ui/Pagination';
import { Input, Select } from '@/components/ui/Input';

const STATUS_VARIANT: Record<string, 'success' | 'gray' | 'warning'> = {
  ACTIVE: 'success',
  INACTIVE: 'gray',
  FULL: 'warning',
};

const STATUS_LABEL: Record<string, string> = {
  ACTIVE: 'Aktif',
  INACTIVE: 'Nonaktif',
  FULL: 'Penuh',
};

interface Extracurricular {
  id: string;
  name: string;
  category: string;
  teacherId: string;
  schedule: string;
  quota: number;
  memberCount: number;
  status: string;
  description?: string;
}

interface AddEkstrakurikulerForm {
  name: string;
  category: string;
  teacherId: string;
  schedule: string;
  quota: number | '';
  description: string;
}

const EMPTY_FORM: AddEkstrakurikulerForm = {
  name: '',
  category: '',
  teacherId: '',
  schedule: '',
  quota: '',
  description: '',
};

export default function ExtracurricularPage() {
  const [data, setData] = useState<Extracurricular[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [meta, setMeta] = useState({ total: 0, totalPages: 1 });
  const [isAddOpen, setIsAddOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [form, setForm] = useState<AddEkstrakurikulerForm>(EMPTY_FORM);

  const fetchData = useCallback(async () => {
    setIsLoading(true);
    try {
      const params = new URLSearchParams({ page: String(page), limit: '20' });
      const res = await api.get(`/extracurricular?${params}`);
      setData(res.data.data || []);
      setMeta(res.data.meta || { total: 0, totalPages: 1 });
    } catch {
      setData([]);
    } finally {
      setIsLoading(false);
    }
  }, [page]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleAdd = async () => {
    setIsSubmitting(true);
    try {
      await api.post('/extracurricular', {
        ...form,
        quota: Number(form.quota),
      });
      setIsAddOpen(false);
      setForm(EMPTY_FORM);
      fetchData();
    } catch {
      // silent
    } finally {
      setIsSubmitting(false);
    }
  };

  const columns = [
    {
      key: 'name',
      header: 'Nama',
      render: (e: Extracurricular) => (
        <div>
          <p className="font-medium text-gray-900">{e.name}</p>
          {e.description && (
            <p className="text-xs text-gray-400 mt-0.5 truncate max-w-xs">{e.description}</p>
          )}
        </div>
      ),
    },
    {
      key: 'category',
      header: 'Kategori',
      render: (e: Extracurricular) => (
        <Badge variant="info" size="sm">{e.category || '—'}</Badge>
      ),
    },
    {
      key: 'teacherId',
      header: 'Pembina',
      render: (e: Extracurricular) => <span className="text-sm text-gray-600">{e.teacherId}</span>,
    },
    {
      key: 'schedule',
      header: 'Jadwal',
      render: (e: Extracurricular) => <span className="text-sm text-gray-600">{e.schedule}</span>,
    },
    {
      key: 'quota',
      header: 'Kuota',
      render: (e: Extracurricular) => (
        <span className="text-sm text-gray-700">{e.quota}</span>
      ),
    },
    {
      key: 'memberCount',
      header: 'Anggota',
      render: (e: Extracurricular) => (
        <span className="text-sm font-medium text-gray-700">{e.memberCount ?? 0}</span>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (e: Extracurricular) => (
        <Badge variant={STATUS_VARIANT[e.status] ?? 'default'} size="sm">
          {STATUS_LABEL[e.status] ?? e.status}
        </Badge>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="p-2 bg-amber-100 rounded-xl">
            <Star className="w-5 h-5 text-amber-600" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-gray-900">Ekstrakurikuler</h1>
            <p className="text-sm text-gray-500 mt-0.5">Kelola kegiatan ekstrakurikuler sekolah</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="secondary" size="sm" leftIcon={<RefreshCw className="w-3.5 h-3.5" />} onClick={fetchData}>
            Refresh
          </Button>
          <Button size="sm" leftIcon={<Plus className="w-3.5 h-3.5" />} onClick={() => setIsAddOpen(true)}>
            Tambah Ekskul
          </Button>
        </div>
      </div>

      {/* Table */}
      <Card padding="none">
        <div className="p-4 border-b border-gray-100">
          <CardHeader title="Daftar Ekstrakurikuler" description={`${meta.total} kegiatan terdaftar`} />
        </div>
        <Table
          columns={columns}
          data={data}
          keyExtractor={(e) => e.id}
          isLoading={isLoading}
          emptyMessage="Belum ada ekstrakurikuler terdaftar"
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
        onClose={() => { setIsAddOpen(false); setForm(EMPTY_FORM); }}
        title="Tambah Ekstrakurikuler"
        description="Daftarkan kegiatan ekstrakurikuler baru"
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
                label="Nama Kegiatan"
                required
                placeholder="Contoh: Paskibra, Basket, Robotika"
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
              />
            </div>
            <Input
              label="Kategori"
              required
              placeholder="Contoh: Olahraga, Seni, Akademik"
              value={form.category}
              onChange={(e) => setForm({ ...form, category: e.target.value })}
            />
            <Input
              label="ID Pembina"
              required
              placeholder="Teacher ID"
              value={form.teacherId}
              onChange={(e) => setForm({ ...form, teacherId: e.target.value })}
            />
            <Input
              label="Jadwal"
              required
              placeholder="Contoh: Jumat, 14.00–16.00"
              value={form.schedule}
              onChange={(e) => setForm({ ...form, schedule: e.target.value })}
            />
            <Input
              label="Kuota Anggota"
              type="number"
              required
              placeholder="30"
              value={String(form.quota)}
              onChange={(e) => setForm({ ...form, quota: e.target.value === '' ? '' : Number(e.target.value) })}
            />
            <div className="col-span-2">
              <Input
                label="Deskripsi"
                placeholder="Deskripsi singkat kegiatan (opsional)"
                value={form.description}
                onChange={(e) => setForm({ ...form, description: e.target.value })}
              />
            </div>
          </div>
        </div>
      </Modal>
    </div>
  );
}
