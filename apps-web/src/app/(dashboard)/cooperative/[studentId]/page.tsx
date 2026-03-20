'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { ArrowLeft, PiggyBank, ArrowDownCircle, ArrowUpCircle, Plus } from 'lucide-react';
import api from '@/lib/api';
import { useAuth } from '@/context/AuthContext';
import type { SavingsAccount, CoopTransaction } from '@/types';
import { Card, CardHeader } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Modal } from '@/components/ui/Modal';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { formatCurrency, formatDateTime } from '@/lib/utils';

export default function CooperativeSavingsPage() {
  const { studentId } = useParams<{ studentId: string }>();
  const router = useRouter();
  const { user } = useAuth();
  const [account, setAccount] = useState<SavingsAccount | null>(null);
  const [transactions, setTransactions] = useState<CoopTransaction[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [depositOpen, setDepositOpen] = useState(false);
  const [withdrawOpen, setWithdrawOpen] = useState(false);
  const [amount, setAmount] = useState('');
  const [description, setDescription] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [createOpen, setCreateOpen] = useState(false);

  const canTransact = ['KASIR_KOPERASI', 'BENDAHARA', 'ADMIN_SEKOLAH'].includes(user?.role || '');

  const fetchData = useCallback(async () => {
    setIsLoading(true);
    try {
      const [accRes, txRes] = await Promise.all([
        api.get(`/cooperative/${studentId}/account`),
        api.get(`/cooperative/${studentId}/transactions`),
      ]);
      setAccount(accRes.data.data);
      setTransactions(txRes.data.data || []);
    } catch {
      setAccount(null);
    } finally {
      setIsLoading(false);
    }
  }, [studentId]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleCreateAccount = async () => {
    setIsSubmitting(true);
    try {
      await api.post(`/cooperative/${studentId}/account`);
      setCreateOpen(false);
      fetchData();
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDeposit = async () => {
    if (!amount) return;
    setIsSubmitting(true);
    try {
      await api.post(`/cooperative/${studentId}/deposit`, {
        amount: Number(amount),
        description: description || 'Setoran tabungan',
      });
      setAmount('');
      setDescription('');
      setDepositOpen(false);
      fetchData();
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleWithdraw = async () => {
    if (!amount) return;
    setIsSubmitting(true);
    try {
      await api.post(`/cooperative/${studentId}/withdraw`, {
        amount: Number(amount),
        description: description || 'Penarikan tabungan',
      });
      setAmount('');
      setDescription('');
      setWithdrawOpen(false);
      fetchData();
    } finally {
      setIsSubmitting(false);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-24">
        <div className="w-7 h-7 border-2 border-primary-500 border-t-transparent rounded-full animate-spin" />
      </div>
    );
  }

  if (!account) {
    return (
      <div className="space-y-6">
        <Button variant="ghost" size="sm" leftIcon={<ArrowLeft className="w-4 h-4" />} onClick={() => router.back()}>
          Kembali
        </Button>
        <Card>
          <div className="text-center py-12">
            <PiggyBank className="w-12 h-12 text-gray-300 mx-auto mb-3" />
            <h2 className="text-base font-semibold text-gray-700">Akun Tabungan Belum Ada</h2>
            <p className="text-sm text-gray-400 mt-1">Siswa ini belum memiliki akun tabungan koperasi</p>
            {canTransact && (
              <Button className="mt-4" onClick={() => setCreateOpen(true)}>
                Buat Akun Tabungan
              </Button>
            )}
          </div>
        </Card>

        <Modal
          isOpen={createOpen}
          onClose={() => setCreateOpen(false)}
          title="Buat Akun Tabungan"
          description={`Buat akun tabungan koperasi untuk siswa ID: ${studentId}`}
          size="sm"
          footer={
            <>
              <Button variant="secondary" onClick={() => setCreateOpen(false)}>Batal</Button>
              <Button onClick={handleCreateAccount} isLoading={isSubmitting}>Buat Akun</Button>
            </>
          }
        >
          <p className="text-sm text-gray-600">
            Akun tabungan akan dibuat dengan saldo awal Rp 0. Setelah dibuat, kasir dapat mencatat setoran dan penarikan.
          </p>
        </Modal>
      </div>
    );
  }

  const txColumns = [
    {
      key: 'type',
      header: 'Jenis',
      render: (tx: CoopTransaction) => {
        const map = {
          DEPOSIT: { label: 'Setoran', variant: 'success' as const, icon: <ArrowDownCircle className="w-3.5 h-3.5" /> },
          WITHDRAWAL: { label: 'Penarikan', variant: 'danger' as const, icon: <ArrowUpCircle className="w-3.5 h-3.5" /> },
          PURCHASE: { label: 'Pembelian', variant: 'warning' as const, icon: <ArrowUpCircle className="w-3.5 h-3.5" /> },
        };
        const { label, variant, icon } = map[tx.type] || { label: tx.type, variant: 'default' as const, icon: null };
        return (
          <Badge variant={variant} size="sm">
            <span className="flex items-center gap-1">{icon}{label}</span>
          </Badge>
        );
      },
    },
    {
      key: 'amount',
      header: 'Jumlah',
      render: (tx: CoopTransaction) => (
        <span className={tx.type === 'DEPOSIT' ? 'text-green-700 font-semibold' : 'text-red-700 font-semibold'}>
          {tx.type === 'DEPOSIT' ? '+' : '-'}{formatCurrency(tx.amount)}
        </span>
      ),
    },
    {
      key: 'balanceAfter',
      header: 'Saldo',
      render: (tx: CoopTransaction) => formatCurrency(tx.balanceAfter),
    },
    {
      key: 'description',
      header: 'Keterangan',
    },
    {
      key: 'createdAt',
      header: 'Waktu',
      render: (tx: CoopTransaction) => formatDateTime(tx.createdAt),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="sm" leftIcon={<ArrowLeft className="w-4 h-4" />} onClick={() => router.back()}>
          Kembali
        </Button>
        <h1 className="text-xl font-bold text-gray-900">Tabungan Koperasi</h1>
      </div>

      {/* Balance card */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card className="md:col-span-1 bg-gradient-to-br from-primary-600 to-primary-800 text-white">
          <div className="flex items-center justify-between mb-4">
            <PiggyBank className="w-7 h-7 text-primary-200" />
            <Badge variant="primary" className="bg-white/20 text-white border-0">Tabungan</Badge>
          </div>
          <p className="text-sm text-primary-200">Saldo Saat Ini</p>
          <p className="text-3xl font-bold mt-1">{formatCurrency(account.balance)}</p>

          {canTransact && (
            <div className="flex gap-2 mt-5">
              <Button
                variant="secondary"
                size="sm"
                className="flex-1 bg-white/10 border-white/20 text-white hover:bg-white/20"
                leftIcon={<ArrowDownCircle className="w-4 h-4" />}
                onClick={() => setDepositOpen(true)}
              >
                Setor
              </Button>
              <Button
                variant="secondary"
                size="sm"
                className="flex-1 bg-white/10 border-white/20 text-white hover:bg-white/20"
                leftIcon={<ArrowUpCircle className="w-4 h-4" />}
                onClick={() => setWithdrawOpen(true)}
              >
                Tarik
              </Button>
            </div>
          )}
        </Card>

        <Card padding="sm">
          <p className="text-sm text-gray-500">Total Setoran</p>
          <p className="text-xl font-bold text-green-700 mt-1">{formatCurrency(account.totalDeposits)}</p>
          <p className="text-xs text-gray-400 mt-1">Sejak akun dibuat</p>
        </Card>
        <Card padding="sm">
          <p className="text-sm text-gray-500">Total Penarikan</p>
          <p className="text-xl font-bold text-red-700 mt-1">{formatCurrency(account.totalWithdrawals)}</p>
          <p className="text-xs text-gray-400 mt-1">Sejak akun dibuat</p>
        </Card>
      </div>

      {/* Transactions */}
      <Card padding="none">
        <CardHeader
          title="Riwayat Transaksi"
          description={`${transactions.length} transaksi terakhir`}
          className="px-6 pt-5 pb-4"
        />
        <Table
          columns={txColumns}
          data={transactions}
          keyExtractor={(tx) => tx.id}
          emptyMessage="Belum ada transaksi"
        />
      </Card>

      {/* Deposit Modal */}
      <Modal
        isOpen={depositOpen}
        onClose={() => { setDepositOpen(false); setAmount(''); setDescription(''); }}
        title="Setoran Tabungan"
        size="sm"
        footer={
          <>
            <Button variant="secondary" onClick={() => setDepositOpen(false)}>Batal</Button>
            <Button onClick={handleDeposit} isLoading={isSubmitting}>Catat Setoran</Button>
          </>
        }
      >
        <div className="space-y-4">
          <Input
            label="Jumlah Setoran (Rp)"
            type="number"
            placeholder="50000"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            leftAddon="Rp"
            required
          />
          <Input
            label="Keterangan"
            placeholder="Setoran rutin..."
            value={description}
            onChange={(e) => setDescription(e.target.value)}
          />
        </div>
      </Modal>

      {/* Withdraw Modal */}
      <Modal
        isOpen={withdrawOpen}
        onClose={() => { setWithdrawOpen(false); setAmount(''); setDescription(''); }}
        title="Penarikan Tabungan"
        description={`Saldo saat ini: ${formatCurrency(account.balance)}`}
        size="sm"
        footer={
          <>
            <Button variant="secondary" onClick={() => setWithdrawOpen(false)}>Batal</Button>
            <Button variant="danger" onClick={handleWithdraw} isLoading={isSubmitting}>Catat Penarikan</Button>
          </>
        }
      >
        <div className="space-y-4">
          <Input
            label="Jumlah Penarikan (Rp)"
            type="number"
            placeholder="50000"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            leftAddon="Rp"
            required
          />
          <Input
            label="Keterangan"
            placeholder="Penarikan untuk keperluan..."
            value={description}
            onChange={(e) => setDescription(e.target.value)}
          />
        </div>
      </Modal>
    </div>
  );
}
