import { prisma } from '../../lib/prisma';
import { NotFoundError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';

export async function createAlumniProfile(schoolId: string, data: { userId: string; studentId?: string; graduationYear: number; major?: string; university?: string; occupation?: string; company?: string; linkedinUrl?: string }) {
  return prisma.alumniProfile.upsert({ where: { userId: data.userId }, create: { ...data, schoolId }, update: data });
}

export async function getAlumni(schoolId: string, query: { graduationYear?: number; search?: string; page?: number; limit?: number }) {
  const { page, limit, skip } = getPaginationParams(query);
  const where: Record<string, unknown> = { schoolId };
  if (query.graduationYear) where.graduationYear = Number(query.graduationYear);
  if (query.search) where.OR = [{ user: { name: { contains: query.search, mode: 'insensitive' } } }, { company: { contains: query.search, mode: 'insensitive' } }];
  const [alumni, total] = await Promise.all([
    prisma.alumniProfile.findMany({ where, skip, take: limit, orderBy: { graduationYear: 'desc' }, include: { user: { select: { name: true, email: true } } } }),
    prisma.alumniProfile.count({ where }),
  ]);
  return { alumni, meta: buildPaginationMeta(total, page, limit) };
}

export async function createPortfolio(schoolId: string, studentId: string, data: { title: string; description?: string; visibility?: string }) {
  return prisma.portfolio.upsert({ where: { studentId }, create: { studentId, schoolId, title: data.title, description: data.description, visibility: data.visibility || 'SCHOOL' }, update: data });
}

export async function addPortfolioItem(portfolioId: string, data: { type: string; title: string; description?: string; mediaUrl?: string; achievedAt?: string; metadata?: Record<string, unknown> }) {
  return prisma.portfolioItem.create({ data: { ...data, portfolioId, achievedAt: data.achievedAt ? new Date(data.achievedAt) : undefined, metadata: (data.metadata || {}) as never } });
}

export async function getPortfolio(schoolId: string, studentId: string) {
  const portfolio = await prisma.portfolio.findFirst({ where: { studentId, schoolId }, include: { items: { orderBy: { achievedAt: 'desc' } }, student: { select: { name: true, nisn: true } } } });
  if (!portfolio) throw new NotFoundError('Portfolio');
  return portfolio;
}

export async function createTracerStudy(schoolId: string, alumniId: string, data: { academicYear: number; currentStatus: string; institution?: string; income?: number; timeToFirstJob?: number; satisfactionScore?: number; feedback?: string }) {
  return prisma.tracerStudy.create({ data: { ...data, schoolId, alumniId } });
}
