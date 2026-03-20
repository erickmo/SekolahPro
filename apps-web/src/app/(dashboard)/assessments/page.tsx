'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Plus, ClipboardList } from 'lucide-react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import api from '@/lib/api';
import { useAuth } from '@/context/AuthContext';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Modal } from '@/components/ui/Modal';
import { Table } from '@/components/ui/Table';
import { Pagination } from '@/components/ui/Pagination';
import { formatDate } from '@/lib/utils';

type TabKey = 'rubrics' | 'projects';

const rubricSchema = z.object({
  subject: z.string().min(1, 'Mata pelajaran wajib diisi'),
  gradeLevel: z.string().min(1, 'Kelas wajib diisi'),
  title: z.string().min(1, 'Judul wajib diisi'),
  criteria: z.string().optional(),
});
type RubricForm = z.infer<typeof rubricSchema>;

interface Rubric {
  id: string;
  title: string;
  subject: string;
  gradeLevel: string;
  createdAt: string;
}

interface Project {
  id: string;
  studentName?: string;
  studentNisn?: string;
  rubricTitle?: string;
  totalScore: number;
  feedback?: string;
  assessedAt?: string;
  createdAt: string;
}

export default function AssessmentsPage() {
  const { user } = useAuth();
  const [activeTab, setActiveTab] = useState<TabKey>('rubrics');

  const [rubrics, setRubrics] = useState<Rubric[]>([]);
  const [rubricLoading, setRubricLoading] = useState(true);
  const [rubricPage, setRubricPage] = useState(1);
  const [rubricMeta, setRubricMeta] = useState({ total: 0, totalPages: 1 });
  const [isRubricOpen, setIsRubricOpen] = useState(false);

  const [projects, setProjects] = useState<Project[]>([]);
  const [projectLoading, setProjectLoading] = useState(true);
  const [projectPage, setProjectPage] = useState(1);
  const [projectMeta, setProjectMeta] = useState({ total: 0, totalPages: 1 });

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<RubricForm>({ resolver: zodResolver(rubricSchema) });

  const fetchRubrics = useCallback(async () => {
    setRubricLoading(true);
    try {
      const res = await api.get('/api/v1/assessments/rubrics', {
        params: { page: rubricPage, limit: 20 },
      });
      setRubrics(res.data.data || []);
      if (res.data.meta) setRubricMeta(res.data.meta);
    } catch {
      // handle silently
    } finally {
      setRubricLoading(false);
    }
  }, [rubricPage]);

  const fetchProjects = useCallback(async () => {
    setProjectLoading(true);
    try {
      const res = await api.get('/api/v1/assessments/projects', {
        params: { page: projectPage, limit: 20 },
      });
      setProjects(res.data.data || []);
      if (res.data.meta) setProjectMeta(res.data.meta);
    } catch {
      // handle silently
    } finally {
      setProjectLoading(false);
    }
  }, [projectPage]);

  useEffect(() => {
    fetchRubrics();
  }, [fetchRubrics]);

  useEffect(() => {
    fetchProjects();
  }, [fetchProjects]);

  const onAddRubric = async (values: RubricForm) => {
    let parsedCriteria: Record<string, unknown> | undefined;
    if (values.criteria) {
      try {
        parsedCriteria = JSON.parse(values.criteria);
      } catch {
        parsedCriteria = { raw: values.criteria };
      }
    }
    await api.post('/api/v1/assessments/rubrics', {
      ...values,
      teacherId: user?.userId,
      criteria: parsedCriteria,
    });
    reset();
    setIsRubricOpen(false);
    fetchRubrics();
  };

  const tabs = [
    { key: 'rubrics' as TabKey, label: 'Rubrik Penilaian', count: rubrics.length },
    { key: 'projects' as TabKey, label: 'Penilaian Proyek', count: projects.length },
  ];

  const rubricColumns = [
    {
      key: 'title',
      header: 'Judul',
      render: (r: Rubric) => <p className="font-medium text-gray-900">{r.title}</p>,
    },
    { key: 'subject', header: 'Mata Pelajaran' },
    { key: 'gradeLevel', header: 'Kelas' },
    {
      key: 'createdAt',
      header: 'Dibuat',
      render: (r: Rubric) => formatDate(r.createdAt),
    },
  ];

  const projectColumns = [
    {
      key: 'studentName',
      header: 'Siswa',
      render: (p: Project) => (
        <div>
          <p className="font-medium text-gray-900">{p.studentName || '—'}</p>
          {p.studentNisn && <p className="text-xs text-gray-400">NISN: {p.studentNisn}</p>}
        </div>
      ),
    },
    {
      key: 'rubricTitle',
      header: 'Rubrik',
      render: (p: Project) => p.rubricTitle || '—',
    },
    {
      key: 'totalScore',
      header: 'Total Skor',
      render: (p: Project) => (
        <span className="font-semibold text-gray-900">{p.totalScore}</span>
      ),
    },
    {
      key: 'feedback',
      header: 'Feedback',
      render: (p: Project) =>
        p.feedback ? (
          <span className="text-sm text-gray-600 line-clamp-2">{p.feedback}</span>
        ) : (
          <span className="text-gray-400">—</span>
        ),
    },
    {
      key: 'assessedAt',
      header: 'Tanggal',
      render: (p: Project) => (p.assessedAt ? formatDate(p.assessedAt) : formatDate(p.createdAt)),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="p-2 bg-primary-100 rounded-lg">
            <ClipboardList className="w-5 h-5 text-primary-600" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-gray-900">Penilaian</h1>
            <p className="text-sm text-gray-500 mt-0.5">Rubrik penilaian dan penilaian proyek siswa</p>
          </div>
        </div>
        {activeTab === 'rubrics' && (
          <Button
            size="sm"
            leftIcon={<Plus className="w-3.5 h-3.5" />}
            onClick={() => setIsRubricOpen(true)}
          >
            Tambah Rubrik
          </Button>
        )}
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

      {/* Rubrik */}
      {activeTab === 'rubrics' && (
        <Card padding="none">
          <Table
            columns={rubricColumns}
            data={rubrics}
            keyExtractor={(r) => r.id}
            isLoading={rubricLoading}
            emptyMessage="Belum ada rubrik penilaian"
          />
          {rubricMeta.total > 0 && (
            <div className="px-4 py-3 border-t border-gray-100">
              <Pagination
                currentPage={rubricPage}
                totalPages={rubricMeta.totalPages}
                total={rubricMeta.total}
                limit={20}
                onPageChange={setRubricPage}
              />
            </div>
          )}
        </Card>
      )}

      {/* Penilaian Proyek */}
      {activeTab === 'projects' && (
        <Card padding="none">
          <Table
            columns={projectColumns}
            data={projects}
            keyExtractor={(p) => p.id}
            isLoading={projectLoading}
            emptyMessage="Belum ada penilaian proyek"
          />
          {projectMeta.total > 0 && (
            <div className="px-4 py-3 border-t border-gray-100">
              <Pagination
                currentPage={projectPage}
                totalPages={projectMeta.totalPages}
                total={projectMeta.total}
                limit={20}
                onPageChange={setProjectPage}
              />
            </div>
          )}
        </Card>
      )}

      {/* Add Rubric Modal */}
      <Modal
        isOpen={isRubricOpen}
        onClose={() => { setIsRubricOpen(false); reset(); }}
        title="Tambah Rubrik Penilaian"
        size="md"
        footer={
          <>
            <Button variant="secondary" onClick={() => { setIsRubricOpen(false); reset(); }}>
              Batal
            </Button>
            <Button form="rubric-form" type="submit" isLoading={isSubmitting}>
              Simpan
            </Button>
          </>
        }
      >
        <form id="rubric-form" onSubmit={handleSubmit(onAddRubric)} className="space-y-4">
          <Input
            label="Judul Rubrik"
            placeholder="Contoh: Rubrik Presentasi Ilmiah"
            required
            {...register('title')}
            error={errors.title?.message}
          />
          <Input
            label="Mata Pelajaran"
            placeholder="Contoh: IPA"
            required
            {...register('subject')}
            error={errors.subject?.message}
          />
          <Input
            label="Kelas"
            placeholder="Contoh: VIII B"
            required
            {...register('gradeLevel')}
            error={errors.gradeLevel?.message}
          />
          <div className="w-full">
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Kriteria <span className="text-xs text-gray-400">(JSON opsional)</span>
            </label>
            <textarea
              rows={4}
              placeholder='[{"nama": "Konten", "bobot": 40}, {"nama": "Presentasi", "bobot": 60}]'
              className="block w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent resize-none"
              {...register('criteria')}
            />
          </div>
        </form>
      </Modal>
    </div>
  );
}
