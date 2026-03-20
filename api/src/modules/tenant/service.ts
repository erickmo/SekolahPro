import { prisma } from '../../lib/prisma';
import { NotFoundError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';
import { TenantTier, TenantStatus } from '@prisma/client';

export async function createTenant(data: {
  schoolId: string;
  subdomain: string;
  tier?: string;
  billingEmail: string;
  trialEndsAt?: string;
}) {
  return prisma.tenant.create({
    data: { schoolId: data.schoolId, subdomain: data.subdomain, tier: (data.tier as TenantTier) || 'FREE', status: 'TRIAL', billingEmail: data.billingEmail, trialEndsAt: data.trialEndsAt ? new Date(data.trialEndsAt) : undefined },
  });
}

export async function listTenants(query: { tier?: string; status?: string; page?: string; limit?: string }) {
  const { page, limit, skip } = getPaginationParams({ page: Number(query.page) || undefined, limit: Number(query.limit) || undefined });
  const where: Record<string, unknown> = {};
  if (query.tier) where.tier = query.tier;
  if (query.status) where.status = query.status;
  const [tenants, total] = await Promise.all([
    prisma.tenant.findMany({ where, skip, take: limit, orderBy: { createdAt: 'desc' } }),
    prisma.tenant.count({ where }),
  ]);
  return { tenants, meta: buildPaginationMeta(total, page, limit) };
}

export async function getTenant(id: string) {
  const tenant = await prisma.tenant.findFirst({ where: { id }, include: { features: true, subscription: true, addons: true } });
  if (!tenant) throw new NotFoundError('Tenant');
  return tenant;
}

export async function upgradePlan(id: string, tier: string) {
  const tenant = await prisma.tenant.findFirst({ where: { id } });
  if (!tenant) throw new NotFoundError('Tenant');
  return prisma.tenant.update({ where: { id }, data: { tier: tier as TenantTier, status: 'ACTIVE' as TenantStatus } });
}

export async function updateStatus(id: string, status: string) {
  const tenant = await prisma.tenant.findFirst({ where: { id } });
  if (!tenant) throw new NotFoundError('Tenant');
  return prisma.tenant.update({ where: { id }, data: { status: status as TenantStatus } });
}

export async function getFeatures(id: string) {
  const tenant = await prisma.tenant.findFirst({ where: { id } });
  if (!tenant) throw new NotFoundError('Tenant');
  return prisma.tenantFeature.findMany({ where: { tenantId: id } });
}

export async function toggleFeature(id: string, moduleId: string, enabled: boolean, updatedBy: string, config?: object) {
  const tenant = await prisma.tenant.findFirst({ where: { id } });
  if (!tenant) throw new NotFoundError('Tenant');
  return prisma.tenantFeature.upsert({
    where: { tenantId_moduleId: { tenantId: id, moduleId } },
    create: { tenantId: id, moduleId, enabled, config: config as unknown as never, updatedBy },
    update: { enabled, config: config as unknown as never, updatedBy },
  });
}

export async function addAddon(tenantId: string, data: { addon: string; quantity?: number; activeFrom: string; activeTo: string }) {
  const tenant = await prisma.tenant.findFirst({ where: { id: tenantId } });
  if (!tenant) throw new NotFoundError('Tenant');
  return prisma.tenantAddon.create({
    data: { tenantId, addon: data.addon, quantity: data.quantity || 1, activeFrom: new Date(data.activeFrom), activeTo: new Date(data.activeTo) },
  });
}

export async function getMyTenant(schoolId: string) {
  const tenant = await prisma.tenant.findFirst({ where: { schoolId }, include: { features: true, subscription: true, addons: true } });
  if (!tenant) throw new NotFoundError('Tenant');
  return tenant;
}
