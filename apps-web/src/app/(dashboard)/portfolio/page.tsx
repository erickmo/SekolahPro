'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { FolderOpen, Search, ArrowLeft, FileText } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import api from '@/lib/api';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { Pagination } from '@/components/ui/Pagination';
import { formatDate } from '@/lib/utils';

interface Portfolio {
  id: string;
  studentId: string;
  studentName: string;
  itemCount: number;
  createdAt: string;
}

interface PortfolioItem {
  id: string;
  title: string;
  type: string;
  description: string;
  createdAt: string;
}

export default function PortfolioPage() {
  const { user } = useAuth();
  const [portfolios, setPortfolios] = useState<Portfolio[]>([]);
  const [items, setItems] = useState<PortfolioItem[]>([]);
  const [selectedPortfolio, setSelectedPortfolio] = useState<Portfolio | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [page, setPage] = useState(1);
  const [meta, setMeta] = useState({ total: 0, totalPages: 1 });

  const fetchPortfolios = useCallback(async () => {
    setIsLoading(true);
    try {
      const params = new URLSearchParams({
        page: String(page),
        limit: '20',
        ...(search && { search }),
      });
      const res = await api.get(`/portfolio?${params}`);
      setPortfolios(res.data.data || []);
      setMeta(res.data.meta || { total: 0, totalPages: 1 });
    } catch {
      setPortfolios([]);
    } finally {
      setIsLoading(false);
    }
  }, [page, search]);

  const fetchItems = useCallback(async (portfolioId: string) => {
    setIsLoading(true);
    try {
      const res = await api.get(`/portfolio/${portfolioId}/items`);
      setItems(res.data.data || []);
    } catch {
      setItems([]);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!selectedPortfolio) {
      fetchPortfolios();
    }
  }, [fetchPortfolios, selectedPortfolio]);

  const handleViewPortfolio = (portfolio: Portfolio) => {
    setSelectedPortfolio(portfolio);
    fetchItems(portfolio.id);
  };

  const handleBack = () => {
    setSelectedPortfolio(null);
    setItems([]);
  };

  const portfolioColumns = [
    {
      key: 'student',
      header: 'Siswa',
      render: (p: Portfolio) => (
        <div>
          <p className="font-medium text-gray-900">{p.studentName || p.studentId}</p>
          <p className="text-xs text-gray-400">ID: {p.studentId}</p>
        </div>
      ),
    },
    {
      key: 'itemCount',
      header: 'Jumlah Karya',
      render: (p: Portfolio) => (
        <div className="flex items-center gap-1.5 text-gray-600">
          <FileText className="w-3.5 h-3.5" />
          <span className="text-sm font-medium">{p.itemCount ?? 0} karya</span>
        </div>
      ),
    },
    {
      key: 'createdAt',
      header: 'Dibuat',
      render: (p: Portfolio) => (
        <span className="text-sm text-gray-500">{formatDate(p.createdAt)}</span>
      ),
    },
    {
      key: 'actions',
      header: '',
      render: (p: Portfolio) => (
        <Button variant="outline" size="sm" onClick={() => handleViewPortfolio(p)}>
          Lihat
        </Button>
      ),
      className: 'w-24',
    },
  ];

  const itemColumns = [
    {
      key: 'title',
      header: 'Judul Karya',
      render: (i: PortfolioItem) => (
        <div>
          <p className="font-medium text-gray-900">{i.title}</p>
          <p className="text-xs text-gray-400 line-clamp-1">{i.description}</p>
        </div>
      ),
    },
    {
      key: 'type',
      header: 'Tipe',
      render: (i: PortfolioItem) => (
        <Badge variant="info">{i.type || '—'}</Badge>
      ),
    },
    {
      key: 'createdAt',
      header: 'Tanggal',
      render: (i: PortfolioItem) => (
        <span className="text-sm text-gray-500">{formatDate(i.createdAt)}</span>
      ),
    },
  ];

  if (selectedPortfolio) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Button variant="secondary" size="sm" leftIcon={<ArrowLeft className="w-3.5 h-3.5" />} onClick={handleBack}>
              Kembali
            </Button>
            <div>
              <h1 className="text-xl font-bold text-gray-900">
                Portofolio — {selectedPortfolio.studentName || selectedPortfolio.studentId}
              </h1>
              <p className="text-sm text-gray-500 mt-0.5">{selectedPortfolio.itemCount ?? 0} karya</p>
            </div>
          </div>
        </div>

        <Card padding="none">
          <Table
            columns={itemColumns}
            data={items}
            keyExtractor={(i) => i.id}
            isLoading={isLoading}
            emptyMessage="Belum ada karya dalam portofolio ini"
          />
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900">Portofolio Digital Siswa</h1>
          <p className="text-sm text-gray-500 mt-0.5">Koleksi karya dan pencapaian siswa</p>
        </div>
      </div>

      <Card padding="none">
        <div className="p-4 border-b border-gray-100">
          <Input
            placeholder="Cari nama siswa..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1); }}
            leftAddon={<Search className="w-4 h-4" />}
            className="max-w-xs"
          />
        </div>

        <Table
          columns={portfolioColumns}
          data={portfolios}
          keyExtractor={(p) => p.id}
          isLoading={isLoading}
          emptyMessage="Belum ada portofolio siswa"
        />

        {meta.total > 0 && (
          <div className="px-4 py-3 border-t border-gray-100">
            <Pagination
              currentPage={page}
              totalPages={meta.totalPages}
              total={meta.total}
              limit={20}
              onPageChange={setPage}
            />
          </div>
        )}
      </Card>
    </div>
  );
}
