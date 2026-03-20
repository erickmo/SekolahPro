import { prisma } from '../../lib/prisma';
import { NotFoundError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';
import { EduContentType, EduContentStatus } from '@prisma/client';

export async function createContent(schoolOrigin: string, authorId: string, data: {
  title: string;
  description: string;
  type: string;
  price?: number;
  licenseType?: string;
  fileUrl?: string;
  previewUrl?: string;
}) {
  return prisma.eduContent.create({
    data: { schoolOrigin, authorId, title: data.title, description: data.description, type: data.type as EduContentType, price: data.price || 0, licenseType: data.licenseType || 'SINGLE_SCHOOL', fileUrl: data.fileUrl, previewUrl: data.previewUrl, status: 'PENDING_REVIEW' as EduContentStatus },
  });
}

export async function listContent(query: { type?: string; page?: string; limit?: string }) {
  const { page, limit, skip } = getPaginationParams({ page: Number(query.page) || undefined, limit: Number(query.limit) || undefined });
  const where: Record<string, unknown> = { status: 'APPROVED' };
  if (query.type) where.type = query.type;
  const [content, total] = await Promise.all([
    prisma.eduContent.findMany({ where, skip, take: limit, orderBy: { createdAt: 'desc' }, select: { id: true, title: true, description: true, type: true, price: true, previewUrl: true, rating: true, downloads: true, authorId: true, schoolOrigin: true, createdAt: true } }),
    prisma.eduContent.count({ where }),
  ]);
  return { content, meta: buildPaginationMeta(total, page, limit) };
}

export async function getContent(id: string) {
  const content = await prisma.eduContent.findFirst({ where: { id } });
  if (!content) throw new NotFoundError('Konten');
  return content;
}

export async function publishContent(id: string, userId: string) {
  const content = await prisma.eduContent.findFirst({ where: { id, authorId: userId } });
  if (!content) throw new NotFoundError('Konten');
  return prisma.eduContent.update({ where: { id }, data: { status: 'APPROVED' as EduContentStatus } });
}

export async function purchaseContent(contentId: string, schoolId: string, buyerId: string) {
  const content = await prisma.eduContent.findFirst({ where: { id: contentId } });
  if (!content) throw new NotFoundError('Konten');
  const purchase = await prisma.contentPurchase.create({
    data: { contentId, schoolId, buyerId, amount: content.price, royaltyPaid: Number(content.price) * 0.7 },
  });
  await prisma.eduContent.update({ where: { id: contentId }, data: { downloads: { increment: 1 } } });
  return purchase;
}

export async function getMyPurchases(schoolId: string, buyerId: string) {
  return prisma.contentPurchase.findMany({
    where: { schoolId, buyerId },
    include: { content: { select: { title: true, type: true, fileUrl: true } } },
    orderBy: { createdAt: 'desc' },
  });
}

export async function getMyContent(authorId: string) {
  return prisma.eduContent.findMany({
    where: { authorId },
    orderBy: { createdAt: 'desc' },
  });
}
