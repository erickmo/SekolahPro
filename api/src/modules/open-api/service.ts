import crypto from 'crypto';
import { prisma } from '../../lib/prisma';
import { NotFoundError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';

export async function createApiKey(schoolId: string, data: { name: string; scopes: string[]; userId: string }) {
  const key = `eds_${crypto.randomBytes(32).toString('hex')}`;
  return prisma.apiKey.create({ data: { name: data.name, key, schoolId, userId: data.userId, scopes: data.scopes, isActive: true } });
}

export async function getApiKeys(schoolId: string) {
  return prisma.apiKey.findMany({ where: { schoolId }, select: { id: true, name: true, scopes: true, isActive: true, lastUsedAt: true, createdAt: true } });
}

export async function revokeApiKey(schoolId: string, id: string) {
  const apiKey = await prisma.apiKey.findFirst({ where: { id, schoolId } });
  if (!apiKey) throw new NotFoundError('API Key');
  return prisma.apiKey.update({ where: { id }, data: { isActive: false } });
}

export async function validateApiKey(key: string) {
  const apiKey = await prisma.apiKey.findUnique({ where: { key } });
  if (!apiKey || !apiKey.isActive) return null;
  await prisma.apiKey.update({ where: { id: apiKey.id }, data: { lastUsedAt: new Date() } });
  return apiKey;
}

export async function getAuditLogs(schoolId: string, query: { userId?: string; action?: string; resource?: string; page?: number; limit?: number }) {
  const { page, limit, skip } = getPaginationParams(query);
  const where: Record<string, unknown> = { schoolId };
  if (query.userId) where.userId = query.userId;
  if (query.action) where.action = query.action;
  if (query.resource) where.resource = query.resource;
  const [logs, total] = await Promise.all([
    prisma.auditLog.findMany({ where, skip, take: limit, orderBy: { createdAt: 'desc' } }),
    prisma.auditLog.count({ where }),
  ]);
  return { logs, meta: buildPaginationMeta(total, page, limit) };
}

export async function createAuditLog(schoolId: string, data: { userId: string; action: string; resource: string; resourceId?: string; details?: Record<string, unknown>; ipAddress?: string }) {
  return prisma.auditLog.create({ data: { ...data, schoolId, details: (data.details || null) as unknown as never } });
}
