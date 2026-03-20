'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { useAuth } from '@/context/AuthContext';
import { Card, CardHeader, StatCard } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import api from '@/lib/api';
import { formatCurrency, formatDate } from '@/lib/utils';
import type { DashboardStats, AttendanceTrend } from '@/types';
import {
  Users,
  UserCheck,
  UserX,
  CreditCard,
  BookOpen,
  TrendingUp,
  AlertTriangle,
  GraduationCap,
  Sparkles,
} from 'lucide-react';
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts';

export default function DashboardPage() {
  const { user } = useAuth();
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [trend, setTrend] = useState<AttendanceTrend[]>([]);
  const [insight, setInsight] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  const canViewInsight =
    user?.role === 'KEPALA_SEKOLAH' || user?.role === 'ADMIN_SEKOLAH';

  const fetchDashboard = useCallback(async () => {
    setIsLoading(true);
    try {
      const [dashRes, trendRes] = await Promise.all([
        api.get('/dashboard'),
        api.get('/dashboard/attendance-trend?days=30'),
      ]);
      setStats(dashRes.data.data);
      setTrend(trendRes.data.data || []);

      if (canViewInsight) {
        try {
          const insightRes = await api.get('/dashboard/insight');
          setInsight(insightRes.data.data?.insight || null);
        } catch {
          // insight is optional
        }
      }
    } catch {
      // handle silently — will show empty state
    } finally {
      setIsLoading(false);
    }
  }, [canViewInsight]);

  useEffect(() => {
    fetchDashboard();
  }, [fetchDashboard]);

  const chartData = trend.map((t) => ({
    date: formatDate(t.date, 'dd MMM'),
    Hadir: t.present,
    Absen: t.absent,
    Terlambat: t.late,
  }));

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="bg-white rounded-xl p-6 animate-pulse border border-gray-100">
              <div className="h-4 bg-gray-200 rounded w-2/3 mb-3" />
              <div className="h-7 bg-gray-200 rounded w-1/2 mb-2" />
              <div className="h-3 bg-gray-100 rounded w-1/3" />
            </div>
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Page header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900">Dashboard</h1>
          <p className="text-sm text-gray-500 mt-0.5">
            Selamat datang, <span className="font-medium">{user?.name}</span> —{' '}
            {formatDate(new Date().toISOString())}
          </p>
        </div>
      </div>

      {/* Stat cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title="Total Siswa"
          value={stats?.totalStudents?.toLocaleString('id-ID') || '0'}
          icon={<Users className="w-5 h-5 text-primary-600" />}
          iconBg="bg-primary-50"
        />
        <StatCard
          title="Hadir Hari Ini"
          value={stats?.presentToday?.toLocaleString('id-ID') || '0'}
          subtitle={`${stats?.attendanceRate?.toFixed(1) || '0'}% kehadiran`}
          icon={<UserCheck className="w-5 h-5 text-green-600" />}
          iconBg="bg-green-50"
        />
        <StatCard
          title="Tagihan Pending"
          value={formatCurrency(stats?.pendingPayments || 0)}
          icon={<CreditCard className="w-5 h-5 text-amber-600" />}
          iconBg="bg-amber-50"
        />
        <StatCard
          title="Siswa Berisiko"
          value={stats?.riskStudents || 0}
          subtitle="Early warning aktif"
          icon={<AlertTriangle className="w-5 h-5 text-red-600" />}
          iconBg="bg-red-50"
        />
      </div>

      {/* Row 2 */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <StatCard
          title="Guru & Staff"
          value={stats?.totalTeachers || 0}
          icon={<GraduationCap className="w-5 h-5 text-indigo-600" />}
          iconBg="bg-indigo-50"
        />
        <StatCard
          title="Kelas Aktif"
          value={stats?.activeClasses || 0}
          icon={<BookOpen className="w-5 h-5 text-purple-600" />}
          iconBg="bg-purple-50"
        />
        <StatCard
          title="Total Pendapatan"
          value={formatCurrency(stats?.totalRevenue || 0)}
          icon={<TrendingUp className="w-5 h-5 text-teal-600" />}
          iconBg="bg-teal-50"
        />
      </div>

      {/* Charts + AI insight row */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Attendance chart */}
        <Card className="lg:col-span-2">
          <CardHeader
            title="Tren Kehadiran (30 Hari)"
            description="Grafik kehadiran siswa per hari"
          />
          {chartData.length > 0 ? (
            <ResponsiveContainer width="100%" height={260}>
              <AreaChart data={chartData} margin={{ top: 5, right: 10, left: -20, bottom: 0 }}>
                <defs>
                  <linearGradient id="hadir" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#4f46e5" stopOpacity={0.2} />
                    <stop offset="95%" stopColor="#4f46e5" stopOpacity={0} />
                  </linearGradient>
                  <linearGradient id="absen" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#ef4444" stopOpacity={0.2} />
                    <stop offset="95%" stopColor="#ef4444" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
                <XAxis dataKey="date" tick={{ fontSize: 11 }} tickLine={false} />
                <YAxis tick={{ fontSize: 11 }} tickLine={false} />
                <Tooltip
                  contentStyle={{ borderRadius: '8px', fontSize: '12px', border: '1px solid #e5e7eb' }}
                />
                <Area type="monotone" dataKey="Hadir" stroke="#4f46e5" strokeWidth={2} fill="url(#hadir)" />
                <Area type="monotone" dataKey="Absen" stroke="#ef4444" strokeWidth={2} fill="url(#absen)" />
              </AreaChart>
            </ResponsiveContainer>
          ) : (
            <div className="h-[260px] flex items-center justify-center text-sm text-gray-400">
              Belum ada data kehadiran
            </div>
          )}
        </Card>

        {/* AI Insight */}
        <Card>
          <CardHeader
            title="AI Insight"
            description="Rangkuman kondisi sekolah"
            action={
              <div className="p-1.5 bg-primary-50 rounded-lg">
                <Sparkles className="w-4 h-4 text-primary-600" />
              </div>
            }
          />
          {!canViewInsight ? (
            <div className="flex flex-col items-center justify-center py-8 text-center">
              <AlertTriangle className="w-8 h-8 text-gray-300 mb-2" />
              <p className="text-sm text-gray-400">
                Fitur ini hanya untuk Kepala Sekolah dan Admin
              </p>
            </div>
          ) : insight ? (
            <div className="prose prose-sm max-w-none">
              <p className="text-sm text-gray-600 leading-relaxed whitespace-pre-wrap">{insight}</p>
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center py-8 text-center">
              <Sparkles className="w-8 h-8 text-gray-200 mb-2" />
              <p className="text-sm text-gray-400">Insight belum tersedia</p>
              <p className="text-xs text-gray-300 mt-1">Update setiap minggu</p>
            </div>
          )}

          {/* Quick metrics */}
          <div className="mt-4 pt-4 border-t border-gray-100 space-y-2">
            <div className="flex items-center justify-between">
              <span className="text-xs text-gray-500">Kehadiran hari ini</span>
              <Badge variant={
                (stats?.attendanceRate || 0) >= 90 ? 'success' :
                (stats?.attendanceRate || 0) >= 75 ? 'warning' : 'danger'
              }>
                {stats?.attendanceRate?.toFixed(1) || '0'}%
              </Badge>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-xs text-gray-500">Siswa tidak hadir</span>
              <Badge variant="danger">{stats?.absentToday || 0} siswa</Badge>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-xs text-gray-500">Peringatan dini</span>
              <Badge variant={stats?.riskStudents ? 'warning' : 'success'}>
                {stats?.riskStudents || 0} siswa
              </Badge>
            </div>
          </div>
        </Card>
      </div>
    </div>
  );
}
