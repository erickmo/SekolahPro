'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Plus, GraduationCap, BookOpen, Users, Calendar, CheckCircle } from 'lucide-react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import api from '@/lib/api';
import { useAuth } from '@/context/AuthContext';
import type { AcademicYear, Subject, SchoolClass, Teacher } from '@/types';
import { Card, CardHeader } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input, Select } from '@/components/ui/Input';
import { Badge } from '@/components/ui/Badge';
import { Modal } from '@/components/ui/Modal';
import { Table } from '@/components/ui/Table';
import { formatDate } from '@/lib/utils';

type TabKey = 'years' | 'subjects' | 'classes' | 'teachers';

const yearSchema = z.object({
  name: z.string().min(1, 'Nama tahun ajaran wajib diisi'),
  startDate: z.string().min(1, 'Tanggal mulai wajib diisi'),
  endDate: z.string().min(1, 'Tanggal selesai wajib diisi'),
});
type YearForm = z.infer<typeof yearSchema>;

const subjectSchema = z.object({
  name: z.string().min(1, 'Nama mata pelajaran wajib diisi'),
  code: z.string().min(1, 'Kode wajib diisi'),
  gradeLevel: z.string().min(1, 'Tingkat kelas wajib diisi'),
  curriculum: z.enum(['K13', 'MERDEKA']),
  weeklyHours: z.number().min(1).max(40),
});
type SubjectForm = z.infer<typeof subjectSchema>;

export default function AcademicPage() {
  const { user } = useAuth();
  const [activeTab, setActiveTab] = useState<TabKey>('years');
  const [years, setYears] = useState<AcademicYear[]>([]);
  const [subjects, setSubjects] = useState<Subject[]>([]);
  const [classes, setClasses] = useState<SchoolClass[]>([]);
  const [teachers, setTeachers] = useState<Teacher[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [addYearOpen, setAddYearOpen] = useState(false);
  const [addSubjectOpen, setAddSubjectOpen] = useState(false);

  const canManage = ['ADMIN_SEKOLAH', 'KEPALA_KURIKULUM', 'KEPALA_SEKOLAH'].includes(user?.role || '');

  const {
    register: regYear,
    handleSubmit: submitYear,
    reset: resetYear,
    formState: { errors: errYear, isSubmitting: submittingYear },
  } = useForm<YearForm>({ resolver: zodResolver(yearSchema) });

  const {
    register: regSubject,
    handleSubmit: submitSubject,
    reset: resetSubject,
    formState: { errors: errSubject, isSubmitting: submittingSubject },
  } = useForm<SubjectForm>({
    resolver: zodResolver(subjectSchema),
    defaultValues: { curriculum: 'MERDEKA', weeklyHours: 2 },
  });

  const fetchData = useCallback(async () => {
    setIsLoading(true);
    try {
      const [yRes, sRes, cRes, tRes] = await Promise.all([
        api.get('/academic/academic-years'),
        api.get('/academic/subjects'),
        api.get('/academic/classes'),
        api.get('/academic/teachers'),
      ]);
      setYears(yRes.data.data || []);
      setSubjects(sRes.data.data || []);
      setClasses(cRes.data.data || []);
      setTeachers(tRes.data.data || []);
    } catch {
      // handle silently
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const onAddYear = async (values: YearForm) => {
    await api.post('/academic/academic-years', values);
    resetYear();
    setAddYearOpen(false);
    fetchData();
  };

  const onAddSubject = async (values: SubjectForm) => {
    await api.post('/academic/subjects', values);
    resetSubject();
    setAddSubjectOpen(false);
    fetchData();
  };

  const tabs = [
    { key: 'years' as TabKey, label: 'Tahun Ajaran', icon: <Calendar className="w-4 h-4" />, count: years.length },
    { key: 'subjects' as TabKey, label: 'Mata Pelajaran', icon: <BookOpen className="w-4 h-4" />, count: subjects.length },
    { key: 'classes' as TabKey, label: 'Kelas', icon: <Users className="w-4 h-4" />, count: classes.length },
    { key: 'teachers' as TabKey, label: 'Guru', icon: <GraduationCap className="w-4 h-4" />, count: teachers.length },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900">Akademik & Kurikulum</h1>
          <p className="text-sm text-gray-500 mt-0.5">Kelola tahun ajaran, mata pelajaran, dan kelas</p>
        </div>
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
            {tab.icon}
            {tab.label}
            <span className={`text-xs px-1.5 py-0.5 rounded-full ${
              activeTab === tab.key ? 'bg-primary-100 text-primary-700' : 'bg-gray-100 text-gray-500'
            }`}>
              {tab.count}
            </span>
          </button>
        ))}
      </div>

      {/* Academic Years */}
      {activeTab === 'years' && (
        <div className="space-y-4">
          <div className="flex justify-between items-center">
            <p className="text-sm text-gray-500">{years.length} tahun ajaran</p>
            {canManage && (
              <Button size="sm" leftIcon={<Plus className="w-3.5 h-3.5" />} onClick={() => setAddYearOpen(true)}>
                Tambah Tahun Ajaran
              </Button>
            )}
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {years.map((y) => (
              <Card key={y.id} padding="sm" className="hover:shadow-card-hover transition-shadow">
                <div className="flex items-start justify-between">
                  <div>
                    <p className="font-semibold text-gray-900">{y.name}</p>
                    <p className="text-xs text-gray-500 mt-1">
                      {formatDate(y.startDate)} — {formatDate(y.endDate)}
                    </p>
                  </div>
                  {y.isActive && (
                    <Badge variant="success" size="sm">
                      <CheckCircle className="w-3 h-3 mr-1" />
                      Aktif
                    </Badge>
                  )}
                </div>
              </Card>
            ))}
            {years.length === 0 && !isLoading && (
              <div className="col-span-3 text-center py-12 text-sm text-gray-400">
                Belum ada data tahun ajaran
              </div>
            )}
          </div>
        </div>
      )}

      {/* Subjects */}
      {activeTab === 'subjects' && (
        <div className="space-y-4">
          <div className="flex justify-between items-center">
            <p className="text-sm text-gray-500">{subjects.length} mata pelajaran</p>
            {canManage && (
              <Button size="sm" leftIcon={<Plus className="w-3.5 h-3.5" />} onClick={() => setAddSubjectOpen(true)}>
                Tambah Mata Pelajaran
              </Button>
            )}
          </div>
          <Card padding="none">
            <Table
              columns={[
                { key: 'name', header: 'Mata Pelajaran', render: (s) => (
                  <div>
                    <p className="font-medium">{s.name}</p>
                    <p className="text-xs text-gray-400">{s.code}</p>
                  </div>
                )},
                { key: 'gradeLevel', header: 'Tingkat' },
                { key: 'curriculum', header: 'Kurikulum', render: (s) => (
                  <Badge variant={s.curriculum === 'MERDEKA' ? 'success' : 'primary'} size="sm">{s.curriculum}</Badge>
                )},
                { key: 'weeklyHours', header: 'Jam/Minggu', render: (s) => `${s.weeklyHours} jam` },
              ]}
              data={subjects}
              keyExtractor={(s) => s.id}
              isLoading={isLoading}
              emptyMessage="Belum ada mata pelajaran"
            />
          </Card>
        </div>
      )}

      {/* Classes */}
      {activeTab === 'classes' && (
        <Card padding="none">
          <Table
            columns={[
              { key: 'name', header: 'Kelas' },
              { key: 'gradeLevel', header: 'Tingkat' },
              { key: 'studentCount', header: 'Jumlah Siswa', render: (c) => (
                <span className="font-medium">{c.studentCount ?? 0} / {c.capacity}</span>
              )},
            ]}
            data={classes}
            keyExtractor={(c) => c.id}
            isLoading={isLoading}
            emptyMessage="Belum ada kelas"
          />
        </Card>
      )}

      {/* Teachers */}
      {activeTab === 'teachers' && (
        <Card padding="none">
          <Table
            columns={[
              { key: 'name', header: 'Nama Guru', render: (t) => (
                <div>
                  <p className="font-medium">{t.name}</p>
                  {t.nip && <p className="text-xs text-gray-400">NIP: {t.nip}</p>}
                </div>
              )},
              { key: 'email', header: 'Email' },
              { key: 'phone', header: 'Telepon', render: (t) => t.phone || '—' },
            ]}
            data={teachers}
            keyExtractor={(t) => t.id}
            isLoading={isLoading}
            emptyMessage="Belum ada data guru"
          />
        </Card>
      )}

      {/* Add Year Modal */}
      <Modal
        isOpen={addYearOpen}
        onClose={() => { setAddYearOpen(false); resetYear(); }}
        title="Tambah Tahun Ajaran"
        size="sm"
        footer={
          <>
            <Button variant="secondary" onClick={() => setAddYearOpen(false)}>Batal</Button>
            <Button form="year-form" type="submit" isLoading={submittingYear}>Simpan</Button>
          </>
        }
      >
        <form id="year-form" onSubmit={submitYear(onAddYear)} className="space-y-4">
          <Input label="Nama Tahun Ajaran" placeholder="2024/2025" required {...regYear('name')} error={errYear.name?.message} />
          <Input label="Tanggal Mulai" type="date" required {...regYear('startDate')} error={errYear.startDate?.message} />
          <Input label="Tanggal Selesai" type="date" required {...regYear('endDate')} error={errYear.endDate?.message} />
        </form>
      </Modal>

      {/* Add Subject Modal */}
      <Modal
        isOpen={addSubjectOpen}
        onClose={() => { setAddSubjectOpen(false); resetSubject(); }}
        title="Tambah Mata Pelajaran"
        size="md"
        footer={
          <>
            <Button variant="secondary" onClick={() => setAddSubjectOpen(false)}>Batal</Button>
            <Button form="subject-form" type="submit" isLoading={submittingSubject}>Simpan</Button>
          </>
        }
      >
        <form id="subject-form" onSubmit={submitSubject(onAddSubject)} className="space-y-4">
          <Input label="Nama Mata Pelajaran" placeholder="Matematika" required {...regSubject('name')} error={errSubject.name?.message} />
          <Input label="Kode" placeholder="MTK" required {...regSubject('code')} error={errSubject.code?.message} />
          <Input label="Tingkat Kelas" placeholder="X / XI / XII" required {...regSubject('gradeLevel')} error={errSubject.gradeLevel?.message} />
          <Select
            label="Kurikulum"
            required
            options={[{ value: 'MERDEKA', label: 'Merdeka Belajar' }, { value: 'K13', label: 'Kurikulum 2013' }]}
            {...regSubject('curriculum')}
            error={errSubject.curriculum?.message}
          />
          <Input
            label="Jam Pelajaran per Minggu"
            type="number"
            min={1}
            max={40}
            required
            {...regSubject('weeklyHours', { valueAsNumber: true })}
            error={errSubject.weeklyHours?.message}
          />
        </form>
      </Modal>
    </div>
  );
}
