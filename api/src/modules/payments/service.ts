import { prisma } from '../../lib/prisma';
import { NotFoundError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';
import { config } from '../../config';
import { logger } from '../../lib/logger';

async function createMidtransTransaction(invoice: { id: string; amount: number; description: string; studentName: string }) {
  if (config.midtrans.serverKey.includes('placeholder')) {
    logger.info('[Midtrans Mock] Create transaction', invoice);
    return { token: `mock-token-${invoice.id}`, redirect_url: `https://app.sandbox.midtrans.com/snap/v2/vtweb/mock-${invoice.id}` };
  }
  const authKey = Buffer.from(`${config.midtrans.serverKey}:`).toString('base64');
  const baseUrl = config.midtrans.isProduction ? 'https://app.midtrans.com' : 'https://app.sandbox.midtrans.com';
  const res = await fetch(`${baseUrl}/snap/v1/transactions`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Authorization: `Basic ${authKey}` },
    body: JSON.stringify({ transaction_details: { order_id: invoice.id, gross_amount: invoice.amount }, customer_details: { first_name: invoice.studentName } }),
  });
  return res.json();
}

export async function generateMonthlyInvoices(schoolId: string, semesterId: string, amount: number, dueDate: string) {
  const students = await prisma.student.findMany({ where: { schoolId, isActive: true } });
  const invoices = await Promise.all(
    students.map((s) =>
      prisma.invoice.create({
        data: {
          schoolId,
          studentId: s.id,
          type: 'SPP',
          amount,
          dueDate: new Date(dueDate),
          semesterId,
          status: 'UNPAID',
          invoiceNumber: `INV-${Date.now()}-${s.id.slice(-4)}`,
        },
      }).catch(() => null),
    ),
  );
  return invoices.filter(Boolean);
}

export async function getInvoices(schoolId: string, query: { studentId?: string; status?: string; page?: number; limit?: number }) {
  const { page, limit, skip } = getPaginationParams(query);
  const where: Record<string, unknown> = { schoolId };
  if (query.studentId) where.studentId = query.studentId;
  if (query.status) where.status = query.status;
  const [invoices, total] = await Promise.all([
    prisma.invoice.findMany({ where, skip, take: limit, orderBy: { createdAt: 'desc' }, include: { student: { select: { name: true, nisn: true } } } }),
    prisma.invoice.count({ where }),
  ]);
  return { invoices, meta: buildPaginationMeta(total, page, limit) };
}

export async function createPaymentToken(schoolId: string, invoiceId: string) {
  const invoice = await prisma.invoice.findFirst({ where: { id: invoiceId, schoolId }, include: { student: { select: { name: true } } } });
  if (!invoice) throw new NotFoundError('Invoice');
  const snap = await createMidtransTransaction({ id: invoice.id, amount: invoice.amount, description: `SPP ${invoice.student.name}`, studentName: invoice.student.name });
  return snap;
}

export async function handlePaymentCallback(data: { order_id: string; transaction_status: string; payment_type: string; gross_amount: string }) {
  const invoice = await prisma.invoice.findUnique({ where: { id: data.order_id } });
  if (!invoice) return;

  const statusMap: Record<string, string> = { settlement: 'PAID', capture: 'PAID', expire: 'OVERDUE', cancel: 'CANCELLED', pending: 'UNPAID' };
  const status = statusMap[data.transaction_status] || 'UNPAID';

  await prisma.$transaction([
    prisma.invoice.update({ where: { id: invoice.id }, data: { status: status as never, paidAt: status === 'PAID' ? new Date() : undefined } }),
    ...(status === 'PAID'
      ? [prisma.payment.create({ data: { invoiceId: invoice.id, schoolId: invoice.schoolId, amount: Number(data.gross_amount), method: data.payment_type, status: 'SUCCESS', transactionId: data.order_id } })]
      : []),
  ]);
}

export async function getFinancialSummary(schoolId: string, query: { startDate?: string; endDate?: string }) {
  const where: Record<string, unknown> = { schoolId, status: 'PAID' };
  if (query.startDate || query.endDate) {
    where.paidAt = {};
    if (query.startDate) (where.paidAt as Record<string, unknown>).gte = new Date(query.startDate);
    if (query.endDate) (where.paidAt as Record<string, unknown>).lte = new Date(query.endDate);
  }
  const paidInvoices = await prisma.invoice.findMany({ where, select: { amount: true } });
  const totalPaid = paidInvoices.reduce((s, i) => s + i.amount, 0);
  const allInvoices = await prisma.invoice.findMany({ where: { schoolId }, select: { amount: true } });
  const totalBilled = allInvoices.reduce((s, i) => s + i.amount, 0);
  const unpaidCount = await prisma.invoice.count({ where: { schoolId, status: 'UNPAID' } });
  return { totalBilled, totalPaid, unpaidCount };
}
