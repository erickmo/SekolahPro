'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Shield, AlertTriangle, Users, User, RefreshCw, Lock, Unlock } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import api from '@/lib/api';
import { Card, CardHeader, StatCard } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { formatDateTime } from '@/lib/utils';

interface SecurityDashboard {
  failedLogins24h: number;
  suspiciousActivity7d: number;
  activeSessions24h: number;
  totalUsers: number;
}

interface SuspiciousEvent {
  id: string;
  action: string;
  resource: string;
  userId: string;
  ipAddress: string;
  createdAt: string;
}

interface MfaStatus {
  enabled: boolean;
}

export default function SecurityPage() {
  const { user } = useAuth();
  const [dashboard, setDashboard] = useState<SecurityDashboard>({
    failedLogins24h: 0,
    suspiciousActivity7d: 0,
    activeSessions24h: 0,
    totalUsers: 0,
  });
  const [suspicious, setSuspicious] = useState<SuspiciousEvent[]>([]);
  const [mfa, setMfa] = useState<MfaStatus>({ enabled: false });
  const [isLoading, setIsLoading] = useState(true);
  const [isLoadingSuspicious, setIsLoadingSuspicious] = useState(true);
  const [isMfaLoading, setIsMfaLoading] = useState(false);

  const fetchDashboard = useCallback(async () => {
    setIsLoading(true);
    try {
      const res = await api.get('/security/dashboard');
      setDashboard(res.data.data || {});
    } catch {
      // silent
    } finally {
      setIsLoading(false);
    }
  }, []);

  const fetchSuspicious = useCallback(async () => {
    setIsLoadingSuspicious(true);
    try {
      const res = await api.get('/security/suspicious');
      setSuspicious(res.data.data || []);
      // Extract MFA status if included
      if (res.data.data?.mfaEnabled !== undefined) {
        setMfa({ enabled: res.data.data.mfaEnabled });
      }
    } catch {
      setSuspicious([]);
    } finally {
      setIsLoadingSuspicious(false);
    }
  }, []);

  const fetchMfaStatus = useCallback(async () => {
    try {
      const res = await api.get('/security/mfa/status');
      setMfa({ enabled: res.data.data?.enabled ?? false });
    } catch {
      // silent
    }
  }, []);

  useEffect(() => {
    fetchDashboard();
    fetchSuspicious();
    fetchMfaStatus();
  }, [fetchDashboard, fetchSuspicious, fetchMfaStatus]);

  const toggleMfa = async () => {
    setIsMfaLoading(true);
    try {
      if (mfa.enabled) {
        await api.post('/security/mfa/disable');
        setMfa({ enabled: false });
      } else {
        await api.post('/security/mfa/enable');
        setMfa({ enabled: true });
      }
    } catch {
      // silent
    } finally {
      setIsMfaLoading(false);
    }
  };

  const suspiciousColumns = [
    {
      key: 'action',
      header: 'Aksi',
      render: (e: SuspiciousEvent) => (
        <Badge variant="danger" size="sm">{e.action}</Badge>
      ),
    },
    {
      key: 'resource',
      header: 'Resource',
      render: (e: SuspiciousEvent) => <span className="text-sm text-gray-700 font-mono">{e.resource}</span>,
    },
    {
      key: 'userId',
      header: 'User ID',
      render: (e: SuspiciousEvent) => <span className="text-sm text-gray-600">{e.userId}</span>,
    },
    {
      key: 'ipAddress',
      header: 'IP Address',
      render: (e: SuspiciousEvent) => <span className="text-sm font-mono text-gray-600">{e.ipAddress}</span>,
    },
    {
      key: 'createdAt',
      header: 'Waktu',
      render: (e: SuspiciousEvent) => (
        <span className="text-xs text-gray-400">{formatDateTime(e.createdAt)}</span>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="p-2 bg-red-100 rounded-xl">
            <Shield className="w-5 h-5 text-red-600" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-gray-900">Security Center</h1>
            <p className="text-sm text-gray-500 mt-0.5">Pantau keamanan dan aktivitas mencurigakan</p>
          </div>
        </div>
        <Button variant="secondary" size="sm" leftIcon={<RefreshCw className="w-3.5 h-3.5" />} onClick={() => { fetchDashboard(); fetchSuspicious(); }}>
          Refresh
        </Button>
      </div>

      {/* Dashboard cards */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title="Gagal Login (24 jam)"
          value={isLoading ? '...' : dashboard.failedLogins24h}
          icon={<AlertTriangle className="w-5 h-5 text-red-600" />}
          iconBg="bg-red-100"
        />
        <StatCard
          title="Aktivitas Mencurigakan (7 hari)"
          value={isLoading ? '...' : dashboard.suspiciousActivity7d}
          icon={<Shield className="w-5 h-5 text-amber-600" />}
          iconBg="bg-amber-100"
        />
        <StatCard
          title="Sesi Aktif (24 jam)"
          value={isLoading ? '...' : dashboard.activeSessions24h}
          icon={<Users className="w-5 h-5 text-blue-600" />}
          iconBg="bg-blue-100"
        />
        <StatCard
          title="Total Pengguna"
          value={isLoading ? '...' : dashboard.totalUsers}
          icon={<User className="w-5 h-5 text-primary-600" />}
          iconBg="bg-primary-100"
        />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Suspicious activity table */}
        <div className="lg:col-span-2">
          <Card padding="none">
            <div className="p-4 border-b border-gray-100">
              <CardHeader title="Aktivitas Mencurigakan" description="Log aktivitas yang terdeteksi sebagai anomali" />
            </div>
            <Table
              columns={suspiciousColumns}
              data={suspicious}
              keyExtractor={(e) => e.id}
              isLoading={isLoadingSuspicious}
              emptyMessage="Tidak ada aktivitas mencurigakan terdeteksi"
            />
          </Card>
        </div>

        {/* MFA card */}
        <div>
          <Card>
            <CardHeader
              title="Autentikasi Dua Faktor"
              description="MFA untuk akun Anda"
            />
            <div className="space-y-4">
              <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                <div>
                  <p className="text-sm font-medium text-gray-700">Status MFA</p>
                  <p className="text-xs text-gray-400 mt-0.5">Akun: {user?.email || '—'}</p>
                </div>
                <Badge variant={mfa.enabled ? 'success' : 'danger'} size="sm">
                  {mfa.enabled ? 'Aktif' : 'Nonaktif'}
                </Badge>
              </div>

              {mfa.enabled ? (
                <div className="p-3 bg-green-50 rounded-lg">
                  <p className="text-xs text-green-700">
                    MFA aktif. Akun Anda terlindungi dengan lapisan keamanan tambahan.
                  </p>
                </div>
              ) : (
                <div className="p-3 bg-amber-50 rounded-lg">
                  <p className="text-xs text-amber-700">
                    MFA belum aktif. Aktifkan untuk meningkatkan keamanan akun Anda.
                  </p>
                </div>
              )}

              <Button
                variant={mfa.enabled ? 'secondary' : 'primary'}
                size="sm"
                className="w-full"
                leftIcon={mfa.enabled ? <Unlock className="w-4 h-4" /> : <Lock className="w-4 h-4" />}
                isLoading={isMfaLoading}
                onClick={toggleMfa}
              >
                {mfa.enabled ? 'Nonaktifkan MFA' : 'Aktifkan MFA'}
              </Button>
            </div>
          </Card>
        </div>
      </div>
    </div>
  );
}
