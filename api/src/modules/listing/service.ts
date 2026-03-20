import { prisma } from '../../lib/prisma';
import { NotFoundError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';
import { ListingCategory, ListingTier, ListingStatus } from '@prisma/client';

export async function registerVendor(data: {
  businessName: string;
  category: string;
  description: string;
  photos?: string[];
  contactPhone: string;
  contactEmail: string;
  website?: string;
  coverageArea?: string[];
  tier?: string;
}) {
  return prisma.listingVendor.create({
    data: { businessName: data.businessName, category: data.category as ListingCategory, description: data.description, photos: data.photos || [], contactPhone: data.contactPhone, contactEmail: data.contactEmail, website: data.website, coverageArea: data.coverageArea || [], tier: (data.tier as ListingTier) || 'BASIC', status: 'PENDING_REVIEW' as ListingStatus },
  });
}

export async function listVendors(query: { category?: string; status?: string; page?: string; limit?: string }) {
  const { page, limit, skip } = getPaginationParams({ page: Number(query.page) || undefined, limit: Number(query.limit) || undefined });
  const where: Record<string, unknown> = {};
  if (query.category) where.category = query.category;
  if (query.status) where.status = query.status;
  const [vendors, total] = await Promise.all([
    prisma.listingVendor.findMany({ where, skip, take: limit, orderBy: { createdAt: 'desc' } }),
    prisma.listingVendor.count({ where }),
  ]);
  return { vendors, meta: buildPaginationMeta(total, page, limit) };
}

export async function getVendor(id: string) {
  const vendor = await prisma.listingVendor.findFirst({ where: { id }, include: { placements: true, billing: true } });
  if (!vendor) throw new NotFoundError('Vendor listing');
  return vendor;
}

export async function approveVendor(id: string, verifiedBy: string) {
  const vendor = await prisma.listingVendor.findFirst({ where: { id } });
  if (!vendor) throw new NotFoundError('Vendor listing');
  return prisma.listingVendor.update({ where: { id }, data: { status: 'ACTIVE' as ListingStatus, verifiedBy, verifiedAt: new Date() } });
}

export async function createPlacement(vendorId: string, data: { schoolId?: string; moduleId: string; position?: number }) {
  const vendor = await prisma.listingVendor.findFirst({ where: { id: vendorId } });
  if (!vendor) throw new NotFoundError('Vendor listing');
  return prisma.listingPlacement.create({
    data: { vendorId, schoolId: data.schoolId, moduleId: data.moduleId, position: data.position || 0, isActive: true },
  });
}

export async function listPlacements(query: { vendorId?: string; schoolId?: string; page?: string; limit?: string }) {
  const { page, limit, skip } = getPaginationParams({ page: Number(query.page) || undefined, limit: Number(query.limit) || undefined });
  const where: Record<string, unknown> = {};
  if (query.vendorId) where.vendorId = query.vendorId;
  if (query.schoolId) where.schoolId = query.schoolId;
  const [placements, total] = await Promise.all([
    prisma.listingPlacement.findMany({ where, skip, take: limit, orderBy: { position: 'asc' }, include: { vendor: { select: { businessName: true, category: true, tier: true } } } }),
    prisma.listingPlacement.count({ where }),
  ]);
  return { placements, meta: buildPaginationMeta(total, page, limit) };
}

export async function addReview(vendorId: string, schoolId: string, userId: string, data: { rating: number; comment?: string }) {
  const vendor = await prisma.listingVendor.findFirst({ where: { id: vendorId } });
  if (!vendor) throw new NotFoundError('Vendor listing');
  return prisma.listingReview.create({ data: { vendorId, schoolId, userId, rating: data.rating, comment: data.comment } });
}

export async function getActivePlacements(schoolId?: string) {
  const where: Record<string, unknown> = { isActive: true };
  if (schoolId) where.schoolId = schoolId;
  return prisma.listingPlacement.findMany({
    where,
    orderBy: { position: 'asc' },
    include: { vendor: { select: { businessName: true, category: true, tier: true, description: true, contactPhone: true, contactEmail: true, website: true } } },
  });
}
