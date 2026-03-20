import { prisma } from '../../lib/prisma';
import { NotFoundError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';

export async function getPlacements(schoolId: string, query: { studentId?: string; status?: string; page?: number; limit?: number }) {
  const { page, limit, skip } = getPaginationParams(query);
  const where: Record<string, unknown> = { schoolId };
  if (query.studentId) where.studentId = query.studentId;
  if (query.status) where.status = query.status;
  const [placements, total] = await Promise.all([
    prisma.internshipPlacement.findMany({ where, skip, take: limit, orderBy: { startDate: 'desc' }, include: { student: { select: { name: true } } } }),
    prisma.internshipPlacement.count({ where }),
  ]);
  return { placements, meta: buildPaginationMeta(total, page, limit) };
}

export async function createPlacement(schoolId: string, data: { studentId: string; companyName: string; companyAddress: string; supervisorName: string; supervisorContact: string; startDate: string; endDate: string; field: string; mentorId?: string }) {
  return prisma.internshipPlacement.create({ data: { ...data, schoolId, startDate: new Date(data.startDate), endDate: new Date(data.endDate), status: 'ACTIVE' } });
}

export async function addJournalEntry(schoolId: string, placementId: string, studentId: string, data: { date: string; activities: string; reflection?: string; photos?: string[] }) {
  const placement = await prisma.internshipPlacement.findFirst({ where: { id: placementId, studentId, schoolId } });
  if (!placement) throw new NotFoundError('Penempatan prakerin');
  return prisma.internshipJournal.create({ data: { placementId, date: new Date(data.date), activities: data.activities, reflection: data.reflection, photos: data.photos || [] } });
}

export async function getJournals(schoolId: string, placementId: string) {
  return prisma.internshipJournal.findMany({ where: { placement: { id: placementId, schoolId } }, orderBy: { date: 'desc' } });
}

export async function submitGrade(schoolId: string, placementId: string, data: { score: number; feedback?: string; competencies?: Record<string, number> }) {
  const placement = await prisma.internshipPlacement.findFirst({ where: { id: placementId, schoolId } });
  if (!placement) throw new NotFoundError('Penempatan prakerin');
  return prisma.internshipPlacement.update({ where: { id: placementId }, data: { score: data.score, supervisorFeedback: data.feedback, status: 'COMPLETED' } });
}
