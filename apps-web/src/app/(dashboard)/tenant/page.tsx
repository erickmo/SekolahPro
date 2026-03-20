'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Building2, Plus, RefreshCw } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import api from '@/lib/api';
import { Card, CardHeader, StatCard } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { Modal } from '@/components/ui/Modal';
import { Pagination } from '@/components/ui/Pagination';
import { Input, Select } from '@/components/ui/Input';
import { formatDate } from '@/lib/utils';

const ALLOWED_ROLES = ['EDS_SUPERADMIN', 'EDS_SUPPORT', 'EDS_SALES'];

const TIER_VARIANT: Record<string, 'gray' | 'info' | 'primary' | 'warning'> = {
  FREE: 'gray',
  BASIC: 'info',
  PRO: 'primary',
  ENTERPRISE: 'warning',
};

const STATUS_VARIANT: Record<string, 'warning' | 'success' | 'danger'> = {
  TRIAL: 'warning',
  ACTIVE: 'success',
  SUSPENDED: 'danger',
};

const STATUS_LABEL: Record<string, string> = {
  TRIAL: 'Trial',
  ACTIVE: 'Aktif',
  SUSPENDED: 'Ditangguhkan',
};

interface Tenant {
  id: string;
  schoolId: string;
  subdomain: string;
  billingEmail: string;
  tier: string;
  status: string;
  createdAt: string;
}

interface TenantStats {
  total: number;
  active: number;
  trial: number;
  suspended: number;
}

interface AddTenantForm {
  schoolId: string;
  subdomain: string;
  billingEmail: string;
  tier: string;
}

export default function TenantPage() {
  const { user } = useAuth();
  const [tenants, setTenants] = useState<Tenant[]>([]);
  const [stats, setStats] = useState<TenantStats>({ total: 0, active: 0, trial: 0, suspended: 0 });
  const [isLoading, setIsLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [meta, setMeta] = useState({ total: 0, totalPages: 1 });
  const [isAddOpen, setIsAddOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [form, setForm] = useState<AddTenantForm>({ schoolId: '', subdomain: '', billingEmail: '', tier: 'FREE' });

  const canAccess = ALLOWED_ROLES.includes(user?.role || '');

  const fetchTenants = useCallback(async () => {
    setIsLoading(true);
    try {
      const params = new URLSearchParams({ page: String(page), limit: '20' });
      const res = await api.get(`/tenant?${params}`);
      const data = res.data.data || [];
      setTenants(data);
      setMeta(res.data.meta || { total: 0, totalPages: 1 });
      const total = data.length;
      const active = data.filter((t: Tenant) => t.status === 'ACTIVE').length;
      const trial = data.filter((t: Tenant) => t.status === 'TRIAL').length;
      const suspended = data.filter((t: Tenant) => t.status === 'SUSPENDED').length;
      setStats({ total: res.data.meta?.total ?? total, active, trial, suspended });
    } catch {
      setTenants([]);
    } finally {
      setIsLoading(false);
    }
  }, [page]);

  useEffect(() => {
    fetchTenants();
  }, [fetchTenants]);

  const handleAdd = async () => {
    setIsSubmitting(true);
    try {
      await api.post('/tenant', form);
      setIsAddOpen(false);
      setForm({ schoolId: '', subdomain: '', billingEmail: '', tier: 'FREE' });
      fetchTenants();
    } catch {
      // silent — produksi: tampilkan toast error
    } finally {
      setIsSubmitting(false);
    }
  };

  const columns = [
    {
      key: 'subdomain',
      header: 'Subdomain',
      render: (t: Tenant) => (
        <div>
          <p className="font-medium text-gray-900">{t.subdomain}.eds.id</p>
          <p className="text-xs text-gray-400">ID: {t.schoolId}</p>
        </div>
      ),
    },
    {
      key: 'tier',
      header: 'Tier',
      render: (t: Tenant) => (
        <Badge variant={TIER_VARIANT[t.tier] ?? 'default'} size="sm">
          {t.tier}
        </Badge>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (t: Tenant) => (
        <Badge variant={STATUS_VARIANT[t.status] ?? 'default'} size="sm">
          {STATUS_LABEL[t.status] ?? t.status}
        </Badge>
      ),
    },
    {
      key: 'billingEmail',
      header: 'Email Tagihan',
      render: (t: Tenant) => <span className="text-sm text-gray-600">{t.billingEmail}</span>,
    },
    {
      key: 'createdAt',
      header: 'Tanggal Daftar',
      render: (t: Tenant) => <span className="text-sm text-gray-500">{formatDate(t.createdAt)}</span>,
    },
  ];

  if (!canAccess) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-gray-500">Anda tidak memiliki akses ke halaman ini.</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="p-2 bg-primary-100 rounded-xl">
            <Building2 className="w-5 h-5 text-primary-600" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-gray-900">Manajemen Tenant</h1>
            <p className="text-sm text-gray-500 mt-0.5">Kelola semua tenant sekolah di platform EDS</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="secondary" size="sm" leftIcon={<RefreshCw className="w-3.5 h-3.5" />} onClick={fetchTenants}>
            Refresh
          </Button>
          <Button size="sm" leftIcon={<Plus className="w-3.5 h-3.5" />} onClick={() => setIsAddOpen(true)}>
            Tambah Tenant
          </Button>
        </div>
      </div>

      {/* Summary cards */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title="Total Tenant"
          value={stats.total}
          icon={<Building2 className="w-5 h-5 text-primary-600" />}
          iconBg="bg-primary-100"
        />
        <StatCard
          title="Aktif"
          value={stats.active}
          icon={<Building2 className="w-5 h-5 text-green-600" />}
          iconBg="bg-green-100"
        />
        <StatCard
          title="Trial"
          value={stats.trial}
          icon={<Building2 className="w-5 h-5 text-amber-600" />}
          iconBg="bg-amber-100"
        />
        <StatCard
          title="Ditangguhkan"
          value={stats.suspended}
          icon={<Building2 className="w-5 h-5 text-red-600" />}
          iconBg="bg-red-100"
        />
      </div>

      {/* Table */}
      <Card padding="none">
        <div className="p-4 border-b border-gray-100">
          <CardHeader title="Daftar Tenant" description="Semua sekolah yang terdaftar di platform EDS" />
        </div>
        <Table
          columns={columns}
          data={tenants}
          keyExtractor={(t) => t.id}
          isLoading={isLoading}
          emptyMessage="Belum ada tenant terdaftar"
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

      {/* Add Modal */}
      <Modal
        isOpen={isAddOpen}
        onClose={() => setIsAddOpen(false)}
        title="Tambah Tenant Baru"
        description="Daftarkan sekolah baru ke platform EDS"
        footer={
          <>
            <Button variant="secondary" onClick={() => setIsAddOpen(false)}>Batal</Button>
            <Button onClick={handleAdd} isLoading={isSubmitting}>Simpan</Button>
          </>
        }
      >
        <div className="space-y-4">
          <Input
            label="School ID"
            required
            placeholder="Contoh: SMAN1JKT"
            value={form.schoolId}
            onChange={(e) => setForm({ ...form, schoolId: e.target.value })}
          />
          <Input
            label="Subdomain"
            required
            placeholder="sman1jkt (tanpa .eds.id)"
            value={form.subdomain}
            onChange={(e) => setForm({ ...form, subdomain: e.target.value })}
          />
          <Input
            label="Email Tagihan"
            type="email"
            required
            placeholder="admin@sekolah.sch.id"
            value={form.billingEmail}
            onChange={(e) => setForm({ ...form, billingEmail: e.target.value })}
          />
          <Select
            label="Tier"
            options={[
              { value: 'FREE', label: 'Free' },
              { value: 'BASIC', label: 'Basic — Rp 299.000/bln' },
              { value: 'PRO', label: 'Pro — Rp 799.000/bln' },
              { value: 'ENTERPRISE', label: 'Enterprise — Custom' },
            ]}
            value={form.tier}
            onChange={(e) => setForm({ ...form, tier: e.target.value })}
          />
        </div>
      </Modal>
    </div>
  );
}
