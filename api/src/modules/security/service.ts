import { prisma } from '../../lib/prisma';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';

export async function getSecurityDashboard(schoolId: string | undefined) {
  const since24h = new Date(Date.now() - 24 * 60 * 60 * 1000);
  const since7d = new Date(Date.now() - 7 * 24 * 60 * 60 * 1000);
  const schoolFilter = schoolId ? { schoolId } : {};

  const [failedLogins, suspiciousCount, activeUsers, totalUsers] = await Promise.all([
    prisma.auditLog.count({ where: { ...schoolFilter, action: 'LOGIN_FAILED', createdAt: { gte: since24h } } }),
    prisma.auditLog.count({ where: { ...schoolFilter, action: { in: ['LOGIN_FAILED', 'UNAUTHORIZED'] }, createdAt: { gte: since7d } } }),
    prisma.user.count({ where: { ...(schoolId ? { schoolId } : {}), updatedAt: { gte: since24h } } }),
    prisma.user.count({ where: schoolId ? { schoolId } : {} }),
  ]);

  return { failedLogins24h: failedLogins, suspiciousActivity7d: suspiciousCount, activeSessions24h: activeUsers, totalUsers };
}

export async function getSuspiciousActivity(schoolId: string | undefined, query: { page?: string; limit?: string }) {
  const { page, limit, skip } = getPaginationParams({ page: Number(query.page) || undefined, limit: Number(query.limit) || undefined });
  const where: Record<string, unknown> = { action: { in: ['LOGIN_FAILED', 'UNAUTHORIZED'] } };
  if (schoolId) where.schoolId = schoolId;
  const [logs, total] = await Promise.all([
    prisma.auditLog.findMany({ where, skip, take: limit, orderBy: { createdAt: 'desc' } }),
    prisma.auditLog.count({ where }),
  ]);
  return { logs, meta: buildPaginationMeta(total, page, limit) };
}

export async function enableMFA(userId: string) {
  return prisma.user.update({ where: { id: userId }, data: { mfaEnabled: true } });
}

export async function disableMFA(userId: string) {
  return prisma.user.update({ where: { id: userId }, data: { mfaEnabled: false, mfaSecret: null } });
}

export async function getActiveSessions(schoolId: string | undefined, query: { page?: string; limit?: string }) {
  const { page, limit, skip } = getPaginationParams({ page: Number(query.page) || undefined, limit: Number(query.limit) || undefined });
  const since24h = new Date(Date.now() - 24 * 60 * 60 * 1000);
  const where: Record<string, unknown> = { updatedAt: { gte: since24h } };
  if (schoolId) where.schoolId = schoolId;
  const [users, total] = await Promise.all([
    prisma.user.findMany({ where, skip, take: limit, orderBy: { updatedAt: 'desc' }, select: { id: true, name: true, email: true, role: true, schoolId: true, updatedAt: true, mfaEnabled: true } }),
    prisma.user.count({ where }),
  ]);
  return { users, meta: buildPaginationMeta(total, page, limit) };
}
