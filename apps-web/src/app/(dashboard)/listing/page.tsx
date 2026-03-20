'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Store, Plus, RefreshCw, CheckCircle } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import api from '@/lib/api';
import { Card, CardHeader } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { Modal } from '@/components/ui/Modal';
import { Pagination } from '@/components/ui/Pagination';
import { Input, Select } from '@/components/ui/Input';

const CATEGORY_LABELS: Record<string, string> = {
  GURU_LES: 'Guru Les / Bimbel',
  TEMPAT_KURSUS: 'Tempat Kursus',
  ANTAR_JEMPUT: 'Antar Jemput',
  KATERING: 'Katering',
  PENERBIT: 'Penerbit / Toko Buku',
  SERAGAM: 'Seragam & Atribut',
  VENDOR_EVENT: 'Vendor Event',
  ASURANSI: 'Asuransi Pendidikan',
  TABUNGAN: 'Tabungan & Investasi',
  ALAT_TULIS: 'Alat Tulis',
};

const STATUS_VARIANT: Record<string, 'warning' | 'success' | 'danger'> = {
  PENDING_REVIEW: 'warning',
  ACTIVE: 'success',
  SUSPENDED: 'danger',
};

const STATUS_LABEL: Record<string, string> = {
  PENDING_REVIEW: 'Menunggu Review',
  ACTIVE: 'Aktif',
  SUSPENDED: 'Ditangguhkan',
};

interface Vendor {
  id: string;
  businessName: string;
  category: string;
  tier: string;
  status: string;
  contactPhone: string;
  contactEmail: string;
}

interface Placement {
  id: string;
  vendorId: string;
  businessName: string;
  schoolId: string;
  category: string;
  startDate: string;
  endDate: string;
}

interface AddVendorForm {
  businessName: string;
  category: string;
  description: string;
  contactPhone: string;
  contactEmail: string;
}

const EMPTY_FORM: AddVendorForm = {
  businessName: '',
  category: 'GURU_LES',
  description: '',
  contactPhone: '',
  contactEmail: '',
};

export default function ListingPage() {
  const { user } = useAuth();
  const [vendors, setVendors] = useState<Vendor[]>([]);
  const [placements, setPlacements] = useState<Placement[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isLoadingPlacements, setIsLoadingPlacements] = useState(true);
  const [page, setPage] = useState(1);
  const [meta, setMeta] = useState({ total: 0, totalPages: 1 });
  const [isAddOpen, setIsAddOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [approvingId, setApprovingId] = useState<string | null>(null);
  const [form, setForm] = useState<AddVendorForm>(EMPTY_FORM);

  const fetchVendors = useCallback(async () => {
    setIsLoading(true);
    try {
      const params = new URLSearchParams({ page: String(page), limit: '20' });
      const res = await api.get(`/listing/vendors?${params}`);
      setVendors(res.data.data || []);
      setMeta(res.data.meta || { total: 0, totalPages: 1 });
    } catch {
      setVendors([]);
    } finally {
      setIsLoading(false);
    }
  }, [page]);

  const fetchPlacements = useCallback(async () => {
    setIsLoadingPlacements(true);
    try {
      const res = await api.get('/listing/active');
      setPlacements(res.data.data || []);
    } catch {
      setPlacements([]);
    } finally {
      setIsLoadingPlacements(false);
    }
  }, []);

  useEffect(() => {
    fetchVendors();
    fetchPlacements();
  }, [fetchVendors, fetchPlacements]);

  const handleApprove = async (id: string) => {
    setApprovingId(id);
    try {
      await api.patch(`/listing/vendors/${id}/approve`);
      fetchVendors();
    } catch {
      // silent
    } finally {
      setApprovingId(null);
    }
  };

  const handleAdd = async () => {
    setIsSubmitting(true);
    try {
      await api.post('/listing/vendors', form);
      setIsAddOpen(false);
      setForm(EMPTY_FORM);
      fetchVendors();
    } catch {
      // silent
    } finally {
      setIsSubmitting(false);
    }
  };

  const vendorColumns = [
    {
      key: 'businessName',
      header: 'Nama Usaha',
      render: (v: Vendor) => (
        <div>
          <p className="font-medium text-gray-900">{v.businessName}</p>
          <p className="text-xs text-gray-400">{v.contactEmail}</p>
        </div>
      ),
    },
    {
      key: 'category',
      header: 'Kategori',
      render: (v: Vendor) => (
        <Badge variant="info" size="sm">{CATEGORY_LABELS[v.category] ?? v.category}</Badge>
      ),
    },
    {
      key: 'tier',
      header: 'Tier',
      render: (v: Vendor) => (
        <Badge variant="default" size="sm">{v.tier || '—'}</Badge>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (v: Vendor) => (
        <Badge variant={STATUS_VARIANT[v.status] ?? 'default'} size="sm">
          {STATUS_LABEL[v.status] ?? v.status}
        </Badge>
      ),
    },
    {
      key: 'actions',
      header: '',
      render: (v: Vendor) =>
        v.status === 'PENDING_REVIEW' ? (
          <Button
            variant="ghost"
            size="sm"
            leftIcon={<CheckCircle className="w-3.5 h-3.5" />}
            isLoading={approvingId === v.id}
            onClick={(e) => { e.stopPropagation(); handleApprove(v.id); }}
          >
            Setujui
          </Button>
        ) : null,
      className: 'w-28',
    },
  ];

  const placementColumns = [
    {
      key: 'businessName',
      header: 'Nama Usaha',
      render: (p: Placement) => <span className="font-medium text-gray-900">{p.businessName}</span>,
    },
    {
      key: 'category',
      header: 'Kategori',
      render: (p: Placement) => (
        <Badge variant="info" size="sm">{CATEGORY_LABELS[p.category] ?? p.category}</Badge>
      ),
    },
    {
      key: 'schoolId',
      header: 'Sekolah',
      render: (p: Placement) => <span className="text-sm text-gray-600">{p.schoolId}</span>,
    },
    {
      key: 'period',
      header: 'Periode',
      render: (p: Placement) => (
        <span className="text-sm text-gray-500">{p.startDate} — {p.endDate}</span>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="p-2 bg-amber-100 rounded-xl">
            <Store className="w-5 h-5 text-amber-600" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-gray-900">Listing Marketplace</h1>
            <p className="text-sm text-gray-500 mt-0.5">Kelola vendor dan penempatan iklan listing</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="secondary" size="sm" leftIcon={<RefreshCw className="w-3.5 h-3.5" />} onClick={() => { fetchVendors(); fetchPlacements(); }}>
            Refresh
          </Button>
          <Button size="sm" leftIcon={<Plus className="w-3.5 h-3.5" />} onClick={() => setIsAddOpen(true)}>
            Tambah Vendor
          </Button>
        </div>
      </div>

      {/* Vendors Table */}
      <Card padding="none">
        <div className="p-4 border-b border-gray-100">
          <CardHeader title="Daftar Vendor" description={`${meta.total} vendor terdaftar`} />
        </div>
        <Table
          columns={vendorColumns}
          data={vendors}
          keyExtractor={(v) => v.id}
          isLoading={isLoading}
          emptyMessage="Belum ada vendor terdaftar"
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

      {/* Active Placements */}
      <Card padding="none">
        <div className="p-4 border-b border-gray-100">
          <CardHeader title="Penempatan Aktif" description="Listing yang sedang tayang di halaman sekolah" />
        </div>
        <Table
          columns={placementColumns}
          data={placements}
          keyExtractor={(p) => p.id}
          isLoading={isLoadingPlacements}
          emptyMessage="Tidak ada penempatan aktif"
        />
      </Card>

      {/* Add Vendor Modal */}
      <Modal
        isOpen={isAddOpen}
        onClose={() => { setIsAddOpen(false); setForm(EMPTY_FORM); }}
        title="Tambah Vendor Baru"
        description="Daftarkan penyedia layanan baru ke marketplace EDS"
        size="lg"
        footer={
          <>
            <Button variant="secondary" onClick={() => { setIsAddOpen(false); setForm(EMPTY_FORM); }}>Batal</Button>
            <Button onClick={handleAdd} isLoading={isSubmitting}>Simpan</Button>
          </>
        }
      >
        <div className="space-y-4">
          <Input
            label="Nama Usaha"
            required
            placeholder="Contoh: Bimbel Cerdas Mandiri"
            value={form.businessName}
            onChange={(e) => setForm({ ...form, businessName: e.target.value })}
          />
          <Select
            label="Kategori"
            options={Object.entries(CATEGORY_LABELS).map(([value, label]) => ({ value, label }))}
            value={form.category}
            onChange={(e) => setForm({ ...form, category: e.target.value })}
          />
          <Input
            label="Deskripsi"
            placeholder="Deskripsi singkat layanan"
            value={form.description}
            onChange={(e) => setForm({ ...form, description: e.target.value })}
          />
          <div className="grid grid-cols-2 gap-4">
            <Input
              label="No. Telepon"
              placeholder="08xxxxxxxxxx"
              value={form.contactPhone}
              onChange={(e) => setForm({ ...form, contactPhone: e.target.value })}
            />
            <Input
              label="Email Kontak"
              type="email"
              placeholder="vendor@email.com"
              value={form.contactEmail}
              onChange={(e) => setForm({ ...form, contactEmail: e.target.value })}
            />
          </div>
        </div>
      </Modal>
    </div>
  );
}
