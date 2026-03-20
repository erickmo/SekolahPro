import { prisma } from '../../lib/prisma';
import { NotFoundError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';

export async function createTicket(schoolId: string, data: { reporterId: string; title: string; description: string; category: string; location?: string; photoUrl?: string; priority?: string }) {
  const ticketNumber = `TKT-${Date.now()}`;
  return prisma.helpdeskTicket.create({ data: { ...data, schoolId, ticketNumber, status: 'OPEN', priority: data.priority || 'MEDIUM' } });
}

export async function getTickets(schoolId: string, query: { status?: string; category?: string; assignedTo?: string; page?: number; limit?: number }) {
  const { page, limit, skip } = getPaginationParams(query);
  const where: Record<string, unknown> = { schoolId };
  if (query.status) where.status = query.status;
  if (query.category) where.category = query.category;
  if (query.assignedTo) where.assignedTo = query.assignedTo;
  const [tickets, total] = await Promise.all([
    prisma.helpdeskTicket.findMany({ where, skip, take: limit, orderBy: { createdAt: 'desc' }, include: { reporter: { select: { name: true } } } }),
    prisma.helpdeskTicket.count({ where }),
  ]);
  return { tickets, meta: buildPaginationMeta(total, page, limit) };
}

export async function updateTicket(schoolId: string, id: string, data: { status?: string; assignedTo?: string; resolution?: string; satisfactionRating?: number }) {
  const ticket = await prisma.helpdeskTicket.findFirst({ where: { id, schoolId } });
  if (!ticket) throw new NotFoundError('Tiket');
  return prisma.helpdeskTicket.update({ where: { id }, data: { ...data, status: data.status as never, ...(data.status === 'RESOLVED' ? { resolvedAt: new Date() } : {}) } });
}

export async function getTicketStats(schoolId: string) {
  const [open, inProgress, resolved, total] = await Promise.all([
    prisma.helpdeskTicket.count({ where: { schoolId, status: 'OPEN' } }),
    prisma.helpdeskTicket.count({ where: { schoolId, status: 'IN_PROGRESS' } }),
    prisma.helpdeskTicket.count({ where: { schoolId, status: 'RESOLVED' } }),
    prisma.helpdeskTicket.count({ where: { schoolId } }),
  ]);
  return { open, inProgress, resolved, total };
}
