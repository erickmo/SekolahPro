'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Database, Plus, RotateCcw, RefreshCw } from 'lucide-react';
import api from '@/lib/api';
import { Card, CardHeader } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { Modal } from '@/components/ui/Modal';
import { formatDateTime } from '@/lib/utils';

interface Snapshot {
  id: string;
  name: string;
  size: string;
  status: 'PENDING' | 'RUNNING' | 'COMPLETED' | 'FAILED';
  createdAt: string;
  completedAt: string | null;
  note: string;
}

const STATUS_VARIANT: Record<string, 'warning' | 'info' | 'success' | 'danger'> = {
  PENDING: 'warning',
  RUNNING: 'info',
  COMPLETED: 'success',
  FAILED: 'danger',
};

const STATUS_LABEL: Record<string, string> = {
  PENDING: 'Menunggu',
  RUNNING: 'Berjalan',
  COMPLETED: 'Selesai',
  FAILED: 'Gagal',
};

export default function BackupPage() {
  const [snapshots, setSnapshots] = useState<Snapshot[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isTriggering, setIsTriggering] = useState(false);
  const [restoreTarget, setRestoreTarget] = useState<Snapshot | null>(null);
  const [isRestoring, setIsRestoring] = useState(false);

  const fetchSnapshots = useCallback(async () => {
    setIsLoading(true);
    try {
      const res = await api.get('/backup/snapshots');
      setSnapshots(res.data.data || []);
    } catch {
      setSnapshots([]);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchSnapshots();
  }, [fetchSnapshots]);

  const handleTriggerBackup = async () => {
    setIsTriggering(true);
    try {
      await api.post('/backup/snapshots');
      fetchSnapshots();
    } catch {
      // silent
    } finally {
      setIsTriggering(false);
    }
  };

  const handleRestore = async () => {
    if (!restoreTarget) return;
    setIsRestoring(true);
    try {
      await api.post(`/backup/snapshots/${restoreTarget.id}/restore`);
      setRestoreTarget(null);
    } catch {
      // silent
    } finally {
      setIsRestoring(false);
    }
  };

  const columns = [
    {
      key: 'name',
      header: 'Nama Snapshot',
      render: (s: Snapshot) => (
        <div>
          <p className="font-medium text-gray-900">{s.name}</p>
          {s.note && <p className="text-xs text-gray-400">{s.note}</p>}
        </div>
      ),
    },
    {
      key: 'size',
      header: 'Ukuran',
      render: (s: Snapshot) => (
        <span className="text-sm text-gray-600">{s.size || '—'}</span>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (s: Snapshot) => (
        <div className="flex items-center gap-1.5">
          {s.status === 'RUNNING' && (
            <span className="w-2 h-2 rounded-full bg-blue-500 animate-pulse" />
          )}
          <Badge variant={STATUS_VARIANT[s.status]} size="sm">
            {STATUS_LABEL[s.status]}
          </Badge>
        </div>
      ),
    },
    {
      key: 'createdAt',
      header: 'Dibuat',
      render: (s: Snapshot) => (
        <span className="text-xs text-gray-400">{formatDateTime(s.createdAt)}</span>
      ),
    },
    {
      key: 'completedAt',
      header: 'Selesai',
      render: (s: Snapshot) => (
        <span className="text-xs text-gray-400">
          {s.completedAt ? formatDateTime(s.completedAt) : '—'}
        </span>
      ),
    },
    {
      key: 'actions',
      header: '',
      render: (s: Snapshot) => (
        <Button
          variant="ghost"
          size="sm"
          leftIcon={<RotateCcw className="w-3.5 h-3.5" />}
          onClick={(e) => { e.stopPropagation(); setRestoreTarget(s); }}
          disabled={s.status !== 'COMPLETED'}
        >
          Pulihkan
        </Button>
      ),
      className: 'w-32',
    },
  ];

  const completedCount = snapshots.filter((s) => s.status === 'COMPLETED').length;
  const runningCount = snapshots.filter((s) => s.status === 'RUNNING').length;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="p-2 bg-teal-100 rounded-xl">
            <Database className="w-5 h-5 text-teal-600" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-gray-900">Backup & Recovery</h1>
            <p className="text-sm text-gray-500 mt-0.5">Kelola snapshot dan pemulihan data sistem</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="secondary"
            size="sm"
            leftIcon={<RefreshCw className="w-3.5 h-3.5" />}
            onClick={fetchSnapshots}
          >
            Refresh
          </Button>
          <Button
            size="sm"
            leftIcon={<Plus className="w-3.5 h-3.5" />}
            onClick={handleTriggerBackup}
            isLoading={isTriggering}
          >
            Backup Sekarang
          </Button>
        </div>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-3 gap-4">
        <Card>
          <p className="text-sm text-gray-500">Total Snapshot</p>
          <p className="text-2xl font-bold text-gray-900 mt-1">{snapshots.length}</p>
        </Card>
        <Card>
          <p className="text-sm text-gray-500">Berhasil</p>
          <p className="text-2xl font-bold text-green-600 mt-1">{completedCount}</p>
        </Card>
        <Card>
          <p className="text-sm text-gray-500">Sedang Berjalan</p>
          <div className="flex items-center gap-2 mt-1">
            {runningCount > 0 && <span className="w-2 h-2 rounded-full bg-blue-500 animate-pulse" />}
            <p className="text-2xl font-bold text-blue-600">{runningCount}</p>
          </div>
        </Card>
      </div>

      {/* Snapshots Table */}
      <Card padding="none">
        <div className="p-4 border-b border-gray-100">
          <CardHeader
            title="Daftar Snapshot"
            description={`${snapshots.length} snapshot tersedia`}
          />
        </div>
        <Table
          columns={columns}
          data={snapshots}
          keyExtractor={(s) => s.id}
          isLoading={isLoading}
          emptyMessage="Belum ada snapshot backup"
        />
      </Card>

      {/* Restore Confirmation Modal */}
      <Modal
        isOpen={!!restoreTarget}
        onClose={() => setRestoreTarget(null)}
        title="Konfirmasi Pemulihan Data"
        description={`Anda akan memulihkan data dari snapshot "${restoreTarget?.name}". Tindakan ini tidak dapat dibatalkan.`}
        footer={
          <>
            <Button variant="secondary" onClick={() => setRestoreTarget(null)}>Batal</Button>
            <Button variant="danger" onClick={handleRestore} isLoading={isRestoring}>
              Ya, Pulihkan Data
            </Button>
          </>
        }
      >
        <div className="p-4 rounded-lg bg-red-50 border border-red-200">
          <p className="text-sm text-red-700 font-medium">Peringatan!</p>
          <p className="text-sm text-red-600 mt-1">
            Pemulihan akan menimpa data saat ini dengan data dari snapshot. Pastikan Anda sudah membuat backup terbaru sebelum melanjutkan.
          </p>
          {restoreTarget && (
            <div className="mt-3 space-y-1">
              <p className="text-xs text-red-500">Snapshot: <span className="font-semibold">{restoreTarget.name}</span></p>
              <p className="text-xs text-red-500">Dibuat: <span className="font-semibold">{formatDateTime(restoreTarget.createdAt)}</span></p>
              <p className="text-xs text-red-500">Ukuran: <span className="font-semibold">{restoreTarget.size || '—'}</span></p>
            </div>
          )}
        </div>
      </Modal>
    </div>
  );
}
