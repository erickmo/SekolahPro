import { prisma } from '../../lib/prisma';
import { NotFoundError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';

export async function createProgram(schoolId: string, data: {
  teacherId: string;
  subject: string;
  gradeLevel: string;
  type: string;
  title: string;
  targetKDs?: string[];
}) {
  return prisma.remediationProgram.create({
    data: { ...data, schoolId, targetKDs: data.targetKDs || [] },
  });
}

export async function listPrograms(schoolId: string, query: { type?: string; page?: string; limit?: string }) {
  const { page, limit, skip } = getPaginationParams({ page: Number(query.page) || undefined, limit: Number(query.limit) || undefined });
  const where: Record<string, unknown> = { schoolId };
  if (query.type) where.type = query.type;
  const [programs, total] = await Promise.all([
    prisma.remediationProgram.findMany({ where, skip, take: limit, orderBy: { createdAt: 'desc' } }),
    prisma.remediationProgram.count({ where }),
  ]);
  return { programs, meta: buildPaginationMeta(total, page, limit) };
}

export async function addSession(schoolId: string, programId: string, data: {
  scheduledAt: string;
  studentIds: string[];
  notes?: string;
}) {
  const program = await prisma.remediationProgram.findFirst({ where: { id: programId, schoolId } });
  if (!program) throw new NotFoundError('Program remedial');
  return prisma.remediationSession.create({
    data: { programId, schoolId, scheduledAt: new Date(data.scheduledAt), studentIds: data.studentIds, attendedIds: [], notes: data.notes },
  });
}

export async function listSessions(schoolId: string, programId: string) {
  const program = await prisma.remediationProgram.findFirst({ where: { id: programId, schoolId } });
  if (!program) throw new NotFoundError('Program remedial');
  return prisma.remediationSession.findMany({ where: { programId, schoolId }, orderBy: { scheduledAt: 'asc' } });
}

export async function updateSession(schoolId: string, id: string, data: { attendedIds?: string[]; notes?: string }) {
  const session = await prisma.remediationSession.findFirst({ where: { id, schoolId } });
  if (!session) throw new NotFoundError('Sesi remedial');
  return prisma.remediationSession.update({ where: { id }, data });
}

export async function getStudentRemediations(schoolId: string, studentId: string) {
  return prisma.remediationSession.findMany({
    where: { schoolId, studentIds: { has: studentId } },
    include: { program: { select: { title: true, subject: true, type: true } } },
    orderBy: { scheduledAt: 'desc' },
  });
}
