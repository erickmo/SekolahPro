import { prisma } from '../../lib/prisma';
import { ConflictError, NotFoundError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';
import { CreateStudentInput, UpdateStudentInput } from './dto';

export async function createStudent(schoolId: string, input: CreateStudentInput) {
  const exists = await prisma.student.findFirst({ where: { nisn: input.nisn, schoolId } });
  if (exists) throw new ConflictError('STUDENT_001: NISN sudah terdaftar');

  const { guardians, ...studentData } = input;
  return prisma.student.create({
    data: {
      ...studentData,
      birthDate: new Date(studentData.birthDate),
      schoolId,
      guardians: guardians ? { create: guardians } : undefined,
    },
    include: { guardians: true },
  });
}

export async function getStudents(schoolId: string, query: { page?: number; limit?: number; search?: string; classId?: string; gender?: string }) {
  const { page, limit, skip } = getPaginationParams(query);
  const where: Record<string, unknown> = { schoolId };
  if (query.search) where.name = { contains: query.search, mode: 'insensitive' };
  if (query.gender) where.gender = query.gender;

  const [students, total] = await Promise.all([
    prisma.student.findMany({ where, skip, take: limit, orderBy: { name: 'asc' }, include: { guardians: true, _count: { select: { enrollments: true } } } }),
    prisma.student.count({ where }),
  ]);
  return { students, meta: buildPaginationMeta(total, page, limit) };
}

export async function getStudent(schoolId: string, id: string) {
  const student = await prisma.student.findFirst({
    where: { id, schoolId },
    include: {
      guardians: true,
      enrollments: { include: { schoolClass: true, academicYear: true } },
      studentCard: true,
      savingsAccount: true,
    },
  });
  if (!student) throw new NotFoundError('Siswa');
  return student;
}

export async function updateStudent(schoolId: string, id: string, input: UpdateStudentInput) {
  await getStudent(schoolId, id);
  const { guardians, ...data } = input;
  return prisma.student.update({
    where: { id },
    data: { ...data, ...(data.birthDate ? { birthDate: new Date(data.birthDate) } : {}) },
  });
}

export async function deleteStudent(schoolId: string, id: string) {
  await getStudent(schoolId, id);
  return prisma.student.delete({ where: { id } });
}

export async function importStudents(schoolId: string, students: CreateStudentInput[]) {
  const results = await Promise.allSettled(students.map((s) => createStudent(schoolId, s)));
  const created = results.filter((r) => r.status === 'fulfilled').length;
  const failed = results.filter((r) => r.status === 'rejected').length;
  return { created, failed, total: students.length };
}

export async function getStudentAttendance(schoolId: string, studentId: string, query: { startDate?: string; endDate?: string }) {
  await getStudent(schoolId, studentId);
  const where: Record<string, unknown> = { studentId };
  if (query.startDate || query.endDate) {
    where.date = {};
    if (query.startDate) (where.date as Record<string, unknown>).gte = new Date(query.startDate);
    if (query.endDate) (where.date as Record<string, unknown>).lte = new Date(query.endDate);
  }
  return prisma.studentAttendance.findMany({ where, orderBy: { date: 'desc' } });
}

export async function recordAttendance(schoolId: string, records: Array<{ studentId: string; date: string; status: string; scheduleId?: string; note?: string }>) {
  const data = records.map((r) => ({ ...r, date: new Date(r.date), schoolId }));
  return prisma.studentAttendance.createMany({ data, skipDuplicates: true });
}
