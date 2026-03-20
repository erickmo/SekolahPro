import { prisma } from '../../lib/prisma';
import { NotFoundError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';

export async function getTeachers(schoolId: string, query: { search?: string; page?: number; limit?: number }) {
  const { page, limit, skip } = getPaginationParams(query);
  const where: Record<string, unknown> = { schoolId };
  if (query.search) where.name = { contains: query.search, mode: 'insensitive' };
  const [teachers, total] = await Promise.all([
    prisma.teacher.findMany({ where, skip, take: limit, orderBy: { name: 'asc' } }),
    prisma.teacher.count({ where }),
  ]);
  return { teachers, meta: buildPaginationMeta(total, page, limit) };
}

export async function recordTeacherAttendance(schoolId: string, records: Array<{ teacherId: string; date: string; status: string; checkIn?: string; checkOut?: string; note?: string }>) {
  const data = records.map((r) => ({ ...r, date: new Date(r.date), schoolId }));
  return prisma.teacherAttendance.createMany({ data, skipDuplicates: true });
}

export async function getTeacherAttendance(schoolId: string, teacherId: string, query: { startDate?: string; endDate?: string }) {
  const where: Record<string, unknown> = { teacherId, schoolId };
  if (query.startDate || query.endDate) {
    where.date = {};
    if (query.startDate) (where.date as Record<string, unknown>).gte = new Date(query.startDate);
    if (query.endDate) (where.date as Record<string, unknown>).lte = new Date(query.endDate);
  }
  return prisma.teacherAttendance.findMany({ where, orderBy: { date: 'desc' } });
}

export async function getAttendanceSummary(schoolId: string, month: number, year: number) {
  const start = new Date(year, month - 1, 1);
  const end = new Date(year, month, 0);
  const attendances = await prisma.teacherAttendance.groupBy({
    by: ['teacherId', 'status'],
    where: { schoolId, date: { gte: start, lte: end } },
    _count: { status: true },
  });
  return { month, year, summary: attendances };
}
