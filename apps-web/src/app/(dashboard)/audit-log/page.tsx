'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Shield, RefreshCw, ChevronLeft, ChevronRight } from 'lucide-react';
import api from '@/lib/api';
import { Card, CardHeader } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { Input, Select } from '@/components/ui/Input';
import { formatDateTime } from '@/lib/utils';

interface AuditLog {
  id: string;
  timestamp: string;
  user: string;
  userId: string;
  resource: string;
  action: string;
  details: string;
  ipAddress: string;
}

interface Pagination {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

const ACTION_VARIANT: Record<string, 'success' | 'danger' | 'warning' | 'info' | 'gray'> = {
  CREATE: 'success',
  DELETE: 'danger',
  UPDATE: 'warning',
  READ: 'info',
  LOGIN: 'gray',
  LOGOUT: 'gray',
  EXPORT: 'warning',
};

const RESOURCE_OPTIONS = [
  { value: '', label: 'Semua Resource' },
  { value: 'student', label: 'Siswa' },
  { value: 'payment', label: 'Pembayaran' },
  { value: 'user', label: 'Pengguna' },
  { value: 'exam', label: 'Ujian' },
  { value: 'grade', label: 'Nilai' },
  { value: 'health', label: 'Kesehatan' },
  { value: 'auth', label: 'Autentikasi' },
];

const ACTION_OPTIONS = [
  { value: '', label: 'Semua Aksi' },
  { value: 'CREATE', label: 'Buat' },
  { value: 'READ', label: 'Baca' },
  { value: 'UPDATE', label: 'Ubah' },
  { value: 'DELETE', label: 'Hapus' },
  { value: 'LOGIN', label: 'Login' },
  { value: 'LOGOUT', label: 'Logout' },
  { value: 'EXPORT', label: 'Ekspor' },
];

export default function AuditLogPage() {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [pagination, setPagination] = useState<Pagination>({ page: 1, limit: 20, total: 0, totalPages: 1 });
  const [isLoading, setIsLoading] = useState(true);
  const [filterResource, setFilterResource] = useState('');
  const [filterAction, setFilterAction] = useState('');

  const fetchLogs = useCallback(async (page: number, resource: string, action: string) => {
    setIsLoading(true);
    try {
      const params = new URLSearchParams({ page: String(page), limit: '20' });
      if (resource) params.set('resource', resource);
      if (action) params.set('action', action);
      const res = await api.get(`/audit-logs?${params}`);
      setLogs(res.data.data || []);
      if (res.data.pagination) setPagination(res.data.pagination);
    } catch {
      setLogs([]);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchLogs(pagination.page, filterResource, filterAction);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [fetchLogs, pagination.page, filterResource, filterAction]);

  const handleFilterChange = () => {
    setPagination((prev) => ({ ...prev, page: 1 }));
  };

  const columns = [
    {
      key: 'timestamp',
      header: 'Waktu',
      render: (log: AuditLog) => (
        <span className="text-xs text-gray-400 whitespace-nowrap">{formatDateTime(log.timestamp)}</span>
      ),
    },
    {
      key: 'user',
      header: 'Pengguna',
      render: (log: AuditLog) => (
        <div>
          <p className="text-sm font-medium text-gray-900">{log.user}</p>
          <p className="text-xs text-gray-400">{log.ipAddress}</p>
        </div>
      ),
    },
    {
      key: 'resource',
      header: 'Resource',
      render: (log: AuditLog) => (
        <Badge variant="info" size="sm">{log.resource}</Badge>
      ),
    },
    {
      key: 'action',
      header: 'Aksi',
      render: (log: AuditLog) => (
        <Badge variant={ACTION_VARIANT[log.action] ?? 'gray'} size="sm">
          {log.action}
        </Badge>
      ),
    },
    {
      key: 'details',
      header: 'Detail',
      render: (log: AuditLog) => (
        <p className="text-xs text-gray-500 max-w-xs truncate" title={log.details}>
          {log.details}
        </p>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="p-2 bg-slate-100 rounded-xl">
            <Shield className="w-5 h-5 text-slate-600" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-gray-900">Audit Log & Compliance</h1>
            <p className="text-sm text-gray-500 mt-0.5">Catatan aktivitas sistem yang tidak dapat diubah</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <div className="flex items-center gap-1.5 px-3 py-1.5 bg-red-50 rounded-lg border border-red-100">
            <span className="text-xs text-red-700 font-medium">Read-Only · Immutable</span>
          </div>
          <Button
            variant="secondary"
            size="sm"
            leftIcon={<RefreshCw className="w-3.5 h-3.5" />}
            onClick={() => fetchLogs(pagination.page, filterResource, filterAction)}
          >
            Refresh
          </Button>
        </div>
      </div>

      {/* Filters */}
      <Card>
        <div className="flex flex-wrap items-end gap-3">
          <div className="flex-1 min-w-40">
            <Select
              label="Filter Resource"
              options={RESOURCE_OPTIONS}
              value={filterResource}
              onChange={(e) => { setFilterResource(e.target.value); handleFilterChange(); }}
            />
          </div>
          <div className="flex-1 min-w-40">
            <Select
              label="Filter Aksi"
              options={ACTION_OPTIONS}
              value={filterAction}
              onChange={(e) => { setFilterAction(e.target.value); handleFilterChange(); }}
            />
          </div>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => { setFilterResource(''); setFilterAction(''); handleFilterChange(); }}
          >
            Reset Filter
          </Button>
        </div>
      </Card>

      {/* Logs Table */}
      <Card padding="none">
        <div className="p-4 border-b border-gray-100">
          <CardHeader
            title="Riwayat Aktivitas"
            description={`${pagination.total.toLocaleString('id-ID')} entri log ditemukan`}
          />
        </div>
        <Table
          columns={columns}
          data={logs}
          keyExtractor={(log) => log.id}
          isLoading={isLoading}
          emptyMessage="Tidak ada log yang cocok dengan filter"
        />
        {/* Pagination */}
        {pagination.totalPages > 1 && (
          <div className="flex items-center justify-between px-4 py-3 border-t border-gray-100">
            <span className="text-xs text-gray-500">
              Halaman {pagination.page} dari {pagination.totalPages}
            </span>
            <div className="flex items-center gap-1">
              <Button
                variant="ghost"
                size="sm"
                leftIcon={<ChevronLeft className="w-3.5 h-3.5" />}
                onClick={() => setPagination((prev) => ({ ...prev, page: Math.max(1, prev.page - 1) }))}
                disabled={pagination.page === 1}
              >
                Sebelumnya
              </Button>
              <Button
                variant="ghost"
                size="sm"
                leftIcon={<ChevronRight className="w-3.5 h-3.5" />}
                onClick={() => setPagination((prev) => ({ ...prev, page: Math.min(prev.totalPages, prev.page + 1) }))}
                disabled={pagination.page === pagination.totalPages}
              >
                Berikutnya
              </Button>
            </div>
          </div>
        )}
      </Card>
    </div>
  );
}
