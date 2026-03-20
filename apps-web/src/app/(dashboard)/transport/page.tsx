'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Plus, Bus, MapPin, Clock } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import api from '@/lib/api';
import { Card, CardHeader } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Table } from '@/components/ui/Table';
import { Pagination } from '@/components/ui/Pagination';
import { Modal } from '@/components/ui/Modal';
import { Badge } from '@/components/ui/Badge';

interface TransportBus {
  id: string;
  plateNumber: string;
  capacity: number;
  driverName: string;
  driverPhone?: string;
  status: 'ACTIVE' | 'MAINTENANCE' | 'INACTIVE';
}

interface TransportRoute {
  id: string;
  name: string;
  origin: string;
  destination: string;
  schedule: string;
  busId?: string;
  busPlate?: string;
}

type TabKey = 'buses' | 'routes';

const BUS_STATUS_CONFIG: Record<string, { label: string; variant: 'success' | 'warning' | 'gray' }> = {
  ACTIVE: { label: 'Aktif', variant: 'success' },
  MAINTENANCE: { label: 'Perawatan', variant: 'warning' },
  INACTIVE: { label: 'Tidak Aktif', variant: 'gray' },
};

export default function TransportPage() {
  const { user } = useAuth();
  const [activeTab, setActiveTab] = useState<TabKey>('buses');
  const [buses, setBuses] = useState<TransportBus[]>([]);
  const [routes, setRoutes] = useState<TransportRoute[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [busPage, setBusPage] = useState(1);
  const [routePage, setRoutePage] = useState(1);
  const [busMeta, setBusMeta] = useState({ total: 0, totalPages: 1 });
  const [routeMeta, setRouteMeta] = useState({ total: 0, totalPages: 1 });
  const [addOpen, setAddOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Form state
  const [fPlate, setFPlate] = useState('');
  const [fCapacity, setFCapacity] = useState('');
  const [fDriver, setFDriver] = useState('');
  const [fDriverPhone, setFDriverPhone] = useState('');

  const fetchBuses = useCallback(async () => {
    setIsLoading(true);
    try {
      const params = new URLSearchParams({ page: String(busPage), limit: '20' });
      const res = await api.get(`/transport/buses?${params}`);
      setBuses(res.data.data || []);
      setBusMeta(res.data.meta || { total: 0, totalPages: 1 });
    } catch {
      setBuses([]);
    } finally {
      setIsLoading(false);
    }
  }, [busPage]);

  const fetchRoutes = useCallback(async () => {
    setIsLoading(true);
    try {
      const params = new URLSearchParams({ page: String(routePage), limit: '20' });
      const res = await api.get(`/transport/routes?${params}`);
      setRoutes(res.data.data || []);
      setRouteMeta(res.data.meta || { total: 0, totalPages: 1 });
    } catch {
      setRoutes([]);
    } finally {
      setIsLoading(false);
    }
  }, [routePage]);

  useEffect(() => { fetchBuses(); }, [fetchBuses]);
  useEffect(() => { fetchRoutes(); }, [fetchRoutes]);

  const resetForm = () => {
    setFPlate(''); setFCapacity(''); setFDriver(''); setFDriverPhone('');
  };

  const handleAddBus = async () => {
    if (!fPlate || !fDriver) return;
    setIsSubmitting(true);
    try {
      await api.post('/transport/buses', {
        plateNumber: fPlate,
        capacity: parseInt(fCapacity) || 0,
        driverName: fDriver,
        driverPhone: fDriverPhone || undefined,
      });
      setAddOpen(false);
      resetForm();
      fetchBuses();
    } finally {
      setIsSubmitting(false);
    }
  };

  const activeBuses = buses.filter((b) => b.status === 'ACTIVE').length;
  const maintenanceBuses = buses.filter((b) => b.status === 'MAINTENANCE').length;

  const busColumns = [
    {
      key: 'plateNumber',
      header: 'Plat Nomor',
      render: (b: TransportBus) => (
        <span className="font-mono font-semibold text-gray-900">{b.plateNumber}</span>
      ),
    },
    {
      key: 'capacity',
      header: 'Kapasitas',
      render: (b: TransportBus) => `${b.capacity} penumpang`,
    },
    {
      key: 'driverName',
      header: 'Pengemudi',
      render: (b: TransportBus) => (
        <div>
          <p className="text-sm font-medium text-gray-900">{b.driverName}</p>
          {b.driverPhone && <p className="text-xs text-gray-400">{b.driverPhone}</p>}
        </div>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (b: TransportBus) => {
        const cfg = BUS_STATUS_CONFIG[b.status] || { label: b.status, variant: 'gray' as const };
        return <Badge variant={cfg.variant} size="sm">{cfg.label}</Badge>;
      },
    },
  ];

  const routeColumns = [
    {
      key: 'name',
      header: 'Rute',
      render: (r: TransportRoute) => (
        <span className="font-medium text-gray-900">{r.name}</span>
      ),
    },
    {
      key: 'origin',
      header: 'Asal',
      render: (r: TransportRoute) => (
        <div className="flex items-center gap-1 text-sm text-gray-700">
          <MapPin className="w-3.5 h-3.5 text-gray-400" />
          {r.origin}
        </div>
      ),
    },
    {
      key: 'destination',
      header: 'Tujuan',
      render: (r: TransportRoute) => (
        <div className="flex items-center gap-1 text-sm text-gray-700">
          <MapPin className="w-3.5 h-3.5 text-primary-400" />
          {r.destination}
        </div>
      ),
    },
    {
      key: 'schedule',
      header: 'Jadwal',
      render: (r: TransportRoute) => (
        <div className="flex items-center gap-1 text-sm text-gray-700">
          <Clock className="w-3.5 h-3.5 text-gray-400" />
          {r.schedule}
        </div>
      ),
    },
    {
      key: 'busPlate',
      header: 'Armada',
      render: (r: TransportRoute) => r.busPlate ? (
        <span className="font-mono text-sm">{r.busPlate}</span>
      ) : <span className="text-gray-300">—</span>,
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900">Transportasi Sekolah</h1>
          <p className="text-sm text-gray-500 mt-0.5">Manajemen armada dan rute transportasi sekolah</p>
        </div>
        {activeTab === 'buses' && (
          <Button
            size="sm"
            leftIcon={<Plus className="w-3.5 h-3.5" />}
            onClick={() => setAddOpen(true)}
          >
            Tambah Armada
          </Button>
        )}
      </div>

      {/* Stats */}
      <div className="grid grid-cols-3 gap-4">
        <Card padding="sm">
          <div className="flex items-center gap-3">
            <div className="p-2.5 bg-primary-50 rounded-lg">
              <Bus className="w-5 h-5 text-primary-600" />
            </div>
            <div>
              <p className="text-xl font-bold text-gray-900">{busMeta.total}</p>
              <p className="text-xs text-gray-500">Total Armada</p>
            </div>
          </div>
        </Card>
        <Card padding="sm">
          <div className="flex items-center gap-3">
            <div className="p-2.5 bg-green-50 rounded-lg">
              <Bus className="w-5 h-5 text-green-600" />
            </div>
            <div>
              <p className="text-xl font-bold text-gray-900">{activeBuses}</p>
              <p className="text-xs text-gray-500">Armada Aktif</p>
            </div>
          </div>
        </Card>
        <Card padding="sm">
          <div className="flex items-center gap-3">
            <div className="p-2.5 bg-blue-50 rounded-lg">
              <MapPin className="w-5 h-5 text-blue-600" />
            </div>
            <div>
              <p className="text-xl font-bold text-gray-900">{routeMeta.total}</p>
              <p className="text-xs text-gray-500">Total Rute</p>
            </div>
          </div>
        </Card>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 bg-gray-100 p-1 rounded-xl w-fit">
        {([
          { key: 'buses', label: 'Armada' },
          { key: 'routes', label: 'Rute' },
        ] as { key: TabKey; label: string }[]).map((t) => (
          <button
            key={t.key}
            onClick={() => setActiveTab(t.key)}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
              activeTab === t.key ? 'bg-white text-primary-700 shadow-sm' : 'text-gray-500'
            }`}
          >
            {t.label}
          </button>
        ))}
      </div>

      {activeTab === 'buses' && (
        <Card padding="none">
          <Table
            columns={busColumns}
            data={buses}
            keyExtractor={(b) => b.id}
            isLoading={isLoading}
            emptyMessage="Belum ada data armada transportasi"
          />
          {busMeta.total > 0 && (
            <div className="px-4 py-3 border-t border-gray-100">
              <Pagination
                currentPage={busPage}
                totalPages={busMeta.totalPages}
                total={busMeta.total}
                limit={20}
                onPageChange={setBusPage}
              />
            </div>
          )}
        </Card>
      )}

      {activeTab === 'routes' && (
        <Card padding="none">
          <Table
            columns={routeColumns}
            data={routes}
            keyExtractor={(r) => r.id}
            isLoading={isLoading}
            emptyMessage="Belum ada data rute transportasi"
          />
          {routeMeta.total > 0 && (
            <div className="px-4 py-3 border-t border-gray-100">
              <Pagination
                currentPage={routePage}
                totalPages={routeMeta.totalPages}
                total={routeMeta.total}
                limit={20}
                onPageChange={setRoutePage}
              />
            </div>
          )}
        </Card>
      )}

      {/* Add Bus Modal */}
      <Modal
        isOpen={addOpen}
        onClose={() => { setAddOpen(false); resetForm(); }}
        title="Tambah Armada Transportasi"
        size="sm"
        footer={
          <>
            <Button variant="secondary" onClick={() => { setAddOpen(false); resetForm(); }}>Batal</Button>
            <Button onClick={handleAddBus} isLoading={isSubmitting}>Simpan</Button>
          </>
        }
      >
        <div className="space-y-4">
          <Input
            label="Plat Nomor"
            placeholder="Contoh: B 1234 ABC"
            required
            value={fPlate}
            onChange={(e) => setFPlate(e.target.value)}
          />
          <Input
            label="Kapasitas (penumpang)"
            type="number"
            placeholder="Contoh: 30"
            value={fCapacity}
            onChange={(e) => setFCapacity(e.target.value)}
          />
          <Input
            label="Nama Pengemudi"
            placeholder="Nama lengkap pengemudi"
            required
            value={fDriver}
            onChange={(e) => setFDriver(e.target.value)}
          />
          <Input
            label="Nomor Telepon Pengemudi (opsional)"
            placeholder="Contoh: 08123456789"
            value={fDriverPhone}
            onChange={(e) => setFDriverPhone(e.target.value)}
          />
        </div>
      </Modal>
    </div>
  );
}
