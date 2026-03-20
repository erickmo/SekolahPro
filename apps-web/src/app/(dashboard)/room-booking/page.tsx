'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Plus, DoorOpen, CalendarCheck, Clock } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import api from '@/lib/api';
import { Card, CardHeader } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Table } from '@/components/ui/Table';
import { Pagination } from '@/components/ui/Pagination';
import { Modal } from '@/components/ui/Modal';
import { Badge } from '@/components/ui/Badge';
import { formatDate } from '@/lib/utils';

interface Room {
  id: string;
  name: string;
  type: string;
  capacity: number;
  facilities?: string;
  isAvailable: boolean;
}

interface RoomBooking {
  id: string;
  requesterName?: string;
  requesterId: string;
  roomId: string;
  roomName?: string;
  date: string;
  startTime: string;
  endTime: string;
  purpose: string;
  status: 'PENDING' | 'APPROVED' | 'REJECTED';
}

type TabKey = 'rooms' | 'bookings';

const BOOKING_STATUS_CONFIG: Record<string, { label: string; variant: 'warning' | 'success' | 'danger' }> = {
  PENDING: { label: 'Menunggu', variant: 'warning' },
  APPROVED: { label: 'Disetujui', variant: 'success' },
  REJECTED: { label: 'Ditolak', variant: 'danger' },
};

export default function RoomBookingPage() {
  const { user } = useAuth();
  const [activeTab, setActiveTab] = useState<TabKey>('rooms');
  const [rooms, setRooms] = useState<Room[]>([]);
  const [bookings, setBookings] = useState<RoomBooking[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [roomPage, setRoomPage] = useState(1);
  const [bookingPage, setBookingPage] = useState(1);
  const [roomMeta, setRoomMeta] = useState({ total: 0, totalPages: 1 });
  const [bookingMeta, setBookingMeta] = useState({ total: 0, totalPages: 1 });
  const [addOpen, setAddOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Form state
  const [fRoomId, setFRoomId] = useState('');
  const [fDate, setFDate] = useState('');
  const [fStartTime, setFStartTime] = useState('');
  const [fEndTime, setFEndTime] = useState('');
  const [fPurpose, setFPurpose] = useState('');

  const fetchRooms = useCallback(async () => {
    setIsLoading(true);
    try {
      const params = new URLSearchParams({ page: String(roomPage), limit: '20' });
      const res = await api.get(`/room-booking/rooms?${params}`);
      setRooms(res.data.data || []);
      setRoomMeta(res.data.meta || { total: 0, totalPages: 1 });
    } catch {
      setRooms([]);
    } finally {
      setIsLoading(false);
    }
  }, [roomPage]);

  const fetchBookings = useCallback(async () => {
    setIsLoading(true);
    try {
      const params = new URLSearchParams({ page: String(bookingPage), limit: '20' });
      const res = await api.get(`/room-booking/bookings?${params}`);
      setBookings(res.data.data || []);
      setBookingMeta(res.data.meta || { total: 0, totalPages: 1 });
    } catch {
      setBookings([]);
    } finally {
      setIsLoading(false);
    }
  }, [bookingPage]);

  useEffect(() => { fetchRooms(); }, [fetchRooms]);
  useEffect(() => { fetchBookings(); }, [fetchBookings]);

  const resetForm = () => {
    setFRoomId(''); setFDate(''); setFStartTime(''); setFEndTime(''); setFPurpose('');
  };

  const handleAddBooking = async () => {
    if (!fRoomId || !fDate || !fStartTime || !fEndTime || !fPurpose) return;
    setIsSubmitting(true);
    try {
      await api.post('/room-booking/bookings', {
        roomId: fRoomId,
        date: fDate,
        startTime: fStartTime,
        endTime: fEndTime,
        purpose: fPurpose,
      });
      setAddOpen(false);
      resetForm();
      fetchBookings();
    } finally {
      setIsSubmitting(false);
    }
  };

  const availableRooms = rooms.filter((r) => r.isAvailable).length;
  const pendingBookings = bookings.filter((b) => b.status === 'PENDING').length;

  const roomColumns = [
    {
      key: 'name',
      header: 'Nama Ruangan',
      render: (r: Room) => <span className="font-medium text-gray-900">{r.name}</span>,
    },
    {
      key: 'type',
      header: 'Tipe',
      render: (r: Room) => r.type || <span className="text-gray-300">—</span>,
    },
    {
      key: 'capacity',
      header: 'Kapasitas',
      render: (r: Room) => `${r.capacity} orang`,
    },
    {
      key: 'facilities',
      header: 'Fasilitas',
      render: (r: Room) => (
        <p className="max-w-xs text-sm text-gray-600 truncate">
          {r.facilities || <span className="text-gray-300">—</span>}
        </p>
      ),
    },
    {
      key: 'isAvailable',
      header: 'Status',
      render: (r: Room) => (
        <Badge variant={r.isAvailable ? 'success' : 'gray'} size="sm">
          {r.isAvailable ? 'Tersedia' : 'Tidak Tersedia'}
        </Badge>
      ),
    },
  ];

  const bookingColumns = [
    {
      key: 'requester',
      header: 'Pemohon',
      render: (b: RoomBooking) => b.requesterName || b.requesterId,
    },
    {
      key: 'roomName',
      header: 'Ruangan',
      render: (b: RoomBooking) => (
        <span className="font-medium text-gray-900">{b.roomName || b.roomId}</span>
      ),
    },
    {
      key: 'date',
      header: 'Tanggal',
      render: (b: RoomBooking) => formatDate(b.date),
    },
    {
      key: 'time',
      header: 'Waktu',
      render: (b: RoomBooking) => (
        <div className="flex items-center gap-1 text-sm text-gray-700">
          <Clock className="w-3.5 h-3.5 text-gray-400" />
          {b.startTime} — {b.endTime}
        </div>
      ),
    },
    {
      key: 'purpose',
      header: 'Keperluan',
      render: (b: RoomBooking) => (
        <p className="max-w-xs text-sm text-gray-700 truncate">{b.purpose}</p>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (b: RoomBooking) => {
        const cfg = BOOKING_STATUS_CONFIG[b.status] || { label: b.status, variant: 'gray' as const };
        return <Badge variant={cfg.variant} size="sm">{cfg.label}</Badge>;
      },
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900">Booking Ruang & Fasilitas</h1>
          <p className="text-sm text-gray-500 mt-0.5">Manajemen peminjaman ruangan dan fasilitas sekolah</p>
        </div>
        <Button
          size="sm"
          leftIcon={<Plus className="w-3.5 h-3.5" />}
          onClick={() => setAddOpen(true)}
        >
          Buat Pemesanan
        </Button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-3 gap-4">
        <Card padding="sm">
          <div className="flex items-center gap-3">
            <div className="p-2.5 bg-primary-50 rounded-lg">
              <DoorOpen className="w-5 h-5 text-primary-600" />
            </div>
            <div>
              <p className="text-xl font-bold text-gray-900">{roomMeta.total}</p>
              <p className="text-xs text-gray-500">Total Ruangan</p>
            </div>
          </div>
        </Card>
        <Card padding="sm">
          <div className="flex items-center gap-3">
            <div className="p-2.5 bg-green-50 rounded-lg">
              <DoorOpen className="w-5 h-5 text-green-600" />
            </div>
            <div>
              <p className="text-xl font-bold text-gray-900">{availableRooms}</p>
              <p className="text-xs text-gray-500">Ruangan Tersedia</p>
            </div>
          </div>
        </Card>
        <Card padding="sm">
          <div className="flex items-center gap-3">
            <div className="p-2.5 bg-amber-50 rounded-lg">
              <CalendarCheck className="w-5 h-5 text-amber-600" />
            </div>
            <div>
              <p className="text-xl font-bold text-gray-900">{pendingBookings}</p>
              <p className="text-xs text-gray-500">Menunggu Persetujuan</p>
            </div>
          </div>
        </Card>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 bg-gray-100 p-1 rounded-xl w-fit">
        {([
          { key: 'rooms', label: 'Daftar Ruangan' },
          { key: 'bookings', label: 'Pemesanan' },
        ] as { key: TabKey; label: string }[]).map((t) => (
          <button
            key={t.key}
            onClick={() => setActiveTab(t.key)}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
              activeTab === t.key ? 'bg-white text-primary-700 shadow-sm' : 'text-gray-500'
            }`}
          >
            {t.label}
            {t.key === 'bookings' && pendingBookings > 0 && (
              <span className="ml-1.5 inline-flex items-center justify-center px-1.5 py-0.5 rounded-full text-xs font-bold bg-amber-500 text-white">
                {pendingBookings}
              </span>
            )}
          </button>
        ))}
      </div>

      {activeTab === 'rooms' && (
        <Card padding="none">
          <Table
            columns={roomColumns}
            data={rooms}
            keyExtractor={(r) => r.id}
            isLoading={isLoading}
            emptyMessage="Belum ada data ruangan"
          />
          {roomMeta.total > 0 && (
            <div className="px-4 py-3 border-t border-gray-100">
              <Pagination
                currentPage={roomPage}
                totalPages={roomMeta.totalPages}
                total={roomMeta.total}
                limit={20}
                onPageChange={setRoomPage}
              />
            </div>
          )}
        </Card>
      )}

      {activeTab === 'bookings' && (
        <Card padding="none">
          <Table
            columns={bookingColumns}
            data={bookings}
            keyExtractor={(b) => b.id}
            isLoading={isLoading}
            emptyMessage="Belum ada pemesanan ruangan"
          />
          {bookingMeta.total > 0 && (
            <div className="px-4 py-3 border-t border-gray-100">
              <Pagination
                currentPage={bookingPage}
                totalPages={bookingMeta.totalPages}
                total={bookingMeta.total}
                limit={20}
                onPageChange={setBookingPage}
              />
            </div>
          )}
        </Card>
      )}

      {/* Add Booking Modal */}
      <Modal
        isOpen={addOpen}
        onClose={() => { setAddOpen(false); resetForm(); }}
        title="Buat Pemesanan Ruangan"
        size="sm"
        footer={
          <>
            <Button variant="secondary" onClick={() => { setAddOpen(false); resetForm(); }}>Batal</Button>
            <Button onClick={handleAddBooking} isLoading={isSubmitting}>Simpan</Button>
          </>
        }
      >
        <div className="space-y-4">
          <Input
            label="ID Ruangan"
            placeholder="Masukkan ID atau nama ruangan"
            required
            value={fRoomId}
            onChange={(e) => setFRoomId(e.target.value)}
          />
          <Input
            label="Tanggal"
            type="date"
            required
            value={fDate}
            onChange={(e) => setFDate(e.target.value)}
          />
          <div className="grid grid-cols-2 gap-4">
            <Input
              label="Waktu Mulai"
              type="time"
              required
              value={fStartTime}
              onChange={(e) => setFStartTime(e.target.value)}
            />
            <Input
              label="Waktu Selesai"
              type="time"
              required
              value={fEndTime}
              onChange={(e) => setFEndTime(e.target.value)}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Keperluan <span className="text-red-500">*</span>
            </label>
            <textarea
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
              rows={3}
              placeholder="Jelaskan keperluan pemesanan ruangan..."
              value={fPurpose}
              onChange={(e) => setFPurpose(e.target.value)}
            />
          </div>
        </div>
      </Modal>
    </div>
  );
}
