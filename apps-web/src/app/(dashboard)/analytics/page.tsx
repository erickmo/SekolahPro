'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { BarChart3, Download, RefreshCw } from 'lucide-react';
import api from '@/lib/api';
import { Card, CardHeader } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { Modal } from '@/components/ui/Modal';
import { Input, Select } from '@/components/ui/Input';

interface OverviewStats {
  totalStudents: number;
  attendanceRate: number;
  averageGrade: number;
  passRate: number;
}

interface TrendPoint {
  label: string;
  value: number;
  period: string;
}

const METRIC_OPTIONS = [
  { value: 'attendance', label: 'Kehadiran' },
  { value: 'grade', label: 'Nilai Rata-rata' },
  { value: 'pass_rate', label: 'Tingkat Kelulusan' },
  { value: 'enrollment', label: 'Jumlah Siswa' },
];

export default function AnalyticsPage() {
  const [overview, setOverview] = useState<OverviewStats | null>(null);
  const [trends, setTrends] = useState<TrendPoint[]>([]);
  const [selectedMetric, setSelectedMetric] = useState('attendance');
  const [isLoadingOverview, setIsLoadingOverview] = useState(true);
  const [isLoadingTrends, setIsLoadingTrends] = useState(false);
  const [isExportOpen, setIsExportOpen] = useState(false);
  const [isExporting, setIsExporting] = useState(false);
  const [exportForm, setExportForm] = useState({ format: 'xlsx', startDate: '', endDate: '' });

  const fetchOverview = useCallback(async () => {
    setIsLoadingOverview(true);
    try {
      const res = await api.get('/analytics/overview');
      setOverview(res.data.data || null);
    } catch {
      setOverview(null);
    } finally {
      setIsLoadingOverview(false);
    }
  }, []);

  const fetchTrends = useCallback(async (metric: string) => {
    setIsLoadingTrends(true);
    try {
      const res = await api.get(`/analytics/trends?metric=${metric}`);
      setTrends(res.data.data || []);
    } catch {
      setTrends([]);
    } finally {
      setIsLoadingTrends(false);
    }
  }, []);

  useEffect(() => {
    fetchOverview();
  }, [fetchOverview]);

  useEffect(() => {
    fetchTrends(selectedMetric);
  }, [selectedMetric, fetchTrends]);

  const handleExport = async () => {
    setIsExporting(true);
    try {
      await api.post('/analytics/export', exportForm);
      setIsExportOpen(false);
    } catch {
      // silent
    } finally {
      setIsExporting(false);
    }
  };

  const statCards = overview
    ? [
        { label: 'Total Siswa', value: overview.totalStudents.toLocaleString('id-ID'), color: 'bg-blue-100 text-blue-700', suffix: 'siswa' },
        { label: 'Tingkat Kehadiran', value: `${overview.attendanceRate.toFixed(1)}%`, color: 'bg-green-100 text-green-700', suffix: '' },
        { label: 'Nilai Rata-rata', value: overview.averageGrade.toFixed(1), color: 'bg-yellow-100 text-yellow-700', suffix: '' },
        { label: 'Tingkat Kelulusan', value: `${overview.passRate.toFixed(1)}%`, color: 'bg-purple-100 text-purple-700', suffix: '' },
      ]
    : [];

  const maxTrendValue = trends.length > 0 ? Math.max(...trends.map((t) => t.value)) : 1;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="p-2 bg-orange-100 rounded-xl">
            <BarChart3 className="w-5 h-5 text-orange-600" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-gray-900">Analitik & Laporan</h1>
            <p className="text-sm text-gray-500 mt-0.5">Data statistik dan tren kinerja sekolah</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="secondary"
            size="sm"
            leftIcon={<RefreshCw className="w-3.5 h-3.5" />}
            onClick={() => { fetchOverview(); fetchTrends(selectedMetric); }}
          >
            Refresh
          </Button>
          <Button
            size="sm"
            leftIcon={<Download className="w-3.5 h-3.5" />}
            onClick={() => setIsExportOpen(true)}
          >
            Ekspor Laporan
          </Button>
        </div>
      </div>

      {/* Overview Stats Cards */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        {isLoadingOverview
          ? Array.from({ length: 4 }).map((_, i) => (
              <Card key={i}>
                <div className="animate-pulse space-y-2">
                  <div className="h-3 bg-gray-200 rounded w-3/4" />
                  <div className="h-8 bg-gray-200 rounded w-1/2" />
                </div>
              </Card>
            ))
          : statCards.map((stat) => (
              <Card key={stat.label}>
                <p className="text-sm text-gray-500">{stat.label}</p>
                <div className="mt-2 flex items-end gap-1">
                  <span className={`text-2xl font-bold px-2 py-0.5 rounded-lg ${stat.color}`}>
                    {stat.value}
                  </span>
                  {stat.suffix && (
                    <span className="text-xs text-gray-400 mb-1">{stat.suffix}</span>
                  )}
                </div>
              </Card>
            ))}
      </div>

      {/* Trends Section */}
      <Card padding="none">
        <div className="p-4 border-b border-gray-100 flex items-center justify-between">
          <CardHeader
            title="Tren Data"
            description="Perubahan data per periode waktu"
          />
          <div className="w-52">
            <Select
              options={METRIC_OPTIONS}
              value={selectedMetric}
              onChange={(e) => setSelectedMetric(e.target.value)}
            />
          </div>
        </div>
        <div className="p-4">
          {isLoadingTrends ? (
            <div className="animate-pulse space-y-3">
              {Array.from({ length: 5 }).map((_, i) => (
                <div key={i} className="h-8 bg-gray-100 rounded" />
              ))}
            </div>
          ) : trends.length === 0 ? (
            <p className="text-sm text-gray-400 text-center py-8">Belum ada data tren tersedia</p>
          ) : (
            <div className="space-y-3">
              {trends.map((point) => (
                <div key={point.period} className="flex items-center gap-3">
                  <span className="text-xs text-gray-500 w-24 shrink-0">{point.label}</span>
                  <div className="flex-1 h-6 bg-gray-100 rounded-full overflow-hidden">
                    <div
                      className="h-full bg-orange-400 rounded-full transition-all duration-500"
                      style={{ width: `${(point.value / maxTrendValue) * 100}%` }}
                    />
                  </div>
                  <span className="text-xs font-semibold text-gray-700 w-16 text-right">
                    {point.value.toLocaleString('id-ID')}
                  </span>
                  <Badge variant="gray" size="sm">{point.period}</Badge>
                </div>
              ))}
            </div>
          )}
        </div>
      </Card>

      {/* Export Modal */}
      <Modal
        isOpen={isExportOpen}
        onClose={() => setIsExportOpen(false)}
        title="Ekspor Laporan Analitik"
        description="Unduh data analitik dalam format yang diinginkan"
        footer={
          <>
            <Button variant="secondary" onClick={() => setIsExportOpen(false)}>Batal</Button>
            <Button onClick={handleExport} isLoading={isExporting}>Ekspor</Button>
          </>
        }
      >
        <div className="space-y-4">
          <Select
            label="Format File"
            options={[
              { value: 'xlsx', label: 'Excel (.xlsx)' },
              { value: 'csv', label: 'CSV (.csv)' },
              { value: 'pdf', label: 'PDF (.pdf)' },
            ]}
            value={exportForm.format}
            onChange={(e) => setExportForm({ ...exportForm, format: e.target.value })}
          />
          <Input
            label="Tanggal Mulai"
            type="date"
            value={exportForm.startDate}
            onChange={(e) => setExportForm({ ...exportForm, startDate: e.target.value })}
          />
          <Input
            label="Tanggal Akhir"
            type="date"
            value={exportForm.endDate}
            onChange={(e) => setExportForm({ ...exportForm, endDate: e.target.value })}
          />
        </div>
      </Modal>
    </div>
  );
}
