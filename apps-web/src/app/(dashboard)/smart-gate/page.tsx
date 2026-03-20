'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { LogIn, RefreshCw } from 'lucide-react';
import api from '@/lib/api';
import { Card, CardHeader, StatCard } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { Pagination } from '@/components/ui/Pagination';
import { Input } from '@/components/ui/Input';
import { formatDateTime } from '@/lib/utils';

interface GateLog {
  id: string;
  personName: string;
  personType: 'STUDENT' | 'TEACHER' | 'STAFF';
  type: 'ENTRY' | 'EXIT';
  timestamp: string;
  gate: string;
  method: 'RFID' | 'FACE' | 'QR';
}

interface GateStatus {
  totalEntry: number;
  totalExit: number;
  currentlyInSchool: number;
}

const TYPE_MAP: Record<string, { label: string; variant: 'success' | 'info' }> = {
  ENTRY: { label: 'Masuk', variant: 'success' },
  EXIT: { label: 'Keluar', variant: 'info' },
};

const METHOD_MAP: Record<string, { label: string; variant: 'primary' | 'warning' | 'gray' }> = {
  RFID: { label: 'RFID', variant: 'primary' },
  FACE: { label: 'Wajah', variant: 'warning' },
  QR: { label: 'QR Code', variant: 'gray' },
};

const PERSON_TYPE_LABELS: Record<string, string> = {
  STUDENT: 'Siswa',
  TEACHER: 'Guru',
  STAFF: 'Staf',
};

export default function SmartGatePage() {
  const [logs, setLogs] = useState<GateLog[]>([]);
  const [gateStatus, setGateStatus] = useState<GateStatus | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [meta, setMeta] = useState({ total: 0, totalPages: 1 });
  const [dateFilter, setDateFilter] = useState(() => {
    const today = new Date();
    return today.toISOString().split('T')[0];
  });

  const fetchData = useCallback(async () => {
    setIsLoading(true);
    try {
      const params = new URLSearchParams({
        page: String(page),
        limit: '20',
        ...(dateFilter && { date: dateFilter }),
      });
      const [logsRes, statusRes] = await Promise.all([
        api.get(`/smart-gate/logs?${params}`),
        api.get('/smart-gate/status'),
      ]);
      setLogs(logsRes.data.data || []);
      setMeta(logsRes.data.meta || { total: 0, totalPages: 1 });
      setGateStatus(statusRes.data || null);
    } catch {
      setLogs([]);
    } finally {
      setIsLoading(false);
    }
  }, [page, dateFilter]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const columns = [
    {
      key: 'personName',
      header: 'Nama',
      render: (l: GateLog) => (
        <div>
          <p className="font-medium text-gray-900">{l.personName}</p>
          <p className="text-xs text-gray-400">{PERSON_TYPE_LABELS[l.personType] || l.personType}</p>
        </div>
      ),
    },
    {
      key: 'type',
      header: 'Tipe',
      render: (l: GateLog) => {
        const t = TYPE_MAP[l.type] || { label: l.type, variant: 'default' as const };
        return (
          <Badge variant={t.variant}>
            {t.label}
          </Badge>
        );
      },
    },
    {
      key: 'timestamp',
      header: 'Waktu',
      render: (l: GateLog) => (
        <span className="text-sm text-gray-700">{formatDateTime(l.timestamp)}</span>
      ),
    },
    {
      key: 'gate',
      header: 'Gerbang',
      render: (l: GateLog) => (
        <span className="text-sm text-gray-700">{l.gate}</span>
      ),
    },
    {
      key: 'method',
      header: 'Metode',
      render: (l: GateLog) => {
        const m = METHOD_MAP[l.method] || { label: l.method, variant: 'default' as const };
        return <Badge variant={m.variant} size="sm">{m.label}</Badge>;
      },
    },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900">Smart Gate & Keamanan</h1>
          <p className="text-sm text-gray-500 mt-0.5">
            Pantau aktivitas masuk dan keluar gerbang sekolah secara real-time
          </p>
        </div>
        <Button
          variant="secondary"
          size="sm"
          leftIcon={<RefreshCw className="w-3.5 h-3.5" />}
          onClick={fetchData}
        >
          Refresh
        </Button>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <StatCard
          title="Total Masuk Hari Ini"
          value={gateStatus?.totalEntry ?? '—'}
          icon={<LogIn className="w-5 h-5 text-green-600" />}
          iconBg="bg-green-100"
        />
        <StatCard
          title="Total Keluar Hari Ini"
          value={gateStatus?.totalExit ?? '—'}
          icon={<LogIn className="w-5 h-5 text-amber-600" style={{ transform: 'scaleX(-1)' }} />}
          iconBg="bg-amber-100"
        />
        <StatCard
          title="Sedang di Sekolah"
          value={gateStatus?.currentlyInSchool ?? '—'}
          icon={<LogIn className="w-5 h-5 text-blue-600" />}
          iconBg="bg-blue-100"
        />
      </div>

      {/* Log Table */}
      <Card padding="none">
        <div className="p-4 border-b border-gray-100 flex items-center justify-between gap-4 flex-wrap">
          <CardHeader
            title="Log Akses Gerbang"
            description={`Menampilkan log tanggal ${dateFilter || 'semua'}`}
          />
          <Input
            type="date"
            value={dateFilter}
            onChange={(e) => { setDateFilter(e.target.value); setPage(1); }}
            className="w-40"
          />
        </div>

        <Table
          columns={columns}
          data={logs}
          keyExtractor={(l) => l.id}
          isLoading={isLoading}
          emptyMessage="Tidak ada log akses untuk tanggal ini"
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
