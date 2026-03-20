'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Plus, Headphones, Search } from 'lucide-react';
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

type TicketStatus = 'OPEN' | 'IN_PROGRESS' | 'RESOLVED' | 'CLOSED';
type TicketPriority = 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL';

interface HelpdeskTicket {
  id: string;
  subject: string;
  category: string;
  status: TicketStatus;
  priority: TicketPriority;
  description: string;
  createdAt: string;
}

const STATUS_MAP: Record<TicketStatus, { label: string; variant: 'info' | 'warning' | 'success' | 'gray' }> = {
  OPEN: { label: 'Terbuka', variant: 'info' },
  IN_PROGRESS: { label: 'Diproses', variant: 'warning' },
  RESOLVED: { label: 'Selesai', variant: 'success' },
  CLOSED: { label: 'Ditutup', variant: 'gray' },
};

const PRIORITY_MAP: Record<TicketPriority, { label: string; variant: 'gray' | 'primary' | 'warning' | 'danger' }> = {
  LOW: { label: 'Rendah', variant: 'gray' },
  MEDIUM: { label: 'Sedang', variant: 'primary' },
  HIGH: { label: 'Tinggi', variant: 'warning' },
  CRITICAL: { label: 'Kritis', variant: 'danger' },
};

export default function HelpdeskPage() {
  const { user } = useAuth();
  const [tickets, setTickets] = useState<HelpdeskTicket[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [page, setPage] = useState(1);
  const [meta, setMeta] = useState({ total: 0, totalPages: 1 });
  const [isAddOpen, setIsAddOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const [form, setForm] = useState({
    subject: '',
    category: '',
    priority: 'MEDIUM' as TicketPriority,
    description: '',
  });

  const fetchData = useCallback(async () => {
    setIsLoading(true);
    try {
      const params = new URLSearchParams({
        page: String(page),
        limit: '20',
        ...(search && { search }),
        ...(statusFilter && { status: statusFilter }),
      });
      const res = await api.get(`/helpdesk/tickets?${params}`);
      setTickets(res.data.data || []);
      setMeta(res.data.meta || { total: 0, totalPages: 1 });
    } catch {
      setTickets([]);
    } finally {
      setIsLoading(false);
    }
  }, [page, search, statusFilter]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const handleSubmit = async () => {
    if (!form.subject || !form.category || !form.description) return;
    setIsSubmitting(true);
    try {
      await api.post('/helpdesk/tickets', form);
      setIsAddOpen(false);
      setForm({ subject: '', category: '', priority: 'MEDIUM', description: '' });
      fetchData();
    } catch {
      // handle
    } finally {
      setIsSubmitting(false);
    }
  };

  const CATEGORIES = [
    'Teknis / IT', 'Fasilitas', 'Administrasi', 'Akademik',
    'Keuangan', 'Keamanan', 'Lainnya',
  ];

  const columns = [
    {
      key: 'subject',
      header: 'Subjek',
      render: (t: HelpdeskTicket) => (
        <div>
          <p className="font-medium text-gray-900">{t.subject}</p>
          <p className="text-xs text-gray-400">{t.category}</p>
        </div>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (t: HelpdeskTicket) => {
        const s = STATUS_MAP[t.status] || { label: t.status, variant: 'default' as const };
        return <Badge variant={s.variant}>{s.label}</Badge>;
      },
    },
    {
      key: 'priority',
      header: 'Prioritas',
      render: (t: HelpdeskTicket) => {
        const p = PRIORITY_MAP[t.priority] || { label: t.priority, variant: 'default' as const };
        return <Badge variant={p.variant}>{p.label}</Badge>;
      },
    },
    {
      key: 'createdAt',
      header: 'Dibuat',
      render: (t: HelpdeskTicket) => (
        <span className="text-sm text-gray-500">{formatDateTime(t.createdAt)}</span>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900">Helpdesk & Tiket</h1>
          <p className="text-sm text-gray-500 mt-0.5">Kelola permintaan bantuan dan laporan masalah</p>
        </div>
        <Button size="sm" leftIcon={<Plus className="w-3.5 h-3.5" />} onClick={() => setIsAddOpen(true)}>
          Buat Tiket
        </Button>
      </div>

      <Card padding="none">
        <div className="p-4 border-b border-gray-100 flex flex-wrap gap-3 items-center">
          <Input
            placeholder="Cari subjek tiket..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1); }}
            leftAddon={<Search className="w-4 h-4" />}
            className="max-w-xs"
          />
          <select
            value={statusFilter}
            onChange={(e) => { setStatusFilter(e.target.value); setPage(1); }}
            className="text-sm border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500"
          >
            <option value="">Semua Status</option>
            <option value="OPEN">Terbuka</option>
            <option value="IN_PROGRESS">Diproses</option>
            <option value="RESOLVED">Selesai</option>
            <option value="CLOSED">Ditutup</option>
          </select>
        </div>

        <Table
          columns={columns}
          data={tickets}
          keyExtractor={(t) => t.id}
          isLoading={isLoading}
          emptyMessage="Tidak ada tiket"
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
        title="Buat Tiket Baru"
        description="Laporkan masalah atau ajukan permintaan bantuan"
        size="md"
        footer={
          <>
            <Button variant="secondary" onClick={() => setIsAddOpen(false)}>Batal</Button>
            <Button
              onClick={handleSubmit}
              isLoading={isSubmitting}
              leftIcon={<Headphones className="w-4 h-4" />}
            >
              Kirim Tiket
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Subjek <span className="text-red-500">*</span></label>
            <input
              className="w-full text-sm border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500"
              placeholder="Deskripsi singkat masalah"
              value={form.subject}
              onChange={(e) => setForm({ ...form, subject: e.target.value })}
            />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Kategori <span className="text-red-500">*</span></label>
              <select
                className="w-full text-sm border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500"
                value={form.category}
                onChange={(e) => setForm({ ...form, category: e.target.value })}
              >
                <option value="">Pilih kategori</option>
                {CATEGORIES.map((c) => (
                  <option key={c} value={c}>{c}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Prioritas</label>
              <select
                className="w-full text-sm border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500"
                value={form.priority}
                onChange={(e) => setForm({ ...form, priority: e.target.value as TicketPriority })}
              >
                <option value="LOW">Rendah</option>
                <option value="MEDIUM">Sedang</option>
                <option value="HIGH">Tinggi</option>
                <option value="CRITICAL">Kritis</option>
              </select>
            </div>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Deskripsi <span className="text-red-500">*</span></label>
            <textarea
              className="w-full text-sm border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500 resize-none"
              rows={4}
              placeholder="Jelaskan masalah secara detail..."
              value={form.description}
              onChange={(e) => setForm({ ...form, description: e.target.value })}
            />
          </div>
        </div>
      </Modal>
    </div>
  );
}
