'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Plus, MessageSquare, Flag, AlertTriangle, CheckCircle2 } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import api from '@/lib/api';
import type { CounselingSession, BullyingReport } from '@/types';
import { Card, CardHeader } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input, Select } from '@/components/ui/Input';
import { Table } from '@/components/ui/Table';
import { Pagination } from '@/components/ui/Pagination';
import { Modal } from '@/components/ui/Modal';
import { SessionStatusBadge, Badge } from '@/components/ui/Badge';
import { formatDateTime, COUNSELING_TYPE_LABELS, BULLYING_CATEGORY_LABELS } from '@/lib/utils';

type TabKey = 'sessions' | 'bullying';

export default function CounselingPage() {
  const { user } = useAuth();
  const [activeTab, setActiveTab] = useState<TabKey>('sessions');
  const [sessions, setSessions] = useState<CounselingSession[]>([]);
  const [reports, setReports] = useState<BullyingReport[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [sessionPage, setSessionPage] = useState(1);
  const [reportPage, setReportPage] = useState(1);
  const [sessionMeta, setSessionMeta] = useState({ total: 0, totalPages: 1 });
  const [reportMeta, setReportMeta] = useState({ total: 0, totalPages: 1 });
  const [addSessionOpen, setAddSessionOpen] = useState(false);
  const [reportOpen, setReportOpen] = useState(false);

  // Session form
  const [sStudentId, setSStudentId] = useState('');
  const [sType, setSType] = useState('ACADEMIC');
  const [sScheduledAt, setSScheduledAt] = useState('');
  const [sNotes, setSNotes] = useState('');

  // Bullying form
  const [bCategory, setBCategory] = useState('VERBAL');
  const [bDescription, setBDescription] = useState('');
  const [bAnonymous, setBAnonymous] = useState(true);

  const [isSubmitting, setIsSubmitting] = useState(false);

  const isBK = ['GURU_BK', 'ADMIN_SEKOLAH', 'KEPALA_SEKOLAH'].includes(user?.role || '');

  const fetchSessions = useCallback(async () => {
    setIsLoading(true);
    try {
      const params = new URLSearchParams({ page: String(sessionPage), limit: '20' });
      const res = await api.get(`/counseling/sessions?${params}`);
      setSessions(res.data.data || []);
      setSessionMeta(res.data.meta || { total: 0, totalPages: 1 });
    } catch {
      setSessions([]);
    } finally {
      setIsLoading(false);
    }
  }, [sessionPage]);

  const fetchReports = useCallback(async () => {
    setIsLoading(true);
    try {
      const params = new URLSearchParams({ page: String(reportPage), limit: '20' });
      const res = await api.get(`/counseling/bullying-reports?${params}`);
      setReports(res.data.data || []);
      setReportMeta(res.data.meta || { total: 0, totalPages: 1 });
    } catch {
      setReports([]);
    } finally {
      setIsLoading(false);
    }
  }, [reportPage]);

  useEffect(() => { fetchSessions(); }, [fetchSessions]);
  useEffect(() => { fetchReports(); }, [fetchReports]);

  const handleAddSession = async () => {
    if (!sStudentId || !sScheduledAt) return;
    setIsSubmitting(true);
    try {
      await api.post('/counseling/sessions', {
        studentId: sStudentId,
        type: sType,
        scheduledAt: sScheduledAt,
        notes: sNotes,
      });
      setAddSessionOpen(false);
      setSStudentId(''); setSType('ACADEMIC'); setSScheduledAt(''); setSNotes('');
      fetchSessions();
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleReport = async () => {
    if (!bDescription) return;
    setIsSubmitting(true);
    try {
      await api.post('/counseling/bullying-reports', {
        category: bCategory,
        description: bDescription,
        isAnonymous: bAnonymous,
      });
      setReportOpen(false);
      setBCategory('VERBAL'); setBDescription(''); setBAnonymous(true);
      fetchReports();
    } finally {
      setIsSubmitting(false);
    }
  };

  const sessionColumns = [
    {
      key: 'studentName',
      header: 'Siswa',
      render: (s: CounselingSession) => s.studentName || s.studentId,
    },
    {
      key: 'type',
      header: 'Jenis',
      render: (s: CounselingSession) => (
        <Badge variant="primary" size="sm">{COUNSELING_TYPE_LABELS[s.type] || s.type}</Badge>
      ),
    },
    {
      key: 'scheduledAt',
      header: 'Jadwal',
      render: (s: CounselingSession) => formatDateTime(s.scheduledAt),
    },
    {
      key: 'counselorName',
      header: 'Konselor',
      render: (s: CounselingSession) => s.counselorName || '—',
    },
    {
      key: 'status',
      header: 'Status',
      render: (s: CounselingSession) => <SessionStatusBadge status={s.status} />,
    },
  ];

  const reportColumns = [
    {
      key: 'category',
      header: 'Kategori',
      render: (r: BullyingReport) => (
        <Badge variant="danger" size="sm">{BULLYING_CATEGORY_LABELS[r.category] || r.category}</Badge>
      ),
    },
    {
      key: 'description',
      header: 'Deskripsi',
      render: (r: BullyingReport) => (
        <p className="max-w-xs text-sm truncate text-gray-700">{r.description}</p>
      ),
    },
    {
      key: 'isAnonymous',
      header: 'Pelapor',
      render: (r: BullyingReport) => (
        <Badge variant={r.isAnonymous ? 'gray' : 'info'} size="sm">
          {r.isAnonymous ? 'Anonim' : 'Teridentifikasi'}
        </Badge>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (r: BullyingReport) => {
        const map: Record<string, { label: string; variant: 'warning' | 'info' | 'success' | 'gray' }> = {
          OPEN: { label: 'Terbuka', variant: 'warning' },
          INVESTIGATING: { label: 'Diinvestigasi', variant: 'info' },
          RESOLVED: { label: 'Selesai', variant: 'success' },
          CLOSED: { label: 'Ditutup', variant: 'gray' },
        };
        const { label, variant } = map[r.status] || { label: r.status, variant: 'gray' as const };
        return <Badge variant={variant} size="sm">{label}</Badge>;
      },
    },
    {
      key: 'createdAt',
      header: 'Dilaporkan',
      render: (r: BullyingReport) => formatDateTime(r.createdAt),
    },
  ];

  const openReports = reports.filter((r) => r.status === 'OPEN').length;
  const completedSessions = sessions.filter((s) => s.status === 'COMPLETED').length;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900">Konseling & BK</h1>
          <p className="text-sm text-gray-500 mt-0.5">Manajemen sesi konseling dan laporan anti-bullying</p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="secondary"
            size="sm"
            leftIcon={<Flag className="w-3.5 h-3.5" />}
            onClick={() => setReportOpen(true)}
          >
            Laporkan Bullying
          </Button>
          <Button
            size="sm"
            leftIcon={<Plus className="w-3.5 h-3.5" />}
            onClick={() => setAddSessionOpen(true)}
          >
            Jadwalkan Konseling
          </Button>
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-3 gap-4">
        <Card padding="sm">
          <div className="flex items-center gap-3">
            <div className="p-2.5 bg-primary-50 rounded-lg">
              <MessageSquare className="w-5 h-5 text-primary-600" />
            </div>
            <div>
              <p className="text-xl font-bold text-gray-900">{sessionMeta.total}</p>
              <p className="text-xs text-gray-500">Total Sesi</p>
            </div>
          </div>
        </Card>
        <Card padding="sm">
          <div className="flex items-center gap-3">
            <div className="p-2.5 bg-green-50 rounded-lg">
              <CheckCircle2 className="w-5 h-5 text-green-600" />
            </div>
            <div>
              <p className="text-xl font-bold text-gray-900">{completedSessions}</p>
              <p className="text-xs text-gray-500">Sesi Selesai</p>
            </div>
          </div>
        </Card>
        <Card padding="sm">
          <div className="flex items-center gap-3">
            <div className="p-2.5 bg-red-50 rounded-lg">
              <AlertTriangle className="w-5 h-5 text-red-600" />
            </div>
            <div>
              <p className="text-xl font-bold text-gray-900">{openReports}</p>
              <p className="text-xs text-gray-500">Laporan Bullying Terbuka</p>
            </div>
          </div>
        </Card>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 bg-gray-100 p-1 rounded-xl w-fit">
        {([
          { key: 'sessions', label: 'Sesi Konseling' },
          { key: 'bullying', label: 'Laporan Bullying' },
        ] as { key: TabKey; label: string }[]).map((t) => (
          <button
            key={t.key}
            onClick={() => setActiveTab(t.key)}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
              activeTab === t.key ? 'bg-white text-primary-700 shadow-sm' : 'text-gray-500'
            }`}
          >
            {t.label}
          </button>
        ))}
      </div>

      {activeTab === 'sessions' && (
        <Card padding="none">
          <Table
            columns={sessionColumns}
            data={sessions}
            keyExtractor={(s) => s.id}
            isLoading={isLoading}
            emptyMessage="Belum ada sesi konseling"
          />
          {sessionMeta.total > 0 && (
            <div className="px-4 py-3 border-t border-gray-100">
              <Pagination
                currentPage={sessionPage}
                totalPages={sessionMeta.totalPages}
                total={sessionMeta.total}
                limit={20}
                onPageChange={setSessionPage}
              />
            </div>
          )}
        </Card>
      )}

      {activeTab === 'bullying' && (
        <Card padding="none">
          {!isBK ? (
            <div className="p-8 text-center">
              <AlertTriangle className="w-10 h-10 text-gray-300 mx-auto mb-2" />
              <p className="text-sm text-gray-500">Hanya Guru BK, Kepala Sekolah, atau Admin yang dapat melihat laporan ini</p>
            </div>
          ) : (
            <>
              <Table
                columns={reportColumns}
                data={reports}
                keyExtractor={(r) => r.id}
                isLoading={isLoading}
                emptyMessage="Tidak ada laporan bullying"
              />
              {reportMeta.total > 0 && (
                <div className="px-4 py-3 border-t border-gray-100">
                  <Pagination
                    currentPage={reportPage}
                    totalPages={reportMeta.totalPages}
                    total={reportMeta.total}
                    limit={20}
                    onPageChange={setReportPage}
                  />
                </div>
              )}
            </>
          )}
        </Card>
      )}

      {/* Add Session Modal */}
      <Modal
        isOpen={addSessionOpen}
        onClose={() => setAddSessionOpen(false)}
        title="Jadwalkan Sesi Konseling"
        size="sm"
        footer={
          <>
            <Button variant="secondary" onClick={() => setAddSessionOpen(false)}>Batal</Button>
            <Button onClick={handleAddSession} isLoading={isSubmitting}>Simpan</Button>
          </>
        }
      >
        <div className="space-y-4">
          <Input
            label="ID Siswa"
            placeholder="ID siswa"
            required
            value={sStudentId}
            onChange={(e) => setSStudentId(e.target.value)}
          />
          <Select
            label="Jenis Konseling"
            value={sType}
            onChange={(e) => setSType(e.target.value)}
            options={[
              { value: 'ACADEMIC', label: 'Akademik' },
              { value: 'PERSONAL', label: 'Personal' },
              { value: 'CAREER', label: 'Karir' },
              { value: 'SOCIAL', label: 'Sosial' },
            ]}
          />
          <Input
            label="Jadwal"
            type="datetime-local"
            required
            value={sScheduledAt}
            onChange={(e) => setSScheduledAt(e.target.value)}
          />
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Catatan Awal</label>
            <textarea
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
              rows={3}
              placeholder="Catatan awal sesi konseling..."
              value={sNotes}
              onChange={(e) => setSNotes(e.target.value)}
            />
          </div>
        </div>
      </Modal>

      {/* Bullying Report Modal */}
      <Modal
        isOpen={reportOpen}
        onClose={() => setReportOpen(false)}
        title="Laporkan Bullying"
        description="Laporan dapat disampaikan secara anonim"
        size="sm"
        footer={
          <>
            <Button variant="secondary" onClick={() => setReportOpen(false)}>Batal</Button>
            <Button variant="danger" onClick={handleReport} isLoading={isSubmitting}>Kirim Laporan</Button>
          </>
        }
      >
        <div className="space-y-4">
          <Select
            label="Kategori Bullying"
            value={bCategory}
            onChange={(e) => setBCategory(e.target.value)}
            options={[
              { value: 'PHYSICAL', label: 'Fisik' },
              { value: 'VERBAL', label: 'Verbal' },
              { value: 'CYBERBULLYING', label: 'Cyberbullying' },
              { value: 'SEXUAL_HARASSMENT', label: 'Pelecehan Seksual' },
              { value: 'OTHER', label: 'Lainnya' },
            ]}
          />
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Deskripsi Kejadian <span className="text-red-500">*</span>
            </label>
            <textarea
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
              rows={4}
              placeholder="Ceritakan kejadian yang terjadi..."
              value={bDescription}
              onChange={(e) => setBDescription(e.target.value)}
            />
          </div>
          <div className="flex items-center gap-3 p-3 bg-gray-50 rounded-lg">
            <input
              type="checkbox"
              id="anonymous"
              checked={bAnonymous}
              onChange={(e) => setBAnonymous(e.target.checked)}
              className="w-4 h-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
            />
            <div>
              <label htmlFor="anonymous" className="text-sm font-medium text-gray-700">
                Laporan Anonim
              </label>
              <p className="text-xs text-gray-400">Identitas pelapor tidak akan ditampilkan</p>
            </div>
          </div>
        </div>
      </Modal>
    </div>
  );
}
