'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Plus, Search, Heart, Activity, AlertCircle } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import api from '@/lib/api';
import type { UKSVisit } from '@/types';
import { Card, CardHeader } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Table } from '@/components/ui/Table';
import { Pagination } from '@/components/ui/Pagination';
import { Modal } from '@/components/ui/Modal';
import { Badge } from '@/components/ui/Badge';
import { formatDateTime } from '@/lib/utils';

export default function HealthPage() {
  const { user } = useAuth();
  const [visits, setVisits] = useState<UKSVisit[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [meta, setMeta] = useState({ total: 0, totalPages: 1 });
  const [addOpen, setAddOpen] = useState(false);

  // Form state
  const [studentId, setStudentId] = useState('');
  const [complaint, setComplaint] = useState('');
  const [diagnosis, setDiagnosis] = useState('');
  const [treatment, setTreatment] = useState('');
  const [medicine, setMedicine] = useState('');
  const [referral, setReferral] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const canRecord = ['PETUGAS_UKS', 'ADMIN_SEKOLAH'].includes(user?.role || '');

  const fetchVisits = useCallback(async () => {
    setIsLoading(true);
    try {
      const params = new URLSearchParams({ page: String(page), limit: '20' });
      const res = await api.get(`/health/uks-visits?${params}`);
      setVisits(res.data.data || []);
      setMeta(res.data.meta || { total: 0, totalPages: 1 });
    } catch {
      setVisits([]);
    } finally {
      setIsLoading(false);
    }
  }, [page]);

  useEffect(() => {
    fetchVisits();
  }, [fetchVisits]);

  const handleAdd = async () => {
    if (!studentId || !complaint) return;
    setIsSubmitting(true);
    try {
      await api.post('/health/uks-visits', {
        studentId,
        date: new Date().toISOString(),
        complaint,
        diagnosis,
        treatment,
        medicine,
        referral,
      });
      setStudentId('');
      setComplaint('');
      setDiagnosis('');
      setTreatment('');
      setMedicine('');
      setReferral(false);
      setAddOpen(false);
      fetchVisits();
    } finally {
      setIsSubmitting(false);
    }
  };

  const columns = [
    {
      key: 'studentName',
      header: 'Siswa',
      render: (v: UKSVisit) => v.studentName || v.studentId,
    },
    {
      key: 'date',
      header: 'Waktu Kunjungan',
      render: (v: UKSVisit) => formatDateTime(v.date),
    },
    {
      key: 'complaint',
      header: 'Keluhan',
      render: (v: UKSVisit) => (
        <p className="max-w-xs text-sm text-gray-700 truncate">{v.complaint}</p>
      ),
    },
    {
      key: 'diagnosis',
      header: 'Diagnosis',
      render: (v: UKSVisit) => v.diagnosis || <span className="text-gray-300">—</span>,
    },
    {
      key: 'referral',
      header: 'Rujukan',
      render: (v: UKSVisit) =>
        v.referral ? (
          <Badge variant="warning" size="sm">Dirujuk</Badge>
        ) : (
          <Badge variant="success" size="sm">Tidak Dirujuk</Badge>
        ),
    },
    {
      key: 'staffName',
      header: 'Petugas',
      render: (v: UKSVisit) => v.staffName || '—',
    },
  ];

  const todayVisits = visits.filter((v) => {
    const today = new Date().toDateString();
    return new Date(v.date).toDateString() === today;
  });

  const referralCount = visits.filter((v) => v.referral).length;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900">UKS & Rekam Medis</h1>
          <p className="text-sm text-gray-500 mt-0.5">Pencatatan kunjungan dan rekam medis siswa</p>
        </div>
        {canRecord && (
          <Button size="sm" leftIcon={<Plus className="w-3.5 h-3.5" />} onClick={() => setAddOpen(true)}>
            Catat Kunjungan
          </Button>
        )}
      </div>

      {/* Stats */}
      <div className="grid grid-cols-3 gap-4">
        <Card padding="sm">
          <div className="flex items-center gap-3">
            <div className="p-2.5 bg-red-50 rounded-lg">
              <Heart className="w-5 h-5 text-red-600" />
            </div>
            <div>
              <p className="text-xl font-bold text-gray-900">{todayVisits.length}</p>
              <p className="text-xs text-gray-500">Kunjungan Hari Ini</p>
            </div>
          </div>
        </Card>
        <Card padding="sm">
          <div className="flex items-center gap-3">
            <div className="p-2.5 bg-primary-50 rounded-lg">
              <Activity className="w-5 h-5 text-primary-600" />
            </div>
            <div>
              <p className="text-xl font-bold text-gray-900">{meta.total}</p>
              <p className="text-xs text-gray-500">Total Kunjungan</p>
            </div>
          </div>
        </Card>
        <Card padding="sm">
          <div className="flex items-center gap-3">
            <div className="p-2.5 bg-amber-50 rounded-lg">
              <AlertCircle className="w-5 h-5 text-amber-600" />
            </div>
            <div>
              <p className="text-xl font-bold text-gray-900">{referralCount}</p>
              <p className="text-xs text-gray-500">Dirujuk</p>
            </div>
          </div>
        </Card>
      </div>

      <Card padding="none">
        <Table
          columns={columns}
          data={visits}
          keyExtractor={(v) => v.id}
          isLoading={isLoading}
          emptyMessage="Belum ada catatan kunjungan UKS"
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

      {/* Add Visit Modal */}
      <Modal
        isOpen={addOpen}
        onClose={() => setAddOpen(false)}
        title="Catat Kunjungan UKS"
        size="md"
        footer={
          <>
            <Button variant="secondary" onClick={() => setAddOpen(false)}>Batal</Button>
            <Button onClick={handleAdd} isLoading={isSubmitting}>Simpan</Button>
          </>
        }
      >
        <div className="space-y-4">
          <Input
            label="ID Siswa"
            placeholder="ID siswa dari SIMS"
            required
            value={studentId}
            onChange={(e) => setStudentId(e.target.value)}
          />
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Keluhan <span className="text-red-500">*</span>
            </label>
            <textarea
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
              rows={2}
              placeholder="Deskripsi keluhan siswa..."
              value={complaint}
              onChange={(e) => setComplaint(e.target.value)}
            />
          </div>
          <Input
            label="Diagnosis"
            placeholder="Diagnosis medis..."
            value={diagnosis}
            onChange={(e) => setDiagnosis(e.target.value)}
          />
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Tindakan</label>
            <textarea
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
              rows={2}
              placeholder="Tindakan yang diberikan..."
              value={treatment}
              onChange={(e) => setTreatment(e.target.value)}
            />
          </div>
          <Input
            label="Obat yang Diberikan"
            placeholder="Nama obat dan dosis..."
            value={medicine}
            onChange={(e) => setMedicine(e.target.value)}
          />
          <div className="flex items-center gap-3">
            <input
              type="checkbox"
              id="referral"
              checked={referral}
              onChange={(e) => setReferral(e.target.checked)}
              className="w-4 h-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
            />
            <label htmlFor="referral" className="text-sm text-gray-700">
              Perlu rujukan ke puskesmas/RS
            </label>
          </div>
        </div>
      </Modal>
    </div>
  );
}
