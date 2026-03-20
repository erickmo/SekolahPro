'use client';

import React from 'react';
import { cn } from '@/lib/utils';
import { ChevronLeft, ChevronRight } from 'lucide-react';

interface PaginationProps {
  currentPage: number;
  totalPages: number;
  total: number;
  limit: number;
  onPageChange: (page: number) => void;
  className?: string;
}

export function Pagination({
  currentPage,
  totalPages,
  total,
  limit,
  onPageChange,
  className,
}: PaginationProps) {
  const from = (currentPage - 1) * limit + 1;
  const to = Math.min(currentPage * limit, total);

  const pages = buildPages(currentPage, totalPages);

  return (
    <div className={cn('flex items-center justify-between', className)}>
      <p className="text-sm text-gray-500">
        Menampilkan <span className="font-medium">{from}</span>–
        <span className="font-medium">{to}</span> dari{' '}
        <span className="font-medium">{total}</span> data
      </p>

      <div className="flex items-center gap-1">
        <button
          onClick={() => onPageChange(currentPage - 1)}
          disabled={currentPage <= 1}
          className={cn(
            'p-1.5 rounded-lg border text-sm transition-colors',
            'disabled:opacity-40 disabled:cursor-not-allowed',
            'hover:bg-gray-100 border-gray-300'
          )}
          aria-label="Sebelumnya"
        >
          <ChevronLeft className="w-4 h-4" />
        </button>

        {pages.map((p, i) =>
          p === '...' ? (
            <span key={`dots-${i}`} className="px-2 text-gray-400 text-sm">
              ...
            </span>
          ) : (
            <button
              key={p}
              onClick={() => onPageChange(Number(p))}
              className={cn(
                'min-w-[32px] h-8 rounded-lg border text-sm font-medium transition-colors',
                Number(p) === currentPage
                  ? 'bg-primary-600 text-white border-primary-600'
                  : 'border-gray-300 text-gray-700 hover:bg-gray-100'
              )}
            >
              {p}
            </button>
          )
        )}

        <button
          onClick={() => onPageChange(currentPage + 1)}
          disabled={currentPage >= totalPages}
          className={cn(
            'p-1.5 rounded-lg border text-sm transition-colors',
            'disabled:opacity-40 disabled:cursor-not-allowed',
            'hover:bg-gray-100 border-gray-300'
          )}
          aria-label="Berikutnya"
        >
          <ChevronRight className="w-4 h-4" />
        </button>
      </div>
    </div>
  );
}

function buildPages(current: number, total: number): (number | '...')[] {
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1);

  const pages: (number | '...')[] = [1];
  if (current > 3) pages.push('...');
  for (let p = Math.max(2, current - 1); p <= Math.min(total - 1, current + 1); p++) {
    pages.push(p);
  }
  if (current < total - 2) pages.push('...');
  pages.push(total);
  return pages;
}
