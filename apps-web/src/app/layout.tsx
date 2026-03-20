import type { Metadata } from 'next';
import './globals.css';
import { AuthProvider } from '@/context/AuthContext';

export const metadata: Metadata = {
  title: {
    default: 'EDS - Ekosistem Digital Sekolah',
    template: '%s | EDS',
  },
  description:
    'Platform terintegrasi yang mendigitalisasi seluruh operasional sekolah — administrasi, akademik, keuangan, kesehatan, dan lebih.',
  keywords: ['sekolah', 'digital', 'manajemen sekolah', 'SIMS', 'EDS'],
  viewport: 'width=device-width, initial-scale=1',
  icons: {
    icon: '/favicon.ico',
  },
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="id">
      <body className="bg-gray-50 text-gray-900">
        <AuthProvider>{children}</AuthProvider>
      </body>
    </html>
  );
}
