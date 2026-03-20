'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Database } from 'lucide-react';
import api from '@/lib/api';
import { Card, CardHeader } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { Modal } from '@/components/ui/Modal';
import { Input } from '@/components/ui/Input';

interface ETLJob {
  id: string;
  type: string;
  status: 'PENDING' | 'RUNNING' | 'COMPLETED' | 'FAILED';
  params: Record<string, unknown>;
  startedAt: string | null;
  completedAt: string | null;
  createdAt: string;
}

interface WarehouseReport {
  id: string;
  name: string;
  description: string;
  lastRunAt: string;
  recordCount: number;
}

const STATUS_MAP: Record<string, { label: string; variant: 'default' | 'warning' | 'success' | 'danger' }> = {
  PENDING: { label: 'Menunggu', variant: 'default' },
  RUNNING: { label: 'Berjalan', variant: 'warning' },
  COMPLETED: { label: 'Selesai', variant: 'success' },
  FAILED: { label: 'Gagal', variant: 'danger' },
};

const JOB_TYPES = ['FULL_SYNC', 'INCREMENTAL_SYNC', 'AGGREGATE', 'EXPORT', 'CLEANUP'];

const DEFAULT_FORM = { type: 'INCREMENTAL_SYNC', params: '{}' };

export default function WarehousePage() {
  const [jobs, setJobs] = useState<ETLJob[]>([]);
  const [reports, setReports] = useState<WarehouseReport[]>([]);
  const [isLoadingJobs, setIsLoadingJobs] = useState(true);
  const [isLoadingReports, setIsLoadingReports] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [form, setForm] = useState(DEFAULT_FORM);
  const [paramsError, setParamsError] = useState('');

  const fetchJobs = useCallback(async () => {
    setIsLoadingJobs(true);
    try {
      const res = await api.get('/warehouse/jobs');
      setJobs(res.data.data || []);
    } catch {
      setJobs([]);
    } finally {
      setIsLoadingJobs(false);
    }
  }, []);

  const fetchReports = useCallback(async () => {
    setIsLoadingReports(true);
    try {
      const res = await api.get('/warehouse/reports');
      setReports(res.data.data || []);
    } catch {
      setReports([]);
    } finally {
      setIsLoadingReports(false);
    }
  }, []);

  useEffect(() => {
    fetchJobs();
    fetchReports();
  }, [fetchJobs, fetchReports]);

  const handleTriggerJob = async () => {
    setParamsError('');
    let parsedParams: Record<string, unknown> = {};
    try {
      parsedParams = JSON.parse(form.params);
    } catch {
      setParamsError('Parameter harus berupa JSON yang valid');
      return;
    }
    setIsSubmitting(true);
    try {
      await api.post('/warehouse/jobs', { type: form.type, params: parsedParams });
      setIsModalOpen(false);
      setForm(DEFAULT_FORM);
      fetchJobs();
    } catch {
      // handle error silently
    } finally {
      setIsSubmitting(false);
    }
  };

  const runningCount = jobs.filter((j) => j.status === 'RUNNING').length;
  const failedCount = jobs.filter((j) => j.status === 'FAILED').length;

  const jobColumns = [
    {
      key: 'type',
      header: 'Tipe Job',
      render: (j: ETLJob) => (
        <p className="font-medium text-gray-900 text-sm">{j.type.replace(/_/g, ' ')}</p>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (j: ETLJob) => {
        const s = STATUS_MAP[j.status] || { label: j.status, variant: 'default' as const };
        return <Badge variant={s.variant}>{s.label}</Badge>;
      },
    },
    {
      key: 'createdAt',
      header: 'Dibuat',
      render: (j: ETLJob) => (
        <span className="text-xs text-gray-500">{new Date(j.createdAt).toLocaleString('id-ID')}</span>
      ),
    },
    {
      key: 'completedAt',
      header: 'Selesai',
      render: (j: ETLJob) => (
        <span className="text-xs text-gray-500">
          {j.completedAt ? new Date(j.completedAt).toLocaleString('id-ID') : '—'}
        </span>
      ),
    },
  ];

  const reportColumns = [
    {
      key: 'name',
      header: 'Nama Laporan',
      render: (r: WarehouseReport) => (
        <p className="font-medium text-gray-900">{r.name}</p>
      ),
    },
    {
      key: 'description',
      header: 'Deskripsi',
      render: (r: WarehouseReport) => (
        <p className="text-sm text-gray-500 line-clamp-1">{r.description}</p>
      ),
    },
    {
      key: 'recordCount',
      header: 'Jumlah Record',
      render: (r: WarehouseReport) => (
        <span className="text-sm font-medium text-gray-700">{r.recordCount?.toLocaleString('id-ID')}</span>
      ),
    },
    {
      key: 'lastRunAt',
      header: 'Terakhir Dijalankan',
      render: (r: WarehouseReport) => (
        <span className="text-xs text-gray-400">{new Date(r.lastRunAt).toLocaleString('id-ID')}</span>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900 flex items-center gap-2">
            <Database className="w-5 h-5 text-primary-600" />
            Data Warehouse & BI
          </h1>
          <p className="text-sm text-gray-500 mt-0.5">
            Kelola ETL jobs dan laporan business intelligence sekolah
          </p>
        </div>
        <Button onClick={() => setIsModalOpen(true)}>Jalankan Job Baru</Button>
      </div>

      <div className="grid grid-cols-3 gap-4">
        <Card>
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-full bg-primary-100 flex items-center justify-center">
              <Database className="w-4 h-4 text-primary-600" />
            </div>
            <div>
              <p className="text-xs text-gray-500">Total Job</p>
              <p className="text-xl font-bold text-gray-900">{jobs.length}</p>
            </div>
          </div>
        </Card>
        <Card>
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-full bg-amber-100 flex items-center justify-center">
              <Database className="w-4 h-4 text-amber-600" />
            </div>
            <div>
              <p className="text-xs text-gray-500">Sedang Berjalan</p>
              <p className="text-xl font-bold text-gray-900">{runningCount}</p>
            </div>
          </div>
        </Card>
        <Card>
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-full bg-red-100 flex items-center justify-center">
              <Database className="w-4 h-4 text-red-600" />
            </div>
            <div>
              <p className="text-xs text-gray-500">Gagal</p>
              <p className="text-xl font-bold text-gray-900">{failedCount}</p>
            </div>
          </div>
        </Card>
      </div>

      <Card padding="none">
        <div className="p-4 border-b border-gray-100">
          <CardHeader
            title="ETL Jobs"
            description="Riwayat dan status pekerjaan sinkronisasi data warehouse"
          />
        </div>
        <Table
          columns={jobColumns}
          data={jobs}
          keyExtractor={(j) => j.id}
          isLoading={isLoadingJobs}
          emptyMessage="Belum ada ETL job yang dijalankan"
        />
      </Card>

      <Card padding="none">
        <div className="p-4 border-b border-gray-100">
          <CardHeader
            title="Laporan BI"
            description="Laporan business intelligence yang tersedia"
          />
        </div>
        <Table
          columns={reportColumns}
          data={reports}
          keyExtractor={(r) => r.id}
          isLoading={isLoadingReports}
          emptyMessage="Belum ada laporan BI tersedia"
        />
      </Card>

      <Modal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        title="Jalankan ETL Job Baru"
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Tipe Job</label>
            <select
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
              value={form.type}
              onChange={(e) => setForm({ ...form, type: e.target.value })}
            >
              {JOB_TYPES.map((t) => (
                <option key={t} value={t}>{t.replace(/_/g, ' ')}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Parameter (JSON)</label>
            <textarea
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-primary-500 min-h-[100px]"
              placeholder='{"startDate": "2026-01-01", "endDate": "2026-03-31"}'
              value={form.params}
              onChange={(e) => { setForm({ ...form, params: e.target.value }); setParamsError(''); }}
            />
            {paramsError && <p className="text-xs text-red-500 mt-1">{paramsError}</p>}
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <Button variant="secondary" onClick={() => setIsModalOpen(false)}>Batal</Button>
            <Button isLoading={isSubmitting} onClick={handleTriggerJob}>
              Jalankan Job
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
