import { prisma } from '../../lib/prisma';
import { NotFoundError, BadRequestError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';

export async function getMenuItems(schoolId: string, date?: string) {
  const where: Record<string, unknown> = { schoolId, isAvailable: true };
  if (date) {
    const d = new Date(date);
    const next = new Date(date);
    next.setDate(next.getDate() + 1);
    where.availableDate = { gte: d, lt: next };
  }
  return prisma.menuItem.findMany({ where, orderBy: { category: 'asc' } });
}

export async function addMenuItem(schoolId: string, data: { name: string; price: number; category: string; imageUrl?: string; availableDate?: string; stock?: number; nutritionInfo?: Record<string, unknown> }) {
  return prisma.menuItem.create({ data: { ...data, schoolId, isAvailable: true, availableDate: data.availableDate ? new Date(data.availableDate) : undefined, stock: data.stock || 100, nutritionInfo: (data.nutritionInfo || {}) as never } });
}

export async function placeOrder(schoolId: string, data: { studentId: string; items: Array<{ menuItemId: string; quantity: number }>; paymentMethod?: string }) {
  const wallet = await prisma.cashlessWallet.findFirst({ where: { ownerId: data.studentId, schoolId } });

  let totalAmount = 0;
  const itemDetails = await Promise.all(
    data.items.map(async (item) => {
      const menu = await prisma.menuItem.findFirst({ where: { id: item.menuItemId, schoolId } });
      if (!menu) throw new NotFoundError(`Menu ${item.menuItemId}`);
      if (menu.stock < item.quantity) throw new BadRequestError('CANTEEN_001', `Stok ${menu.name} tidak cukup`);
      totalAmount += menu.price * item.quantity;
      return { menuItemId: item.menuItemId, quantity: item.quantity, unitPrice: menu.price };
    }),
  );

  if (wallet && (data.paymentMethod === 'WALLET' || !data.paymentMethod)) {
    if (wallet.balance < totalAmount) throw new BadRequestError('CANTEEN_002', 'Saldo EDS Wallet tidak cukup');
  }

  const order = await prisma.$transaction(async (tx) => {
    const o = await tx.menuOrder.create({ data: { schoolId, studentId: data.studentId, totalAmount, status: 'PENDING', paymentMethod: data.paymentMethod || 'WALLET', items: { create: itemDetails } } });
    for (const item of data.items) {
      await tx.menuItem.update({ where: { id: item.menuItemId }, data: { stock: { decrement: item.quantity } } });
    }
    if (wallet && (data.paymentMethod === 'WALLET' || !data.paymentMethod)) {
      await tx.cashlessWallet.update({ where: { id: wallet.id }, data: { balance: { decrement: totalAmount } } });
      await tx.cashlessTransaction.create({ data: { walletId: wallet.id, type: 'DEBIT', amount: totalAmount, description: `Order kantin #${o.id}`, schoolId } });
    }
    return o;
  });
  return order;
}

export async function getOrders(schoolId: string, query: { studentId?: string; status?: string; date?: string; page?: number; limit?: number }) {
  const { page, limit, skip } = getPaginationParams(query);
  const where: Record<string, unknown> = { schoolId };
  if (query.studentId) where.studentId = query.studentId;
  if (query.status) where.status = query.status;
  const [orders, total] = await Promise.all([
    prisma.menuOrder.findMany({ where, skip, take: limit, orderBy: { createdAt: 'desc' }, include: { items: { include: { menuItem: { select: { name: true } } } }, student: { select: { name: true } } } }),
    prisma.menuOrder.count({ where }),
  ]);
  return { orders, meta: buildPaginationMeta(total, page, limit) };
}

export async function topUpWallet(schoolId: string, ownerId: string, amount: number, topUpBy: string) {
  let wallet = await prisma.cashlessWallet.findFirst({ where: { ownerId, schoolId } });
  if (!wallet) wallet = await prisma.cashlessWallet.create({ data: { ownerId, schoolId, balance: 0 } });

  const [updated] = await prisma.$transaction([
    prisma.cashlessWallet.update({ where: { id: wallet.id }, data: { balance: { increment: amount } } }),
    prisma.cashlessTransaction.create({ data: { walletId: wallet.id, type: 'CREDIT', amount, description: `Top up oleh ${topUpBy}`, schoolId } }),
  ]);
  return updated;
}

export async function getWallet(schoolId: string, ownerId: string) {
  return prisma.cashlessWallet.findFirst({ where: { ownerId, schoolId } });
}
