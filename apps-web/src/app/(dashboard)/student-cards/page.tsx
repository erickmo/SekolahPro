'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { CreditCard } from 'lucide-react';
import api from '@/lib/api';
import { Card, CardHeader } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { Modal } from '@/components/ui/Modal';
import { Input } from '@/components/ui/Input';

interface StudentCard {
  id: string;
  studentId: string;
  studentName: string;
  nisn: string;
  status: 'ACTIVE' | 'INACTIVE' | 'EXPIRED';
  issuedAt: string;
  expiresAt: string;
}

const STATUS_MAP: Record<string, { label: string; variant: 'success' | 'default' | 'danger' }> = {
  ACTIVE: { label: 'Aktif', variant: 'success' },
  INACTIVE: { label: 'Tidak Aktif', variant: 'default' },
  EXPIRED: { label: 'Kadaluarsa', variant: 'danger' },
};

export default function StudentCardsPage() {
  const [cards, setCards] = useState<StudentCard[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [downloadingId, setDownloadingId] = useState<string | null>(null);
  const [studentId, setStudentId] = useState('');

  const fetchCards = useCallback(async () => {
    setIsLoading(true);
    try {
      const res = await api.get('/student-cards');
      setCards(res.data.data || []);
    } catch {
      setCards([]);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchCards();
  }, [fetchCards]);

  const handleIssue = async () => {
    if (!studentId.trim()) return;
    setIsSubmitting(true);
    try {
      await api.post('/student-cards', { studentId });
      setIsModalOpen(false);
      setStudentId('');
      fetchCards();
    } catch {
      // handle error silently
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDownload = async (id: string) => {
    setDownloadingId(id);
    try {
      const res = await api.get(`/student-cards/${id}/download`, { responseType: 'blob' });
      const url = window.URL.createObjectURL(new Blob([res.data]));
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', `kartu-siswa-${id}.pdf`);
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.URL.revokeObjectURL(url);
    } catch {
      // handle error silently
    } finally {
      setDownloadingId(null);
    }
  };

  const activeCount = cards.filter((c) => c.status === 'ACTIVE').length;
  const expiredCount = cards.filter((c) => c.status === 'EXPIRED').length;

  const columns = [
    {
      key: 'student',
      header: 'Nama Siswa',
      render: (c: StudentCard) => (
        <div>
          <p className="font-medium text-gray-900">{c.studentName}</p>
          <p className="text-xs text-gray-400">NISN: {c.nisn}</p>
        </div>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (c: StudentCard) => {
        const s = STATUS_MAP[c.status] || { label: c.status, variant: 'default' as const };
        return <Badge variant={s.variant}>{s.label}</Badge>;
      },
    },
    {
      key: 'issuedAt',
      header: 'Tanggal Terbit',
      render: (c: StudentCard) => (
        <span className="text-sm text-gray-600">{new Date(c.issuedAt).toLocaleDateString('id-ID')}</span>
      ),
    },
    {
      key: 'expiresAt',
      header: 'Berlaku Hingga',
      render: (c: StudentCard) => (
        <span className="text-sm text-gray-600">{new Date(c.expiresAt).toLocaleDateString('id-ID')}</span>
      ),
    },
    {
      key: 'actions',
      header: '',
      render: (c: StudentCard) => (
        <Button
          variant="ghost"
          size="sm"
          leftIcon={<CreditCard className="w-3.5 h-3.5" />}
          isLoading={downloadingId === c.id}
          onClick={() => handleDownload(c.id)}
        >
          Unduh
        </Button>
      ),
      className: 'w-28',
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900 flex items-center gap-2">
            <CreditCard className="w-5 h-5 text-primary-600" />
            Kartu Siswa Digital
          </h1>
          <p className="text-sm text-gray-500 mt-0.5">
            Kelola dan cetak kartu identitas digital siswa
          </p>
        </div>
        <Button onClick={() => setIsModalOpen(true)}>Terbitkan Kartu Baru</Button>
      </div>

      <div className="grid grid-cols-3 gap-4">
        <Card>
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-full bg-primary-100 flex items-center justify-center">
              <CreditCard className="w-4 h-4 text-primary-600" />
            </div>
            <div>
              <p className="text-xs text-gray-500">Total Kartu</p>
              <p className="text-xl font-bold text-gray-900">{cards.length}</p>
            </div>
          </div>
        </Card>
        <Card>
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-full bg-green-100 flex items-center justify-center">
              <CreditCard className="w-4 h-4 text-green-600" />
            </div>
            <div>
              <p className="text-xs text-gray-500">Aktif</p>
              <p className="text-xl font-bold text-gray-900">{activeCount}</p>
            </div>
          </div>
        </Card>
        <Card>
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-full bg-red-100 flex items-center justify-center">
              <CreditCard className="w-4 h-4 text-red-600" />
            </div>
            <div>
              <p className="text-xs text-gray-500">Kadaluarsa</p>
              <p className="text-xl font-bold text-gray-900">{expiredCount}</p>
            </div>
          </div>
        </Card>
      </div>

      <Card padding="none">
        <div className="p-4 border-b border-gray-100">
          <CardHeader
            title="Daftar Kartu Siswa"
            description="Semua kartu siswa digital yang diterbitkan"
          />
        </div>
        <Table
          columns={columns}
          data={cards}
          keyExtractor={(c) => c.id}
          isLoading={isLoading}
          emptyMessage="Belum ada kartu siswa yang diterbitkan"
        />
      </Card>

      <Modal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        title="Terbitkan Kartu Siswa Baru"
      >
        <div className="space-y-4">
          <Input
            label="ID Siswa"
            placeholder="Masukkan ID siswa"
            value={studentId}
            onChange={(e) => setStudentId(e.target.value)}
          />
          <p className="text-xs text-gray-500">
            Masukkan ID siswa yang terdaftar di sistem untuk menerbitkan kartu digital baru.
          </p>
          <div className="flex justify-end gap-2 pt-2">
            <Button variant="secondary" onClick={() => setIsModalOpen(false)}>Batal</Button>
            <Button isLoading={isSubmitting} onClick={handleIssue} disabled={!studentId.trim()}>
              Terbitkan Kartu
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
