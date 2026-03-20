'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Plus, Search, Upload, RefreshCw, User } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import api from '@/lib/api';
import type { Student } from '@/types';
import { Card, CardHeader } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input, Select } from '@/components/ui/Input';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { Modal } from '@/components/ui/Modal';
import { Pagination } from '@/components/ui/Pagination';
import { RiskScoreBadge } from '@/components/ui/Badge';
import { formatDate, genderLabel } from '@/lib/utils';

const createStudentSchema = z.object({
  name: z.string().min(2, 'Nama minimal 2 karakter'),
  nisn: z.string().min(10, 'NISN harus 10 digit').max(10, 'NISN harus 10 digit'),
  nik: z.string().optional(),
  birthDate: z.string().min(1, 'Tanggal lahir wajib diisi'),
  birthPlace: z.string().optional(),
  gender: z.enum(['MALE', 'FEMALE']),
  address: z.string().optional(),
  guardianName: z.string().min(2, 'Nama wali wajib diisi'),
  guardianPhone: z.string().min(10, 'Nomor telepon minimal 10 digit'),
  guardianRelationship: z.string().default('PARENT'),
});

type CreateStudentForm = z.infer<typeof createStudentSchema>;

export default function StudentsPage() {
  const router = useRouter();
  const { user } = useAuth();
  const [students, setStudents] = useState<Student[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [page, setPage] = useState(1);
  const [meta, setMeta] = useState({ total: 0, totalPages: 1 });
  const [isAddOpen, setIsAddOpen] = useState(false);

  const canCreate = ['ADMIN_SEKOLAH', 'OPERATOR_SIMS', 'TATA_USAHA'].includes(user?.role || '');

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<CreateStudentForm>({
    resolver: zodResolver(createStudentSchema),
    defaultValues: { gender: 'MALE', guardianRelationship: 'PARENT' },
  });

  const fetchStudents = useCallback(async () => {
    setIsLoading(true);
    try {
      const params = new URLSearchParams({
        page: String(page),
        limit: '20',
        ...(search && { search }),
      });
      const res = await api.get(`/students?${params}`);
      setStudents(res.data.data || []);
      setMeta(res.data.meta || { total: 0, totalPages: 1 });
    } catch {
      setStudents([]);
    } finally {
      setIsLoading(false);
    }
  }, [page, search]);

  useEffect(() => {
    fetchStudents();
  }, [fetchStudents]);

  const onSearch = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setPage(1);
    fetchStudents();
  };

  const onAddStudent = async (values: CreateStudentForm) => {
    await api.post('/students', {
      name: values.name,
      nisn: values.nisn,
      nik: values.nik,
      birthDate: values.birthDate,
      birthPlace: values.birthPlace,
      gender: values.gender,
      address: values.address,
      guardians: [
        {
          name: values.guardianName,
          relationship: values.guardianRelationship,
          phone: values.guardianPhone,
        },
      ],
    });
    reset();
    setIsAddOpen(false);
    fetchStudents();
  };

  const columns = [
    {
      key: 'photo',
      header: '',
      render: (s: Student) => (
        <div className="w-8 h-8 rounded-full bg-primary-100 flex items-center justify-center flex-shrink-0">
          {s.photoUrl ? (
            <img src={s.photoUrl} alt={s.name} className="w-8 h-8 rounded-full object-cover" />
          ) : (
            <User className="w-4 h-4 text-primary-500" />
          )}
        </div>
      ),
      className: 'w-10',
    },
    {
      key: 'name',
      header: 'Nama Siswa',
      render: (s: Student) => (
        <div>
          <p className="font-medium text-gray-900">{s.name}</p>
          <p className="text-xs text-gray-400">NISN: {s.nisn}</p>
        </div>
      ),
    },
    {
      key: 'gender',
      header: 'Jenis Kelamin',
      render: (s: Student) => (
        <Badge variant={s.gender === 'MALE' ? 'info' : 'primary'} size="sm">
          {genderLabel(s.gender)}
        </Badge>
      ),
    },
    {
      key: 'birthDate',
      header: 'Tanggal Lahir',
      render: (s: Student) => formatDate(s.birthDate),
    },
    {
      key: 'currentClass',
      header: 'Kelas',
      render: (s: Student) => s.currentClass || <span className="text-gray-400">—</span>,
    },
    {
      key: 'riskScore',
      header: 'Risiko',
      render: (s: Student) =>
        s.riskScore !== undefined ? (
          <RiskScoreBadge score={s.riskScore} />
        ) : (
          <span className="text-gray-300 text-xs">—</span>
        ),
    },
    {
      key: 'actions',
      header: '',
      render: (s: Student) => (
        <Button
          variant="ghost"
          size="sm"
          onClick={(e) => {
            e.stopPropagation();
            router.push(`/dashboard/students/${s.id}`);
          }}
        >
          Detail
        </Button>
      ),
      className: 'w-20',
    },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900">Data Siswa</h1>
          <p className="text-sm text-gray-500 mt-0.5">
            Manajemen data siswa sekolah — {meta.total} siswa terdaftar
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="secondary" size="sm" leftIcon={<RefreshCw className="w-3.5 h-3.5" />} onClick={fetchStudents}>
            Refresh
          </Button>
          {canCreate && (
            <>
              <Button variant="secondary" size="sm" leftIcon={<Upload className="w-3.5 h-3.5" />}>
                Import Excel
              </Button>
              <Button size="sm" leftIcon={<Plus className="w-3.5 h-3.5" />} onClick={() => setIsAddOpen(true)}>
                Tambah Siswa
              </Button>
            </>
          )}
        </div>
      </div>

      <Card padding="none">
        {/* Search bar */}
        <div className="p-4 border-b border-gray-100">
          <form onSubmit={onSearch} className="flex gap-3 max-w-md">
            <Input
              placeholder="Cari nama atau NISN..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              leftAddon={<Search className="w-4 h-4" />}
            />
            <Button type="submit" size="sm">Cari</Button>
          </form>
        </div>

        {/* Table */}
        <Table
          columns={columns}
          data={students}
          keyExtractor={(s) => s.id}
          isLoading={isLoading}
          emptyMessage="Tidak ada data siswa"
          onRowClick={(s) => router.push(`/dashboard/students/${s.id}`)}
        />

        {/* Pagination */}
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

      {/* Add Student Modal */}
      <Modal
        isOpen={isAddOpen}
        onClose={() => { setIsAddOpen(false); reset(); }}
        title="Tambah Siswa Baru"
        description="Isi data siswa dengan lengkap dan benar"
        size="lg"
        footer={
          <>
            <Button variant="secondary" onClick={() => { setIsAddOpen(false); reset(); }}>
              Batal
            </Button>
            <Button
              form="add-student-form"
              type="submit"
              isLoading={isSubmitting}
            >
              Simpan
            </Button>
          </>
        }
      >
        <form id="add-student-form" onSubmit={handleSubmit(onAddStudent)} className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="col-span-2">
              <Input label="Nama Lengkap" required {...register('name')} error={errors.name?.message} />
            </div>
            <Input label="NISN" placeholder="10 digit" required {...register('nisn')} error={errors.nisn?.message} />
            <Input label="NIK" placeholder="16 digit (opsional)" {...register('nik')} error={errors.nik?.message} />
            <Input label="Tanggal Lahir" type="date" required {...register('birthDate')} error={errors.birthDate?.message} />
            <Input label="Tempat Lahir" {...register('birthPlace')} error={errors.birthPlace?.message} />
            <Select
              label="Jenis Kelamin"
              required
              options={[{ value: 'MALE', label: 'Laki-laki' }, { value: 'FEMALE', label: 'Perempuan' }]}
              {...register('gender')}
              error={errors.gender?.message}
            />
            <div />
            <div className="col-span-2">
              <Input label="Alamat" {...register('address')} error={errors.address?.message} />
            </div>
          </div>

          <div className="pt-2 border-t border-gray-100">
            <p className="text-sm font-medium text-gray-700 mb-3">Data Wali Siswa</p>
            <div className="grid grid-cols-2 gap-4">
              <div className="col-span-2">
                <Input label="Nama Wali" required {...register('guardianName')} error={errors.guardianName?.message} />
              </div>
              <Input label="No. Telepon Wali" required {...register('guardianPhone')} error={errors.guardianPhone?.message} />
              <Select
                label="Hubungan"
                options={[
                  { value: 'PARENT', label: 'Orang Tua' },
                  { value: 'GUARDIAN', label: 'Wali' },
                  { value: 'SIBLING', label: 'Saudara' },
                  { value: 'OTHER', label: 'Lainnya' },
                ]}
                {...register('guardianRelationship')}
              />
            </div>
          </div>
        </form>
      </Modal>
    </div>
  );
}
