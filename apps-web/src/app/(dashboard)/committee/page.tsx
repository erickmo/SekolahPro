'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Users, Plus, RefreshCw, ChevronRight, Calendar } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import api from '@/lib/api';
import { Card, CardHeader } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { Modal } from '@/components/ui/Modal';
import { Input } from '@/components/ui/Input';
import { formatDate } from '@/lib/utils';

interface Meeting {
  id: string;
  title: string;
  date: string;
  status: 'SCHEDULED' | 'COMPLETED' | 'CANCELLED';
  decisions: string[];
}

interface CommitteeBoard {
  id: string;
  periodStart: string;
  periodEnd: string;
  memberCount: number;
  meetingCount: number;
  status: 'ACTIVE' | 'INACTIVE';
  meetings?: Meeting[];
}

const MEETING_STATUS_MAP: Record<string, { label: string; variant: 'primary' | 'success' | 'gray' }> = {
  SCHEDULED: { label: 'Terjadwal', variant: 'primary' },
  COMPLETED: { label: 'Selesai', variant: 'success' },
  CANCELLED: { label: 'Dibatalkan', variant: 'gray' },
};

const EMPTY_FORM = { periodStart: '', periodEnd: '' };

export default function CommitteePage() {
  const { user } = useAuth();
  const [boards, setBoards] = useState<CommitteeBoard[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isAddOpen, setIsAddOpen] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [selectedBoard, setSelectedBoard] = useState<CommitteeBoard | null>(null);
  const [isDetailOpen, setIsDetailOpen] = useState(false);
  const [isLoadingDetail, setIsLoadingDetail] = useState(false);
  const [form, setForm] = useState(EMPTY_FORM);
  const [errors, setErrors] = useState<Partial<typeof EMPTY_FORM>>({});

  const canManage = ['ADMIN_SEKOLAH', 'KEPALA_SEKOLAH', 'TATA_USAHA'].includes(user?.role || '');

  const fetchData = useCallback(async () => {
    setIsLoading(true);
    try {
      const res = await api.get('/committee/boards');
      setBoards(res.data.data || []);
    } catch {
      setBoards([]);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => { fetchData(); }, [fetchData]);

  const handleBoardClick = async (board: CommitteeBoard) => {
    setSelectedBoard(board);
    setIsDetailOpen(true);
    setIsLoadingDetail(true);
    try {
      const res = await api.get(`/committee/boards/${board.id}/meetings`);
      setSelectedBoard((prev) => prev ? { ...prev, meetings: res.data.data || [] } : null);
    } catch {
      // handle
    } finally {
      setIsLoadingDetail(false);
    }
  };

  const validate = () => {
    const e: Partial<typeof EMPTY_FORM> = {};
    if (!form.periodStart) e.periodStart = 'Tanggal mulai periode wajib diisi';
    if (!form.periodEnd) e.periodEnd = 'Tanggal selesai periode wajib diisi';
    setErrors(e);
    return Object.keys(e).length === 0;
  };

  const handleSave = async () => {
    if (!validate()) return;
    setIsSaving(true);
    try {
      await api.post('/committee/boards', form);
      setIsAddOpen(false);
      setForm(EMPTY_FORM);
      setErrors({});
      fetchData();
    } catch {
      // handle
    } finally {
      setIsSaving(false);
    }
  };

  const handleClose = () => {
    setIsAddOpen(false);
    setForm(EMPTY_FORM);
    setErrors({});
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900">Komite Sekolah Digital</h1>
          <p className="text-sm text-gray-500 mt-0.5">
            Kelola kepengurusan komite sekolah, rapat, dan keputusan
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="secondary"
            size="sm"
            leftIcon={<RefreshCw className="w-3.5 h-3.5" />}
            onClick={fetchData}
          >
            Refresh
          </Button>
          {canManage && (
            <Button
              size="sm"
              leftIcon={<Plus className="w-3.5 h-3.5" />}
              onClick={() => setIsAddOpen(true)}
            >
              Tambah Kepengurusan
            </Button>
          )}
        </div>
      </div>

      {/* Boards Grid */}
      {isLoading ? (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {[1, 2, 3].map((i) => (
            <div key={i} className="h-36 bg-gray-100 rounded-xl animate-pulse" />
          ))}
        </div>
      ) : boards.length === 0 ? (
        <Card>
          <div className="flex flex-col items-center justify-center py-12 text-center">
            <Users className="w-10 h-10 text-gray-300 mb-3" />
            <p className="text-sm text-gray-500">Belum ada data kepengurusan komite</p>
            {canManage && (
              <Button
                size="sm"
                className="mt-4"
                leftIcon={<Plus className="w-3.5 h-3.5" />}
                onClick={() => setIsAddOpen(true)}
              >
                Tambah Kepengurusan
              </Button>
            )}
          </div>
        </Card>
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {boards.map((board) => (
            <button
              key={board.id}
              className="text-left bg-white rounded-xl shadow-card border border-gray-100 p-5 hover:shadow-card-hover hover:border-primary-200 transition-all duration-200 group"
              onClick={() => handleBoardClick(board)}
            >
              <div className="flex items-start justify-between mb-3">
                <div className="p-2.5 rounded-lg bg-primary-100">
                  <Users className="w-5 h-5 text-primary-600" />
                </div>
                <div className="flex items-center gap-2">
                  <Badge variant={board.status === 'ACTIVE' ? 'success' : 'gray'} size="sm">
                    {board.status === 'ACTIVE' ? 'Aktif' : 'Tidak Aktif'}
                  </Badge>
                  <ChevronRight className="w-4 h-4 text-gray-300 group-hover:text-primary-400 transition-colors" />
                </div>
              </div>
              <p className="text-sm font-semibold text-gray-900 mb-1">
                Periode {formatDate(board.periodStart)} — {formatDate(board.periodEnd)}
              </p>
              <div className="flex items-center gap-4 mt-3">
                <div className="flex items-center gap-1.5 text-xs text-gray-500">
                  <Users className="w-3.5 h-3.5" />
                  <span>{board.memberCount} anggota</span>
                </div>
                <div className="flex items-center gap-1.5 text-xs text-gray-500">
                  <Calendar className="w-3.5 h-3.5" />
                  <span>{board.meetingCount} rapat</span>
                </div>
              </div>
            </button>
          ))}
        </div>
      )}

      {/* Add Modal */}
      <Modal
        isOpen={isAddOpen}
        onClose={handleClose}
        title="Tambah Kepengurusan Komite"
        description="Buat periode kepengurusan komite sekolah baru"
        size="sm"
        footer={
          <>
            <Button variant="secondary" onClick={handleClose}>Batal</Button>
            <Button
              leftIcon={<Users className="w-4 h-4" />}
              isLoading={isSaving}
              onClick={handleSave}
            >
              Simpan
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          <Input
            label="Tanggal Mulai Periode"
            type="date"
            required
            value={form.periodStart}
            onChange={(e) => setForm((f) => ({ ...f, periodStart: e.target.value }))}
            error={errors.periodStart}
          />
          <Input
            label="Tanggal Selesai Periode"
            type="date"
            required
            value={form.periodEnd}
            onChange={(e) => setForm((f) => ({ ...f, periodEnd: e.target.value }))}
            error={errors.periodEnd}
          />
        </div>
      </Modal>

      {/* Detail Modal */}
      <Modal
        isOpen={isDetailOpen}
        onClose={() => { setIsDetailOpen(false); setSelectedBoard(null); }}
        title={
          selectedBoard
            ? `Komite ${formatDate(selectedBoard.periodStart)} — ${formatDate(selectedBoard.periodEnd)}`
            : 'Detail Komite'
        }
        description="Rapat dan keputusan komite sekolah"
        size="lg"
        footer={
          <Button
            variant="secondary"
            onClick={() => { setIsDetailOpen(false); setSelectedBoard(null); }}
          >
            Tutup
          </Button>
        }
      >
        {isLoadingDetail ? (
          <div className="py-8 text-center text-sm text-gray-400">Memuat data rapat...</div>
        ) : !selectedBoard?.meetings || selectedBoard.meetings.length === 0 ? (
          <div className="py-8 text-center text-sm text-gray-400">Belum ada rapat tercatat</div>
        ) : (
          <ul className="space-y-3">
            {selectedBoard.meetings.map((m) => {
              const ms = MEETING_STATUS_MAP[m.status] || { label: m.status, variant: 'default' as const };
              return (
                <li key={m.id} className="border border-gray-100 rounded-lg p-4">
                  <div className="flex items-start justify-between gap-2 mb-2">
                    <p className="text-sm font-medium text-gray-900">{m.title}</p>
                    <Badge variant={ms.variant} size="sm">{ms.label}</Badge>
                  </div>
                  <p className="text-xs text-gray-400 mb-2">{formatDate(m.date)}</p>
                  {m.decisions && m.decisions.length > 0 && (
                    <div>
                      <p className="text-xs font-medium text-gray-600 mb-1">Keputusan:</p>
                      <ul className="space-y-1">
                        {m.decisions.map((d, i) => (
                          <li key={i} className="text-xs text-gray-500 flex items-start gap-1.5">
                            <span className="mt-1 w-1 h-1 rounded-full bg-gray-400 flex-shrink-0" />
                            {d}
                          </li>
                        ))}
                      </ul>
                    </div>
                  )}
                </li>
              );
            })}
          </ul>
        )}
      </Modal>
    </div>
  );
}
