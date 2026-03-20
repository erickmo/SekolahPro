'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Plus, FileEdit, Eye, Send } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import api from '@/lib/api';
import { Card, CardHeader } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input, Select } from '@/components/ui/Input';
import { Table } from '@/components/ui/Table';
import { Pagination } from '@/components/ui/Pagination';
import { Modal } from '@/components/ui/Modal';
import { Badge } from '@/components/ui/Badge';
import { formatDate, formatNumber } from '@/lib/utils';

interface BlogPost {
  id: string;
  title: string;
  category: string;
  authorId: string;
  authorName?: string;
  status: 'DRAFT' | 'PUBLISHED' | 'ARCHIVED';
  viewCount: number;
  createdAt: string;
  publishedAt?: string;
}

const STATUS_CONFIG: Record<string, { label: string; variant: 'gray' | 'success' | 'danger' }> = {
  DRAFT: { label: 'Draf', variant: 'gray' },
  PUBLISHED: { label: 'Diterbitkan', variant: 'success' },
  ARCHIVED: { label: 'Diarsipkan', variant: 'danger' },
};

export default function SchoolBlogPage() {
  const { user } = useAuth();
  const [posts, setPosts] = useState<BlogPost[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [meta, setMeta] = useState({ total: 0, totalPages: 1 });
  const [addOpen, setAddOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [publishingId, setPublishingId] = useState<string | null>(null);

  // Form state
  const [fTitle, setFTitle] = useState('');
  const [fContent, setFContent] = useState('');
  const [fCategory, setFCategory] = useState('');
  const [fFeaturedImage, setFFeaturedImage] = useState('');
  const [fStatus, setFStatus] = useState<'DRAFT' | 'PUBLISHED'>('DRAFT');

  const fetchPosts = useCallback(async () => {
    setIsLoading(true);
    try {
      const params = new URLSearchParams({ page: String(page), limit: '20' });
      const res = await api.get(`/blog/posts?${params}`);
      setPosts(res.data.data || []);
      setMeta(res.data.meta || { total: 0, totalPages: 1 });
    } catch {
      setPosts([]);
    } finally {
      setIsLoading(false);
    }
  }, [page]);

  useEffect(() => { fetchPosts(); }, [fetchPosts]);

  const resetForm = () => {
    setFTitle(''); setFContent(''); setFCategory('');
    setFFeaturedImage(''); setFStatus('DRAFT');
  };

  const handleAdd = async () => {
    if (!fTitle || !fContent) return;
    setIsSubmitting(true);
    try {
      await api.post('/blog/posts', {
        title: fTitle,
        content: fContent,
        category: fCategory || undefined,
        featuredImage: fFeaturedImage || undefined,
        status: fStatus,
      });
      setAddOpen(false);
      resetForm();
      fetchPosts();
    } finally {
      setIsSubmitting(false);
    }
  };

  const handlePublish = async (postId: string) => {
    setPublishingId(postId);
    try {
      await api.patch(`/blog/posts/${postId}`, { status: 'PUBLISHED' });
      fetchPosts();
    } finally {
      setPublishingId(null);
    }
  };

  const publishedCount = posts.filter((p) => p.status === 'PUBLISHED').length;
  const draftCount = posts.filter((p) => p.status === 'DRAFT').length;
  const totalViews = posts.reduce((sum, p) => sum + (p.viewCount || 0), 0);

  const columns = [
    {
      key: 'title',
      header: 'Judul',
      render: (p: BlogPost) => (
        <p className="font-medium text-gray-900 max-w-xs truncate">{p.title}</p>
      ),
    },
    {
      key: 'category',
      header: 'Kategori',
      render: (p: BlogPost) => p.category || <span className="text-gray-300">—</span>,
    },
    {
      key: 'authorId',
      header: 'Penulis',
      render: (p: BlogPost) => (
        <span className="text-sm text-gray-600">{p.authorName || p.authorId}</span>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (p: BlogPost) => {
        const cfg = STATUS_CONFIG[p.status] || { label: p.status, variant: 'gray' as const };
        return <Badge variant={cfg.variant} size="sm">{cfg.label}</Badge>;
      },
    },
    {
      key: 'viewCount',
      header: 'Dilihat',
      render: (p: BlogPost) => (
        <div className="flex items-center gap-1 text-sm text-gray-600">
          <Eye className="w-3.5 h-3.5" />
          {formatNumber(p.viewCount || 0)}
        </div>
      ),
    },
    {
      key: 'createdAt',
      header: 'Tanggal',
      render: (p: BlogPost) => formatDate(p.publishedAt || p.createdAt),
    },
    {
      key: 'actions',
      header: '',
      render: (p: BlogPost) => {
        if (p.status !== 'DRAFT') return null;
        return (
          <Button
            size="xs"
            variant="secondary"
            leftIcon={<Send className="w-3 h-3" />}
            isLoading={publishingId === p.id}
            onClick={() => handlePublish(p.id)}
          >
            Terbitkan
          </Button>
        );
      },
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900">Blog & Majalah Sekolah</h1>
          <p className="text-sm text-gray-500 mt-0.5">Publikasi artikel dan berita sekolah</p>
        </div>
        <Button
          size="sm"
          leftIcon={<Plus className="w-3.5 h-3.5" />}
          onClick={() => setAddOpen(true)}
        >
          Tulis Artikel
        </Button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-3 gap-4">
        <Card padding="sm">
          <div className="flex items-center gap-3">
            <div className="p-2.5 bg-green-50 rounded-lg">
              <FileEdit className="w-5 h-5 text-green-600" />
            </div>
            <div>
              <p className="text-xl font-bold text-gray-900">{publishedCount}</p>
              <p className="text-xs text-gray-500">Artikel Diterbitkan</p>
            </div>
          </div>
        </Card>
        <Card padding="sm">
          <div className="flex items-center gap-3">
            <div className="p-2.5 bg-gray-100 rounded-lg">
              <FileEdit className="w-5 h-5 text-gray-500" />
            </div>
            <div>
              <p className="text-xl font-bold text-gray-900">{draftCount}</p>
              <p className="text-xs text-gray-500">Draf</p>
            </div>
          </div>
        </Card>
        <Card padding="sm">
          <div className="flex items-center gap-3">
            <div className="p-2.5 bg-blue-50 rounded-lg">
              <Eye className="w-5 h-5 text-blue-600" />
            </div>
            <div>
              <p className="text-xl font-bold text-gray-900">{formatNumber(totalViews)}</p>
              <p className="text-xs text-gray-500">Total Tampilan</p>
            </div>
          </div>
        </Card>
      </div>

      <Card padding="none">
        <Table
          columns={columns}
          data={posts}
          keyExtractor={(p) => p.id}
          isLoading={isLoading}
          emptyMessage="Belum ada artikel blog"
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

      {/* Add Post Modal */}
      <Modal
        isOpen={addOpen}
        onClose={() => { setAddOpen(false); resetForm(); }}
        title="Tulis Artikel Baru"
        size="lg"
        footer={
          <>
            <Button variant="secondary" onClick={() => { setAddOpen(false); resetForm(); }}>Batal</Button>
            <Button onClick={handleAdd} isLoading={isSubmitting}>Simpan</Button>
          </>
        }
      >
        <div className="space-y-4">
          <Input
            label="Judul Artikel"
            placeholder="Masukkan judul artikel..."
            required
            value={fTitle}
            onChange={(e) => setFTitle(e.target.value)}
          />
          <div className="grid grid-cols-2 gap-4">
            <Input
              label="Kategori"
              placeholder="Contoh: Berita, Akademik, Event..."
              value={fCategory}
              onChange={(e) => setFCategory(e.target.value)}
            />
            <Select
              label="Status"
              value={fStatus}
              onChange={(e) => setFStatus(e.target.value as 'DRAFT' | 'PUBLISHED')}
              options={[
                { value: 'DRAFT', label: 'Draf' },
                { value: 'PUBLISHED', label: 'Terbitkan Langsung' },
              ]}
            />
          </div>
          <Input
            label="URL Gambar Utama (opsional)"
            placeholder="https://..."
            value={fFeaturedImage}
            onChange={(e) => setFFeaturedImage(e.target.value)}
          />
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Konten Artikel <span className="text-red-500">*</span>
            </label>
            <textarea
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
              rows={8}
              placeholder="Tulis konten artikel di sini..."
              value={fContent}
              onChange={(e) => setFContent(e.target.value)}
            />
          </div>
        </div>
      </Modal>
    </div>
  );
}
