import { prisma } from '../../lib/prisma';
import { NotFoundError, ConflictError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';

export async function getExtraCurriculars(schoolId: string) {
  return prisma.extracurricular.findMany({ where: { schoolId, isActive: true }, include: { _count: { select: { enrollments: true } } } });
}

export async function createExtraCurricular(schoolId: string, data: { name: string; description?: string; supervisorId?: string; schedule?: string; maxMembers?: number; category?: string }) {
  return prisma.extracurricular.create({ data: { ...data, schoolId, isActive: true } });
}

export async function enrollStudent(schoolId: string, data: { studentId: string; extraId: string; academicYearId: string }) {
  const extra = await prisma.extracurricular.findFirst({ where: { id: data.extraId, schoolId } });
  if (!extra) throw new NotFoundError('Ekstrakurikuler');
  const exists = await prisma.extraEnrollment.findFirst({ where: { studentId: data.studentId, extraId: data.extraId, academicYearId: data.academicYearId } });
  if (exists) throw new ConflictError('Siswa sudah terdaftar di ekskul ini');
  return prisma.extraEnrollment.create({ data: { ...data, schoolId, status: 'ACTIVE' } });
}

export async function recordActivity(schoolId: string, data: { extraId: string; date: string; title: string; description?: string; location?: string }) {
  return prisma.extraActivity.create({ data: { ...data, schoolId, date: new Date(data.date) } });
}

export async function recordAttendance(schoolId: string, activityId: string, records: Array<{ studentId: string; status: string }>) {
  const data = records.map((r) => ({ ...r, activityId, schoolId }));
  return prisma.extraAttendance.createMany({ data, skipDuplicates: true });
}

export async function getEnrollments(schoolId: string, query: { extraId?: string; studentId?: string; page?: number; limit?: number }) {
  const { page, limit, skip } = getPaginationParams(query);
  const where: Record<string, unknown> = { schoolId };
  if (query.extraId) where.extraId = query.extraId;
  if (query.studentId) where.studentId = query.studentId;
  const [enrollments, total] = await Promise.all([
    prisma.extraEnrollment.findMany({ where, skip, take: limit, include: { student: { select: { name: true } }, extracurricular: { select: { name: true } } } }),
    prisma.extraEnrollment.count({ where }),
  ]);
  return { enrollments, meta: buildPaginationMeta(total, page, limit) };
}
