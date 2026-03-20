import { prisma } from '../../lib/prisma';
import { NotFoundError, BadRequestError, ConflictError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';

export async function createAccount(schoolId: string, studentId: string) {
  const student = await prisma.student.findFirst({ where: { id: studentId, schoolId } });
  if (!student) throw new NotFoundError('Siswa');
  const exists = await prisma.savingsAccount.findUnique({ where: { studentId } });
  if (exists) throw new ConflictError('Akun tabungan sudah ada');
  return prisma.savingsAccount.create({ data: { studentId, schoolId, balance: 0, accountNumber: `SAV-${Date.now()}-${studentId.slice(-4)}` } });
}

export async function getAccount(schoolId: string, studentId: string) {
  const account = await prisma.savingsAccount.findFirst({ where: { studentId, schoolId }, include: { student: { select: { name: true, nisn: true } } } });
  if (!account) throw new NotFoundError('Akun tabungan');
  return account;
}

export async function deposit(schoolId: string, studentId: string, amount: number, note?: string, cashierId?: string) {
  const account = await getAccount(schoolId, studentId);
  const [updated, tx] = await prisma.$transaction([
    prisma.savingsAccount.update({ where: { id: account.id }, data: { balance: { increment: amount } } }),
    prisma.savingsTransaction.create({ data: { accountId: account.id, schoolId, type: 'DEPOSIT', amount, balanceAfter: account.balance + amount, note, cashierId } }),
  ]);
  return { account: updated, transaction: tx };
}

export async function withdraw(schoolId: string, studentId: string, amount: number, note?: string, cashierId?: string) {
  const account = await getAccount(schoolId, studentId);
  const minBalance = 10000;
  if (account.balance - amount < minBalance) throw new BadRequestError('COOP_001', `Saldo tidak mencukupi. Minimum saldo Rp ${minBalance.toLocaleString()}`);
  const [updated, tx] = await prisma.$transaction([
    prisma.savingsAccount.update({ where: { id: account.id }, data: { balance: { decrement: amount } } }),
    prisma.savingsTransaction.create({ data: { accountId: account.id, schoolId, type: 'WITHDRAWAL', amount, balanceAfter: account.balance - amount, note, cashierId } }),
  ]);
  return { account: updated, transaction: tx };
}

export async function getTransactions(schoolId: string, studentId: string, query: { page?: number; limit?: number }) {
  const account = await getAccount(schoolId, studentId);
  const { page, limit, skip } = getPaginationParams(query);
  const [transactions, total] = await Promise.all([
    prisma.savingsTransaction.findMany({ where: { accountId: account.id }, skip, take: limit, orderBy: { createdAt: 'desc' } }),
    prisma.savingsTransaction.count({ where: { accountId: account.id } }),
  ]);
  return { transactions, meta: buildPaginationMeta(total, page, limit) };
}

export async function getProducts(schoolId: string) {
  return prisma.coopProduct.findMany({ where: { schoolId, isActive: true }, orderBy: { name: 'asc' } });
}

export async function createProduct(schoolId: string, data: { name: string; price: number; stock: number; category?: string; imageUrl?: string }) {
  return prisma.coopProduct.create({ data: { ...data, schoolId } });
}

export async function getDailyReport(schoolId: string, date: string) {
  const start = new Date(date);
  const end = new Date(date);
  end.setDate(end.getDate() + 1);

  const transactions = await prisma.savingsTransaction.findMany({
    where: { account: { schoolId }, createdAt: { gte: start, lt: end } },
    include: { account: { include: { student: { select: { name: true } } } } },
  });

  const totalDeposit = transactions.filter((t) => t.type === 'DEPOSIT').reduce((sum, t) => sum + t.amount, 0);
  const totalWithdraw = transactions.filter((t) => t.type === 'WITHDRAWAL').reduce((sum, t) => sum + t.amount, 0);

  return { date, totalDeposit, totalWithdraw, transactionCount: transactions.length, transactions };
}
