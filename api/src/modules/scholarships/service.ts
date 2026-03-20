import { prisma } from '../../lib/prisma';
import { NotFoundError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';

export async function getPrograms(schoolId: string) {
  return prisma.scholarshipProgram.findMany({ where: { schoolId, isActive: true }, orderBy: { createdAt: 'desc' } });
}

export async function createProgram(schoolId: string, data: { name: string; description?: string; type: string; amount?: number; quota?: number; criteria?: Record<string, unknown>; deadline?: string; academicYearId: string }) {
  return prisma.scholarshipProgram.create({ data: { ...data, schoolId, isActive: true, deadline: data.deadline ? new Date(data.deadline) : undefined, criteria: (data.criteria || {}) as unknown as never } });
}

export async function applyScholarship(schoolId: string, data: { programId: string; studentId: string; essay?: string; documents?: Record<string, string> }) {
  const program = await prisma.scholarshipProgram.findFirst({ where: { id: data.programId, schoolId } });
  if (!program) throw new NotFoundError('Program beasiswa');
  return prisma.scholarshipApplication.create({ data: { ...data, schoolId, status: 'SUBMITTED', documents: (data.documents || {}) as unknown as never } });
}

export async function getApplications(schoolId: string, query: { programId?: string; status?: string; page?: number; limit?: number }) {
  const { page, limit, skip } = getPaginationParams(query);
  const where: Record<string, unknown> = { schoolId };
  if (query.programId) where.programId = query.programId;
  if (query.status) where.status = query.status;
  const [applications, total] = await Promise.all([
    prisma.scholarshipApplication.findMany({ where, skip, take: limit, orderBy: { createdAt: 'desc' }, include: { student: { select: { name: true, nisn: true } }, program: { select: { name: true, type: true } } } }),
    prisma.scholarshipApplication.count({ where }),
  ]);
  return { applications, meta: buildPaginationMeta(total, page, limit) };
}

export async function updateApplicationStatus(schoolId: string, id: string, status: string, reviewNotes?: string) {
  const app = await prisma.scholarshipApplication.findFirst({ where: { id, schoolId } });
  if (!app) throw new NotFoundError('Pendaftaran beasiswa');
  return prisma.scholarshipApplication.update({ where: { id }, data: { status: status as never, reviewNotes } });
}
