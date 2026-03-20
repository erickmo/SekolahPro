'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Megaphone, Plus, RefreshCw, Search } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import api from '@/lib/api';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { Modal } from '@/components/ui/Modal';
import { Pagination } from '@/components/ui/Pagination';
import { Input, Select } from '@/components/ui/Input';
import { formatDate, truncate } from '@/lib/utils';

type Priority = 'LOW' | 'MEDIUM' | 'HIGH' | 'URGENT';

interface Announcement {
  id: string;
  title: string;
  content: string;
  priority: Priority;
  targetRoles: string[];
  createdAt: string;
  authorName: string;
}

const PRIORITY_MAP: Record<Priority, { label: string; variant: 'gray' | 'info' | 'warning' | 'danger' }> = {
  LOW: { label: 'Rendah', variant: 'gray' },
  MEDIUM: { label: 'Normal', variant: 'info' },
  HIGH: { label: 'Penting', variant: 'warning' },
  URGENT: { label: 'Mendesak', variant: 'danger' },
};

const ALL_ROLES = [
  { value: 'SEMUA', label: 'Semua Pengguna' },
  { value: 'GURU', label: 'Guru' },
  { value: 'SISWA', label: 'Siswa' },
  { value: 'ORANG_TUA', label: 'Orang Tua' },
  { value: 'ADMIN_SEKOLAH', label: 'Admin Sekolah' },
  { value: 'KEPALA_SEKOLAH', label: 'Kepala Sekolah' },
  { value: 'TATA_USAHA', label: 'Tata Usaha' },
  { value: 'BENDAHARA', label: 'Bendahara' },
];

const EMPTY_FORM = {
  title: '',
  content: '',
  priority: 'MEDIUM' as Priority,
  targetRoles: [] as string[],
};

export default function AnnouncementsPage() {
  const { user } = useAuth();
  const [items, setItems] = useState<Announcement[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [page, setPage] = useState(1);
  const [meta, setMeta] = useState({ total: 0, totalPages: 1 });
  const [isAddOpen, setIsAddOpen] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [form, setForm] = useState(EMPTY_FORM);
  const [errors, setErrors] = useState<{ title?: string; content?: string; targetRoles?: string }>({});

  const canManage = ['ADMIN_SEKOLAH', 'KEPALA_SEKOLAH', 'TATA_USAHA', 'GURU'].includes(user?.role || '');

  const fetchData = useCallback(async () => {
    setIsLoading(true);
    try {
      const params = new URLSearchParams({
        page: String(page),
        limit: '20',
        ...(search && { search }),
      });
      const res = await api.get(`/notifications/announcements?${params}`);
      setItems(res.data.data || []);
      setMeta(res.data.meta || { total: 0, totalPages: 1 });
    } catch {
      setItems([]);
    } finally {
      setIsLoading(false);
    }
  }, [page, search]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const validate = () => {
    const e: typeof errors = {};
    if (!form.title.trim()) e.title = 'Judul pengumuman wajib diisi';
    if (!form.content.trim()) e.content = 'Isi pengumuman wajib diisi';
    if (form.targetRoles.length === 0) e.targetRoles = 'Pilih minimal satu target penerima';
    setErrors(e);
    return Object.keys(e).length === 0;
  };

  const handleSave = async () => {
    if (!validate()) return;
    setIsSaving(true);
    try {
      await api.post('/notifications/announcements', form);
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

  const toggleRole = (role: string) => {
    setForm((f) => ({
      ...f,
      targetRoles: f.targetRoles.includes(role)
        ? f.targetRoles.filter((r) => r !== role)
        : [...f.targetRoles, role],
    }));
  };

  const columns = [
    {
      key: 'title',
      header: 'Judul',
      render: (a: Announcement) => (
        <div>
          <p className="font-medium text-gray-900">{a.title}</p>
          <p className="text-xs text-gray-400 mt-0.5">{a.authorName}</p>
        </div>
      ),
    },
    {
      key: 'content',
      header: 'Isi',
      render: (a: Announcement) => (
        <p className="text-sm text-gray-600 max-w-xs">{truncate(a.content, 100)}</p>
      ),
    },
    {
      key: 'priority',
      header: 'Prioritas',
      render: (a: Announcement) => {
        const p = PRIORITY_MAP[a.priority] || { label: a.priority, variant: 'default' as const };
        return <Badge variant={p.variant}>{p.label}</Badge>;
      },
    },
    {
      key: 'targetRoles',
      header: 'Target',
      render: (a: Announcement) => (
        <div className="flex flex-wrap gap-1">
          {(a.targetRoles || []).slice(0, 2).map((role) => (
            <Badge key={role} variant="gray" size="sm">
              {ALL_ROLES.find((r) => r.value === role)?.label || role}
            </Badge>
          ))}
          {(a.targetRoles || []).length > 2 && (
            <span className="text-xs text-gray-400">+{a.targetRoles.length - 2}</span>
          )}
        </div>
      ),
    },
    {
      key: 'createdAt',
      header: 'Tanggal',
      render: (a: Announcement) => (
        <span className="text-sm text-gray-500">{formatDate(a.createdAt)}</span>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900">Pengumuman</h1>
          <p className="text-sm text-gray-500 mt-0.5">
            Kelola pengumuman sekolah untuk seluruh pengguna — {meta.total} pengumuman
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
              Buat Pengumuman
            </Button>
          )}
        </div>
      </div>

      {/* Table Card */}
      <Card padding="none">
        <div className="p-4 border-b border-gray-100">
          <form
            onSubmit={(e) => { e.preventDefault(); setPage(1); fetchData(); }}
            className="flex gap-3 max-w-sm"
          >
            <Input
              placeholder="Cari judul pengumuman..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              leftAddon={<Search className="w-4 h-4" />}
            />
            <Button type="submit" size="sm">Cari</Button>
          </form>
        </div>

        <Table
          columns={columns}
          data={items}
          keyExtractor={(a) => a.id}
          isLoading={isLoading}
          emptyMessage="Belum ada pengumuman"
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
        title="Buat Pengumuman"
        description="Pengumuman akan dikirimkan ke pengguna sesuai target yang dipilih"
        size="lg"
        footer={
          <>
            <Button variant="secondary" onClick={handleClose}>Batal</Button>
            <Button
              leftIcon={<Megaphone className="w-4 h-4" />}
              isLoading={isSaving}
              onClick={handleSave}
            >
              Publikasikan
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          <Input
            label="Judul Pengumuman"
            placeholder="Masukkan judul pengumuman"
            required
            value={form.title}
            onChange={(e) => setForm((f) => ({ ...f, title: e.target.value }))}
            error={errors.title}
          />
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Isi Pengumuman <span className="text-red-500">*</span>
            </label>
            <textarea
              className="w-full text-sm border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500 resize-none"
              rows={4}
              placeholder="Tulis isi pengumuman di sini..."
              value={form.content}
              onChange={(e) => setForm((f) => ({ ...f, content: e.target.value }))}
            />
            {errors.content && (
              <p className="mt-1 text-xs text-red-500">{errors.content}</p>
            )}
          </div>
          <Select
            label="Prioritas"
            options={[
              { value: 'LOW', label: 'Rendah' },
              { value: 'MEDIUM', label: 'Normal' },
              { value: 'HIGH', label: 'Penting' },
              { value: 'URGENT', label: 'Mendesak' },
            ]}
            value={form.priority}
            onChange={(e) => setForm((f) => ({ ...f, priority: e.target.value as Priority }))}
          />
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Target Penerima <span className="text-red-500">*</span>
            </label>
            <div className="flex flex-wrap gap-2">
              {ALL_ROLES.map((role) => {
                const isSelected = form.targetRoles.includes(role.value);
                return (
                  <button
                    key={role.value}
                    type="button"
                    onClick={() => toggleRole(role.value)}
                    className={`text-xs px-3 py-1.5 rounded-full border font-medium transition-colors ${
                      isSelected
                        ? 'bg-primary-600 border-primary-600 text-white'
                        : 'bg-white border-gray-300 text-gray-600 hover:border-primary-400 hover:text-primary-600'
                    }`}
                  >
                    {role.label}
                  </button>
                );
              })}
            </div>
            {errors.targetRoles && (
              <p className="mt-1 text-xs text-red-500">{errors.targetRoles}</p>
            )}
          </div>
        </div>
      </Modal>
    </div>
  );
}
