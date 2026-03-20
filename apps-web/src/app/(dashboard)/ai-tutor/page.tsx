'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Bot, Plus, Search } from 'lucide-react';
import api from '@/lib/api';
import { Card, CardHeader } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { Modal } from '@/components/ui/Modal';
import { Input } from '@/components/ui/Input';
import { formatDateTime } from '@/lib/utils';

interface TutorSession {
  id: string;
  studentId: string;
  studentName: string;
  subject: string;
  question: string;
  answer: string | null;
  createdAt: string;
}

interface PlagiarismResult {
  score: number;
  status: 'ORIGINAL' | 'SUSPICIOUS' | 'PLAGIARIZED';
  details: string;
}

interface SessionForm {
  studentId: string;
  subject: string;
  question: string;
}

interface PlagiarismForm {
  text: string;
  studentId: string;
}

const EMPTY_SESSION_FORM: SessionForm = { studentId: '', subject: '', question: '' };
const EMPTY_PLAGIARISM_FORM: PlagiarismForm = { text: '', studentId: '' };

const STATUS_VARIANT: Record<string, 'success' | 'warning' | 'danger'> = {
  ORIGINAL: 'success',
  SUSPICIOUS: 'warning',
  PLAGIARIZED: 'danger',
};

const STATUS_LABEL: Record<string, string> = {
  ORIGINAL: 'Original',
  SUSPICIOUS: 'Mencurigakan',
  PLAGIARIZED: 'Plagiat',
};

export default function AiTutorPage() {
  const [sessions, setSessions] = useState<TutorSession[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isSessionOpen, setIsSessionOpen] = useState(false);
  const [isPlagiarismOpen, setIsPlagiarismOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [sessionForm, setSessionForm] = useState<SessionForm>(EMPTY_SESSION_FORM);
  const [plagiarismForm, setPlagiarismForm] = useState<PlagiarismForm>(EMPTY_PLAGIARISM_FORM);
  const [plagiarismResult, setPlagiarismResult] = useState<PlagiarismResult | null>(null);

  const fetchSessions = useCallback(async () => {
    setIsLoading(true);
    try {
      const res = await api.get('/ai-tutor/sessions');
      setSessions(res.data.data || []);
    } catch {
      setSessions([]);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchSessions();
  }, [fetchSessions]);

  const handleAddSession = async () => {
    setIsSubmitting(true);
    try {
      await api.post('/ai-tutor/sessions', sessionForm);
      setIsSessionOpen(false);
      setSessionForm(EMPTY_SESSION_FORM);
      fetchSessions();
    } catch {
      // silent
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleCheckPlagiarism = async () => {
    setIsSubmitting(true);
    try {
      const res = await api.post('/ai-tutor/check-plagiarism', plagiarismForm);
      setPlagiarismResult(res.data.data || null);
    } catch {
      setPlagiarismResult(null);
    } finally {
      setIsSubmitting(false);
    }
  };

  const sessionColumns = [
    {
      key: 'student',
      header: 'Siswa',
      render: (s: TutorSession) => (
        <div>
          <p className="font-medium text-gray-900">{s.studentName}</p>
          <p className="text-xs text-gray-400">{s.studentId}</p>
        </div>
      ),
    },
    {
      key: 'subject',
      header: 'Mata Pelajaran',
      render: (s: TutorSession) => (
        <Badge variant="info" size="sm">{s.subject}</Badge>
      ),
    },
    {
      key: 'question',
      header: 'Pertanyaan',
      render: (s: TutorSession) => (
        <p className="text-sm text-gray-600 max-w-xs truncate">{s.question}</p>
      ),
    },
    {
      key: 'status',
      header: 'Jawaban',
      render: (s: TutorSession) => (
        <Badge variant={s.answer ? 'success' : 'warning'} size="sm">
          {s.answer ? 'Sudah Dijawab' : 'Menunggu'}
        </Badge>
      ),
    },
    {
      key: 'createdAt',
      header: 'Waktu',
      render: (s: TutorSession) => (
        <span className="text-xs text-gray-400">{formatDateTime(s.createdAt)}</span>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="p-2 bg-purple-100 rounded-xl">
            <Bot className="w-5 h-5 text-purple-600" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-gray-900">AI Tutor & Anti-Plagiarisme</h1>
            <p className="text-sm text-gray-500 mt-0.5">Sesi tanya-jawab AI dan pemeriksaan orisinalitas teks</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="secondary"
            size="sm"
            leftIcon={<Search className="w-3.5 h-3.5" />}
            onClick={() => { setPlagiarismResult(null); setIsPlagiarismOpen(true); }}
          >
            Cek Plagiarisme
          </Button>
          <Button
            size="sm"
            leftIcon={<Plus className="w-3.5 h-3.5" />}
            onClick={() => setIsSessionOpen(true)}
          >
            Sesi Baru
          </Button>
        </div>
      </div>

      {/* Sessions Table */}
      <Card padding="none">
        <div className="p-4 border-b border-gray-100">
          <CardHeader
            title="Sesi AI Tutor"
            description={`${sessions.length} sesi terdaftar`}
          />
        </div>
        <Table
          columns={sessionColumns}
          data={sessions}
          keyExtractor={(s) => s.id}
          isLoading={isLoading}
          emptyMessage="Belum ada sesi AI Tutor"
        />
      </Card>

      {/* New Session Modal */}
      <Modal
        isOpen={isSessionOpen}
        onClose={() => { setIsSessionOpen(false); setSessionForm(EMPTY_SESSION_FORM); }}
        title="Sesi Tanya-Jawab AI Tutor"
        description="Ajukan pertanyaan kepada AI Tutor untuk siswa"
        footer={
          <>
            <Button variant="secondary" onClick={() => { setIsSessionOpen(false); setSessionForm(EMPTY_SESSION_FORM); }}>
              Batal
            </Button>
            <Button onClick={handleAddSession} isLoading={isSubmitting}>
              Kirim Pertanyaan
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          <Input
            label="ID Siswa"
            required
            placeholder="Contoh: STD-2024-001"
            value={sessionForm.studentId}
            onChange={(e) => setSessionForm({ ...sessionForm, studentId: e.target.value })}
          />
          <Input
            label="Mata Pelajaran"
            required
            placeholder="Contoh: Matematika, Fisika, Bahasa Indonesia"
            value={sessionForm.subject}
            onChange={(e) => setSessionForm({ ...sessionForm, subject: e.target.value })}
          />
          <Input
            label="Pertanyaan"
            required
            placeholder="Tuliskan pertanyaan siswa di sini..."
            value={sessionForm.question}
            onChange={(e) => setSessionForm({ ...sessionForm, question: e.target.value })}
          />
        </div>
      </Modal>

      {/* Plagiarism Check Modal */}
      <Modal
        isOpen={isPlagiarismOpen}
        onClose={() => { setIsPlagiarismOpen(false); setPlagiarismForm(EMPTY_PLAGIARISM_FORM); setPlagiarismResult(null); }}
        title="Cek Plagiarisme"
        description="Periksa orisinalitas teks karya siswa"
        footer={
          <>
            <Button variant="secondary" onClick={() => { setIsPlagiarismOpen(false); setPlagiarismForm(EMPTY_PLAGIARISM_FORM); setPlagiarismResult(null); }}>
              Tutup
            </Button>
            <Button onClick={handleCheckPlagiarism} isLoading={isSubmitting}>
              Periksa
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          <Input
            label="ID Siswa"
            required
            placeholder="Contoh: STD-2024-001"
            value={plagiarismForm.studentId}
            onChange={(e) => setPlagiarismForm({ ...plagiarismForm, studentId: e.target.value })}
          />
          <Input
            label="Teks yang Diperiksa"
            required
            placeholder="Tempelkan teks karya siswa di sini..."
            value={plagiarismForm.text}
            onChange={(e) => setPlagiarismForm({ ...plagiarismForm, text: e.target.value })}
          />
          {plagiarismResult && (
            <div className="p-4 rounded-lg bg-gray-50 border border-gray-200 space-y-2">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium text-gray-700">Hasil Pemeriksaan</span>
                <Badge variant={STATUS_VARIANT[plagiarismResult.status]} size="sm">
                  {STATUS_LABEL[plagiarismResult.status]}
                </Badge>
              </div>
              <p className="text-sm text-gray-500">
                Skor Kemiripan: <span className="font-semibold text-gray-900">{plagiarismResult.score}%</span>
              </p>
              <p className="text-xs text-gray-400">{plagiarismResult.details}</p>
            </div>
          )}
        </div>
      </Modal>
    </div>
  );
}
