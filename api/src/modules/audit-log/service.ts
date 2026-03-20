import { prisma } from '../../lib/prisma';
import { NotFoundError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';

export async function listAuditLogs(schoolId: string | undefined, query: {
  resource?: string;
  userId?: string;
  startDate?: string;
  endDate?: string;
  page?: string;
  limit?: string;
}) {
  const { page, limit, skip } = getPaginationParams({ page: Number(query.page) || undefined, limit: Number(query.limit) || undefined });
  const where: Record<string, unknown> = {};
  if (schoolId) where.schoolId = schoolId;
  if (query.resource) where.resource = query.resource;
  if (query.userId) where.userId = query.userId;
  if (query.startDate || query.endDate) {
    const createdAt: Record<string, unknown> = {};
    if (query.startDate) createdAt.gte = new Date(query.startDate);
    if (query.endDate) createdAt.lte = new Date(query.endDate);
    where.createdAt = createdAt;
  }
  const [logs, total] = await Promise.all([
    prisma.auditLog.findMany({ where, skip, take: limit, orderBy: { createdAt: 'desc' } }),
    prisma.auditLog.count({ where }),
  ]);
  return { logs, meta: buildPaginationMeta(total, page, limit) };
}

export async function getAuditLog(id: string) {
  const log = await prisma.auditLog.findFirst({ where: { id } });
  if (!log) throw new NotFoundError('Audit log');
  return log;
}

export async function logAudit(params: {
  schoolId: string;
  userId: string;
  action: string;
  resource: string;
  resourceId?: string;
  details?: object;
  ipAddress?: string;
}) {
  await prisma.auditLog.create({
    data: {
      schoolId: params.schoolId,
      userId: params.userId,
      action: params.action,
      resource: params.resource,
      resourceId: params.resourceId,
      details: (params.details || {}) as unknown as never,
      ipAddress: params.ipAddress,
    },
  }).catch(() => null);
}
