import { prisma } from '../../lib/prisma';
import { NotFoundError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';

export async function bookSession(schoolId: string, data: { studentId: string; counselorId: string; scheduledAt: string; topic?: string }) {
  return prisma.counselingSession.create({ data: { ...data, schoolId, scheduledAt: new Date(data.scheduledAt), status: 'SCHEDULED' } });
}

export async function getSessions(schoolId: string, query: { counselorId?: string; studentId?: string; status?: string; page?: number; limit?: number }) {
  const { page, limit, skip } = getPaginationParams(query);
  const where: Record<string, unknown> = { schoolId };
  if (query.counselorId) where.counselorId = query.counselorId;
  if (query.studentId) where.studentId = query.studentId;
  if (query.status) where.status = query.status;
  const [sessions, total] = await Promise.all([
    prisma.counselingSession.findMany({ where, skip, take: limit, orderBy: { scheduledAt: 'desc' }, include: { student: { select: { name: true } } } }),
    prisma.counselingSession.count({ where }),
  ]);
  return { sessions, meta: buildPaginationMeta(total, page, limit) };
}

export async function updateSession(schoolId: string, id: string, data: { notes?: string; status?: string; outcome?: string }) {
  const session = await prisma.counselingSession.findFirst({ where: { id, schoolId } });
  if (!session) throw new NotFoundError('Sesi konseling');
  return prisma.counselingSession.update({ where: { id }, data: { notes: data.notes, outcome: data.outcome, status: data.status as never, ...(data.status === 'COMPLETED' ? { completedAt: new Date() } : {}) } });
}

export async function reportBullying(schoolId: string, data: { reporterId?: string; victimId?: string; perpetratorDesc?: string; type: string; description: string; location?: string; isAnonymous?: boolean }) {
  return prisma.bullyingReport.create({ data: { ...data, schoolId, status: 'REPORTED', incidentDate: new Date() } });
}

export async function getBullyingReports(schoolId: string, query: { status?: string; page?: number; limit?: number }) {
  const { page, limit, skip } = getPaginationParams(query);
  const where: Record<string, unknown> = { schoolId };
  if (query.status) where.status = query.status;
  const [reports, total] = await Promise.all([
    prisma.bullyingReport.findMany({ where, skip, take: limit, orderBy: { createdAt: 'desc' } }),
    prisma.bullyingReport.count({ where }),
  ]);
  return { reports, meta: buildPaginationMeta(total, page, limit) };
}

export async function updateBullyingReport(schoolId: string, id: string, data: { status?: string; actionTaken?: string; handledBy?: string; resolvedAt?: string }) {
  const report = await prisma.bullyingReport.findFirst({ where: { id, schoolId } });
  if (!report) throw new NotFoundError('Laporan');
  return prisma.bullyingReport.update({ where: { id }, data: { ...data, ...(data.resolvedAt ? { resolvedAt: new Date(data.resolvedAt) } : {}) } });
}
