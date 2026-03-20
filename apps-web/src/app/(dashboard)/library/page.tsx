'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Plus, Search, BookOpen, Library, RotateCcw, AlertCircle } from 'lucide-react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import api from '@/lib/api';
import { useAuth } from '@/context/AuthContext';
import type { Book, BookLoan } from '@/types';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Table } from '@/components/ui/Table';
import { LoanStatusBadge } from '@/components/ui/Badge';
import { Pagination } from '@/components/ui/Pagination';
import { Modal } from '@/components/ui/Modal';
import { formatDate } from '@/lib/utils';

const addBookSchema = z.object({
  title: z.string().min(1, 'Judul buku wajib diisi'),
  author: z.string().min(1, 'Pengarang wajib diisi'),
  isbn: z.string().optional(),
  publisher: z.string().optional(),
  publishYear: z.number().optional(),
  category: z.string().min(1, 'Kategori wajib diisi'),
  stock: z.number().min(1, 'Stok minimal 1'),
});
type AddBookForm = z.infer<typeof addBookSchema>;

const borrowSchema = z.object({
  bookId: z.string().min(1, 'Pilih buku'),
  studentId: z.string().min(1, 'ID siswa wajib diisi'),
  dueDate: z.string().min(1, 'Tanggal kembali wajib diisi'),
});
type BorrowForm = z.infer<typeof borrowSchema>;

type TabKey = 'books' | 'loans';

export default function LibraryPage() {
  const { user } = useAuth();
  const [activeTab, setActiveTab] = useState<TabKey>('books');
  const [books, setBooks] = useState<Book[]>([]);
  const [loans, setLoans] = useState<BookLoan[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [searchBook, setSearchBook] = useState('');
  const [bookPage, setBookPage] = useState(1);
  const [loanPage, setLoanPage] = useState(1);
  const [bookMeta, setBookMeta] = useState({ total: 0, totalPages: 1 });
  const [loanMeta, setLoanMeta] = useState({ total: 0, totalPages: 1 });
  const [addBookOpen, setAddBookOpen] = useState(false);
  const [borrowOpen, setBorrowOpen] = useState(false);

  const isPustakawan = ['PUSTAKAWAN', 'ADMIN_SEKOLAH'].includes(user?.role || '');

  const {
    register: regBook,
    handleSubmit: submitBook,
    reset: resetBook,
    formState: { errors: errBook, isSubmitting: submittingBook },
  } = useForm<AddBookForm>({
    resolver: zodResolver(addBookSchema),
    defaultValues: { stock: 1 },
  });

  const {
    register: regBorrow,
    handleSubmit: submitBorrow,
    reset: resetBorrow,
    formState: { errors: errBorrow, isSubmitting: submittingBorrow },
  } = useForm<BorrowForm>({ resolver: zodResolver(borrowSchema) });

  const fetchBooks = useCallback(async () => {
    setIsLoading(true);
    try {
      const params = new URLSearchParams({
        page: String(bookPage),
        limit: '20',
        ...(searchBook && { search: searchBook }),
      });
      const res = await api.get(`/library/books?${params}`);
      setBooks(res.data.data || []);
      setBookMeta(res.data.meta || { total: 0, totalPages: 1 });
    } catch {
      setBooks([]);
    } finally {
      setIsLoading(false);
    }
  }, [bookPage, searchBook]);

  const fetchLoans = useCallback(async () => {
    setIsLoading(true);
    try {
      const params = new URLSearchParams({ page: String(loanPage), limit: '20' });
      const res = await api.get(`/library/loans?${params}`);
      setLoans(res.data.data || []);
      setLoanMeta(res.data.meta || { total: 0, totalPages: 1 });
    } catch {
      setLoans([]);
    } finally {
      setIsLoading(false);
    }
  }, [loanPage]);

  useEffect(() => { fetchBooks(); }, [fetchBooks]);
  useEffect(() => { fetchLoans(); }, [fetchLoans]);

  const onAddBook = async (values: AddBookForm) => {
    await api.post('/library/books', values);
    resetBook();
    setAddBookOpen(false);
    fetchBooks();
  };

  const onBorrow = async (values: BorrowForm) => {
    await api.post('/library/loans', values);
    resetBorrow();
    setBorrowOpen(false);
    fetchLoans();
  };

  const onReturn = async (loanId: string) => {
    await api.put(`/library/loans/${loanId}/return`);
    fetchLoans();
  };

  const bookColumns = [
    {
      key: 'title',
      header: 'Judul Buku',
      render: (b: Book) => (
        <div>
          <p className="font-medium text-gray-900">{b.title}</p>
          <p className="text-xs text-gray-400">{b.author}</p>
        </div>
      ),
    },
    { key: 'category', header: 'Kategori' },
    {
      key: 'stock',
      header: 'Stok',
      render: (b: Book) => (
        <div className="flex items-center gap-2">
          <span className="font-medium">{b.available}/{b.stock}</span>
          {b.available === 0 && (
            <AlertCircle className="w-4 h-4 text-red-500" />
          )}
        </div>
      ),
    },
    {
      key: 'isbn',
      header: 'ISBN',
      render: (b: Book) => b.isbn || '—',
    },
  ];

  const loanColumns = [
    {
      key: 'bookTitle',
      header: 'Buku',
      render: (l: BookLoan) => l.bookTitle || l.bookId,
    },
    {
      key: 'studentName',
      header: 'Peminjam',
      render: (l: BookLoan) => l.studentName || l.studentId,
    },
    {
      key: 'borrowedAt',
      header: 'Tanggal Pinjam',
      render: (l: BookLoan) => formatDate(l.borrowedAt),
    },
    {
      key: 'dueDate',
      header: 'Tenggat',
      render: (l: BookLoan) => formatDate(l.dueDate),
    },
    {
      key: 'status',
      header: 'Status',
      render: (l: BookLoan) => <LoanStatusBadge status={l.status} />,
    },
    {
      key: 'actions',
      header: '',
      render: (l: BookLoan) =>
        l.status === 'BORROWED' && isPustakawan ? (
          <Button
            variant="outline"
            size="sm"
            leftIcon={<RotateCcw className="w-3.5 h-3.5" />}
            onClick={() => onReturn(l.id)}
          >
            Kembalikan
          </Button>
        ) : null,
      className: 'w-36',
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900">Perpustakaan Digital</h1>
          <p className="text-sm text-gray-500 mt-0.5">Kelola koleksi buku dan peminjaman</p>
        </div>
        {isPustakawan && (
          <div className="flex gap-2">
            <Button variant="secondary" size="sm" leftIcon={<BookOpen className="w-3.5 h-3.5" />} onClick={() => setBorrowOpen(true)}>
              Catat Pinjaman
            </Button>
            <Button size="sm" leftIcon={<Plus className="w-3.5 h-3.5" />} onClick={() => setAddBookOpen(true)}>
              Tambah Buku
            </Button>
          </div>
        )}
      </div>

      {/* Summary */}
      <div className="grid grid-cols-3 gap-4">
        <Card padding="sm">
          <div className="flex items-center gap-3">
            <div className="p-2.5 bg-primary-50 rounded-lg">
              <Library className="w-5 h-5 text-primary-600" />
            </div>
            <div>
              <p className="text-xl font-bold text-gray-900">{bookMeta.total}</p>
              <p className="text-xs text-gray-500">Total Koleksi</p>
            </div>
          </div>
        </Card>
        <Card padding="sm">
          <div className="flex items-center gap-3">
            <div className="p-2.5 bg-amber-50 rounded-lg">
              <BookOpen className="w-5 h-5 text-amber-600" />
            </div>
            <div>
              <p className="text-xl font-bold text-gray-900">
                {loans.filter((l) => l.status === 'BORROWED').length}
              </p>
              <p className="text-xs text-gray-500">Sedang Dipinjam</p>
            </div>
          </div>
        </Card>
        <Card padding="sm">
          <div className="flex items-center gap-3">
            <div className="p-2.5 bg-red-50 rounded-lg">
              <AlertCircle className="w-5 h-5 text-red-600" />
            </div>
            <div>
              <p className="text-xl font-bold text-gray-900">
                {loans.filter((l) => l.status === 'OVERDUE').length}
              </p>
              <p className="text-xs text-gray-500">Terlambat</p>
            </div>
          </div>
        </Card>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 bg-gray-100 p-1 rounded-xl w-fit">
        {(['books', 'loans'] as TabKey[]).map((t) => (
          <button
            key={t}
            onClick={() => setActiveTab(t)}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
              activeTab === t ? 'bg-white text-primary-700 shadow-sm' : 'text-gray-500'
            }`}
          >
            {t === 'books' ? 'Koleksi Buku' : 'Daftar Pinjaman'}
          </button>
        ))}
      </div>

      {activeTab === 'books' && (
        <Card padding="none">
          <div className="p-4 border-b border-gray-100">
            <Input
              placeholder="Cari judul atau pengarang..."
              value={searchBook}
              onChange={(e) => { setSearchBook(e.target.value); setBookPage(1); }}
              leftAddon={<Search className="w-4 h-4" />}
              className="max-w-xs"
            />
          </div>
          <Table columns={bookColumns} data={books} keyExtractor={(b) => b.id} isLoading={isLoading} emptyMessage="Belum ada koleksi buku" />
          {bookMeta.total > 0 && (
            <div className="px-4 py-3 border-t border-gray-100">
              <Pagination currentPage={bookPage} totalPages={bookMeta.totalPages} total={bookMeta.total} limit={20} onPageChange={setBookPage} />
            </div>
          )}
        </Card>
      )}

      {activeTab === 'loans' && (
        <Card padding="none">
          <Table columns={loanColumns} data={loans} keyExtractor={(l) => l.id} isLoading={isLoading} emptyMessage="Tidak ada data peminjaman" />
          {loanMeta.total > 0 && (
            <div className="px-4 py-3 border-t border-gray-100">
              <Pagination currentPage={loanPage} totalPages={loanMeta.totalPages} total={loanMeta.total} limit={20} onPageChange={setLoanPage} />
            </div>
          )}
        </Card>
      )}

      {/* Add Book Modal */}
      <Modal
        isOpen={addBookOpen}
        onClose={() => { setAddBookOpen(false); resetBook(); }}
        title="Tambah Buku Baru"
        size="md"
        footer={
          <>
            <Button variant="secondary" onClick={() => setAddBookOpen(false)}>Batal</Button>
            <Button form="book-form" type="submit" isLoading={submittingBook}>Simpan</Button>
          </>
        }
      >
        <form id="book-form" onSubmit={submitBook(onAddBook)} className="space-y-4">
          <Input label="Judul Buku" required {...regBook('title')} error={errBook.title?.message} />
          <Input label="Pengarang" required {...regBook('author')} error={errBook.author?.message} />
          <div className="grid grid-cols-2 gap-3">
            <Input label="ISBN" {...regBook('isbn')} />
            <Input label="Penerbit" {...regBook('publisher')} />
            <Input label="Tahun Terbit" type="number" {...regBook('publishYear', { valueAsNumber: true })} />
            <Input label="Stok" type="number" min={1} required {...regBook('stock', { valueAsNumber: true })} error={errBook.stock?.message} />
          </div>
          <Input label="Kategori" placeholder="Fiksi / Referensi / Pelajaran..." required {...regBook('category')} error={errBook.category?.message} />
        </form>
      </Modal>

      {/* Borrow Modal */}
      <Modal
        isOpen={borrowOpen}
        onClose={() => { setBorrowOpen(false); resetBorrow(); }}
        title="Catat Peminjaman Buku"
        size="sm"
        footer={
          <>
            <Button variant="secondary" onClick={() => setBorrowOpen(false)}>Batal</Button>
            <Button form="borrow-form" type="submit" isLoading={submittingBorrow}>Simpan</Button>
          </>
        }
      >
        <form id="borrow-form" onSubmit={submitBorrow(onBorrow)} className="space-y-4">
          <Input label="ID Buku" placeholder="ID buku dari sistem" required {...regBorrow('bookId')} error={errBorrow.bookId?.message} />
          <Input label="ID Siswa" placeholder="ID siswa" required {...regBorrow('studentId')} error={errBorrow.studentId?.message} />
          <Input label="Tenggat Pengembalian" type="date" required {...regBorrow('dueDate')} error={errBorrow.dueDate?.message} />
        </form>
      </Modal>
    </div>
  );
}
