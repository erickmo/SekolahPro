'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Plus, GraduationCap, Search } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import api from '@/lib/api';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { Pagination } from '@/components/ui/Pagination';
import { Modal } from '@/components/ui/Modal';

interface AlumniProfile {
  id: string;
  studentId: string;
  studentName: string;
  graduationYear: number;
  currentJob: string;
  currentCity: string;
  createdAt: string;
}

export default function AlumniPage() {
  const { user } = useAuth();
  const [items, setItems] = useState<AlumniProfile[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [page, setPage] = useState(1);
  const [meta, setMeta] = useState({ total: 0, totalPages: 1 });
  const [isAddOpen, setIsAddOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const [form, setForm] = useState({
    studentId: '',
    graduationYear: String(new Date().getFullYear()),
    currentJob: '',
    currentCity: '',
  });

  const canManage = ['ADMIN_SEKOLAH', 'TATA_USAHA', 'OPERATOR_SIMS'].includes(user?.role || '');

  const fetchData = useCallback(async () => {
    setIsLoading(true);
    try {
      const params = new URLSearchParams({
        page: String(page),
        limit: '20',
        ...(search && { search }),
      });
      const res = await api.get(`/alumni/profiles?${params}`);
      setItems(res.data.data || []);
      setMeta(res.data.meta || { total: 0, totalPages: 1 });
    } catch {
      setItems([]);
    } finally {
      setIsLoading(false);
    }
  }, [page, search]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const handleSubmit = async () => {
    if (!form.studentId || !form.graduationYear) return;
    setIsSubmitting(true);
    try {
      await api.post('/alumni/profiles', {
        ...form,
        graduationYear: Number(form.graduationYear),
      });
      setIsAddOpen(false);
      setForm({ studentId: '', graduationYear: String(new Date().getFullYear()), currentJob: '', currentCity: '' });
      fetchData();
    } catch {
      // handle
    } finally {
      setIsSubmitting(false);
    }
  };

  const columns = [
    {
      key: 'name',
      header: 'Alumni',
      render: (a: AlumniProfile) => (
        <div>
          <p className="font-medium text-gray-900">{a.studentName || a.studentId}</p>
          <p className="text-xs text-gray-400">ID: {a.studentId}</p>
        </div>
      ),
    },
    {
      key: 'graduationYear',
      header: 'Tahun Lulus',
      render: (a: AlumniProfile) => (
        <Badge variant="primary">{a.graduationYear}</Badge>
      ),
    },
    {
      key: 'currentJob',
      header: 'Pekerjaan',
      render: (a: AlumniProfile) => (
        <span className="text-sm text-gray-700">{a.currentJob || '—'}</span>
      ),
    },
    {
      key: 'currentCity',
      header: 'Kota',
      render: (a: AlumniProfile) => (
        <span className="text-sm text-gray-700">{a.currentCity || '—'}</span>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900">Alumni & Tracer Study</h1>
          <p className="text-sm text-gray-500 mt-0.5">Kelola data alumni dan lacak perkembangan mereka</p>
        </div>
        {canManage && (
          <Button size="sm" leftIcon={<Plus className="w-3.5 h-3.5" />} onClick={() => setIsAddOpen(true)}>
            Tambah Alumni
          </Button>
        )}
      </div>

      <Card padding="none">
        <div className="p-4 border-b border-gray-100">
          <Input
            placeholder="Cari nama atau ID alumni..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1); }}
            leftAddon={<Search className="w-4 h-4" />}
            className="max-w-xs"
          />
        </div>

        <Table
          columns={columns}
          data={items}
          keyExtractor={(a) => a.id}
          isLoading={isLoading}
          emptyMessage="Belum ada data alumni"
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
        title="Tambah Profil Alumni"
        description="Daftarkan data alumni ke sistem"
        size="sm"
        footer={
          <>
            <Button variant="secondary" onClick={() => setIsAddOpen(false)}>Batal</Button>
            <Button
              onClick={handleSubmit}
              isLoading={isSubmitting}
              leftIcon={<GraduationCap className="w-4 h-4" />}
            >
              Simpan
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">ID Siswa <span className="text-red-500">*</span></label>
            <input
              className="w-full text-sm border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500"
              placeholder="ID siswa dari sistem"
              value={form.studentId}
              onChange={(e) => setForm({ ...form, studentId: e.target.value })}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Tahun Lulus <span className="text-red-500">*</span></label>
            <input
              type="number"
              className="w-full text-sm border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500"
              placeholder="2024"
              min={1990}
              max={2099}
              value={form.graduationYear}
              onChange={(e) => setForm({ ...form, graduationYear: e.target.value })}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Pekerjaan Saat Ini</label>
            <input
              className="w-full text-sm border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500"
              placeholder="Contoh: Software Engineer"
              value={form.currentJob}
              onChange={(e) => setForm({ ...form, currentJob: e.target.value })}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Kota Domisili</label>
            <input
              className="w-full text-sm border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500"
              placeholder="Contoh: Jakarta"
              value={form.currentCity}
              onChange={(e) => setForm({ ...form, currentCity: e.target.value })}
            />
          </div>
        </div>
      </Modal>
    </div>
  );
}
