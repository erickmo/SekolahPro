'use client';

import React, { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import Link from 'next/link';
import { Eye, EyeOff, School } from 'lucide-react';
import api from '@/lib/api';
import { Button } from '@/components/ui/Button';
import { Input, Select } from '@/components/ui/Input';

const registerSchema = z
  .object({
    name: z.string().min(2, 'Nama minimal 2 karakter'),
    email: z.string().email('Format email tidak valid'),
    password: z.string().min(8, 'Password minimal 8 karakter'),
    confirmPassword: z.string(),
    schoolId: z.string().min(1, 'School ID wajib diisi'),
    role: z.string().min(1, 'Role wajib dipilih'),
  })
  .refine((d) => d.password === d.confirmPassword, {
    message: 'Konfirmasi password tidak cocok',
    path: ['confirmPassword'],
  });

type RegisterForm = z.infer<typeof registerSchema>;

const ROLES = [
  { value: 'ADMIN_SEKOLAH', label: 'Admin Sekolah' },
  { value: 'KEPALA_SEKOLAH', label: 'Kepala Sekolah' },
  { value: 'KEPALA_KURIKULUM', label: 'Kepala Kurikulum' },
  { value: 'BENDAHARA', label: 'Bendahara' },
  { value: 'GURU', label: 'Guru' },
  { value: 'WALI_KELAS', label: 'Wali Kelas' },
  { value: 'GURU_BK', label: 'Guru BK' },
  { value: 'OPERATOR_SIMS', label: 'Operator SIMS' },
  { value: 'KASIR_KOPERASI', label: 'Kasir Koperasi' },
  { value: 'PUSTAKAWAN', label: 'Pustakawan' },
  { value: 'PETUGAS_UKS', label: 'Petugas UKS' },
  { value: 'TATA_USAHA', label: 'Tata Usaha' },
];

export default function RegisterPage() {
  const router = useRouter();
  const [showPassword, setShowPassword] = useState(false);
  const [serverError, setServerError] = useState('');
  const [success, setSuccess] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<RegisterForm>({
    resolver: zodResolver(registerSchema),
    defaultValues: { name: '', email: '', password: '', confirmPassword: '', schoolId: '', role: '' },
  });

  const onSubmit = async (values: RegisterForm) => {
    setServerError('');
    try {
      await api.post('/auth/register', {
        name: values.name,
        email: values.email,
        password: values.password,
        schoolId: values.schoolId,
        role: values.role,
      });
      setSuccess(true);
      setTimeout(() => router.push('/login'), 2000);
    } catch (err: unknown) {
      const axiosErr = err as { response?: { data?: { error?: { message?: string } } } };
      setServerError(
        axiosErr?.response?.data?.error?.message || 'Gagal mendaftar. Coba lagi.'
      );
    }
  };

  if (success) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-primary-900 via-primary-800 to-indigo-900 flex items-center justify-center p-4">
        <div className="bg-white rounded-2xl shadow-2xl p-8 max-w-md w-full text-center">
          <div className="w-14 h-14 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <svg className="w-7 h-7 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <h2 className="text-lg font-semibold text-gray-900">Registrasi Berhasil!</h2>
          <p className="text-gray-500 mt-2 text-sm">Akun Anda telah dibuat. Mengalihkan ke halaman login...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-primary-900 via-primary-800 to-indigo-900 flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-14 h-14 bg-white rounded-2xl shadow-lg mb-4">
            <School className="w-8 h-8 text-primary-600" />
          </div>
          <h1 className="text-2xl font-bold text-white">Ekosistem Digital Sekolah</h1>
          <p className="text-primary-200 mt-1 text-sm">Buat akun baru</p>
        </div>

        <div className="bg-white rounded-2xl shadow-2xl p-8">
          <h2 className="text-lg font-semibold text-gray-900 mb-6">Daftar Akun</h2>

          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
            <Input
              label="Nama Lengkap"
              placeholder="Budi Santoso"
              required
              {...register('name')}
              error={errors.name?.message}
            />

            <Input
              label="Email"
              type="email"
              placeholder="nama@sekolah.sch.id"
              autoComplete="email"
              required
              {...register('email')}
              error={errors.email?.message}
            />

            <Input
              label="School ID"
              placeholder="ID sekolah dari administrator"
              hint="Dapatkan dari Admin Sekolah Anda"
              required
              {...register('schoolId')}
              error={errors.schoolId?.message}
            />

            <Select
              label="Role / Jabatan"
              placeholder="-- Pilih Jabatan --"
              required
              options={ROLES}
              {...register('role')}
              error={errors.role?.message}
            />

            <Input
              label="Password"
              type={showPassword ? 'text' : 'password'}
              placeholder="Min. 8 karakter"
              autoComplete="new-password"
              required
              rightAddon={
                <button
                  type="button"
                  onClick={() => setShowPassword((v) => !v)}
                  className="text-gray-400 hover:text-gray-600"
                  tabIndex={-1}
                >
                  {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                </button>
              }
              {...register('password')}
              error={errors.password?.message}
            />

            <Input
              label="Konfirmasi Password"
              type="password"
              placeholder="Ulangi password"
              autoComplete="new-password"
              required
              {...register('confirmPassword')}
              error={errors.confirmPassword?.message}
            />

            {serverError && (
              <div className="rounded-lg bg-red-50 border border-red-200 px-4 py-3">
                <p className="text-sm text-red-700">{serverError}</p>
              </div>
            )}

            <Button type="submit" className="w-full mt-2" size="lg" isLoading={isSubmitting}>
              Daftar
            </Button>
          </form>

          <div className="mt-6 text-center">
            <p className="text-sm text-gray-500">
              Sudah punya akun?{' '}
              <Link href="/login" className="text-primary-600 font-medium hover:underline">
                Masuk di sini
              </Link>
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
