'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { AlertTriangle, RefreshCw, CheckCircle } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import api from '@/lib/api';
import { Card, CardHeader, StatCard } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { Pagination } from '@/components/ui/Pagination';
import { formatDateTime } from '@/lib/utils';

interface EwsAlert {
  id: string;
  studentName: string;
  alertType: string;
  riskLevel: 'HIGH' | 'MEDIUM' | 'LOW';
  message: string;
  status: 'ACTIVE' | 'RESOLVED';
  createdAt: string;
}

interface RiskStudent {
  id: string;
  name: string;
  class: string;
  riskScore: number;
  riskFactors: string[];
}

const RISK_LEVEL_MAP: Record<string, { label: string; variant: 'danger' | 'warning' | 'success' }> = {
  HIGH: { label: 'Tinggi', variant: 'danger' },
  MEDIUM: { label: 'Sedang', variant: 'warning' },
  LOW: { label: 'Rendah', variant: 'success' },
};

const ALERT_TYPE_LABELS: Record<string, string> = {
  ATTENDANCE: 'Kehadiran',
  ACADEMIC: 'Akademik',
  BEHAVIOR: 'Perilaku',
  HEALTH: 'Kesehatan',
  FINANCIAL: 'Keuangan',
};

export default function EwsPage() {
  const { user } = useAuth();
  const [alerts, setAlerts] = useState<EwsAlert[]>([]);
  const [riskStudents, setRiskStudents] = useState<RiskStudent[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isLoadingRisk, setIsLoadingRisk] = useState(true);
  const [page, setPage] = useState(1);
  const [meta, setMeta] = useState({ total: 0, totalPages: 1 });
  const [resolvingId, setResolvingId] = useState<string | null>(null);

  const canResolve = ['ADMIN_SEKOLAH', 'KEPALA_SEKOLAH', 'GURU_BK', 'WALI_KELAS'].includes(user?.role || '');

  const fetchAlerts = useCallback(async () => {
    setIsLoading(true);
    try {
      const res = await api.get(`/ews/alerts?page=${page}&limit=20`);
      setAlerts(res.data.data || []);
      setMeta(res.data.meta || { total: 0, totalPages: 1 });
    } catch {
      setAlerts([]);
    } finally {
      setIsLoading(false);
    }
  }, [page]);

  const fetchRiskStudents = useCallback(async () => {
    setIsLoadingRisk(true);
    try {
      const res = await api.get('/ews/risk-students');
      setRiskStudents(res.data.data || []);
    } catch {
      setRiskStudents([]);
    } finally {
      setIsLoadingRisk(false);
    }
  }, []);

  useEffect(() => {
    fetchAlerts();
    fetchRiskStudents();
  }, [fetchAlerts, fetchRiskStudents]);

  const handleResolve = async (id: string) => {
    setResolvingId(id);
    try {
      await api.patch(`/ews/alerts/${id}/resolve`, {});
      fetchAlerts();
    } catch {
      // handle
    } finally {
      setResolvingId(null);
    }
  };

  const activeAlerts = alerts.filter((a) => a.status === 'ACTIVE').length;
  const highRiskCount = alerts.filter((a) => a.riskLevel === 'HIGH').length;

  const alertColumns = [
    {
      key: 'student',
      header: 'Nama Siswa',
      render: (a: EwsAlert) => (
        <p className="font-medium text-gray-900">{a.studentName}</p>
      ),
    },
    {
      key: 'alertType',
      header: 'Jenis Alert',
      render: (a: EwsAlert) => (
        <span className="text-sm text-gray-700">
          {ALERT_TYPE_LABELS[a.alertType] || a.alertType}
        </span>
      ),
    },
    {
      key: 'riskLevel',
      header: 'Tingkat Risiko',
      render: (a: EwsAlert) => {
        const r = RISK_LEVEL_MAP[a.riskLevel] || { label: a.riskLevel, variant: 'default' as const };
        return <Badge variant={r.variant}>{r.label}</Badge>;
      },
    },
    {
      key: 'message',
      header: 'Pesan',
      render: (a: EwsAlert) => (
        <p className="text-sm text-gray-600 max-w-xs line-clamp-2">{a.message}</p>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (a: EwsAlert) => (
        <Badge variant={a.status === 'ACTIVE' ? 'warning' : 'success'}>
          {a.status === 'ACTIVE' ? 'Aktif' : 'Selesai'}
        </Badge>
      ),
    },
    {
      key: 'createdAt',
      header: 'Waktu',
      render: (a: EwsAlert) => (
        <span className="text-xs text-gray-400">{formatDateTime(a.createdAt)}</span>
      ),
    },
    ...(canResolve
      ? [
          {
            key: 'actions',
            header: '',
            render: (a: EwsAlert) =>
              a.status === 'ACTIVE' ? (
                <Button
                  variant="ghost"
                  size="sm"
                  leftIcon={<CheckCircle className="w-3.5 h-3.5" />}
                  isLoading={resolvingId === a.id}
                  onClick={() => handleResolve(a.id)}
                >
                  Selesaikan
                </Button>
              ) : (
                <span className="text-xs text-gray-300">—</span>
              ),
            className: 'w-32',
          },
        ]
      : []),
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900">Early Warning System (EWS)</h1>
          <p className="text-sm text-gray-500 mt-0.5">
            Deteksi dini siswa berisiko berdasarkan data akademik, kehadiran, dan perilaku
          </p>
        </div>
        <Button
          variant="secondary"
          size="sm"
          leftIcon={<RefreshCw className="w-3.5 h-3.5" />}
          onClick={() => { fetchAlerts(); fetchRiskStudents(); }}
        >
          Refresh
        </Button>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <StatCard
          title="Alert Aktif"
          value={activeAlerts}
          icon={<AlertTriangle className="w-5 h-5 text-amber-600" />}
          iconBg="bg-amber-100"
        />
        <StatCard
          title="Risiko Tinggi"
          value={highRiskCount}
          icon={<AlertTriangle className="w-5 h-5 text-red-600" />}
          iconBg="bg-red-100"
        />
        <StatCard
          title="Siswa Dipantau"
          value={riskStudents.length}
          icon={<AlertTriangle className="w-5 h-5 text-primary-600" />}
          iconBg="bg-primary-100"
        />
      </div>

      <div className="grid grid-cols-1 gap-6 xl:grid-cols-3">
        {/* Alerts Table */}
        <div className="xl:col-span-2">
          <Card padding="none">
            <div className="p-4 border-b border-gray-100">
              <CardHeader
                title="Daftar Alert"
                description="Alert aktif dan riwayat peringatan dini siswa"
              />
            </div>

            <Table
              columns={alertColumns}
              data={alerts}
              keyExtractor={(a) => a.id}
              isLoading={isLoading}
              emptyMessage="Tidak ada alert saat ini"
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

        {/* Risk Students List */}
        <div>
          <Card padding="none">
            <div className="p-4 border-b border-gray-100">
              <CardHeader
                title="Siswa Berisiko"
                description="Siswa dengan skor risiko tertinggi"
              />
            </div>

            {isLoadingRisk ? (
              <div className="p-6 text-center text-sm text-gray-400">Memuat data...</div>
            ) : riskStudents.length === 0 ? (
              <div className="p-6 text-center text-sm text-gray-400">
                Tidak ada siswa berisiko saat ini
              </div>
            ) : (
              <ul className="divide-y divide-gray-50">
                {riskStudents.map((s) => {
                  let scoreVariant: 'danger' | 'warning' | 'success' = 'success';
                  if (s.riskScore >= 70) scoreVariant = 'danger';
                  else if (s.riskScore >= 40) scoreVariant = 'warning';

                  return (
                    <li key={s.id} className="flex items-start gap-3 px-4 py-3">
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-medium text-gray-900">{s.name}</p>
                        <p className="text-xs text-gray-400">{s.class}</p>
                        {s.riskFactors?.length > 0 && (
                          <p className="text-xs text-gray-500 mt-0.5 line-clamp-1">
                            {s.riskFactors.join(', ')}
                          </p>
                        )}
                      </div>
                      <Badge variant={scoreVariant} size="sm">
                        {Math.round(s.riskScore)}
                      </Badge>
                    </li>
                  );
                })}
              </ul>
            )}
          </Card>
        </div>
      </div>
    </div>
  );
}
