'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Plus, AlertOctagon, ShieldAlert, CheckCircle2, Clock } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import api from '@/lib/api';
import { Card, CardHeader } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input, Select } from '@/components/ui/Input';
import { Table } from '@/components/ui/Table';
import { Pagination } from '@/components/ui/Pagination';
import { Modal } from '@/components/ui/Modal';
import { Badge } from '@/components/ui/Badge';
import { formatDate, formatDateTime } from '@/lib/utils';

interface AntiBullyingReport {
  id: string;
  type: string;
  description: string;
  location: string;
  incidentDate: string;
  status: 'OPEN' | 'INVESTIGATING' | 'RESOLVED' | 'CLOSED';
  isAnonymous: boolean;
  reportedBy?: string;
  createdAt: string;
}

const TYPE_LABELS: Record<string, string> = {
  VERBAL: 'Verbal',
  PHYSICAL: 'Fisik',
  CYBERBULLYING: 'Cyberbullying',
  SEXUAL: 'Seksual',
  OTHER: 'Lainnya',
};

const STATUS_CONFIG: Record<string, { label: string; variant: 'danger' | 'warning' | 'success' | 'gray' }> = {
  OPEN: { label: 'Terbuka', variant: 'danger' },
  INVESTIGATING: { label: 'Investigasi', variant: 'warning' },
  RESOLVED: { label: 'Selesai', variant: 'success' },
  CLOSED: { label: 'Ditutup', variant: 'gray' },
};

export default function AntiBullyingPage() {
  const { user } = useAuth();
  const [reports, setReports] = useState<AntiBullyingReport[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [meta, setMeta] = useState({ total: 0, totalPages: 1 });
  const [addOpen, setAddOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Form state
  const [fType, setFType] = useState('VERBAL');
  const [fDescription, setFDescription] = useState('');
  const [fLocation, setFLocation] = useState('');
  const [fDate, setFDate] = useState('');
  const [fAnonymous, setFAnonymous] = useState(true);
  const [fReportedBy, setFReportedBy] = useState('');

  const fetchReports = useCallback(async () => {
    setIsLoading(true);
    try {
      const params = new URLSearchParams({ page: String(page), limit: '20' });
      const res = await api.get(`/anti-bullying/reports?${params}`);
      setReports(res.data.data || []);
      setMeta(res.data.meta || { total: 0, totalPages: 1 });
    } catch {
      setReports([]);
    } finally {
      setIsLoading(false);
    }
  }, [page]);

  useEffect(() => { fetchReports(); }, [fetchReports]);

  const resetForm = () => {
    setFType('VERBAL');
    setFDescription('');
    setFLocation('');
    setFDate('');
    setFAnonymous(true);
    setFReportedBy('');
  };

  const handleAdd = async () => {
    if (!fDescription || !fDate) return;
    setIsSubmitting(true);
    try {
      await api.post('/anti-bullying/reports', {
        type: fType,
        description: fDescription,
        location: fLocation,
        incidentDate: fDate,
        isAnonymous: fAnonymous,
        reportedBy: fAnonymous ? undefined : fReportedBy || undefined,
      });
      setAddOpen(false);
      resetForm();
      fetchReports();
    } finally {
      setIsSubmitting(false);
    }
  };

  const openCount = reports.filter((r) => r.status === 'OPEN').length;
  const resolvedCount = reports.filter((r) => r.status === 'RESOLVED').length;

  const columns = [
    {
      key: 'type',
      header: 'Jenis',
      render: (r: AntiBullyingReport) => (
        <Badge variant="danger" size="sm">{TYPE_LABELS[r.type] || r.type}</Badge>
      ),
    },
    {
      key: 'description',
      header: 'Deskripsi',
      render: (r: AntiBullyingReport) => (
        <p className="max-w-xs text-sm text-gray-700 truncate">{r.description}</p>
      ),
    },
    {
      key: 'location',
      header: 'Lokasi',
      render: (r: AntiBullyingReport) => r.location || <span className="text-gray-300">—</span>,
    },
    {
      key: 'incidentDate',
      header: 'Tanggal',
      render: (r: AntiBullyingReport) => formatDate(r.incidentDate),
    },
    {
      key: 'status',
      header: 'Status',
      render: (r: AntiBullyingReport) => {
        const cfg = STATUS_CONFIG[r.status] || { label: r.status, variant: 'gray' as const };
        return <Badge variant={cfg.variant} size="sm">{cfg.label}</Badge>;
      },
    },
    {
      key: 'isAnonymous',
      header: 'Anonim',
      render: (r: AntiBullyingReport) => (
        <span className="text-sm text-gray-700">{r.isAnonymous ? '✓' : '✗'}</span>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900">Anti-Bullying & Keamanan</h1>
          <p className="text-sm text-gray-500 mt-0.5">Laporan dan penanganan kasus bullying di sekolah</p>
        </div>
        <Button
          size="sm"
          leftIcon={<Plus className="w-3.5 h-3.5" />}
          onClick={() => setAddOpen(true)}
        >
          Buat Laporan
        </Button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-3 gap-4">
        <Card padding="sm">
          <div className="flex items-center gap-3">
            <div className="p-2.5 bg-red-50 rounded-lg">
              <AlertOctagon className="w-5 h-5 text-red-600" />
            </div>
            <div>
              <p className="text-xl font-bold text-gray-900">{meta.total}</p>
              <p className="text-xs text-gray-500">Total Laporan</p>
            </div>
          </div>
        </Card>
        <Card padding="sm">
          <div className="flex items-center gap-3">
            <div className="p-2.5 bg-amber-50 rounded-lg">
              <Clock className="w-5 h-5 text-amber-600" />
            </div>
            <div>
              <p className="text-xl font-bold text-gray-900">{openCount}</p>
              <p className="text-xs text-gray-500">Laporan Terbuka</p>
            </div>
          </div>
        </Card>
        <Card padding="sm">
          <div className="flex items-center gap-3">
            <div className="p-2.5 bg-green-50 rounded-lg">
              <CheckCircle2 className="w-5 h-5 text-green-600" />
            </div>
            <div>
              <p className="text-xl font-bold text-gray-900">{resolvedCount}</p>
              <p className="text-xs text-gray-500">Laporan Selesai</p>
            </div>
          </div>
        </Card>
      </div>

      <Card padding="none">
        <Table
          columns={columns}
          data={reports}
          keyExtractor={(r) => r.id}
          isLoading={isLoading}
          emptyMessage="Belum ada laporan bullying"
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

      {/* Add Report Modal */}
      <Modal
        isOpen={addOpen}
        onClose={() => { setAddOpen(false); resetForm(); }}
        title="Buat Laporan Bullying"
        description="Laporan dapat disampaikan secara anonim untuk keamanan pelapor"
        size="md"
        footer={
          <>
            <Button variant="secondary" onClick={() => { setAddOpen(false); resetForm(); }}>Batal</Button>
            <Button variant="danger" onClick={handleAdd} isLoading={isSubmitting}>Kirim Laporan</Button>
          </>
        }
      >
        <div className="space-y-4">
          <Select
            label="Jenis Bullying"
            value={fType}
            onChange={(e) => setFType(e.target.value)}
            options={[
              { value: 'VERBAL', label: 'Verbal' },
              { value: 'PHYSICAL', label: 'Fisik' },
              { value: 'CYBERBULLYING', label: 'Cyberbullying' },
              { value: 'SEXUAL', label: 'Seksual' },
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
              placeholder="Ceritakan kejadian secara detail..."
              value={fDescription}
              onChange={(e) => setFDescription(e.target.value)}
            />
          </div>
          <Input
            label="Lokasi Kejadian"
            placeholder="Contoh: Kantin, Kelas 8A, Toilet lantai 2..."
            value={fLocation}
            onChange={(e) => setFLocation(e.target.value)}
          />
          <Input
            label="Tanggal Kejadian"
            type="date"
            required
            value={fDate}
            onChange={(e) => setFDate(e.target.value)}
          />
          <div className="flex items-center gap-3 p-3 bg-gray-50 rounded-lg">
            <input
              type="checkbox"
              id="isAnonymous"
              checked={fAnonymous}
              onChange={(e) => setFAnonymous(e.target.checked)}
              className="w-4 h-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
            />
            <div>
              <label htmlFor="isAnonymous" className="text-sm font-medium text-gray-700">
                Laporan Anonim
              </label>
              <p className="text-xs text-gray-400">Identitas pelapor tidak akan ditampilkan</p>
            </div>
          </div>
          {!fAnonymous && (
            <Input
              label="Nama Pelapor (opsional)"
              placeholder="Nama lengkap pelapor..."
              value={fReportedBy}
              onChange={(e) => setFReportedBy(e.target.value)}
            />
          )}
        </div>
      </Modal>
    </div>
  );
}
