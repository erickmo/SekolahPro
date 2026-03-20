'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Heart } from 'lucide-react';
import api from '@/lib/api';
import { Card, CardHeader } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { Modal } from '@/components/ui/Modal';
import { Input } from '@/components/ui/Input';

interface SentimentResult {
  id: string;
  text: string;
  sourceType: string;
  sentiment: 'POSITIVE' | 'NEGATIVE' | 'NEUTRAL';
  score: number;
  analyzedAt: string;
}

const SENTIMENT_MAP: Record<string, { label: string; variant: 'success' | 'danger' | 'default' }> = {
  POSITIVE: { label: 'Positif', variant: 'success' },
  NEGATIVE: { label: 'Negatif', variant: 'danger' },
  NEUTRAL: { label: 'Netral', variant: 'default' },
};

const SOURCE_TYPE_LABELS: Record<string, string> = {
  FORUM: 'Forum',
  FEEDBACK: 'Umpan Balik',
  SURVEY: 'Survei',
  CHAT: 'Pesan',
  OTHER: 'Lainnya',
};

export default function SentimentPage() {
  const [reports, setReports] = useState<SentimentResult[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [formText, setFormText] = useState('');
  const [formSourceType, setFormSourceType] = useState('FEEDBACK');

  const fetchReports = useCallback(async () => {
    setIsLoading(true);
    try {
      const res = await api.get('/sentiment/reports');
      setReports(res.data.data || []);
    } catch {
      setReports([]);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchReports();
  }, [fetchReports]);

  const handleAnalyze = async () => {
    if (!formText.trim()) return;
    setIsSubmitting(true);
    try {
      await api.post('/sentiment/analyze', { text: formText, sourceType: formSourceType });
      setIsModalOpen(false);
      setFormText('');
      setFormSourceType('FEEDBACK');
      fetchReports();
    } catch {
      // handle error silently
    } finally {
      setIsSubmitting(false);
    }
  };

  const positiveCount = reports.filter((r) => r.sentiment === 'POSITIVE').length;
  const negativeCount = reports.filter((r) => r.sentiment === 'NEGATIVE').length;
  const neutralCount = reports.filter((r) => r.sentiment === 'NEUTRAL').length;

  const columns = [
    {
      key: 'text',
      header: 'Teks',
      render: (r: SentimentResult) => (
        <p className="text-sm text-gray-700 max-w-sm line-clamp-2">{r.text}</p>
      ),
    },
    {
      key: 'sourceType',
      header: 'Sumber',
      render: (r: SentimentResult) => (
        <span className="text-sm text-gray-600">{SOURCE_TYPE_LABELS[r.sourceType] || r.sourceType}</span>
      ),
    },
    {
      key: 'sentiment',
      header: 'Sentimen',
      render: (r: SentimentResult) => {
        const s = SENTIMENT_MAP[r.sentiment] || { label: r.sentiment, variant: 'default' as const };
        return <Badge variant={s.variant}>{s.label}</Badge>;
      },
    },
    {
      key: 'score',
      header: 'Skor',
      render: (r: SentimentResult) => (
        <span className="text-sm font-medium text-gray-900">{(r.score * 100).toFixed(1)}%</span>
      ),
    },
    {
      key: 'analyzedAt',
      header: 'Waktu Analisis',
      render: (r: SentimentResult) => (
        <span className="text-xs text-gray-400">{new Date(r.analyzedAt).toLocaleDateString('id-ID')}</span>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900 flex items-center gap-2">
            <Heart className="w-5 h-5 text-pink-500" />
            Analisis Sentimen
          </h1>
          <p className="text-sm text-gray-500 mt-0.5">
            Analisis sentimen teks dari forum, umpan balik, dan survei sekolah
          </p>
        </div>
        <Button onClick={() => setIsModalOpen(true)}>Analisis Teks Baru</Button>
      </div>

      <div className="grid grid-cols-3 gap-4">
        <Card>
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-full bg-green-100 flex items-center justify-center">
              <Heart className="w-4 h-4 text-green-600" />
            </div>
            <div>
              <p className="text-xs text-gray-500">Positif</p>
              <p className="text-xl font-bold text-gray-900">{positiveCount}</p>
            </div>
          </div>
        </Card>
        <Card>
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-full bg-red-100 flex items-center justify-center">
              <Heart className="w-4 h-4 text-red-600" />
            </div>
            <div>
              <p className="text-xs text-gray-500">Negatif</p>
              <p className="text-xl font-bold text-gray-900">{negativeCount}</p>
            </div>
          </div>
        </Card>
        <Card>
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-full bg-gray-100 flex items-center justify-center">
              <Heart className="w-4 h-4 text-gray-500" />
            </div>
            <div>
              <p className="text-xs text-gray-500">Netral</p>
              <p className="text-xl font-bold text-gray-900">{neutralCount}</p>
            </div>
          </div>
        </Card>
      </div>

      <Card padding="none">
        <div className="p-4 border-b border-gray-100">
          <CardHeader
            title="Riwayat Analisis Sentimen"
            description="Hasil analisis sentimen dari berbagai sumber teks"
          />
        </div>
        <Table
          columns={columns}
          data={reports}
          keyExtractor={(r) => r.id}
          isLoading={isLoading}
          emptyMessage="Belum ada analisis sentimen"
        />
      </Card>

      <Modal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        title="Analisis Teks Baru"
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Jenis Sumber</label>
            <select
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
              value={formSourceType}
              onChange={(e) => setFormSourceType(e.target.value)}
            >
              {Object.entries(SOURCE_TYPE_LABELS).map(([val, label]) => (
                <option key={val} value={val}>{label}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Teks yang Dianalisis</label>
            <textarea
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 min-h-[120px]"
              placeholder="Masukkan teks untuk dianalisis sentimennya..."
              value={formText}
              onChange={(e) => setFormText(e.target.value)}
            />
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <Button variant="secondary" onClick={() => setIsModalOpen(false)}>Batal</Button>
            <Button isLoading={isSubmitting} onClick={handleAnalyze} disabled={!formText.trim()}>
              Analisis
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
