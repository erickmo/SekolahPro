import { prisma } from '../../lib/prisma';
import { NotFoundError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';

export async function getAssets(schoolId: string, query: { category?: string; status?: string; search?: string; page?: number; limit?: number }) {
  const { page, limit, skip } = getPaginationParams(query);
  const where: Record<string, unknown> = { schoolId };
  if (query.category) where.category = query.category;
  if (query.status) where.status = query.status;
  if (query.search) where.name = { contains: query.search, mode: 'insensitive' };
  const [assets, total] = await Promise.all([
    prisma.assetItem.findMany({ where, skip, take: limit, orderBy: { name: 'asc' } }),
    prisma.assetItem.count({ where }),
  ]);
  return { assets, meta: buildPaginationMeta(total, page, limit) };
}

export async function createAsset(schoolId: string, data: { name: string; category: string; code?: string; purchaseDate?: string; purchasePrice?: number; condition: string; location?: string; quantity?: number }) {
  return prisma.assetItem.create({ data: { ...data, schoolId, status: 'ACTIVE', purchaseDate: data.purchaseDate ? new Date(data.purchaseDate) : undefined, quantity: data.quantity || 1 } });
}

export async function updateAsset(schoolId: string, id: string, data: Partial<{ condition: string; status: string; location: string }>) {
  const asset = await prisma.assetItem.findFirst({ where: { id, schoolId } });
  if (!asset) throw new NotFoundError('Aset');
  return prisma.assetItem.update({ where: { id }, data });
}

export async function scheduleMaintenance(schoolId: string, data: { assetId: string; type: string; scheduledDate: string; technician?: string; description?: string; cost?: number }) {
  const asset = await prisma.assetItem.findFirst({ where: { id: data.assetId, schoolId } });
  if (!asset) throw new NotFoundError('Aset');
  return prisma.assetMaintenance.create({ data: { ...data, schoolId, scheduledDate: new Date(data.scheduledDate), status: 'SCHEDULED' } });
}

export async function getMaintenances(schoolId: string, query: { assetId?: string; status?: string; page?: number; limit?: number }) {
  const { page, limit, skip } = getPaginationParams(query);
  const where: Record<string, unknown> = { schoolId };
  if (query.assetId) where.assetId = query.assetId;
  if (query.status) where.status = query.status;
  const [maintenances, total] = await Promise.all([
    prisma.assetMaintenance.findMany({ where, skip, take: limit, orderBy: { scheduledDate: 'desc' }, include: { asset: { select: { name: true, code: true } } } }),
    prisma.assetMaintenance.count({ where }),
  ]);
  return { maintenances, meta: buildPaginationMeta(total, page, limit) };
}
