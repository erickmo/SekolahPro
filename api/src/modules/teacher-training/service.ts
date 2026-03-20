import { prisma } from '../../lib/prisma';
import { NotFoundError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';

export async function getPrograms(schoolId: string) {
  return prisma.trainingProgram.findMany({ where: { schoolId, isActive: true }, orderBy: { startDate: 'desc' } });
}

export async function createProgram(schoolId: string, data: { title: string; description?: string; type: string; provider?: string; startDate: string; endDate: string; maxParticipants?: number; cpdPoints?: number; isOnline?: boolean }) {
  return prisma.trainingProgram.create({ data: { ...data, schoolId, isActive: true, startDate: new Date(data.startDate), endDate: new Date(data.endDate), maxParticipants: data.maxParticipants || 30, cpdPoints: data.cpdPoints || 0, isOnline: data.isOnline || false } });
}

export async function enroll(schoolId: string, programId: string, teacherId: string) {
  const program = await prisma.trainingProgram.findFirst({ where: { id: programId, schoolId } });
  if (!program) throw new NotFoundError('Program pelatihan');
  return prisma.trainingEnrollment.upsert({ where: { programId_teacherId: { programId, teacherId } }, create: { programId, teacherId, schoolId, status: 'ENROLLED' }, update: {} });
}

export async function getEnrollments(schoolId: string, query: { teacherId?: string; programId?: string; status?: string; page?: number; limit?: number }) {
  const { page, limit, skip } = getPaginationParams(query);
  const where: Record<string, unknown> = { schoolId };
  if (query.teacherId) where.teacherId = query.teacherId;
  if (query.programId) where.programId = query.programId;
  if (query.status) where.status = query.status;
  const [enrollments, total] = await Promise.all([
    prisma.trainingEnrollment.findMany({ where, skip, take: limit, include: { program: { select: { title: true, type: true, cpdPoints: true } }, teacher: { select: { name: true } } } }),
    prisma.trainingEnrollment.count({ where }),
  ]);
  return { enrollments, meta: buildPaginationMeta(total, page, limit) };
}

export async function completTraining(schoolId: string, enrollmentId: string, data: { certificateUrl?: string; score?: number }) {
  const enrollment = await prisma.trainingEnrollment.findFirst({ where: { id: enrollmentId, schoolId } });
  if (!enrollment) throw new NotFoundError('Enrollment');
  return prisma.trainingEnrollment.update({ where: { id: enrollmentId }, data: { status: 'COMPLETED', completedAt: new Date(), ...data } });
}
