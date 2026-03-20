'use client';

import React, { useState, useEffect } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import Link from 'next/link';
import { Eye, EyeOff, School } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';

const loginSchema = z.object({
  subdomain: z.string().optional(),
  email: z.string().email('Format email tidak valid'),
  password: z.string().min(6, 'Password minimal 6 karakter'),
});

type LoginForm = z.infer<typeof loginSchema>;

export default function LoginPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { login, isAuthenticated } = useAuth();
  const [showPassword, setShowPassword] = useState(false);
  const [serverError, setServerError] = useState('');

  const redirect = searchParams.get('redirect') || '/dashboard';

  useEffect(() => {
    if (isAuthenticated) router.replace(redirect);
  }, [isAuthenticated, redirect, router]);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginForm>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      subdomain: '',
      email: '',
      password: '',
    },
  });

  const onSubmit = async (values: LoginForm) => {
    setServerError('');
    try {
      await login({
        email: values.email,
        password: values.password,
        subdomain: values.subdomain || undefined,
      });

      // Set cookie for middleware
      document.cookie = `eds_access_token=${localStorage.getItem('eds_access_token')}; path=/; SameSite=Lax`;

      router.push(redirect);
    } catch (err: unknown) {
      const axiosErr = err as { response?: { data?: { error?: { message?: string } } } };
      setServerError(
        axiosErr?.response?.data?.error?.message || 'Email atau password salah'
      );
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-primary-900 via-primary-800 to-indigo-900 flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        {/* Logo */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-14 h-14 bg-white rounded-2xl shadow-lg mb-4">
            <School className="w-8 h-8 text-primary-600" />
          </div>
          <h1 className="text-2xl font-bold text-white">Ekosistem Digital Sekolah</h1>
          <p className="text-primary-200 mt-1 text-sm">Masuk ke platform sekolah Anda</p>
        </div>

        {/* Card */}
        <div className="bg-white rounded-2xl shadow-2xl p-8">
          <h2 className="text-lg font-semibold text-gray-900 mb-6">Masuk</h2>

          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
            {/* Subdomain */}
            <Input
              label="Subdomain Sekolah"
              placeholder="smkn1kota"
              hint="Opsional — isi jika Anda tahu subdomain sekolah"
              rightAddon={
                <span className="text-gray-400 text-xs">.eds.id</span>
              }
              {...register('subdomain')}
              error={errors.subdomain?.message}
            />

            {/* Email */}
            <Input
              label="Email"
              type="email"
              placeholder="nama@sekolah.sch.id"
              autoComplete="email"
              required
              {...register('email')}
              error={errors.email?.message}
            />

            {/* Password */}
            <div className="relative">
              <Input
                label="Password"
                type={showPassword ? 'text' : 'password'}
                placeholder="••••••••"
                autoComplete="current-password"
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
            </div>

            {/* Server error */}
            {serverError && (
              <div className="rounded-lg bg-red-50 border border-red-200 px-4 py-3">
                <p className="text-sm text-red-700">{serverError}</p>
              </div>
            )}

            <Button
              type="submit"
              className="w-full mt-2"
              size="lg"
              isLoading={isSubmitting}
            >
              Masuk
            </Button>
          </form>

          <div className="mt-6 text-center">
            <p className="text-sm text-gray-500">
              Belum punya akun?{' '}
              <Link href="/register" className="text-primary-600 font-medium hover:underline">
                Daftar sekarang
              </Link>
            </p>
          </div>
        </div>

        <p className="text-center text-xs text-primary-300 mt-6">
          &copy; {new Date().getFullYear()} EDS Platform. Seluruh hak dilindungi.
        </p>
      </div>
    </div>
  );
}
