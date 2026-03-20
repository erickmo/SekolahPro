import { prisma } from '../../lib/prisma';
import { NotFoundError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';

export async function createLessonPlan(schoolId: string, data: {
  teacherId: string;
  subject: string;
  gradeLevel: string;
  topic: string;
  curriculum: string;
  durationMins: number;
  content: object;
}) {
  return prisma.lessonPlan.create({
    data: { ...data, schoolId, status: 'DRAFT', content: data.content as unknown as never },
  });
}

export async function listLessonPlans(schoolId: string, query: { teacherId?: string; subject?: string; status?: string; page?: string; limit?: string }) {
  const { page, limit, skip } = getPaginationParams({ page: Number(query.page) || undefined, limit: Number(query.limit) || undefined });
  const where: Record<string, unknown> = { schoolId };
  if (query.teacherId) where.teacherId = query.teacherId;
  if (query.subject) where.subject = query.subject;
  if (query.status) where.status = query.status;
  const [plans, total] = await Promise.all([
    prisma.lessonPlan.findMany({ where, skip, take: limit, orderBy: { createdAt: 'desc' } }),
    prisma.lessonPlan.count({ where }),
  ]);
  return { plans, meta: buildPaginationMeta(total, page, limit) };
}

export async function getLessonPlan(schoolId: string, id: string) {
  const plan = await prisma.lessonPlan.findFirst({ where: { id, schoolId }, include: { versions: { orderBy: { version: 'desc' } } } });
  if (!plan) throw new NotFoundError('RPP');
  return plan;
}

export async function updateLessonPlan(schoolId: string, id: string, data: { topic?: string; curriculum?: string; durationMins?: number; content?: object }) {
  const plan = await prisma.lessonPlan.findFirst({ where: { id, schoolId } });
  if (!plan) throw new NotFoundError('RPP');
  const updateData: Record<string, unknown> = { ...data };
  if (data.content) updateData.content = data.content as unknown as never;
  return prisma.lessonPlan.update({ where: { id }, data: updateData });
}

export async function publishLessonPlan(schoolId: string, id: string, reviewedBy: string) {
  const plan = await prisma.lessonPlan.findFirst({ where: { id, schoolId } });
  if (!plan) throw new NotFoundError('RPP');
  const versionCount = await prisma.lessonPlanVersion.count({ where: { lessonPlanId: id } });
  await prisma.lessonPlanVersion.create({
    data: { lessonPlanId: id, version: versionCount + 1, content: plan.content as unknown as never, savedBy: reviewedBy },
  });
  return prisma.lessonPlan.update({ where: { id }, data: { status: 'PUBLISHED', reviewedBy, reviewedAt: new Date() } });
}

export async function getVersions(schoolId: string, id: string) {
  const plan = await prisma.lessonPlan.findFirst({ where: { id, schoolId } });
  if (!plan) throw new NotFoundError('RPP');
  return prisma.lessonPlanVersion.findMany({ where: { lessonPlanId: id }, orderBy: { version: 'desc' } });
}
