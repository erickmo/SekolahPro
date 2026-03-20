import { prisma } from '../../lib/prisma';
import { NotFoundError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';

export async function createEvent(schoolId: string, data: {
  title: string;
  description?: string;
  startDate: string;
  endDate: string;
  location?: string;
  type: string;
}) {
  return prisma.schoolEvent.create({
    data: { schoolId, title: data.title, description: data.description, startDate: new Date(data.startDate), endDate: new Date(data.endDate), location: data.location, type: data.type, status: 'PLANNING' },
  });
}

export async function listEvents(schoolId: string, query: { type?: string; status?: string; page?: string; limit?: string }) {
  const { page, limit, skip } = getPaginationParams({ page: Number(query.page) || undefined, limit: Number(query.limit) || undefined });
  const where: Record<string, unknown> = { schoolId };
  if (query.type) where.type = query.type;
  if (query.status) where.status = query.status;
  const [events, total] = await Promise.all([
    prisma.schoolEvent.findMany({ where, skip, take: limit, orderBy: { startDate: 'desc' } }),
    prisma.schoolEvent.count({ where }),
  ]);
  return { events, meta: buildPaginationMeta(total, page, limit) };
}

export async function getEvent(schoolId: string, id: string) {
  const event = await prisma.schoolEvent.findFirst({
    where: { id, schoolId },
    include: { tasks: true, budgets: true, committees: true },
  });
  if (!event) throw new NotFoundError('Event');
  return event;
}

export async function updateEvent(schoolId: string, id: string, data: { title?: string; status?: string; location?: string; description?: string }) {
  const event = await prisma.schoolEvent.findFirst({ where: { id, schoolId } });
  if (!event) throw new NotFoundError('Event');
  return prisma.schoolEvent.update({ where: { id }, data });
}

export async function addTask(schoolId: string, eventId: string, data: { title: string; assigneeId?: string; dueDate?: string }) {
  const event = await prisma.schoolEvent.findFirst({ where: { id: eventId, schoolId } });
  if (!event) throw new NotFoundError('Event');
  return prisma.eventTask.create({
    data: { eventId, title: data.title, assigneeId: data.assigneeId, dueDate: data.dueDate ? new Date(data.dueDate) : undefined, status: 'TODO' },
  });
}

export async function updateTask(schoolId: string, eventId: string, taskId: string, data: { status?: string; title?: string }) {
  const event = await prisma.schoolEvent.findFirst({ where: { id: eventId, schoolId } });
  if (!event) throw new NotFoundError('Event');
  return prisma.eventTask.update({ where: { id: taskId }, data });
}

export async function addBudgetItem(schoolId: string, eventId: string, data: { category: string; planned: number; notes?: string }) {
  const event = await prisma.schoolEvent.findFirst({ where: { id: eventId, schoolId } });
  if (!event) throw new NotFoundError('Event');
  return prisma.eventBudget.create({
    data: { eventId, category: data.category, planned: data.planned, actual: 0, notes: data.notes },
  });
}

export async function addCommitteeMember(schoolId: string, eventId: string, data: { userId: string; role: string }) {
  const event = await prisma.schoolEvent.findFirst({ where: { id: eventId, schoolId } });
  if (!event) throw new NotFoundError('Event');
  return prisma.eventCommittee.create({ data: { eventId, userId: data.userId, role: data.role } });
}
