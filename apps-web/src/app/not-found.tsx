import Link from 'next/link';

export default function NotFound() {
  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <div className="text-center max-w-md">
        <p className="text-7xl font-bold text-gray-200 select-none">404</p>
        <h1 className="mt-4 text-xl font-bold text-gray-800">Halaman Tidak Ditemukan</h1>
        <p className="mt-2 text-sm text-gray-500">
          Halaman yang Anda cari tidak ada atau telah dipindahkan.
        </p>
        <Link
          href="/dashboard"
          className="mt-6 inline-flex items-center px-5 py-2.5 bg-primary-600 text-white rounded-lg text-sm font-medium hover:bg-primary-700 transition-colors"
        >
          Kembali ke Dashboard
        </Link>
      </div>
    </div>
  );
}
