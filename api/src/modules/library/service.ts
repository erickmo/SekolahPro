import { prisma } from '../../lib/prisma';
import { NotFoundError, BadRequestError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';

export async function getBooks(schoolId: string, query: { search?: string; category?: string; available?: string; page?: number; limit?: number }) {
  const { page, limit, skip } = getPaginationParams(query);
  const where: Record<string, unknown> = { schoolId };
  if (query.search) where.title = { contains: query.search, mode: 'insensitive' };
  if (query.category) where.category = query.category;
  if (query.available === 'true') where.availableCopies = { gt: 0 };
  const [books, total] = await Promise.all([
    prisma.libraryBook.findMany({ where, skip, take: limit, orderBy: { title: 'asc' } }),
    prisma.libraryBook.count({ where }),
  ]);
  return { books, meta: buildPaginationMeta(total, page, limit) };
}

export async function addBook(schoolId: string, data: { title: string; author: string; isbn?: string; category: string; totalCopies?: number; publishYear?: number; coverUrl?: string; stock?: number }) {
  const copies = data.totalCopies ?? data.stock ?? 1;
  return prisma.libraryBook.create({ data: { title: data.title, author: data.author, isbn: data.isbn, category: data.category, publishYear: data.publishYear, coverUrl: data.coverUrl, totalCopies: copies, availableCopies: copies, schoolId } });
}

export async function borrowBook(schoolId: string, data: { bookId: string; borrowerId: string; dueDate: string }) {
  const book = await prisma.libraryBook.findFirst({ where: { id: data.bookId, schoolId } });
  if (!book) throw new NotFoundError('Buku');
  if (book.availableCopies <= 0) throw new BadRequestError('LIBRARY_001', 'Buku tidak tersedia');

  const [loan] = await prisma.$transaction([
    prisma.libraryLoan.create({ data: { bookId: data.bookId, borrowerId: data.borrowerId, schoolId, dueDate: new Date(data.dueDate), status: 'BORROWED' } }),
    prisma.libraryBook.update({ where: { id: data.bookId }, data: { availableCopies: { decrement: 1 } } }),
  ]);
  return loan;
}

export async function returnBook(schoolId: string, loanId: string) {
  const loan = await prisma.libraryLoan.findFirst({ where: { id: loanId, schoolId, status: 'BORROWED' } });
  if (!loan) throw new NotFoundError('Peminjaman');

  const returnDate = new Date();
  const fine = returnDate > loan.dueDate ? Math.ceil((returnDate.getTime() - loan.dueDate.getTime()) / (1000 * 60 * 60 * 24)) * 1000 : 0;

  const [updated] = await prisma.$transaction([
    prisma.libraryLoan.update({ where: { id: loanId }, data: { status: 'RETURNED', returnDate, fine } }),
    prisma.libraryBook.update({ where: { id: loan.bookId }, data: { availableCopies: { increment: 1 } } }),
  ]);
  return updated;
}

export async function getLoans(schoolId: string, query: { borrowerId?: string; status?: string; page?: number; limit?: number }) {
  const { page, limit, skip } = getPaginationParams(query);
  const where: Record<string, unknown> = { schoolId };
  if (query.borrowerId) where.borrowerId = query.borrowerId;
  if (query.status) where.status = query.status;
  const [loans, total] = await Promise.all([
    prisma.libraryLoan.findMany({ where, skip, take: limit, orderBy: { createdAt: 'desc' }, include: { book: { select: { title: true, author: true } } } }),
    prisma.libraryLoan.count({ where }),
  ]);
  return { loans, meta: buildPaginationMeta(total, page, limit) };
}
