import { useState, useEffect, useCallback } from 'react';
import api from '@/lib/api';
import type { PaginationMeta } from '@/types';

interface UsePaginatedFetchOptions {
  url: string;
  params?: Record<string, string | number | undefined>;
  enabled?: boolean;
}

interface UsePaginatedFetchResult<T> {
  data: T[];
  meta: PaginationMeta;
  isLoading: boolean;
  error: string | null;
  refetch: () => void;
  page: number;
  setPage: (p: number) => void;
}

const DEFAULT_META: PaginationMeta = {
  total: 0,
  page: 1,
  limit: 20,
  totalPages: 1,
};

export function usePaginatedFetch<T>({
  url,
  params = {},
  enabled = true,
}: UsePaginatedFetchOptions): UsePaginatedFetchResult<T> {
  const [data, setData] = useState<T[]>([]);
  const [meta, setMeta] = useState<PaginationMeta>(DEFAULT_META);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);

  const fetch = useCallback(async () => {
    if (!enabled) return;
    setIsLoading(true);
    setError(null);
    try {
      const searchParams = new URLSearchParams();
      searchParams.set('page', String(page));
      searchParams.set('limit', '20');
      Object.entries(params).forEach(([k, v]) => {
        if (v !== undefined && v !== '') searchParams.set(k, String(v));
      });

      const res = await api.get(`${url}?${searchParams}`);
      setData(res.data.data || []);
      setMeta(res.data.meta || DEFAULT_META);
    } catch (err: unknown) {
      const axiosErr = err as { response?: { data?: { error?: { message?: string } } } };
      setError(axiosErr?.response?.data?.error?.message || 'Gagal memuat data');
      setData([]);
    } finally {
      setIsLoading(false);
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [url, page, enabled, JSON.stringify(params)]);

  useEffect(() => {
    fetch();
  }, [fetch]);

  return { data, meta, isLoading, error, refetch: fetch, page, setPage };
}
