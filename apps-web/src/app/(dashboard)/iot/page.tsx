'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Wifi, Plus, RefreshCw } from 'lucide-react';
import api from '@/lib/api';
import { Card, CardHeader } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { Modal } from '@/components/ui/Modal';
import { Input, Select } from '@/components/ui/Input';
import { formatDateTime } from '@/lib/utils';
import { cn } from '@/lib/utils';

const DEVICE_TYPE_LABELS: Record<string, string> = {
  TEMPERATURE_HUMIDITY: 'Suhu & Kelembaban',
  SMART_METER_ELECTRIC: 'Smart Meter Listrik',
  SMART_METER_WATER: 'Smart Meter Air',
  MOTION: 'Sensor Gerak',
  DOOR_SENSOR: 'Sensor Pintu',
  PANIC_BUTTON: 'Tombol Panik',
  AIR_QUALITY: 'Kualitas Udara',
};

const DEVICE_TYPE_VARIANT: Record<string, 'info' | 'warning' | 'success' | 'danger' | 'primary' | 'gray'> = {
  TEMPERATURE_HUMIDITY: 'info',
  SMART_METER_ELECTRIC: 'warning',
  SMART_METER_WATER: 'primary',
  MOTION: 'success',
  DOOR_SENSOR: 'gray',
  PANIC_BUTTON: 'danger',
  AIR_QUALITY: 'info',
};

interface IoTDevice {
  id: string;
  name: string;
  type: string;
  location: string;
  mqttTopic: string;
  isActive: boolean;
  lastPing: string | null;
}

interface IoTReading {
  id: string;
  deviceId: string;
  value: number;
  unit: string;
  createdAt: string;
}

interface AddDeviceForm {
  name: string;
  type: string;
  location: string;
  mqttTopic: string;
}

const EMPTY_FORM: AddDeviceForm = {
  name: '',
  type: 'TEMPERATURE_HUMIDITY',
  location: '',
  mqttTopic: '',
};

export default function IoTPage() {
  const [devices, setDevices] = useState<IoTDevice[]>([]);
  const [readings, setReadings] = useState<IoTReading[]>([]);
  const [selectedDeviceId, setSelectedDeviceId] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isLoadingReadings, setIsLoadingReadings] = useState(false);
  const [isAddOpen, setIsAddOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [form, setForm] = useState<AddDeviceForm>(EMPTY_FORM);

  const fetchDevices = useCallback(async () => {
    setIsLoading(true);
    try {
      const res = await api.get('/iot/devices');
      const data = res.data.data || [];
      setDevices(data);
      if (!selectedDeviceId && data.length > 0) {
        setSelectedDeviceId(data[0].id);
      }
    } catch {
      setDevices([]);
    } finally {
      setIsLoading(false);
    }
  }, [selectedDeviceId]);

  const fetchReadings = useCallback(async (deviceId: string) => {
    setIsLoadingReadings(true);
    try {
      const params = new URLSearchParams({ deviceId, limit: '20' });
      const res = await api.get(`/iot/readings?${params}`);
      setReadings(res.data.data || []);
    } catch {
      setReadings([]);
    } finally {
      setIsLoadingReadings(false);
    }
  }, []);

  useEffect(() => {
    fetchDevices();
  }, [fetchDevices]);

  useEffect(() => {
    if (selectedDeviceId) {
      fetchReadings(selectedDeviceId);
    }
  }, [selectedDeviceId, fetchReadings]);

  const handleAdd = async () => {
    setIsSubmitting(true);
    try {
      await api.post('/iot/devices', form);
      setIsAddOpen(false);
      setForm(EMPTY_FORM);
      fetchDevices();
    } catch {
      // silent
    } finally {
      setIsSubmitting(false);
    }
  };

  const selectedDevice = devices.find((d) => d.id === selectedDeviceId);

  const deviceColumns = [
    {
      key: 'name',
      header: 'Nama',
      render: (d: IoTDevice) => (
        <div>
          <p className="font-medium text-gray-900">{d.name}</p>
          <p className="text-xs text-gray-400 font-mono">{d.mqttTopic}</p>
        </div>
      ),
    },
    {
      key: 'type',
      header: 'Tipe',
      render: (d: IoTDevice) => (
        <Badge variant={DEVICE_TYPE_VARIANT[d.type] ?? 'default'} size="sm">
          {DEVICE_TYPE_LABELS[d.type] ?? d.type}
        </Badge>
      ),
    },
    {
      key: 'location',
      header: 'Lokasi',
      render: (d: IoTDevice) => <span className="text-sm text-gray-600">{d.location}</span>,
    },
    {
      key: 'status',
      header: 'Status',
      render: (d: IoTDevice) => (
        <div className="flex items-center gap-1.5">
          <span className={cn('w-2 h-2 rounded-full', d.isActive ? 'bg-green-500' : 'bg-red-400')} />
          <Badge variant={d.isActive ? 'success' : 'danger'} size="sm">
            {d.isActive ? 'Aktif' : 'Nonaktif'}
          </Badge>
        </div>
      ),
    },
    {
      key: 'lastPing',
      header: 'Last Ping',
      render: (d: IoTDevice) => (
        <span className="text-xs text-gray-400">
          {d.lastPing ? formatDateTime(d.lastPing) : '—'}
        </span>
      ),
    },
    {
      key: 'actions',
      header: '',
      render: (d: IoTDevice) => (
        <Button
          variant="ghost"
          size="sm"
          onClick={(e) => { e.stopPropagation(); setSelectedDeviceId(d.id); }}
        >
          Lihat Data
        </Button>
      ),
      className: 'w-24',
    },
  ];

  const readingColumns = [
    {
      key: 'value',
      header: 'Nilai',
      render: (r: IoTReading) => (
        <span className="font-semibold text-gray-900">{r.value}</span>
      ),
    },
    {
      key: 'unit',
      header: 'Satuan',
      render: (r: IoTReading) => <span className="text-sm text-gray-500">{r.unit}</span>,
    },
    {
      key: 'createdAt',
      header: 'Waktu',
      render: (r: IoTReading) => (
        <span className="text-xs text-gray-400">{formatDateTime(r.createdAt)}</span>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="p-2 bg-blue-100 rounded-xl">
            <Wifi className="w-5 h-5 text-blue-600" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-gray-900">IoT & Smart Campus</h1>
            <p className="text-sm text-gray-500 mt-0.5">Pantau perangkat sensor dan pembacaan real-time</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          {/* Real-time indicator */}
          <div className="flex items-center gap-1.5 px-3 py-1.5 bg-green-50 rounded-lg border border-green-100">
            <span className="w-2 h-2 rounded-full bg-green-500 animate-pulse" />
            <span className="text-xs text-green-700 font-medium">Live</span>
          </div>
          <Button variant="secondary" size="sm" leftIcon={<RefreshCw className="w-3.5 h-3.5" />} onClick={fetchDevices}>
            Refresh
          </Button>
          <Button size="sm" leftIcon={<Plus className="w-3.5 h-3.5" />} onClick={() => setIsAddOpen(true)}>
            Tambah Perangkat
          </Button>
        </div>
      </div>

      {/* Devices Table */}
      <Card padding="none">
        <div className="p-4 border-b border-gray-100">
          <CardHeader title="Daftar Perangkat" description={`${devices.length} perangkat terdaftar`} />
        </div>
        <Table
          columns={deviceColumns}
          data={devices}
          keyExtractor={(d) => d.id}
          isLoading={isLoading}
          emptyMessage="Belum ada perangkat IoT terdaftar"
          onRowClick={(d) => setSelectedDeviceId(d.id)}
        />
      </Card>

      {/* Readings */}
      <Card padding="none">
        <div className="p-4 border-b border-gray-100">
          <CardHeader
            title={selectedDevice ? `Pembacaan — ${selectedDevice.name}` : 'Pembacaan Sensor'}
            description={
              selectedDevice
                ? `Lokasi: ${selectedDevice.location} · Topik: ${selectedDevice.mqttTopic}`
                : 'Pilih perangkat untuk melihat data pembacaan'
            }
          />
        </div>
        <Table
          columns={readingColumns}
          data={readings}
          keyExtractor={(r) => r.id}
          isLoading={isLoadingReadings}
          emptyMessage={selectedDeviceId ? 'Belum ada data pembacaan' : 'Pilih perangkat terlebih dahulu'}
        />
      </Card>

      {/* Add Device Modal */}
      <Modal
        isOpen={isAddOpen}
        onClose={() => { setIsAddOpen(false); setForm(EMPTY_FORM); }}
        title="Tambah Perangkat IoT"
        description="Daftarkan perangkat sensor baru ke sistem"
        footer={
          <>
            <Button variant="secondary" onClick={() => { setIsAddOpen(false); setForm(EMPTY_FORM); }}>Batal</Button>
            <Button onClick={handleAdd} isLoading={isSubmitting}>Simpan</Button>
          </>
        }
      >
        <div className="space-y-4">
          <Input
            label="Nama Perangkat"
            required
            placeholder="Contoh: Sensor Ruang Kelas 1A"
            value={form.name}
            onChange={(e) => setForm({ ...form, name: e.target.value })}
          />
          <Select
            label="Tipe Perangkat"
            options={Object.entries(DEVICE_TYPE_LABELS).map(([value, label]) => ({ value, label }))}
            value={form.type}
            onChange={(e) => setForm({ ...form, type: e.target.value })}
          />
          <Input
            label="Lokasi"
            required
            placeholder="Contoh: Ruang Kelas 1A, Lantai 1"
            value={form.location}
            onChange={(e) => setForm({ ...form, location: e.target.value })}
          />
          <Input
            label="MQTT Topic"
            required
            placeholder="Contoh: school/sensor/kelas-1a/temp"
            value={form.mqttTopic}
            onChange={(e) => setForm({ ...form, mqttTopic: e.target.value })}
          />
        </div>
      </Modal>
    </div>
  );
}
