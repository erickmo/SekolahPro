'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { BookOpen } from 'lucide-react';
import api from '@/lib/api';
import { Card, CardHeader } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { Modal } from '@/components/ui/Modal';
import { Input } from '@/components/ui/Input';

interface TutoringProvider {
  id: string;
  name: string;
  subject: string;
  rating: number;
  ratePerHour: number;
}

interface TutoringSession {
  id: string;
  providerName: string;
  studentName: string;
  subject: string;
  scheduledAt: string;
  durationMins: number;
  status: 'SCHEDULED' | 'ONGOING' | 'COMPLETED' | 'CANCELLED';
}

const STATUS_MAP: Record<string, { label: string; variant: 'primary' | 'warning' | 'success' | 'danger' }> = {
  SCHEDULED: { label: 'Terjadwal', variant: 'primary' },
  ONGOING: { label: 'Berlangsung', variant: 'warning' },
  COMPLETED: { label: 'Selesai', variant: 'success' },
  CANCELLED: { label: 'Dibatalkan', variant: 'danger' },
};

const DEFAULT_FORM = { providerId: '', studentId: '', subject: '', scheduledAt: '', durationMins: '60' };

export default function TutoringPage() {
  const [sessions, setSessions] = useState<TutoringSession[]>([]);
  const [providers, setProviders] = useState<TutoringProvider[]>([]);
  const [isLoadingSessions, setIsLoadingSessions] = useState(true);
  const [isLoadingProviders, setIsLoadingProviders] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [form, setForm] = useState(DEFAULT_FORM);

  const fetchSessions = useCallback(async () => {
    setIsLoadingSessions(true);
    try {
      const res = await api.get('/tutoring/sessions');
      setSessions(res.data.data || []);
    } catch {
      setSessions([]);
    } finally {
      setIsLoadingSessions(false);
    }
  }, []);

  const fetchProviders = useCallback(async () => {
    setIsLoadingProviders(true);
    try {
      const res = await api.get('/tutoring/providers');
      setProviders(res.data.data || []);
    } catch {
      setProviders([]);
    } finally {
      setIsLoadingProviders(false);
    }
  }, []);

  useEffect(() => {
    fetchSessions();
    fetchProviders();
  }, [fetchSessions, fetchProviders]);

  const handleCreateSession = async () => {
    if (!form.providerId || !form.studentId || !form.subject || !form.scheduledAt) return;
    setIsSubmitting(true);
    try {
      await api.post('/tutoring/sessions', {
        ...form,
        durationMins: parseInt(form.durationMins, 10),
      });
      setIsModalOpen(false);
      setForm(DEFAULT_FORM);
      fetchSessions();
    } catch {
      // handle error silently
    } finally {
      setIsSubmitting(false);
    }
  };

  const sessionColumns = [
    {
      key: 'student',
      header: 'Siswa',
      render: (s: TutoringSession) => (
        <p className="font-medium text-gray-900">{s.studentName}</p>
      ),
    },
    {
      key: 'provider',
      header: 'Guru Les',
      render: (s: TutoringSession) => (
        <span className="text-sm text-gray-600">{s.providerName}</span>
      ),
    },
    {
      key: 'subject',
      header: 'Mata Pelajaran',
      render: (s: TutoringSession) => (
        <span className="text-sm text-gray-700">{s.subject}</span>
      ),
    },
    {
      key: 'scheduledAt',
      header: 'Jadwal',
      render: (s: TutoringSession) => (
        <span className="text-sm text-gray-600">
          {new Date(s.scheduledAt).toLocaleString('id-ID', { dateStyle: 'medium', timeStyle: 'short' })}
        </span>
      ),
    },
    {
      key: 'duration',
      header: 'Durasi',
      render: (s: TutoringSession) => (
        <span className="text-sm text-gray-600">{s.durationMins} menit</span>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (s: TutoringSession) => {
        const st = STATUS_MAP[s.status] || { label: s.status, variant: 'default' as const };
        return <Badge variant={st.variant}>{st.label}</Badge>;
      },
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900 flex items-center gap-2">
            <BookOpen className="w-5 h-5 text-primary-600" />
            Guru Les & Bimbingan Belajar
          </h1>
          <p className="text-sm text-gray-500 mt-0.5">
            Kelola sesi les dan bimbel siswa dengan guru les terdaftar
          </p>
        </div>
        <Button onClick={() => setIsModalOpen(true)}>Buat Sesi Baru</Button>
      </div>

      <div className="grid grid-cols-1 gap-6 xl:grid-cols-3">
        <div className="xl:col-span-2">
          <Card padding="none">
            <div className="p-4 border-b border-gray-100">
              <CardHeader
                title="Sesi Les"
                description="Daftar sesi bimbingan belajar yang terjadwal"
              />
            </div>
            <Table
              columns={sessionColumns}
              data={sessions}
              keyExtractor={(s) => s.id}
              isLoading={isLoadingSessions}
              emptyMessage="Belum ada sesi les terjadwal"
            />
          </Card>
        </div>

        <div>
          <Card padding="none">
            <div className="p-4 border-b border-gray-100">
              <CardHeader
                title="Guru Les Tersedia"
                description="Daftar provider bimbel aktif"
              />
            </div>
            {isLoadingProviders ? (
              <div className="p-6 text-center text-sm text-gray-400">Memuat data...</div>
            ) : providers.length === 0 ? (
              <div className="p-6 text-center text-sm text-gray-400">Belum ada guru les terdaftar</div>
            ) : (
              <ul className="divide-y divide-gray-50">
                {providers.map((p) => (
                  <li key={p.id} className="flex items-start gap-3 px-4 py-3">
                    <div className="w-8 h-8 rounded-full bg-primary-100 flex items-center justify-center shrink-0">
                      <BookOpen className="w-4 h-4 text-primary-600" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium text-gray-900">{p.name}</p>
                      <p className="text-xs text-gray-500">{p.subject}</p>
                      <p className="text-xs text-gray-400">
                        Rp {p.ratePerHour?.toLocaleString('id-ID')}/jam · ⭐ {p.rating?.toFixed(1)}
                      </p>
                    </div>
                  </li>
                ))}
              </ul>
            )}
          </Card>
        </div>
      </div>

      <Modal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        title="Buat Sesi Bimbingan Belajar"
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Guru Les</label>
            <select
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
              value={form.providerId}
              onChange={(e) => setForm({ ...form, providerId: e.target.value })}
            >
              <option value="">Pilih guru les...</option>
              {providers.map((p) => (
                <option key={p.id} value={p.id}>{p.name} — {p.subject}</option>
              ))}
            </select>
          </div>
          <Input
            label="ID Siswa"
            placeholder="Masukkan ID siswa"
            value={form.studentId}
            onChange={(e) => setForm({ ...form, studentId: e.target.value })}
          />
          <Input
            label="Mata Pelajaran"
            placeholder="Contoh: Matematika, Fisika"
            value={form.subject}
            onChange={(e) => setForm({ ...form, subject: e.target.value })}
          />
          <Input
            label="Jadwal"
            type="datetime-local"
            value={form.scheduledAt}
            onChange={(e) => setForm({ ...form, scheduledAt: e.target.value })}
          />
          <Input
            label="Durasi (menit)"
            type="number"
            placeholder="60"
            value={form.durationMins}
            onChange={(e) => setForm({ ...form, durationMins: e.target.value })}
          />
          <div className="flex justify-end gap-2 pt-2">
            <Button variant="secondary" onClick={() => setIsModalOpen(false)}>Batal</Button>
            <Button isLoading={isSubmitting} onClick={handleCreateSession}>
              Buat Sesi
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
