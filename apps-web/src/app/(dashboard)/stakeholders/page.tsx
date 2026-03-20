'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Users } from 'lucide-react';
import api from '@/lib/api';
import { Card, CardHeader } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { Modal } from '@/components/ui/Modal';
import { Input } from '@/components/ui/Input';

interface Stakeholder {
  id: string;
  name: string;
  type: 'PUBLISHER' | 'SPONSOR' | 'PARTNER';
  contactEmail: string;
  phone: string;
  createdAt: string;
}

const TYPE_MAP: Record<string, { label: string; variant: 'primary' | 'warning' | 'success' }> = {
  PUBLISHER: { label: 'Penerbit', variant: 'primary' },
  SPONSOR: { label: 'Sponsor', variant: 'warning' },
  PARTNER: { label: 'Mitra', variant: 'success' },
};

const DEFAULT_FORM = { name: '', type: 'PARTNER', contactEmail: '', phone: '' };

export default function StakeholdersPage() {
  const [stakeholders, setStakeholders] = useState<Stakeholder[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState(DEFAULT_FORM);

  const fetchStakeholders = useCallback(async () => {
    setIsLoading(true);
    try {
      const res = await api.get('/stakeholders');
      setStakeholders(res.data.data || []);
    } catch {
      setStakeholders([]);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchStakeholders();
  }, [fetchStakeholders]);

  const openAddModal = () => {
    setEditingId(null);
    setForm(DEFAULT_FORM);
    setIsModalOpen(true);
  };

  const openEditModal = (s: Stakeholder) => {
    setEditingId(s.id);
    setForm({ name: s.name, type: s.type, contactEmail: s.contactEmail, phone: s.phone });
    setIsModalOpen(true);
  };

  const handleSubmit = async () => {
    if (!form.name.trim() || !form.contactEmail.trim()) return;
    setIsSubmitting(true);
    try {
      if (editingId) {
        await api.patch(`/stakeholders/${editingId}`, form);
      } else {
        await api.post('/stakeholders', form);
      }
      setIsModalOpen(false);
      setForm(DEFAULT_FORM);
      fetchStakeholders();
    } catch {
      // handle error silently
    } finally {
      setIsSubmitting(false);
    }
  };

  const columns = [
    {
      key: 'name',
      header: 'Nama',
      render: (s: Stakeholder) => (
        <p className="font-medium text-gray-900">{s.name}</p>
      ),
    },
    {
      key: 'type',
      header: 'Tipe',
      render: (s: Stakeholder) => {
        const t = TYPE_MAP[s.type] || { label: s.type, variant: 'default' as const };
        return <Badge variant={t.variant}>{t.label}</Badge>;
      },
    },
    {
      key: 'contactEmail',
      header: 'Email Kontak',
      render: (s: Stakeholder) => (
        <a href={`mailto:${s.contactEmail}`} className="text-sm text-primary-600 hover:underline">
          {s.contactEmail}
        </a>
      ),
    },
    {
      key: 'phone',
      header: 'Telepon',
      render: (s: Stakeholder) => (
        <span className="text-sm text-gray-600">{s.phone || '—'}</span>
      ),
    },
    {
      key: 'createdAt',
      header: 'Terdaftar',
      render: (s: Stakeholder) => (
        <span className="text-xs text-gray-400">{new Date(s.createdAt).toLocaleDateString('id-ID')}</span>
      ),
    },
    {
      key: 'actions',
      header: '',
      render: (s: Stakeholder) => (
        <Button variant="ghost" size="sm" onClick={() => openEditModal(s)}>
          Edit
        </Button>
      ),
      className: 'w-20',
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900 flex items-center gap-2">
            <Users className="w-5 h-5 text-primary-600" />
            Stakeholder & Penerbit
          </h1>
          <p className="text-sm text-gray-500 mt-0.5">
            Kelola penerbit, sponsor, dan mitra sekolah
          </p>
        </div>
        <Button onClick={openAddModal}>Tambah Stakeholder</Button>
      </div>

      <Card padding="none">
        <div className="p-4 border-b border-gray-100">
          <CardHeader
            title="Daftar Stakeholder"
            description="Semua penerbit, sponsor, dan mitra yang terdaftar"
          />
        </div>
        <Table
          columns={columns}
          data={stakeholders}
          keyExtractor={(s) => s.id}
          isLoading={isLoading}
          emptyMessage="Belum ada stakeholder yang terdaftar"
        />
      </Card>

      <Modal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        title={editingId ? 'Edit Stakeholder' : 'Tambah Stakeholder'}
      >
        <div className="space-y-4">
          <Input
            label="Nama"
            placeholder="Nama perusahaan atau individu"
            value={form.name}
            onChange={(e) => setForm({ ...form, name: e.target.value })}
          />
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Tipe</label>
            <select
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
              value={form.type}
              onChange={(e) => setForm({ ...form, type: e.target.value })}
            >
              <option value="PUBLISHER">Penerbit</option>
              <option value="SPONSOR">Sponsor</option>
              <option value="PARTNER">Mitra</option>
            </select>
          </div>
          <Input
            label="Email Kontak"
            type="email"
            placeholder="email@contoh.com"
            value={form.contactEmail}
            onChange={(e) => setForm({ ...form, contactEmail: e.target.value })}
          />
          <Input
            label="Nomor Telepon"
            placeholder="08xxxxxxxxxx"
            value={form.phone}
            onChange={(e) => setForm({ ...form, phone: e.target.value })}
          />
          <div className="flex justify-end gap-2 pt-2">
            <Button variant="secondary" onClick={() => setIsModalOpen(false)}>Batal</Button>
            <Button isLoading={isSubmitting} onClick={handleSubmit}>
              {editingId ? 'Simpan Perubahan' : 'Tambah'}
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
