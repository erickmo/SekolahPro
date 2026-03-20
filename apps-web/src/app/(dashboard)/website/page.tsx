'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Globe } from 'lucide-react';
import api from '@/lib/api';
import { Card, CardHeader } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { Modal } from '@/components/ui/Modal';
import { Input } from '@/components/ui/Input';

interface WebPage {
  id: string;
  slug: string;
  title: string;
  content: string;
  isPublished: boolean;
  updatedAt: string;
}

const DEFAULT_FORM = { slug: '', title: '', content: '', isPublished: false };

export default function WebsitePage() {
  const [pages, setPages] = useState<WebPage[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState(DEFAULT_FORM);

  const fetchPages = useCallback(async () => {
    setIsLoading(true);
    try {
      const res = await api.get('/website/pages');
      setPages(res.data.data || []);
    } catch {
      setPages([]);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchPages();
  }, [fetchPages]);

  const openAddModal = () => {
    setEditingId(null);
    setForm(DEFAULT_FORM);
    setIsModalOpen(true);
  };

  const openEditModal = (p: WebPage) => {
    setEditingId(p.id);
    setForm({ slug: p.slug, title: p.title, content: p.content, isPublished: p.isPublished });
    setIsModalOpen(true);
  };

  const handleSubmit = async () => {
    if (!form.title.trim() || !form.content.trim()) return;
    setIsSubmitting(true);
    try {
      if (editingId) {
        await api.patch(`/website/pages/${editingId}`, {
          title: form.title,
          content: form.content,
          isPublished: form.isPublished,
        });
      } else {
        await api.post('/website/pages', form);
      }
      setIsModalOpen(false);
      setForm(DEFAULT_FORM);
      fetchPages();
    } catch {
      // handle error silently
    } finally {
      setIsSubmitting(false);
    }
  };

  const publishedCount = pages.filter((p) => p.isPublished).length;
  const draftCount = pages.filter((p) => !p.isPublished).length;

  const columns = [
    {
      key: 'title',
      header: 'Judul Halaman',
      render: (p: WebPage) => (
        <div>
          <p className="font-medium text-gray-900">{p.title}</p>
          <p className="text-xs text-gray-400">/{p.slug}</p>
        </div>
      ),
    },
    {
      key: 'isPublished',
      header: 'Status',
      render: (p: WebPage) => (
        <Badge variant={p.isPublished ? 'success' : 'default'}>
          {p.isPublished ? 'Diterbitkan' : 'Draf'}
        </Badge>
      ),
    },
    {
      key: 'content',
      header: 'Pratinjau Konten',
      render: (p: WebPage) => (
        <p className="text-sm text-gray-500 max-w-xs line-clamp-1">{p.content}</p>
      ),
    },
    {
      key: 'updatedAt',
      header: 'Terakhir Diperbarui',
      render: (p: WebPage) => (
        <span className="text-xs text-gray-400">{new Date(p.updatedAt).toLocaleDateString('id-ID')}</span>
      ),
    },
    {
      key: 'actions',
      header: '',
      render: (p: WebPage) => (
        <Button variant="ghost" size="sm" onClick={() => openEditModal(p)}>
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
            <Globe className="w-5 h-5 text-primary-600" />
            Website Sekolah
          </h1>
          <p className="text-sm text-gray-500 mt-0.5">
            Kelola halaman konten website publik sekolah
          </p>
        </div>
        <Button onClick={openAddModal}>Tambah Halaman</Button>
      </div>

      <div className="grid grid-cols-3 gap-4">
        <Card>
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-full bg-primary-100 flex items-center justify-center">
              <Globe className="w-4 h-4 text-primary-600" />
            </div>
            <div>
              <p className="text-xs text-gray-500">Total Halaman</p>
              <p className="text-xl font-bold text-gray-900">{pages.length}</p>
            </div>
          </div>
        </Card>
        <Card>
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-full bg-green-100 flex items-center justify-center">
              <Globe className="w-4 h-4 text-green-600" />
            </div>
            <div>
              <p className="text-xs text-gray-500">Diterbitkan</p>
              <p className="text-xl font-bold text-gray-900">{publishedCount}</p>
            </div>
          </div>
        </Card>
        <Card>
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-full bg-gray-100 flex items-center justify-center">
              <Globe className="w-4 h-4 text-gray-500" />
            </div>
            <div>
              <p className="text-xs text-gray-500">Draf</p>
              <p className="text-xl font-bold text-gray-900">{draftCount}</p>
            </div>
          </div>
        </Card>
      </div>

      <Card padding="none">
        <div className="p-4 border-b border-gray-100">
          <CardHeader
            title="Daftar Halaman"
            description="Semua halaman konten website sekolah"
          />
        </div>
        <Table
          columns={columns}
          data={pages}
          keyExtractor={(p) => p.id}
          isLoading={isLoading}
          emptyMessage="Belum ada halaman website yang dibuat"
        />
      </Card>

      <Modal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        title={editingId ? 'Edit Halaman' : 'Tambah Halaman Baru'}
      >
        <div className="space-y-4">
          {!editingId && (
            <Input
              label="Slug URL"
              placeholder="contoh: tentang-kami"
              value={form.slug}
              onChange={(e) => setForm({ ...form, slug: e.target.value.toLowerCase().replace(/\s+/g, '-') })}
              helperText="Akan digunakan sebagai URL halaman: /slug"
            />
          )}
          <Input
            label="Judul Halaman"
            placeholder="Judul halaman"
            value={form.title}
            onChange={(e) => setForm({ ...form, title: e.target.value })}
          />
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Konten</label>
            <textarea
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 min-h-[160px]"
              placeholder="Tulis konten halaman di sini..."
              value={form.content}
              onChange={(e) => setForm({ ...form, content: e.target.value })}
            />
          </div>
          {editingId && (
            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                className="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                checked={form.isPublished}
                onChange={(e) => setForm({ ...form, isPublished: e.target.checked })}
              />
              <span className="text-sm text-gray-700">Terbitkan halaman ini</span>
            </label>
          )}
          <div className="flex justify-end gap-2 pt-2">
            <Button variant="secondary" onClick={() => setIsModalOpen(false)}>Batal</Button>
            <Button isLoading={isSubmitting} onClick={handleSubmit}>
              {editingId ? 'Simpan Perubahan' : 'Tambah Halaman'}
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
