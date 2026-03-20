'use client';

import React, { useState, useCallback } from 'react';
import { Plus, Activity, Search } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import api from '@/lib/api';
import { Card, CardHeader } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Table } from '@/components/ui/Table';
import { Pagination } from '@/components/ui/Pagination';
import { Modal } from '@/components/ui/Modal';
import { Badge } from '@/components/ui/Badge';
import { formatDateTime } from '@/lib/utils';

interface NutritionRecord {
  id: string;
  studentId: string;
  weight: number;
  height: number;
  bmi: number;
  nutritionStatus: 'NORMAL' | 'KURANG' | 'LEBIH' | 'OBESITAS';
  measuredAt: string;
}

const STATUS_CONFIG: Record<string, { label: string; variant: 'success' | 'warning' | 'danger' | 'gray' }> = {
  NORMAL: { label: 'Normal', variant: 'success' },
  KURANG: { label: 'Kurang', variant: 'warning' },
  LEBIH: { label: 'Lebih', variant: 'danger' },
  OBESITAS: { label: 'Obesitas', variant: 'danger' },
};

export default function NutritionPage() {
  const { user } = useAuth();
  const [searchId, setSearchId] = useState('');
  const [activeStudentId, setActiveStudentId] = useState('');
  const [records, setRecords] = useState<NutritionRecord[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [meta, setMeta] = useState({ total: 0, totalPages: 1 });
  const [addOpen, setAddOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [searched, setSearched] = useState(false);

  // Form state
  const [fStudentId, setFStudentId] = useState('');
  const [fWeight, setFWeight] = useState('');
  const [fHeight, setFHeight] = useState('');
  const [fMeasuredAt, setFMeasuredAt] = useState('');

  const fetchRecords = useCallback(async (studentId: string, pageNum = 1) => {
    if (!studentId) return;
    setIsLoading(true);
    try {
      const params = new URLSearchParams({ page: String(pageNum), limit: '20' });
      const res = await api.get(`/health/nutrition/${studentId}?${params}`);
      setRecords(res.data.data || []);
      setMeta(res.data.meta || { total: 0, totalPages: 1 });
    } catch {
      setRecords([]);
    } finally {
      setIsLoading(false);
    }
  }, []);

  const handleSearch = () => {
    if (!searchId.trim()) return;
    setActiveStudentId(searchId.trim());
    setPage(1);
    setSearched(true);
    fetchRecords(searchId.trim(), 1);
  };

  const handlePageChange = (p: number) => {
    setPage(p);
    fetchRecords(activeStudentId, p);
  };

  const resetForm = () => {
    setFStudentId('');
    setFWeight('');
    setFHeight('');
    setFMeasuredAt('');
  };

  const handleAdd = async () => {
    if (!fStudentId || !fWeight || !fHeight || !fMeasuredAt) return;
    setIsSubmitting(true);
    try {
      await api.post('/health/nutrition', {
        studentId: fStudentId,
        weight: parseFloat(fWeight),
        height: parseFloat(fHeight),
        measuredAt: fMeasuredAt,
      });
      setAddOpen(false);
      resetForm();
      if (activeStudentId === fStudentId) {
        fetchRecords(activeStudentId, page);
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  const columns = [
    {
      key: 'measuredAt',
      header: 'Tanggal',
      render: (r: NutritionRecord) => formatDateTime(r.measuredAt),
    },
    {
      key: 'weight',
      header: 'BB (kg)',
      render: (r: NutritionRecord) => <span className="font-medium">{r.weight}</span>,
    },
    {
      key: 'height',
      header: 'TB (cm)',
      render: (r: NutritionRecord) => r.height,
    },
    {
      key: 'bmi',
      header: 'BMI',
      render: (r: NutritionRecord) => r.bmi?.toFixed(1) ?? '—',
    },
    {
      key: 'nutritionStatus',
      header: 'Status Gizi',
      render: (r: NutritionRecord) => {
        const cfg = STATUS_CONFIG[r.nutritionStatus] || { label: r.nutritionStatus, variant: 'gray' as const };
        return <Badge variant={cfg.variant} size="sm">{cfg.label}</Badge>;
      },
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900">Pemantauan Gizi Siswa</h1>
          <p className="text-sm text-gray-500 mt-0.5">Rekam dan pantau status gizi siswa</p>
        </div>
        <Button
          size="sm"
          leftIcon={<Plus className="w-3.5 h-3.5" />}
          onClick={() => setAddOpen(true)}
        >
          Tambah Data Gizi
        </Button>
      </div>

      {/* Search Box */}
      <Card padding="sm">
        <div className="flex items-end gap-3">
          <div className="flex-1">
            <Input
              label="Cari Data Siswa"
              placeholder="Masukkan ID Siswa..."
              value={searchId}
              onChange={(e) => setSearchId(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
            />
          </div>
          <Button
            leftIcon={<Search className="w-4 h-4" />}
            onClick={handleSearch}
            disabled={!searchId.trim()}
          >
            Cari
          </Button>
        </div>
      </Card>

      {!searched && (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <div className="p-4 bg-primary-50 rounded-full mb-4">
            <Activity className="w-8 h-8 text-primary-400" />
          </div>
          <p className="text-sm font-medium text-gray-600">Cari Data Siswa</p>
          <p className="text-xs text-gray-400 mt-1">Masukkan ID siswa untuk melihat riwayat data gizi</p>
        </div>
      )}

      {searched && (
        <>
          <div className="flex items-center gap-2">
            <Activity className="w-4 h-4 text-primary-600" />
            <p className="text-sm font-medium text-gray-700">
              Data gizi untuk siswa: <span className="text-primary-700">{activeStudentId}</span>
            </p>
            <Badge variant="info" size="sm">{meta.total} catatan</Badge>
          </div>

          <Card padding="none">
            <Table
              columns={columns}
              data={records}
              keyExtractor={(r) => r.id}
              isLoading={isLoading}
              emptyMessage="Belum ada data gizi untuk siswa ini"
            />
            {meta.total > 0 && (
              <div className="px-4 py-3 border-t border-gray-100">
                <Pagination
                  currentPage={page}
                  totalPages={meta.totalPages}
                  total={meta.total}
                  limit={20}
                  onPageChange={handlePageChange}
                />
              </div>
            )}
          </Card>
        </>
      )}

      {/* Add Nutrition Modal */}
      <Modal
        isOpen={addOpen}
        onClose={() => { setAddOpen(false); resetForm(); }}
        title="Tambah Data Gizi Siswa"
        size="sm"
        footer={
          <>
            <Button variant="secondary" onClick={() => { setAddOpen(false); resetForm(); }}>Batal</Button>
            <Button onClick={handleAdd} isLoading={isSubmitting}>Simpan</Button>
          </>
        }
      >
        <div className="space-y-4">
          <Input
            label="ID Siswa"
            placeholder="ID siswa dari SIMS"
            required
            value={fStudentId}
            onChange={(e) => setFStudentId(e.target.value)}
          />
          <div className="grid grid-cols-2 gap-4">
            <Input
              label="Berat Badan (kg)"
              type="number"
              placeholder="Contoh: 45"
              required
              value={fWeight}
              onChange={(e) => setFWeight(e.target.value)}
            />
            <Input
              label="Tinggi Badan (cm)"
              type="number"
              placeholder="Contoh: 155"
              required
              value={fHeight}
              onChange={(e) => setFHeight(e.target.value)}
            />
          </div>
          <Input
            label="Tanggal & Waktu Pengukuran"
            type="datetime-local"
            required
            value={fMeasuredAt}
            onChange={(e) => setFMeasuredAt(e.target.value)}
          />
        </div>
      </Modal>
    </div>
  );
}
