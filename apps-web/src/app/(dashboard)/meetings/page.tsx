'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Plus, Video, ExternalLink, Calendar } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import api from '@/lib/api';
import { Card, CardHeader } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input, Select } from '@/components/ui/Input';
import { Table } from '@/components/ui/Table';
import { Pagination } from '@/components/ui/Pagination';
import { Modal } from '@/components/ui/Modal';
import { Badge } from '@/components/ui/Badge';
import { formatDateTime } from '@/lib/utils';

interface Meeting {
  id: string;
  title: string;
  type: 'STAFF_MEETING' | 'PARENT_MEETING' | 'COMMITTEE_MEETING' | 'OTHER';
  scheduledAt: string;
  durationMins: number;
  status: 'SCHEDULED' | 'IN_PROGRESS' | 'COMPLETED' | 'CANCELLED';
  videoCallUrl?: string;
  description?: string;
}

const TYPE_LABELS: Record<string, string> = {
  STAFF_MEETING: 'Rapat Staf',
  PARENT_MEETING: 'Rapat Orang Tua',
  COMMITTEE_MEETING: 'Rapat Komite',
  OTHER: 'Lainnya',
};

const STATUS_CONFIG: Record<string, { label: string; variant: 'info' | 'warning' | 'success' | 'danger' }> = {
  SCHEDULED: { label: 'Terjadwal', variant: 'info' },
  IN_PROGRESS: { label: 'Sedang Berlangsung', variant: 'warning' },
  COMPLETED: { label: 'Selesai', variant: 'success' },
  CANCELLED: { label: 'Dibatalkan', variant: 'danger' },
};

export default function MeetingsPage() {
  const { user } = useAuth();
  const [meetings, setMeetings] = useState<Meeting[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [meta, setMeta] = useState({ total: 0, totalPages: 1 });
  const [addOpen, setAddOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Form state
  const [fTitle, setFTitle] = useState('');
  const [fType, setFType] = useState('STAFF_MEETING');
  const [fScheduledAt, setFScheduledAt] = useState('');
  const [fDuration, setFDuration] = useState('60');
  const [fVideoUrl, setFVideoUrl] = useState('');
  const [fDescription, setFDescription] = useState('');

  const fetchMeetings = useCallback(async () => {
    setIsLoading(true);
    try {
      const params = new URLSearchParams({ page: String(page), limit: '20' });
      const res = await api.get(`/meetings?${params}`);
      setMeetings(res.data.data || []);
      setMeta(res.data.meta || { total: 0, totalPages: 1 });
    } catch {
      setMeetings([]);
    } finally {
      setIsLoading(false);
    }
  }, [page]);

  useEffect(() => { fetchMeetings(); }, [fetchMeetings]);

  const resetForm = () => {
    setFTitle(''); setFType('STAFF_MEETING'); setFScheduledAt('');
    setFDuration('60'); setFVideoUrl(''); setFDescription('');
  };

  const handleAdd = async () => {
    if (!fTitle || !fScheduledAt) return;
    setIsSubmitting(true);
    try {
      await api.post('/meetings', {
        title: fTitle,
        type: fType,
        scheduledAt: fScheduledAt,
        durationMins: parseInt(fDuration) || 60,
        videoCallUrl: fVideoUrl || undefined,
        description: fDescription || undefined,
      });
      setAddOpen(false);
      resetForm();
      fetchMeetings();
    } finally {
      setIsSubmitting(false);
    }
  };

  const scheduledCount = meetings.filter((m) => m.status === 'SCHEDULED').length;
  const inProgressCount = meetings.filter((m) => m.status === 'IN_PROGRESS').length;

  const columns = [
    {
      key: 'title',
      header: 'Judul',
      render: (m: Meeting) => <span className="font-medium text-gray-900">{m.title}</span>,
    },
    {
      key: 'type',
      header: 'Tipe',
      render: (m: Meeting) => (
        <Badge variant="primary" size="sm">{TYPE_LABELS[m.type] || m.type}</Badge>
      ),
    },
    {
      key: 'scheduledAt',
      header: 'Jadwal',
      render: (m: Meeting) => formatDateTime(m.scheduledAt),
    },
    {
      key: 'durationMins',
      header: 'Durasi',
      render: (m: Meeting) => `${m.durationMins} menit`,
    },
    {
      key: 'status',
      header: 'Status',
      render: (m: Meeting) => {
        const cfg = STATUS_CONFIG[m.status] || { label: m.status, variant: 'gray' as const };
        return <Badge variant={cfg.variant} size="sm">{cfg.label}</Badge>;
      },
    },
    {
      key: 'videoCallUrl',
      header: 'Tautan',
      render: (m: Meeting) => {
        if (!m.videoCallUrl) return <span className="text-gray-300">—</span>;
        const canJoin = m.status === 'SCHEDULED' || m.status === 'IN_PROGRESS';
        return canJoin ? (
          <Button
            size="xs"
            variant="secondary"
            leftIcon={<ExternalLink className="w-3 h-3" />}
            onClick={() => window.open(m.videoCallUrl, '_blank')}
          >
            Bergabung
          </Button>
        ) : (
          <span className="text-xs text-gray-400">Tidak aktif</span>
        );
      },
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900">Rapat & Video Conference</h1>
          <p className="text-sm text-gray-500 mt-0.5">Jadwal dan manajemen rapat sekolah</p>
        </div>
        <Button
          size="sm"
          leftIcon={<Plus className="w-3.5 h-3.5" />}
          onClick={() => setAddOpen(true)}
        >
          Buat Rapat
        </Button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-3 gap-4">
        <Card padding="sm">
          <div className="flex items-center gap-3">
            <div className="p-2.5 bg-primary-50 rounded-lg">
              <Calendar className="w-5 h-5 text-primary-600" />
            </div>
            <div>
              <p className="text-xl font-bold text-gray-900">{meta.total}</p>
              <p className="text-xs text-gray-500">Total Rapat</p>
            </div>
          </div>
        </Card>
        <Card padding="sm">
          <div className="flex items-center gap-3">
            <div className="p-2.5 bg-blue-50 rounded-lg">
              <Video className="w-5 h-5 text-blue-600" />
            </div>
            <div>
              <p className="text-xl font-bold text-gray-900">{scheduledCount}</p>
              <p className="text-xs text-gray-500">Terjadwal</p>
            </div>
          </div>
        </Card>
        <Card padding="sm">
          <div className="flex items-center gap-3">
            <div className="p-2.5 bg-amber-50 rounded-lg">
              <Video className="w-5 h-5 text-amber-600" />
            </div>
            <div>
              <p className="text-xl font-bold text-gray-900">{inProgressCount}</p>
              <p className="text-xs text-gray-500">Sedang Berlangsung</p>
            </div>
          </div>
        </Card>
      </div>

      <Card padding="none">
        <Table
          columns={columns}
          data={meetings}
          keyExtractor={(m) => m.id}
          isLoading={isLoading}
          emptyMessage="Belum ada jadwal rapat"
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

      {/* Add Meeting Modal */}
      <Modal
        isOpen={addOpen}
        onClose={() => { setAddOpen(false); resetForm(); }}
        title="Buat Jadwal Rapat"
        size="md"
        footer={
          <>
            <Button variant="secondary" onClick={() => { setAddOpen(false); resetForm(); }}>Batal</Button>
            <Button onClick={handleAdd} isLoading={isSubmitting}>Simpan</Button>
          </>
        }
      >
        <div className="space-y-4">
          <Input
            label="Judul Rapat"
            placeholder="Contoh: Rapat Koordinasi Semester Genap"
            required
            value={fTitle}
            onChange={(e) => setFTitle(e.target.value)}
          />
          <Select
            label="Tipe Rapat"
            value={fType}
            onChange={(e) => setFType(e.target.value)}
            options={[
              { value: 'STAFF_MEETING', label: 'Rapat Staf' },
              { value: 'PARENT_MEETING', label: 'Rapat Orang Tua' },
              { value: 'COMMITTEE_MEETING', label: 'Rapat Komite' },
              { value: 'OTHER', label: 'Lainnya' },
            ]}
          />
          <div className="grid grid-cols-2 gap-4">
            <Input
              label="Jadwal"
              type="datetime-local"
              required
              value={fScheduledAt}
              onChange={(e) => setFScheduledAt(e.target.value)}
            />
            <Input
              label="Durasi (menit)"
              type="number"
              placeholder="60"
              value={fDuration}
              onChange={(e) => setFDuration(e.target.value)}
            />
          </div>
          <Input
            label="Tautan Video Call (opsional)"
            placeholder="https://meet.jitsi.si/..."
            value={fVideoUrl}
            onChange={(e) => setFVideoUrl(e.target.value)}
          />
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Deskripsi (opsional)</label>
            <textarea
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
              rows={3}
              placeholder="Agenda dan informasi rapat..."
              value={fDescription}
              onChange={(e) => setFDescription(e.target.value)}
            />
          </div>
        </div>
      </Modal>
    </div>
  );
}
