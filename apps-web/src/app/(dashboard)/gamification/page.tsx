'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Plus, Trophy } from 'lucide-react';
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

type TabKey = 'rules' | 'leaderboard';

const ruleSchema = z.object({
  type: z.enum([
    'ATTENDANCE',
    'HOMEWORK',
    'EXAM_SCORE',
    'READING',
    'FORUM',
    'EXTRACURRICULAR',
    'REMEDIAL',
    'CUSTOM',
  ]),
  points: z.number().min(1, 'Poin minimal 1'),
  description: z.string().min(1, 'Keterangan wajib diisi'),
});
type RuleForm = z.infer<typeof ruleSchema>;

interface GamificationRule {
  id: string;
  type: string;
  points: number;
  description: string;
  isActive: boolean;
}

interface LeaderboardEntry {
  id: string;
  studentName?: string;
  studentNisn?: string;
  totalPoints: number;
}

const TYPE_LABELS: Record<string, string> = {
  ATTENDANCE: 'Kehadiran',
  HOMEWORK: 'Tugas Rumah',
  EXAM_SCORE: 'Nilai Ujian',
  READING: 'Membaca',
  FORUM: 'Forum Diskusi',
  EXTRACURRICULAR: 'Ekstrakurikuler',
  REMEDIAL: 'Remedial Tuntas',
  CUSTOM: 'Kustom',
};

export default function GamificationPage() {
  const [activeTab, setActiveTab] = useState<TabKey>('rules');

  const [rules, setRules] = useState<GamificationRule[]>([]);
  const [rulesLoading, setRulesLoading] = useState(true);
  const [isAddOpen, setIsAddOpen] = useState(false);

  const [leaderboard, setLeaderboard] = useState<LeaderboardEntry[]>([]);
  const [lbLoading, setLbLoading] = useState(true);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<RuleForm>({
    resolver: zodResolver(ruleSchema),
    defaultValues: { type: 'ATTENDANCE', points: 10 },
  });

  const fetchRules = useCallback(async () => {
    setRulesLoading(true);
    try {
      const res = await api.get('/api/v1/gamification/rules');
      setRules(res.data.data || []);
    } catch {
      // handle silently
    } finally {
      setRulesLoading(false);
    }
  }, []);

  const fetchLeaderboard = useCallback(async () => {
    setLbLoading(true);
    try {
      const res = await api.get('/api/v1/gamification/leaderboard');
      setLeaderboard(res.data.data || []);
    } catch {
      // handle silently
    } finally {
      setLbLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchRules();
    fetchLeaderboard();
  }, [fetchRules, fetchLeaderboard]);

  const onAddRule = async (values: RuleForm) => {
    await api.post('/api/v1/gamification/rules', values);
    reset();
    setIsAddOpen(false);
    fetchRules();
  };

  const tabs = [
    { key: 'rules' as TabKey, label: 'Aturan Poin', count: rules.length },
    { key: 'leaderboard' as TabKey, label: 'Papan Peringkat', count: leaderboard.length },
  ];

  const ruleColumns = [
    {
      key: 'type',
      header: 'Tipe',
      render: (r: GamificationRule) => (
        <span className="font-medium text-gray-900">{TYPE_LABELS[r.type] ?? r.type}</span>
      ),
    },
    {
      key: 'points',
      header: 'Poin',
      render: (r: GamificationRule) => (
        <span className="inline-flex items-center gap-1 font-semibold text-primary-600">
          +{r.points}
        </span>
      ),
    },
    { key: 'description', header: 'Keterangan' },
    {
      key: 'isActive',
      header: 'Aktif',
      render: (r: GamificationRule) => (
        <Badge variant={r.isActive ? 'success' : 'gray'} size="sm">
          {r.isActive ? 'Aktif' : 'Tidak Aktif'}
        </Badge>
      ),
    },
  ];

  const leaderboardColumns = [
    {
      key: 'rank',
      header: '#',
      render: (_: LeaderboardEntry, index?: number) => {
        const rank = (index ?? 0) + 1;
        if (rank === 1) return <span className="text-lg">🥇</span>;
        if (rank === 2) return <span className="text-lg">🥈</span>;
        if (rank === 3) return <span className="text-lg">🥉</span>;
        return <span className="text-sm font-medium text-gray-500">{rank}</span>;
      },
    },
    {
      key: 'studentName',
      header: 'Nama Siswa',
      render: (entry: LeaderboardEntry) => (
        <p className="font-medium text-gray-900">{entry.studentName || '—'}</p>
      ),
    },
    {
      key: 'studentNisn',
      header: 'NISN',
      render: (entry: LeaderboardEntry) => (
        <span className="font-mono text-sm text-gray-500">{entry.studentNisn || '—'}</span>
      ),
    },
    {
      key: 'totalPoints',
      header: 'Total Poin',
      render: (entry: LeaderboardEntry) => (
        <span className="font-bold text-primary-600">{entry.totalPoints}</span>
      ),
    },
    {
      key: 'level',
      header: 'Level',
      render: (entry: LeaderboardEntry) => {
        const level = Math.floor(entry.totalPoints / 100) + 1;
        return (
          <Badge variant="primary" size="sm">
            Level {level}
          </Badge>
        );
      },
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="p-2 bg-amber-100 rounded-lg">
            <Trophy className="w-5 h-5 text-amber-600" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-gray-900">Gamifikasi & Loyalitas</h1>
            <p className="text-sm text-gray-500 mt-0.5">Aturan poin dan papan peringkat siswa</p>
          </div>
        </div>
        {activeTab === 'rules' && (
          <Button
            size="sm"
            leftIcon={<Plus className="w-3.5 h-3.5" />}
            onClick={() => setIsAddOpen(true)}
          >
            Tambah Aturan
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

      {/* Rules */}
      {activeTab === 'rules' && (
        <Card padding="none">
          <Table
            columns={ruleColumns}
            data={rules}
            keyExtractor={(r) => r.id}
            isLoading={rulesLoading}
            emptyMessage="Belum ada aturan poin yang dikonfigurasi"
          />
        </Card>
      )}

      {/* Leaderboard */}
      {activeTab === 'leaderboard' && (
        <Card padding="none">
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  {leaderboardColumns.map((col) => (
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
                {lbLoading ? (
                  <tr>
                    <td colSpan={leaderboardColumns.length} className="px-4 py-12 text-center">
                      <div className="flex items-center justify-center gap-2 text-gray-400">
                        <div className="w-5 h-5 border-2 border-primary-500 border-t-transparent rounded-full animate-spin" />
                        <span className="text-sm">Memuat data...</span>
                      </div>
                    </td>
                  </tr>
                ) : leaderboard.length === 0 ? (
                  <tr>
                    <td
                      colSpan={leaderboardColumns.length}
                      className="px-4 py-12 text-center text-sm text-gray-400"
                    >
                      Belum ada data papan peringkat
                    </td>
                  </tr>
                ) : (
                  leaderboard.map((entry, index) => (
                    <tr
                      key={entry.id}
                      className={`hover:bg-gray-50 transition-colors ${
                        index < 3 ? 'bg-amber-50/30' : ''
                      }`}
                    >
                      {leaderboardColumns.map((col) => (
                        <td key={col.key} className="px-4 py-3 text-sm text-gray-700">
                          {col.render(entry, index)}
                        </td>
                      ))}
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </Card>
      )}

      {/* Add Rule Modal */}
      <Modal
        isOpen={isAddOpen}
        onClose={() => { setIsAddOpen(false); reset(); }}
        title="Tambah Aturan Poin"
        size="md"
        footer={
          <>
            <Button variant="secondary" onClick={() => { setIsAddOpen(false); reset(); }}>
              Batal
            </Button>
            <Button form="rule-form" type="submit" isLoading={isSubmitting}>
              Simpan
            </Button>
          </>
        }
      >
        <form id="rule-form" onSubmit={handleSubmit(onAddRule)} className="space-y-4">
          <Select
            label="Tipe Aktivitas"
            required
            options={Object.entries(TYPE_LABELS).map(([value, label]) => ({ value, label }))}
            {...register('type')}
            error={errors.type?.message}
          />
          <Input
            label="Poin"
            type="number"
            min={1}
            required
            placeholder="Contoh: 10"
            {...register('points', { valueAsNumber: true })}
            error={errors.points?.message}
          />
          <Input
            label="Keterangan"
            placeholder="Contoh: Hadir tepat waktu selama seminggu penuh"
            required
            {...register('description')}
            error={errors.description?.message}
          />
        </form>
      </Modal>
    </div>
  );
}
