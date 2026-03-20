'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Briefcase, Plus, RefreshCw, UserCheck, UserX } from 'lucide-react';
import api from '@/lib/api';
import { Card, CardHeader, StatCard } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { Modal } from '@/components/ui/Modal';
import { Pagination } from '@/components/ui/Pagination';
import { Input } from '@/components/ui/Input';

const STATUS_VARIANT: Record<string, 'success' | 'gray' | 'warning' | 'danger'> = {
  ACTIVE: 'success',
  INACTIVE: 'gray',
  ON_LEAVE: 'warning',
  RESIGNED: 'danger',
};

const STATUS_LABEL: Record<string, string> = {
  ACTIVE: 'Aktif',
  INACTIVE: 'Nonaktif',
  ON_LEAVE: 'Cuti',
  RESIGNED: 'Resign',
};

interface Teacher {
  id: string;
  name: string;
  nip: string;
  email: string;
  phone: string;
  subjects: string[];
  position: string;
  status: string;
}

interface AttendanceSummary {
  presentToday: number;
  absentToday: number;
}

interface AddTeacherForm {
  name: string;
  nip: string;
  email: string;
  phone: string;
  subjects: string;
  position: string;
}

const EMPTY_FORM: AddTeacherForm = {
  name: '',
  nip: '',
  email: '',
  phone: '',
  subjects: '',
  position: '',
};

export default function HrPage() {
  const [teachers, setTeachers] = useState<Teacher[]>([]);
  const [attendance, setAttendance] = useState<AttendanceSummary>({ presentToday: 0, absentToday: 0 });
  const [isLoading, setIsLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [meta, setMeta] = useState({ total: 0, totalPages: 1 });
  const [isAddOpen, setIsAddOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [form, setForm] = useState<AddTeacherForm>(EMPTY_FORM);

  const fetchTeachers = useCallback(async () => {
    setIsLoading(true);
    try {
      const params = new URLSearchParams({ page: String(page), limit: '20' });
      const res = await api.get(`/hr/teachers?${params}`);
      setTeachers(res.data.data || []);
      setMeta(res.data.meta || { total: 0, totalPages: 1 });
    } catch {
      setTeachers([]);
    } finally {
      setIsLoading(false);
    }
  }, [page]);

  const fetchAttendance = useCallback(async () => {
    try {
      const res = await api.get('/hr/attendance/summary/today');
      setAttendance(res.data.data || { presentToday: 0, absentToday: 0 });
    } catch {
      // silent
    }
  }, []);

  useEffect(() => {
    fetchTeachers();
    fetchAttendance();
  }, [fetchTeachers, fetchAttendance]);

  const handleAdd = async () => {
    setIsSubmitting(true);
    try {
      await api.post('/hr/teachers', {
        name: form.name,
        nip: form.nip,
        email: form.email,
        phone: form.phone,
        subjects: form.subjects
          .split(',')
          .map((s) => s.trim())
          .filter(Boolean),
        position: form.position,
      });
      setIsAddOpen(false);
      setForm(EMPTY_FORM);
      fetchTeachers();
    } catch {
      // silent
    } finally {
      setIsSubmitting(false);
    }
  };

  const columns = [
    {
      key: 'name',
      header: 'Nama',
      render: (t: Teacher) => (
        <div>
          <p className="font-medium text-gray-900">{t.name}</p>
          <p className="text-xs text-gray-400">{t.email}</p>
        </div>
      ),
    },
    {
      key: 'nip',
      header: 'NIP',
      render: (t: Teacher) => <span className="text-sm font-mono text-gray-600">{t.nip || '—'}</span>,
    },
    {
      key: 'subjects',
      header: 'Mata Pelajaran',
      render: (t: Teacher) => (
        <div className="flex flex-wrap gap-1">
          {(t.subjects || []).slice(0, 2).map((s) => (
            <Badge key={s} variant="info" size="sm">{s}</Badge>
          ))}
          {(t.subjects || []).length > 2 && (
            <Badge variant="gray" size="sm">+{t.subjects.length - 2}</Badge>
          )}
          {(!t.subjects || t.subjects.length === 0) && (
            <span className="text-gray-400 text-xs">—</span>
          )}
        </div>
      ),
    },
    {
      key: 'position',
      header: 'Jabatan',
      render: (t: Teacher) => <span className="text-sm text-gray-600">{t.position || '—'}</span>,
    },
    {
      key: 'status',
      header: 'Status',
      render: (t: Teacher) => (
        <Badge variant={STATUS_VARIANT[t.status] ?? 'default'} size="sm">
          {STATUS_LABEL[t.status] ?? t.status}
        </Badge>
      ),
    },
    {
      key: 'phone',
      header: 'Telepon',
      render: (t: Teacher) => <span className="text-sm text-gray-500">{t.phone || '—'}</span>,
    },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="p-2 bg-indigo-100 rounded-xl">
            <Briefcase className="w-5 h-5 text-indigo-600" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-gray-900">SDM & Kepegawaian</h1>
            <p className="text-sm text-gray-500 mt-0.5">Kelola data guru dan tenaga kependidikan</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="secondary" size="sm" leftIcon={<RefreshCw className="w-3.5 h-3.5" />} onClick={() => { fetchTeachers(); fetchAttendance(); }}>
            Refresh
          </Button>
          <Button size="sm" leftIcon={<Plus className="w-3.5 h-3.5" />} onClick={() => setIsAddOpen(true)}>
            Tambah Guru
          </Button>
        </div>
      </div>

      {/* Attendance summary */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title="Total Guru"
          value={meta.total}
          icon={<Briefcase className="w-5 h-5 text-indigo-600" />}
          iconBg="bg-indigo-100"
        />
        <StatCard
          title="Hadir Hari Ini"
          value={attendance.presentToday}
          icon={<UserCheck className="w-5 h-5 text-green-600" />}
          iconBg="bg-green-100"
        />
        <StatCard
          title="Tidak Hadir Hari Ini"
          value={attendance.absentToday}
          icon={<UserX className="w-5 h-5 text-red-600" />}
          iconBg="bg-red-100"
        />
        <StatCard
          title="Tingkat Kehadiran"
          value={
            meta.total > 0
              ? `${Math.round((attendance.presentToday / meta.total) * 100)}%`
              : '—'
          }
          icon={<Briefcase className="w-5 h-5 text-primary-600" />}
          iconBg="bg-primary-100"
        />
      </div>

      {/* Table */}
      <Card padding="none">
        <div className="p-4 border-b border-gray-100">
          <CardHeader title="Daftar Guru & Staf" description={`${meta.total} tenaga pendidik & kependidikan`} />
        </div>
        <Table
          columns={columns}
          data={teachers}
          keyExtractor={(t) => t.id}
          isLoading={isLoading}
          emptyMessage="Belum ada data guru"
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

      {/* Add Teacher Modal */}
      <Modal
        isOpen={isAddOpen}
        onClose={() => { setIsAddOpen(false); setForm(EMPTY_FORM); }}
        title="Tambah Guru / Staf"
        description="Daftarkan tenaga pendidik atau kependidikan baru"
        size="lg"
        footer={
          <>
            <Button variant="secondary" onClick={() => { setIsAddOpen(false); setForm(EMPTY_FORM); }}>Batal</Button>
            <Button onClick={handleAdd} isLoading={isSubmitting}>Simpan</Button>
          </>
        }
      >
        <div className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="col-span-2">
              <Input
                label="Nama Lengkap"
                required
                placeholder="Nama sesuai KTP"
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
              />
            </div>
            <Input
              label="NIP"
              placeholder="Nomor Induk Pegawai"
              value={form.nip}
              onChange={(e) => setForm({ ...form, nip: e.target.value })}
            />
            <Input
              label="No. Telepon"
              placeholder="08xxxxxxxxxx"
              value={form.phone}
              onChange={(e) => setForm({ ...form, phone: e.target.value })}
            />
            <div className="col-span-2">
              <Input
                label="Email"
                type="email"
                required
                placeholder="guru@sekolah.sch.id"
                value={form.email}
                onChange={(e) => setForm({ ...form, email: e.target.value })}
              />
            </div>
            <div className="col-span-2">
              <Input
                label="Mata Pelajaran"
                placeholder="Pisahkan dengan koma, contoh: Matematika, Fisika"
                value={form.subjects}
                onChange={(e) => setForm({ ...form, subjects: e.target.value })}
              />
            </div>
            <div className="col-span-2">
              <Input
                label="Jabatan"
                placeholder="Contoh: Guru, Wali Kelas, Kepala Sekolah"
                value={form.position}
                onChange={(e) => setForm({ ...form, position: e.target.value })}
              />
            </div>
          </div>
        </div>
      </Modal>
    </div>
  );
}
