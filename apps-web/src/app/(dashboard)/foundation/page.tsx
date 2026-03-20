'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Building2, RefreshCw } from 'lucide-react';
import api from '@/lib/api';
import { Card, CardHeader } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';

interface FoundationSummary {
  totalSchools: number;
  totalStudents: number;
  totalTeachers: number;
  totalRevenue: number;
  activeSchools: number;
}

interface School {
  id: string;
  name: string;
  level: string;
  city: string;
  totalStudents: number;
  totalTeachers: number;
  subscriptionTier: string;
  isActive: boolean;
}

const TIER_VARIANT: Record<string, 'gray' | 'info' | 'success' | 'warning'> = {
  FREE: 'gray',
  BASIC: 'info',
  PRO: 'success',
  ENTERPRISE: 'warning',
};

const LEVEL_LABEL: Record<string, string> = {
  SD: 'SD',
  SMP: 'SMP',
  SMA: 'SMA',
  SMK: 'SMK',
};

export default function FoundationPage() {
  const [summary, setSummary] = useState<FoundationSummary | null>(null);
  const [schools, setSchools] = useState<School[]>([]);
  const [isLoadingSummary, setIsLoadingSummary] = useState(true);
  const [isLoadingSchools, setIsLoadingSchools] = useState(true);

  const fetchSummary = useCallback(async () => {
    setIsLoadingSummary(true);
    try {
      const res = await api.get('/foundation/summary');
      setSummary(res.data.data || null);
    } catch {
      setSummary(null);
    } finally {
      setIsLoadingSummary(false);
    }
  }, []);

  const fetchSchools = useCallback(async () => {
    setIsLoadingSchools(true);
    try {
      const res = await api.get('/foundation/schools');
      setSchools(res.data.data || []);
    } catch {
      setSchools([]);
    } finally {
      setIsLoadingSchools(false);
    }
  }, []);

  useEffect(() => {
    fetchSummary();
    fetchSchools();
  }, [fetchSummary, fetchSchools]);

  const handleRefresh = () => {
    fetchSummary();
    fetchSchools();
  };

  const summaryCards = summary
    ? [
        { label: 'Total Sekolah', value: summary.totalSchools.toLocaleString('id-ID'), color: 'text-blue-700 bg-blue-50' },
        { label: 'Sekolah Aktif', value: summary.activeSchools.toLocaleString('id-ID'), color: 'text-green-700 bg-green-50' },
        { label: 'Total Siswa', value: summary.totalStudents.toLocaleString('id-ID'), color: 'text-purple-700 bg-purple-50' },
        { label: 'Total Guru', value: summary.totalTeachers.toLocaleString('id-ID'), color: 'text-orange-700 bg-orange-50' },
        {
          label: 'Total Pendapatan',
          value: `Rp ${(summary.totalRevenue / 1_000_000).toFixed(1)}jt`,
          color: 'text-teal-700 bg-teal-50',
        },
      ]
    : [];

  const columns = [
    {
      key: 'name',
      header: 'Nama Sekolah',
      render: (s: School) => (
        <div>
          <p className="font-medium text-gray-900">{s.name}</p>
          <p className="text-xs text-gray-400">{s.city}</p>
        </div>
      ),
    },
    {
      key: 'level',
      header: 'Jenjang',
      render: (s: School) => (
        <Badge variant="info" size="sm">{LEVEL_LABEL[s.level] ?? s.level}</Badge>
      ),
    },
    {
      key: 'students',
      header: 'Siswa',
      render: (s: School) => (
        <span className="text-sm font-semibold text-gray-700">
          {s.totalStudents.toLocaleString('id-ID')}
        </span>
      ),
    },
    {
      key: 'teachers',
      header: 'Guru',
      render: (s: School) => (
        <span className="text-sm text-gray-600">{s.totalTeachers.toLocaleString('id-ID')}</span>
      ),
    },
    {
      key: 'tier',
      header: 'Paket',
      render: (s: School) => (
        <Badge variant={TIER_VARIANT[s.subscriptionTier] ?? 'gray'} size="sm">
          {s.subscriptionTier}
        </Badge>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (s: School) => (
        <Badge variant={s.isActive ? 'success' : 'danger'} size="sm">
          {s.isActive ? 'Aktif' : 'Nonaktif'}
        </Badge>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="p-2 bg-indigo-100 rounded-xl">
            <Building2 className="w-5 h-5 text-indigo-600" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-gray-900">Dashboard Yayasan</h1>
            <p className="text-sm text-gray-500 mt-0.5">Ringkasan dan daftar sekolah di bawah yayasan</p>
          </div>
        </div>
        <Button
          variant="secondary"
          size="sm"
          leftIcon={<RefreshCw className="w-3.5 h-3.5" />}
          onClick={handleRefresh}
        >
          Refresh
        </Button>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-2 lg:grid-cols-5 gap-4">
        {isLoadingSummary
          ? Array.from({ length: 5 }).map((_, i) => (
              <Card key={i}>
                <div className="animate-pulse space-y-2">
                  <div className="h-3 bg-gray-200 rounded w-3/4" />
                  <div className="h-8 bg-gray-200 rounded w-1/2" />
                </div>
              </Card>
            ))
          : summaryCards.map((stat) => (
              <Card key={stat.label}>
                <p className="text-sm text-gray-500">{stat.label}</p>
                <p className={`text-2xl font-bold mt-2 px-2 py-0.5 rounded-lg inline-block ${stat.color}`}>
                  {stat.value}
                </p>
              </Card>
            ))}
      </div>

      {/* Schools Table */}
      <Card padding="none">
        <div className="p-4 border-b border-gray-100">
          <CardHeader
            title="Daftar Sekolah"
            description={`${schools.length} sekolah terdaftar dalam yayasan`}
          />
        </div>
        <Table
          columns={columns}
          data={schools}
          keyExtractor={(s) => s.id}
          isLoading={isLoadingSchools}
          emptyMessage="Belum ada sekolah terdaftar dalam yayasan"
        />
      </Card>
    </div>
  );
}
