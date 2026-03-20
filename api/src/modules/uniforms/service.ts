import { prisma } from '../../lib/prisma';
import { NotFoundError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';

export async function createItem(schoolId: string, data: {
  name: string;
  description?: string;
  price: number;
  sizes: string[];
  imageUrl?: string;
  vendorId?: string;
}) {
  return prisma.uniformItem.create({
    data: { schoolId, name: data.name, description: data.description, price: data.price, sizes: data.sizes || [], imageUrl: data.imageUrl, vendorId: data.vendorId, stock: {} as unknown as never, isActive: true },
  });
}

export async function listItems(schoolId: string, query: { page?: string; limit?: string }) {
  const { page, limit, skip } = getPaginationParams({ page: Number(query.page) || undefined, limit: Number(query.limit) || undefined });
  const [items, total] = await Promise.all([
    prisma.uniformItem.findMany({ where: { schoolId, isActive: true }, skip, take: limit, orderBy: { createdAt: 'desc' } }),
    prisma.uniformItem.count({ where: { schoolId, isActive: true } }),
  ]);
  return { items, meta: buildPaginationMeta(total, page, limit) };
}

export async function updateStock(schoolId: string, id: string, stock: object) {
  const item = await prisma.uniformItem.findFirst({ where: { id, schoolId } });
  if (!item) throw new NotFoundError('Item seragam');
  return prisma.uniformItem.update({ where: { id }, data: { stock: stock as unknown as never } });
}

export async function createOrder(schoolId: string, data: {
  itemId: string;
  studentId: string;
  size: string;
  quantity: number;
  totalPrice: number;
}) {
  const item = await prisma.uniformItem.findFirst({ where: { id: data.itemId, schoolId } });
  if (!item) throw new NotFoundError('Item seragam');
  return prisma.uniformOrder.create({
    data: { schoolId, itemId: data.itemId, studentId: data.studentId, size: data.size, quantity: data.quantity, totalPrice: data.totalPrice, status: 'PENDING' },
  });
}

export async function listOrders(schoolId: string, query: { studentId?: string; status?: string; page?: string; limit?: string }) {
  const { page, limit, skip } = getPaginationParams({ page: Number(query.page) || undefined, limit: Number(query.limit) || undefined });
  const where: Record<string, unknown> = { schoolId };
  if (query.studentId) where.studentId = query.studentId;
  if (query.status) where.status = query.status;
  const [orders, total] = await Promise.all([
    prisma.uniformOrder.findMany({ where, skip, take: limit, orderBy: { createdAt: 'desc' }, include: { item: { select: { name: true } } } }),
    prisma.uniformOrder.count({ where }),
  ]);
  return { orders, meta: buildPaginationMeta(total, page, limit) };
}

export async function getOrder(schoolId: string, id: string) {
  const order = await prisma.uniformOrder.findFirst({ where: { id, schoolId }, include: { item: true } });
  if (!order) throw new NotFoundError('Order seragam');
  return order;
}

export async function updateOrderStatus(schoolId: string, id: string, status: string) {
  const order = await prisma.uniformOrder.findFirst({ where: { id, schoolId } });
  if (!order) throw new NotFoundError('Order seragam');
  const updateData: Record<string, unknown> = { status };
  if (status === 'PAID') updateData.paidAt = new Date();
  if (status === 'DELIVERED') updateData.deliveredAt = new Date();
  return prisma.uniformOrder.update({ where: { id }, data: updateData });
}
