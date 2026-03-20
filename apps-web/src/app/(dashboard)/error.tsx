'use client';

import { useEffect } from 'react';
import { AlertCircle } from 'lucide-react';

export default function DashboardError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error('Dashboard error:', error);
  }, [error]);

  return (
    <div className="flex items-center justify-center py-24">
      <div className="text-center max-w-sm">
        <AlertCircle className="w-12 h-12 text-red-400 mx-auto mb-3" />
        <h2 className="text-base font-semibold text-gray-800">Terjadi Kesalahan</h2>
        <p className="text-sm text-gray-500 mt-1">
          {error.message || 'Gagal memuat halaman. Silakan coba lagi.'}
        </p>
        <button
          onClick={reset}
          className="mt-4 px-5 py-2 bg-primary-600 text-white rounded-lg text-sm font-medium hover:bg-primary-700 transition-colors"
        >
          Coba Lagi
        </button>
      </div>
    </div>
  );
}
