'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { RefreshCw } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import api from '@/lib/api';
import { Card, CardHeader, StatCard } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { Pagination } from '@/components/ui/Pagination';
import { formatDateTime } from '@/lib/utils';

interface SyncRecord {
  id: string;
  type: string;
  status: 'SUCCESS' | 'FAILED' | 'PENDING';
  recordsSynced: number;
  createdAt: string;
  errorMessage?: string;
}

interface SyncStatus {
  lastSyncAt: string | null;
  totalStudentsSynced: number;
  syncErrors: number;
}

const STATUS_MAP: Record<string, { label: string; variant: 'success' | 'danger' | 'warning' }> = {
  SUCCESS: { label: 'Berhasil', variant: 'success' },
  FAILED: { label: 'Gagal', variant: 'danger' },
  PENDING: { label: 'Menunggu', variant: 'warning' },
};

const TYPE_LABELS: Record<string, string> = {
  STUDENTS: 'Data Siswa',
  TEACHERS: 'Data Guru',
  CLASSES: 'Data Kelas',
  ATTENDANCE: 'Kehadiran',
  GRADES: 'Nilai',
  FULL: 'Sinkronisasi Penuh',
};

export default function DapodikPage() {
  const { user } = useAuth();
  const [history, setHistory] = useState<SyncRecord[]>([]);
  const [status, setStatus] = useState<SyncStatus | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isSyncing, setIsSyncing] = useState(false);
  const [page, setPage] = useState(1);
  const [meta, setMeta] = useState({ total: 0, totalPages: 1 });

  const canSync = ['ADMIN_SEKOLAH', 'OPERATOR_SIMS', 'TATA_USAHA'].includes(user?.role || '');

  const fetchData = useCallback(async () => {
    setIsLoading(true);
    try {
      const [historyRes, statusRes] = await Promise.all([
        api.get(`/dapodik/sync-status?page=${page}&limit=20`),
        api.get('/dapodik/sync-status'),
      ]);
      setHistory(historyRes.data.data || []);
      setMeta(historyRes.data.meta || { total: 0, totalPages: 1 });
      setStatus(statusRes.data.summary || null);
    } catch {
      setHistory([]);
    } finally {
      setIsLoading(false);
    }
  }, [page]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const handleSync = async () => {
    setIsSyncing(true);
    try {
      await api.post('/dapodik/sync', {});
      await fetchData();
    } catch {
      // handle
    } finally {
      setIsSyncing(false);
    }
  };

  const columns = [
    {
      key: 'type',
      header: 'Jenis Sinkronisasi',
      render: (r: SyncRecord) => (
        <span className="font-medium text-gray-900">
          {TYPE_LABELS[r.type] || r.type}
        </span>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (r: SyncRecord) => {
        const s = STATUS_MAP[r.status] || { label: r.status, variant: 'default' as const };
        return <Badge variant={s.variant}>{s.label}</Badge>;
      },
    },
    {
      key: 'recordsSynced',
      header: 'Data Tersinkronisasi',
      render: (r: SyncRecord) => (
        <span className="text-sm text-gray-700">{r.recordsSynced.toLocaleString('id-ID')} record</span>
      ),
    },
    {
      key: 'createdAt',
      header: 'Waktu Sinkronisasi',
      render: (r: SyncRecord) => (
        <span className="text-sm text-gray-500">{formatDateTime(r.createdAt)}</span>
      ),
    },
    {
      key: 'errorMessage',
      header: 'Keterangan',
      render: (r: SyncRecord) =>
        r.errorMessage ? (
          <span className="text-xs text-red-500 line-clamp-1">{r.errorMessage}</span>
        ) : (
          <span className="text-gray-300 text-xs">—</span>
        ),
    },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900">Dapodik / EMIS</h1>
          <p className="text-sm text-gray-500 mt-0.5">
            Sinkronisasi data sekolah ke sistem nasional Dapodik dan EMIS
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
          {canSync && (
            <Button
              size="sm"
              leftIcon={<RefreshCw className={`w-3.5 h-3.5 ${isSyncing ? 'animate-spin' : ''}`} />}
              onClick={handleSync}
              isLoading={isSyncing}
            >
              Sinkronisasi Sekarang
            </Button>
          )}
        </div>
      </div>

      {/* Status Cards */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <StatCard
          title="Sinkronisasi Terakhir"
          value={status?.lastSyncAt ? formatDateTime(status.lastSyncAt) : '—'}
          icon={<RefreshCw className="w-5 h-5 text-primary-600" />}
          iconBg="bg-primary-100"
        />
        <StatCard
          title="Total Siswa Tersinkronisasi"
          value={status?.totalStudentsSynced?.toLocaleString('id-ID') ?? '—'}
          icon={<RefreshCw className="w-5 h-5 text-green-600" />}
          iconBg="bg-green-100"
        />
        <StatCard
          title="Error Sinkronisasi"
          value={status?.syncErrors ?? '—'}
          icon={<RefreshCw className="w-5 h-5 text-red-600" />}
          iconBg="bg-red-100"
        />
      </div>

      {/* History Table */}
      <Card padding="none">
        <div className="p-4 border-b border-gray-100">
          <CardHeader
            title="Riwayat Sinkronisasi"
            description="Log semua aktivitas sinkronisasi data ke Dapodik"
          />
        </div>

        <Table
          columns={columns}
          data={history}
          keyExtractor={(r) => r.id}
          isLoading={isLoading}
          emptyMessage="Belum ada riwayat sinkronisasi"
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
    </div>
  );
}
