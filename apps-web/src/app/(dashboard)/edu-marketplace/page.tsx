'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Plus, ShoppingBag, Star } from 'lucide-react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import api from '@/lib/api';
import { useAuth } from '@/context/AuthContext';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input, Select } from '@/components/ui/Input';
import { Badge } from '@/components/ui/Badge';
import { Modal } from '@/components/ui/Modal';
import { Table } from '@/components/ui/Table';
import { Pagination } from '@/components/ui/Pagination';
import { formatCurrency } from '@/lib/utils';

type TabKey = 'public' | 'my-content';

const contentSchema = z.object({
  title: z.string().min(1, 'Judul wajib diisi'),
  description: z.string().optional(),
  type: z.enum(['MODULE', 'VIDEO', 'RPP', 'RUBRIC', 'EXAM_PACK']),
  price: z.number().min(0, 'Harga tidak boleh negatif'),
  licenseType: z.enum(['SINGLE_SCHOOL', 'OPEN']),
});
type ContentForm = z.infer<typeof contentSchema>;

interface EduContent {
  id: string;
  title: string;
  type: string;
  price: number;
  averageRating?: number;
  downloadCount?: number;
  status?: string;
  creatorName?: string;
}

const TYPE_LABELS: Record<string, string> = {
  MODULE: 'Modul',
  VIDEO: 'Video',
  RPP: 'RPP',
  RUBRIC: 'Rubrik',
  EXAM_PACK: 'Paket Ujian',
};

const TYPE_VARIANTS: Record<string, 'primary' | 'info' | 'success' | 'warning' | 'danger'> = {
  MODULE: 'primary',
  VIDEO: 'info',
  RPP: 'success',
  RUBRIC: 'warning',
  EXAM_PACK: 'danger',
};

const STATUS_MAP: Record<string, { label: string; variant: 'gray' | 'warning' | 'success' | 'danger' }> = {
  DRAFT: { label: 'Draft', variant: 'gray' },
  PENDING_REVIEW: { label: 'Menunggu Review', variant: 'warning' },
  PUBLISHED: { label: 'Diterbitkan', variant: 'success' },
  REJECTED: { label: 'Ditolak', variant: 'danger' },
};

export default function EduMarketplacePage() {
  const { user } = useAuth();
  const [activeTab, setActiveTab] = useState<TabKey>('public');

  const [publicItems, setPublicItems] = useState<EduContent[]>([]);
  const [publicLoading, setPublicLoading] = useState(true);
  const [publicPage, setPublicPage] = useState(1);
  const [publicMeta, setPublicMeta] = useState({ total: 0, totalPages: 1 });

  const [myItems, setMyItems] = useState<EduContent[]>([]);
  const [myLoading, setMyLoading] = useState(true);
  const [myPage, setMyPage] = useState(1);
  const [myMeta, setMyMeta] = useState({ total: 0, totalPages: 1 });

  const [isAddOpen, setIsAddOpen] = useState(false);
  const [purchasing, setPurchasing] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<ContentForm>({
    resolver: zodResolver(contentSchema),
    defaultValues: { type: 'MODULE', price: 0, licenseType: 'SINGLE_SCHOOL' },
  });

  const fetchPublic = useCallback(async () => {
    setPublicLoading(true);
    try {
      const res = await api.get('/api/v1/edu-marketplace/content', {
        params: { page: publicPage, limit: 20 },
      });
      setPublicItems(res.data.data || []);
      if (res.data.meta) setPublicMeta(res.data.meta);
    } catch {
      // handle silently
    } finally {
      setPublicLoading(false);
    }
  }, [publicPage]);

  const fetchMyContent = useCallback(async () => {
    setMyLoading(true);
    try {
      const res = await api.get('/api/v1/edu-marketplace/my-content', {
        params: { page: myPage, limit: 20 },
      });
      setMyItems(res.data.data || []);
      if (res.data.meta) setMyMeta(res.data.meta);
    } catch {
      // handle silently
    } finally {
      setMyLoading(false);
    }
  }, [myPage]);

  useEffect(() => {
    fetchPublic();
  }, [fetchPublic]);

  useEffect(() => {
    fetchMyContent();
  }, [fetchMyContent]);

  const onAddContent = async (values: ContentForm) => {
    await api.post('/api/v1/edu-marketplace/content', {
      ...values,
      creatorId: user?.userId,
    });
    reset();
    setIsAddOpen(false);
    fetchMyContent();
    fetchPublic();
  };

  const handlePurchase = async (contentId: string) => {
    setPurchasing(contentId);
    try {
      await api.post('/api/v1/edu-marketplace/purchase', { contentId });
      fetchPublic();
    } catch {
      // handle silently
    } finally {
      setPurchasing(null);
    }
  };

  const tabs = [
    { key: 'public' as TabKey, label: 'Semua Konten', count: publicMeta.total },
    { key: 'my-content' as TabKey, label: 'Konten Saya', count: myMeta.total },
  ];

  const publicColumns = [
    {
      key: 'title',
      header: 'Judul',
      render: (item: EduContent) => (
        <div>
          <p className="font-medium text-gray-900">{item.title}</p>
          {item.creatorName && (
            <p className="text-xs text-gray-400">oleh {item.creatorName}</p>
          )}
        </div>
      ),
    },
    {
      key: 'type',
      header: 'Tipe',
      render: (item: EduContent) => (
        <Badge variant={TYPE_VARIANTS[item.type] ?? 'default'} size="sm">
          {TYPE_LABELS[item.type] ?? item.type}
        </Badge>
      ),
    },
    {
      key: 'price',
      header: 'Harga',
      render: (item: EduContent) =>
        item.price === 0 ? (
          <Badge variant="success" size="sm">Gratis</Badge>
        ) : (
          <span className="font-semibold text-gray-900">{formatCurrency(item.price)}</span>
        ),
    },
    {
      key: 'averageRating',
      header: 'Rating',
      render: (item: EduContent) =>
        item.averageRating !== undefined ? (
          <span className="flex items-center gap-1 text-sm text-amber-600">
            <Star className="w-3.5 h-3.5 fill-amber-400 text-amber-400" />
            {item.averageRating.toFixed(1)}
          </span>
        ) : (
          <span className="text-gray-400 text-sm">—</span>
        ),
    },
    {
      key: 'downloadCount',
      header: 'Downloads',
      render: (item: EduContent) =>
        item.downloadCount !== undefined ? (
          <span className="text-sm text-gray-600">{item.downloadCount}</span>
        ) : (
          <span className="text-gray-400">—</span>
        ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (item: EduContent) => {
        const s = item.status ?? 'PUBLISHED';
        const map = STATUS_MAP[s] ?? { label: s, variant: 'gray' as const };
        return <Badge variant={map.variant} size="sm">{map.label}</Badge>;
      },
    },
    {
      key: 'actions',
      header: '',
      render: (item: EduContent) => (
        <Button
          variant="outline"
          size="sm"
          isLoading={purchasing === item.id}
          onClick={() => handlePurchase(item.id)}
        >
          {item.price === 0 ? 'Unduh' : 'Beli'}
        </Button>
      ),
      className: 'w-24',
    },
  ];

  const myColumns = [
    {
      key: 'title',
      header: 'Judul',
      render: (item: EduContent) => (
        <p className="font-medium text-gray-900">{item.title}</p>
      ),
    },
    {
      key: 'type',
      header: 'Tipe',
      render: (item: EduContent) => (
        <Badge variant={TYPE_VARIANTS[item.type] ?? 'default'} size="sm">
          {TYPE_LABELS[item.type] ?? item.type}
        </Badge>
      ),
    },
    {
      key: 'price',
      header: 'Harga',
      render: (item: EduContent) =>
        item.price === 0 ? (
          <Badge variant="success" size="sm">Gratis</Badge>
        ) : (
          <span className="font-semibold text-gray-900">{formatCurrency(item.price)}</span>
        ),
    },
    {
      key: 'averageRating',
      header: 'Rating',
      render: (item: EduContent) =>
        item.averageRating !== undefined ? (
          <span className="flex items-center gap-1 text-sm text-amber-600">
            <Star className="w-3.5 h-3.5 fill-amber-400 text-amber-400" />
            {item.averageRating.toFixed(1)}
          </span>
        ) : (
          <span className="text-gray-400 text-sm">—</span>
        ),
    },
    {
      key: 'downloadCount',
      header: 'Downloads',
      render: (item: EduContent) =>
        item.downloadCount !== undefined ? `${item.downloadCount}` : '—',
    },
    {
      key: 'status',
      header: 'Status',
      render: (item: EduContent) => {
        const s = item.status ?? 'DRAFT';
        const map = STATUS_MAP[s] ?? { label: s, variant: 'gray' as const };
        return <Badge variant={map.variant} size="sm">{map.label}</Badge>;
      },
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="p-2 bg-primary-100 rounded-lg">
            <ShoppingBag className="w-5 h-5 text-primary-600" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-gray-900">Marketplace Konten Edukatif</h1>
            <p className="text-sm text-gray-500 mt-0.5">Temukan dan bagikan konten pembelajaran berkualitas</p>
          </div>
        </div>
        <Button
          size="sm"
          leftIcon={<Plus className="w-3.5 h-3.5" />}
          onClick={() => setIsAddOpen(true)}
        >
          Unggah Konten
        </Button>
      </div>

      {/* Tabs */}
      <div className="flex gap-2 border-b border-gray-200">
        {tabs.map((tab) => (
          <button
            key={tab.key}
            onClick={() => setActiveTab(tab.key)}
            className={`flex items-center gap-2 px-4 py-3 text-sm font-medium border-b-2 transition-colors -mb-px ${
              activeTab === tab.key
                ? 'border-primary-600 text-primary-600'
                : 'border-transparent text-gray-500 hover:text-gray-700'
            }`}
          >
            {tab.label}
            <span
              className={`text-xs px-1.5 py-0.5 rounded-full ${
                activeTab === tab.key
                  ? 'bg-primary-100 text-primary-700'
                  : 'bg-gray-100 text-gray-500'
              }`}
            >
              {tab.count}
            </span>
          </button>
        ))}
      </div>

      {/* Public Content */}
      {activeTab === 'public' && (
        <Card padding="none">
          <Table
            columns={publicColumns}
            data={publicItems}
            keyExtractor={(item) => item.id}
            isLoading={publicLoading}
            emptyMessage="Belum ada konten yang tersedia"
          />
          {publicMeta.total > 0 && (
            <div className="px-4 py-3 border-t border-gray-100">
              <Pagination
                currentPage={publicPage}
                totalPages={publicMeta.totalPages}
                total={publicMeta.total}
                limit={20}
                onPageChange={setPublicPage}
              />
            </div>
          )}
        </Card>
      )}

      {/* My Content */}
      {activeTab === 'my-content' && (
        <Card padding="none">
          <Table
            columns={myColumns}
            data={myItems}
            keyExtractor={(item) => item.id}
            isLoading={myLoading}
            emptyMessage="Anda belum mengunggah konten apapun"
          />
          {myMeta.total > 0 && (
            <div className="px-4 py-3 border-t border-gray-100">
              <Pagination
                currentPage={myPage}
                totalPages={myMeta.totalPages}
                total={myMeta.total}
                limit={20}
                onPageChange={setMyPage}
              />
            </div>
          )}
        </Card>
      )}

      {/* Add Content Modal */}
      <Modal
        isOpen={isAddOpen}
        onClose={() => { setIsAddOpen(false); reset(); }}
        title="Unggah Konten Edukatif"
        description="Konten akan diverifikasi sebelum diterbitkan di marketplace"
        size="md"
        footer={
          <>
            <Button variant="secondary" onClick={() => { setIsAddOpen(false); reset(); }}>
              Batal
            </Button>
            <Button form="edu-content-form" type="submit" isLoading={isSubmitting}>
              Unggah
            </Button>
          </>
        }
      >
        <form id="edu-content-form" onSubmit={handleSubmit(onAddContent)} className="space-y-4">
          <Input
            label="Judul Konten"
            placeholder="Contoh: Paket Soal Ujian Matematika Kelas X"
            required
            {...register('title')}
            error={errors.title?.message}
          />
          <div className="w-full">
            <label className="block text-sm font-medium text-gray-700 mb-1">Deskripsi</label>
            <textarea
              rows={3}
              placeholder="Deskripsi singkat konten yang diunggah..."
              className="block w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent resize-none"
              {...register('description')}
            />
          </div>
          <Select
            label="Tipe Konten"
            required
            options={[
              { value: 'MODULE', label: 'Modul Pembelajaran' },
              { value: 'VIDEO', label: 'Video' },
              { value: 'RPP', label: 'RPP' },
              { value: 'RUBRIC', label: 'Rubrik Penilaian' },
              { value: 'EXAM_PACK', label: 'Paket Ujian' },
            ]}
            {...register('type')}
            error={errors.type?.message}
          />
          <Input
            label="Harga (Rp)"
            type="number"
            min={0}
            required
            hint="Isi 0 untuk konten gratis"
            leftAddon="Rp"
            {...register('price', { valueAsNumber: true })}
            error={errors.price?.message}
          />
          <Select
            label="Tipe Lisensi"
            required
            options={[
              { value: 'SINGLE_SCHOOL', label: 'Per Sekolah (Single School)' },
              { value: 'OPEN', label: 'Terbuka (Open License)' },
            ]}
            {...register('licenseType')}
            error={errors.licenseType?.message}
          />
        </form>
      </Modal>
    </div>
  );
}
