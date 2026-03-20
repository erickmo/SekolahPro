'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Plus, Bell, Search } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import api from '@/lib/api';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { Pagination } from '@/components/ui/Pagination';
import { Modal } from '@/components/ui/Modal';
import { formatDateTime } from '@/lib/utils';

type NotifType = 'INFO' | 'WARNING' | 'URGENT';

interface Notification {
  id: string;
  title: string;
  body: string;
  type: NotifType;
  status: string;
  targetRole: string;
  createdAt: string;
}

const TYPE_MAP: Record<NotifType, { label: string; variant: 'info' | 'warning' | 'danger' }> = {
  INFO: { label: 'Info', variant: 'info' },
  WARNING: { label: 'Peringatan', variant: 'warning' },
  URGENT: { label: 'Mendesak', variant: 'danger' },
};

export default function NotificationsPage() {
  const { user } = useAuth();
  const [items, setItems] = useState<Notification[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [page, setPage] = useState(1);
  const [meta, setMeta] = useState({ total: 0, totalPages: 1 });
  const [isAddOpen, setIsAddOpen] = useState(false);
  const [isSending, setIsSending] = useState(false);

  const [form, setForm] = useState({
    title: '',
    body: '',
    type: 'INFO' as NotifType,
    targetRole: '',
  });

  const canManage = ['ADMIN_SEKOLAH', 'KEPALA_SEKOLAH', 'TATA_USAHA'].includes(user?.role || '');

  const fetchData = useCallback(async () => {
    setIsLoading(true);
    try {
      const params = new URLSearchParams({
        page: String(page),
        limit: '20',
        ...(search && { search }),
      });
      const res = await api.get(`/notifications?${params}`);
      setItems(res.data.data || []);
      setMeta(res.data.meta || { total: 0, totalPages: 1 });
    } catch {
      setItems([]);
    } finally {
      setIsLoading(false);
    }
  }, [page, search]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const handleSend = async () => {
    if (!form.title || !form.body || !form.targetRole) return;
    setIsSending(true);
    try {
      await api.post('/notifications/send', form);
      setIsAddOpen(false);
      setForm({ title: '', body: '', type: 'INFO', targetRole: '' });
      fetchData();
    } catch {
      // handle
    } finally {
      setIsSending(false);
    }
  };

  const columns = [
    {
      key: 'title',
      header: 'Judul',
      render: (n: Notification) => (
        <div>
          <p className="font-medium text-gray-900">{n.title}</p>
          <p className="text-xs text-gray-400 line-clamp-1">{n.body}</p>
        </div>
      ),
    },
    {
      key: 'type',
      header: 'Tipe',
      render: (n: Notification) => {
        const t = TYPE_MAP[n.type] || { label: n.type, variant: 'default' as const };
        return <Badge variant={t.variant}>{t.label}</Badge>;
      },
    },
    {
      key: 'targetRole',
      header: 'Target',
      render: (n: Notification) => (
        <span className="text-sm text-gray-700">{n.targetRole || '—'}</span>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (n: Notification) => (
        <Badge variant={n.status === 'SENT' ? 'success' : 'warning'}>
          {n.status === 'SENT' ? 'Terkirim' : n.status}
        </Badge>
      ),
    },
    {
      key: 'createdAt',
      header: 'Waktu',
      render: (n: Notification) => (
        <span className="text-sm text-gray-500">{formatDateTime(n.createdAt)}</span>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900">Notifikasi</h1>
          <p className="text-sm text-gray-500 mt-0.5">Kelola dan kirim notifikasi ke pengguna</p>
        </div>
        {canManage && (
          <Button size="sm" leftIcon={<Plus className="w-3.5 h-3.5" />} onClick={() => setIsAddOpen(true)}>
            Kirim Notifikasi
          </Button>
        )}
      </div>

      <Card padding="none">
        <div className="p-4 border-b border-gray-100">
          <Input
            placeholder="Cari judul notifikasi..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1); }}
            leftAddon={<Search className="w-4 h-4" />}
            className="max-w-xs"
          />
        </div>

        <Table
          columns={columns}
          data={items}
          keyExtractor={(n) => n.id}
          isLoading={isLoading}
          emptyMessage="Tidak ada notifikasi"
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
        title="Kirim Notifikasi"
        description="Buat dan kirim notifikasi ke pengguna"
        size="md"
        footer={
          <>
            <Button variant="secondary" onClick={() => setIsAddOpen(false)}>Batal</Button>
            <Button
              onClick={handleSend}
              isLoading={isSending}
              leftIcon={<Bell className="w-4 h-4" />}
            >
              Kirim
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Judul <span className="text-red-500">*</span></label>
            <input
              className="w-full text-sm border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500"
              placeholder="Judul notifikasi"
              value={form.title}
              onChange={(e) => setForm({ ...form, title: e.target.value })}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Isi Pesan <span className="text-red-500">*</span></label>
            <textarea
              className="w-full text-sm border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500 resize-none"
              rows={3}
              placeholder="Isi notifikasi..."
              value={form.body}
              onChange={(e) => setForm({ ...form, body: e.target.value })}
            />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Tipe</label>
              <select
                className="w-full text-sm border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500"
                value={form.type}
                onChange={(e) => setForm({ ...form, type: e.target.value as NotifType })}
              >
                <option value="INFO">Info</option>
                <option value="WARNING">Peringatan</option>
                <option value="URGENT">Mendesak</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Target Role <span className="text-red-500">*</span></label>
              <select
                className="w-full text-sm border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500"
                value={form.targetRole}
                onChange={(e) => setForm({ ...form, targetRole: e.target.value })}
              >
                <option value="">Pilih role</option>
                <option value="SEMUA">Semua</option>
                <option value="GURU">Guru</option>
                <option value="SISWA">Siswa</option>
                <option value="ORANG_TUA">Orang Tua</option>
                <option value="ADMIN_SEKOLAH">Admin Sekolah</option>
                <option value="KEPALA_SEKOLAH">Kepala Sekolah</option>
              </select>
            </div>
          </div>
        </div>
      </Modal>
    </div>
  );
}
