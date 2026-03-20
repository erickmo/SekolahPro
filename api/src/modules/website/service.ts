import { prisma } from '../../lib/prisma';
import { NotFoundError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';

export async function getArticles(schoolId: string, query: { type?: string; search?: string; page?: number; limit?: number }) {
  const { page, limit, skip } = getPaginationParams(query);
  const where: Record<string, unknown> = { schoolId, isPublished: true };
  if (query.type) where.type = query.type;
  if (query.search) where.title = { contains: query.search, mode: 'insensitive' };
  const [articles, total] = await Promise.all([
    prisma.article.findMany({ where, skip, take: limit, orderBy: { publishedAt: 'desc' }, include: { author: { select: { name: true } } } }),
    prisma.article.count({ where }),
  ]);
  return { articles, meta: buildPaginationMeta(total, page, limit) };
}

export async function getArticle(schoolId: string, slug: string) {
  const article = await prisma.article.findFirst({ where: { schoolId, slug, isPublished: true }, include: { author: { select: { name: true } } } });
  if (!article) throw new NotFoundError('Artikel');
  return article;
}

export async function createArticle(schoolId: string, authorId: string, data: { title: string; content: string; type: string; excerpt?: string; imageUrl?: string; tags?: string[]; isPublished?: boolean }) {
  const slug = data.title.toLowerCase().replace(/\s+/g, '-').replace(/[^a-z0-9-]/g, '') + '-' + Date.now();
  return prisma.article.create({ data: { ...data, schoolId, authorId, slug, tags: data.tags || [], publishedAt: data.isPublished ? new Date() : undefined } });
}

export async function updateArticle(schoolId: string, id: string, data: Partial<{ title: string; content: string; isPublished: boolean; imageUrl: string; excerpt: string }>) {
  const article = await prisma.article.findFirst({ where: { id, schoolId } });
  if (!article) throw new NotFoundError('Artikel');
  return prisma.article.update({ where: { id }, data: { ...data, ...(data.isPublished && !article.publishedAt ? { publishedAt: new Date() } : {}) } });
}

export async function getAnnouncements(schoolId: string) {
  return prisma.announcement.findMany({ where: { schoolId, isActive: true }, orderBy: { createdAt: 'desc' } });
}

export async function createAnnouncement(schoolId: string, data: { title: string; content: string; targetAudience: string; expiresAt?: string; createdBy: string }) {
  return prisma.announcement.create({ data: { ...data, schoolId, isActive: true, expiresAt: data.expiresAt ? new Date(data.expiresAt) : undefined } });
}

export async function getGallery(schoolId: string) {
  return prisma.galleryItem.findMany({ where: { schoolId }, orderBy: { createdAt: 'desc' } });
}

export async function addGalleryItem(schoolId: string, data: { title: string; mediaUrl: string; mediaType: string; category?: string }) {
  return prisma.galleryItem.create({ data: { ...data, schoolId } });
}
