'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Plus, Calendar, ChevronDown, ChevronUp } from 'lucide-react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import api from '@/lib/api';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input, Select } from '@/components/ui/Input';
import { Badge } from '@/components/ui/Badge';
import { Modal } from '@/components/ui/Modal';
import { Table } from '@/components/ui/Table';
import { Pagination } from '@/components/ui/Pagination';
import { formatDate, formatCurrency } from '@/lib/utils';

const schema = z.object({
  title: z.string().min(1, 'Nama event wajib diisi'),
  description: z.string().optional(),
  startDate: z.string().min(1, 'Tanggal mulai wajib diisi'),
  endDate: z.string().min(1, 'Tanggal selesai wajib diisi'),
  location: z.string().optional(),
  type: z.enum(['UPACARA', 'PENTAS', 'OLAHRAGA', 'WISATA', 'LAINNYA']),
});
type FormValues = z.infer<typeof schema>;

interface EventTask {
  id: string;
  title: string;
  status: string;
  assigneeName?: string;
}

interface EventBudget {
  id: string;
  description: string;
  amount: number;
  type: string;
}

interface SchoolEvent {
  id: string;
  title: string;
  type: string;
  startDate: string;
  status: string;
  tasks?: EventTask[];
  budgets?: EventBudget[];
}

const STATUS_MAP: Record<string, { label: string; variant: 'gray' | 'info' | 'success' }> = {
  PLANNING: { label: 'Perencanaan', variant: 'gray' },
  ACTIVE: { label: 'Aktif', variant: 'info' },
  COMPLETED: { label: 'Selesai', variant: 'success' },
};

const TYPE_LABELS: Record<string, string> = {
  UPACARA: 'Upacara',
  PENTAS: 'Pentas Seni',
  OLAHRAGA: 'Olahraga',
  WISATA: 'Wisata',
  LAINNYA: 'Lainnya',
};

export default function EventsPage() {
  const [items, setItems] = useState<SchoolEvent[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [meta, setMeta] = useState({ total: 0, totalPages: 1 });
  const [isAddOpen, setIsAddOpen] = useState(false);
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const [detailData, setDetailData] = useState<Record<string, { tasks: EventTask[]; budgets: EventBudget[] }>>({});
  const [detailLoading, setDetailLoading] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { type: 'LAINNYA' },
  });

  const fetchData = useCallback(async () => {
    setIsLoading(true);
    try {
      const res = await api.get('/api/v1/events', { params: { page, limit: 20 } });
      setItems(res.data.data || []);
      if (res.data.meta) setMeta(res.data.meta);
    } catch {
      // handle silently
    } finally {
      setIsLoading(false);
    }
  }, [page]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const toggleExpand = async (eventId: string) => {
    if (expandedId === eventId) {
      setExpandedId(null);
      return;
    }
    setExpandedId(eventId);
    if (detailData[eventId]) return;

    setDetailLoading(eventId);
    try {
      const [tasksRes, budgetsRes] = await Promise.allSettled([
        api.get(`/api/v1/events/${eventId}/tasks`),
        api.get(`/api/v1/events/${eventId}/budget`),
      ]);
      setDetailData((prev) => ({
        ...prev,
        [eventId]: {
          tasks: tasksRes.status === 'fulfilled' ? tasksRes.value.data.data || [] : [],
          budgets: budgetsRes.status === 'fulfilled' ? budgetsRes.value.data.data || [] : [],
        },
      }));
    } catch {
      // handle silently
    } finally {
      setDetailLoading(null);
    }
  };

  const onSubmit = async (values: FormValues) => {
    await api.post('/api/v1/events', values);
    reset();
    setIsAddOpen(false);
    fetchData();
  };

  const columns = [
    {
      key: 'title',
      header: 'Nama Event',
      render: (item: SchoolEvent) => (
        <button
          onClick={() => toggleExpand(item.id)}
          className="flex items-center gap-1 text-left font-medium text-primary-600 hover:text-primary-800 transition-colors"
        >
          {item.title}
          {expandedId === item.id ? (
            <ChevronUp className="w-3.5 h-3.5" />
          ) : (
            <ChevronDown className="w-3.5 h-3.5" />
          )}
        </button>
      ),
    },
    {
      key: 'type',
      header: 'Tipe',
      render: (item: SchoolEvent) => TYPE_LABELS[item.type] ?? item.type,
    },
    {
      key: 'startDate',
      header: 'Tanggal Mulai',
      render: (item: SchoolEvent) => formatDate(item.startDate),
    },
    {
      key: 'status',
      header: 'Status',
      render: (item: SchoolEvent) => {
        const map = STATUS_MAP[item.status] ?? { label: item.status, variant: 'gray' as const };
        return <Badge variant={map.variant} size="sm">{map.label}</Badge>;
      },
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="p-2 bg-primary-100 rounded-lg">
            <Calendar className="w-5 h-5 text-primary-600" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-gray-900">Event & Kepanitiaan</h1>
            <p className="text-sm text-gray-500 mt-0.5">Kelola acara dan kegiatan sekolah</p>
          </div>
        </div>
        <Button
          size="sm"
          leftIcon={<Plus className="w-3.5 h-3.5" />}
          onClick={() => setIsAddOpen(true)}
        >
          Tambah Event
        </Button>
      </div>

      <Card padding="none">
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                {columns.map((col) => (
                  <th
                    key={col.key}
                    className="px-4 py-3 text-left text-xs font-semibold text-gray-500 uppercase tracking-wider"
                  >
                    {col.header}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-100">
              {isLoading ? (
                <tr>
                  <td colSpan={columns.length} className="px-4 py-12 text-center">
                    <div className="flex items-center justify-center gap-2 text-gray-400">
                      <div className="w-5 h-5 border-2 border-primary-500 border-t-transparent rounded-full animate-spin" />
                      <span className="text-sm">Memuat data...</span>
                    </div>
                  </td>
                </tr>
              ) : items.length === 0 ? (
                <tr>
                  <td colSpan={columns.length} className="px-4 py-12 text-center text-sm text-gray-400">
                    Belum ada event yang dibuat
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <React.Fragment key={item.id}>
                    <tr className="hover:bg-gray-50 transition-colors">
                      {columns.map((col) => (
                        <td key={col.key} className="px-4 py-3 text-sm text-gray-700">
                          {col.render(item)}
                        </td>
                      ))}
                    </tr>
                    {expandedId === item.id && (
                      <tr>
                        <td colSpan={columns.length} className="px-4 py-4 bg-gray-50 border-t border-gray-100">
                          {detailLoading === item.id ? (
                            <div className="flex items-center gap-2 text-sm text-gray-400 py-2">
                              <div className="w-4 h-4 border-2 border-primary-500 border-t-transparent rounded-full animate-spin" />
                              Memuat detail...
                            </div>
                          ) : (
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                              {/* Tasks */}
                              <div>
                                <p className="text-xs font-semibold text-gray-500 uppercase mb-2">
                                  Tugas Panitia
                                </p>
                                {(detailData[item.id]?.tasks ?? []).length === 0 ? (
                                  <p className="text-sm text-gray-400">Belum ada tugas</p>
                                ) : (
                                  <ul className="space-y-1">
                                    {detailData[item.id].tasks.map((task) => (
                                      <li key={task.id} className="flex items-center gap-2 text-sm">
                                        <Badge
                                          variant={task.status === 'DONE' ? 'success' : 'warning'}
                                          size="sm"
                                        >
                                          {task.status}
                                        </Badge>
                                        <span className="text-gray-700">{task.title}</span>
                                        {task.assigneeName && (
                                          <span className="text-gray-400 text-xs">— {task.assigneeName}</span>
                                        )}
                                      </li>
                                    ))}
                                  </ul>
                                )}
                              </div>
                              {/* Budget */}
                              <div>
                                <p className="text-xs font-semibold text-gray-500 uppercase mb-2">
                                  Anggaran
                                </p>
                                {(detailData[item.id]?.budgets ?? []).length === 0 ? (
                                  <p className="text-sm text-gray-400">Belum ada anggaran</p>
                                ) : (
                                  <ul className="space-y-1">
                                    {detailData[item.id].budgets.map((b) => (
                                      <li key={b.id} className="flex justify-between text-sm">
                                        <span className="text-gray-700">{b.description}</span>
                                        <span
                                          className={`font-semibold ${
                                            b.type === 'INCOME' ? 'text-green-600' : 'text-red-600'
                                          }`}
                                        >
                                          {b.type === 'INCOME' ? '+' : '-'}{formatCurrency(b.amount)}
                                        </span>
                                      </li>
                                    ))}
                                  </ul>
                                )}
                              </div>
                            </div>
                          )}
                        </td>
                      </tr>
                    )}
                  </React.Fragment>
                ))
              )}
            </tbody>
          </table>
        </div>
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
        onClose={() => { setIsAddOpen(false); reset(); }}
        title="Tambah Event"
        size="md"
        footer={
          <>
            <Button variant="secondary" onClick={() => { setIsAddOpen(false); reset(); }}>
              Batal
            </Button>
            <Button form="event-form" type="submit" isLoading={isSubmitting}>
              Simpan
            </Button>
          </>
        }
      >
        <form id="event-form" onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <Input
            label="Nama Event"
            placeholder="Contoh: Peringatan HUT RI ke-80"
            required
            {...register('title')}
            error={errors.title?.message}
          />
          <div className="w-full">
            <label className="block text-sm font-medium text-gray-700 mb-1">Deskripsi</label>
            <textarea
              rows={3}
              placeholder="Deskripsi singkat event..."
              className="block w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent resize-none"
              {...register('description')}
            />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <Input
              label="Tanggal Mulai"
              type="date"
              required
              {...register('startDate')}
              error={errors.startDate?.message}
            />
            <Input
              label="Tanggal Selesai"
              type="date"
              required
              {...register('endDate')}
              error={errors.endDate?.message}
            />
          </div>
          <Input
            label="Lokasi"
            placeholder="Contoh: Lapangan Utama"
            {...register('location')}
          />
          <Select
            label="Tipe Event"
            required
            options={[
              { value: 'UPACARA', label: 'Upacara' },
              { value: 'PENTAS', label: 'Pentas Seni' },
              { value: 'OLAHRAGA', label: 'Olahraga' },
              { value: 'WISATA', label: 'Wisata' },
              { value: 'LAINNYA', label: 'Lainnya' },
            ]}
            {...register('type')}
            error={errors.type?.message}
          />
        </form>
      </Modal>
    </div>
  );
}
