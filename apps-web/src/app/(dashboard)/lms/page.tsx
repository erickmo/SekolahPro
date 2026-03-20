'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { GraduationCap, Plus, RefreshCw, RotateCcw } from 'lucide-react';
import api from '@/lib/api';
import { Card, CardHeader } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { Modal } from '@/components/ui/Modal';
import { Input } from '@/components/ui/Input';
import { formatDateTime } from '@/lib/utils';

interface LmsProvider {
  id: string;
  name: string;
  url: string;
  syncStatus: 'SYNCED' | 'PENDING' | 'FAILED' | 'NEVER';
  lastSyncAt: string | null;
  totalCourses: number;
  isActive: boolean;
}

interface AddProviderForm {
  name: string;
  url: string;
  apiKey: string;
}

const EMPTY_FORM: AddProviderForm = { name: '', url: '', apiKey: '' };

const SYNC_VARIANT: Record<string, 'success' | 'warning' | 'danger' | 'gray'> = {
  SYNCED: 'success',
  PENDING: 'warning',
  FAILED: 'danger',
  NEVER: 'gray',
};

const SYNC_LABEL: Record<string, string> = {
  SYNCED: 'Tersinkron',
  PENDING: 'Menunggu',
  FAILED: 'Gagal',
  NEVER: 'Belum Pernah',
};

export default function LmsPage() {
  const [providers, setProviders] = useState<LmsProvider[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isAddOpen, setIsAddOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [syncingId, setSyncingId] = useState<string | null>(null);
  const [form, setForm] = useState<AddProviderForm>(EMPTY_FORM);

  const fetchProviders = useCallback(async () => {
    setIsLoading(true);
    try {
      const res = await api.get('/lms/providers');
      setProviders(res.data.data || []);
    } catch {
      setProviders([]);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchProviders();
  }, [fetchProviders]);

  const handleAddProvider = async () => {
    setIsSubmitting(true);
    try {
      await api.post('/lms/providers', form);
      setIsAddOpen(false);
      setForm(EMPTY_FORM);
      fetchProviders();
    } catch {
      // silent
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleSync = async (providerId: string) => {
    setSyncingId(providerId);
    try {
      await api.post('/lms/sync', { providerId });
      fetchProviders();
    } catch {
      // silent
    } finally {
      setSyncingId(null);
    }
  };

  const columns = [
    {
      key: 'name',
      header: 'Nama LMS',
      render: (p: LmsProvider) => (
        <div>
          <p className="font-medium text-gray-900">{p.name}</p>
          <p className="text-xs text-gray-400 font-mono">{p.url}</p>
        </div>
      ),
    },
    {
      key: 'courses',
      header: 'Kursus',
      render: (p: LmsProvider) => (
        <span className="text-sm font-semibold text-gray-700">
          {p.totalCourses.toLocaleString('id-ID')}
        </span>
      ),
    },
    {
      key: 'syncStatus',
      header: 'Status Sinkronisasi',
      render: (p: LmsProvider) => (
        <div className="flex items-center gap-1.5">
          {p.syncStatus === 'PENDING' && (
            <span className="w-2 h-2 rounded-full bg-yellow-400 animate-pulse" />
          )}
          <Badge variant={SYNC_VARIANT[p.syncStatus]} size="sm">
            {SYNC_LABEL[p.syncStatus]}
          </Badge>
        </div>
      ),
    },
    {
      key: 'lastSync',
      header: 'Sinkronisasi Terakhir',
      render: (p: LmsProvider) => (
        <span className="text-xs text-gray-400">
          {p.lastSyncAt ? formatDateTime(p.lastSyncAt) : '—'}
        </span>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (p: LmsProvider) => (
        <Badge variant={p.isActive ? 'success' : 'danger'} size="sm">
          {p.isActive ? 'Aktif' : 'Nonaktif'}
        </Badge>
      ),
    },
    {
      key: 'actions',
      header: '',
      render: (p: LmsProvider) => (
        <Button
          variant="ghost"
          size="sm"
          leftIcon={<RotateCcw className="w-3.5 h-3.5" />}
          onClick={(e) => { e.stopPropagation(); handleSync(p.id); }}
          isLoading={syncingId === p.id}
          disabled={!p.isActive || syncingId !== null}
        >
          Sinkron
        </Button>
      ),
      className: 'w-28',
    },
  ];

  const syncedCount = providers.filter((p) => p.syncStatus === 'SYNCED').length;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="p-2 bg-emerald-100 rounded-xl">
            <GraduationCap className="w-5 h-5 text-emerald-600" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-gray-900">Integrasi LMS</h1>
            <p className="text-sm text-gray-500 mt-0.5">Kelola penyedia Learning Management System dan sinkronisasi data</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="secondary"
            size="sm"
            leftIcon={<RefreshCw className="w-3.5 h-3.5" />}
            onClick={fetchProviders}
          >
            Refresh
          </Button>
          <Button
            size="sm"
            leftIcon={<Plus className="w-3.5 h-3.5" />}
            onClick={() => setIsAddOpen(true)}
          >
            Tambah LMS
          </Button>
        </div>
      </div>

      {/* Summary */}
      <div className="grid grid-cols-3 gap-4">
        <Card>
          <p className="text-sm text-gray-500">Total Provider</p>
          <p className="text-2xl font-bold text-gray-900 mt-1">{providers.length}</p>
        </Card>
        <Card>
          <p className="text-sm text-gray-500">Tersinkron</p>
          <p className="text-2xl font-bold text-green-600 mt-1">{syncedCount}</p>
        </Card>
        <Card>
          <p className="text-sm text-gray-500">Total Kursus</p>
          <p className="text-2xl font-bold text-emerald-600 mt-1">
            {providers.reduce((sum, p) => sum + p.totalCourses, 0).toLocaleString('id-ID')}
          </p>
        </Card>
      </div>

      {/* Providers Table */}
      <Card padding="none">
        <div className="p-4 border-b border-gray-100">
          <CardHeader
            title="Daftar Provider LMS"
            description={`${providers.length} provider terdaftar`}
          />
        </div>
        <Table
          columns={columns}
          data={providers}
          keyExtractor={(p) => p.id}
          isLoading={isLoading}
          emptyMessage="Belum ada provider LMS terdaftar"
        />
      </Card>

      {/* Add Provider Modal */}
      <Modal
        isOpen={isAddOpen}
        onClose={() => { setIsAddOpen(false); setForm(EMPTY_FORM); }}
        title="Tambah Provider LMS"
        description="Daftarkan LMS baru untuk diintegrasikan dengan sistem"
        footer={
          <>
            <Button variant="secondary" onClick={() => { setIsAddOpen(false); setForm(EMPTY_FORM); }}>
              Batal
            </Button>
            <Button onClick={handleAddProvider} isLoading={isSubmitting}>
              Simpan
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          <Input
            label="Nama LMS"
            required
            placeholder="Contoh: Moodle Sekolah, Google Classroom"
            value={form.name}
            onChange={(e) => setForm({ ...form, name: e.target.value })}
          />
          <Input
            label="URL LMS"
            required
            placeholder="Contoh: https://lms.sekolah.sch.id"
            value={form.url}
            onChange={(e) => setForm({ ...form, url: e.target.value })}
          />
          <Input
            label="API Key"
            required
            type="password"
            placeholder="Masukkan API key dari LMS"
            value={form.apiKey}
            onChange={(e) => setForm({ ...form, apiKey: e.target.value })}
          />
        </div>
      </Modal>
    </div>
  );
}
