'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Plus, MessageSquare, Search } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import api from '@/lib/api';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { Pagination } from '@/components/ui/Pagination';
import { Modal } from '@/components/ui/Modal';
import { formatDateTime } from '@/lib/utils';

interface ForumPost {
  id: string;
  title: string;
  content: string;
  category: string;
  authorName: string;
  repliesCount: number;
  createdAt: string;
}

export default function ForumPage() {
  const { user } = useAuth();
  const [posts, setPosts] = useState<ForumPost[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [page, setPage] = useState(1);
  const [meta, setMeta] = useState({ total: 0, totalPages: 1 });
  const [isAddOpen, setIsAddOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const [form, setForm] = useState({
    title: '',
    content: '',
    category: '',
  });

  const fetchData = useCallback(async () => {
    setIsLoading(true);
    try {
      const params = new URLSearchParams({
        page: String(page),
        limit: '20',
        ...(search && { search }),
      });
      const res = await api.get(`/forum/posts?${params}`);
      setPosts(res.data.data || []);
      setMeta(res.data.meta || { total: 0, totalPages: 1 });
    } catch {
      setPosts([]);
    } finally {
      setIsLoading(false);
    }
  }, [page, search]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const handleSubmit = async () => {
    if (!form.title || !form.content || !form.category) return;
    setIsSubmitting(true);
    try {
      await api.post('/forum/posts', form);
      setIsAddOpen(false);
      setForm({ title: '', content: '', category: '' });
      fetchData();
    } catch {
      // handle
    } finally {
      setIsSubmitting(false);
    }
  };

  const CATEGORIES = [
    'Akademik', 'Kegiatan Sekolah', 'Pengumuman', 'Diskusi Umum',
    'Orang Tua & Wali', 'Prestasi', 'Lainnya',
  ];

  const columns = [
    {
      key: 'title',
      header: 'Topik',
      render: (p: ForumPost) => (
        <div>
          <p className="font-medium text-gray-900">{p.title}</p>
          <p className="text-xs text-gray-400 line-clamp-1">{p.content}</p>
        </div>
      ),
    },
    {
      key: 'category',
      header: 'Kategori',
      render: (p: ForumPost) => (
        <Badge variant="primary">{p.category}</Badge>
      ),
    },
    {
      key: 'author',
      header: 'Penulis',
      render: (p: ForumPost) => (
        <span className="text-sm text-gray-700">{p.authorName || '—'}</span>
      ),
    },
    {
      key: 'repliesCount',
      header: 'Balasan',
      render: (p: ForumPost) => (
        <div className="flex items-center gap-1.5 text-gray-600">
          <MessageSquare className="w-3.5 h-3.5" />
          <span className="text-sm font-medium">{p.repliesCount ?? 0}</span>
        </div>
      ),
    },
    {
      key: 'createdAt',
      header: 'Waktu',
      render: (p: ForumPost) => (
        <span className="text-sm text-gray-500">{formatDateTime(p.createdAt)}</span>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900">Forum & Komunitas</h1>
          <p className="text-sm text-gray-500 mt-0.5">Diskusi dan berbagi informasi antar warga sekolah</p>
        </div>
        <Button size="sm" leftIcon={<Plus className="w-3.5 h-3.5" />} onClick={() => setIsAddOpen(true)}>
          Buat Topik
        </Button>
      </div>

      <Card padding="none">
        <div className="p-4 border-b border-gray-100">
          <Input
            placeholder="Cari topik..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1); }}
            leftAddon={<Search className="w-4 h-4" />}
            className="max-w-xs"
          />
        </div>

        <Table
          columns={columns}
          data={posts}
          keyExtractor={(p) => p.id}
          isLoading={isLoading}
          emptyMessage="Belum ada topik diskusi"
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

      <Modal
        isOpen={isAddOpen}
        onClose={() => setIsAddOpen(false)}
        title="Buat Topik Baru"
        description="Mulai diskusi atau bagikan informasi"
        size="md"
        footer={
          <>
            <Button variant="secondary" onClick={() => setIsAddOpen(false)}>Batal</Button>
            <Button
              onClick={handleSubmit}
              isLoading={isSubmitting}
              leftIcon={<MessageSquare className="w-4 h-4" />}
            >
              Posting
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Judul <span className="text-red-500">*</span></label>
            <input
              className="w-full text-sm border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500"
              placeholder="Judul topik diskusi"
              value={form.title}
              onChange={(e) => setForm({ ...form, title: e.target.value })}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Kategori <span className="text-red-500">*</span></label>
            <select
              className="w-full text-sm border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500"
              value={form.category}
              onChange={(e) => setForm({ ...form, category: e.target.value })}
            >
              <option value="">Pilih kategori</option>
              {CATEGORIES.map((c) => (
                <option key={c} value={c}>{c}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Isi <span className="text-red-500">*</span></label>
            <textarea
              className="w-full text-sm border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500 resize-none"
              rows={5}
              placeholder="Tulis isi diskusi di sini..."
              value={form.content}
              onChange={(e) => setForm({ ...form, content: e.target.value })}
            />
          </div>
        </div>
      </Modal>
    </div>
  );
}
