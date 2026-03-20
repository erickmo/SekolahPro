import { prisma } from '../../lib/prisma';
import { NotFoundError, ConflictError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';

// Academic Years
export async function createAcademicYear(schoolId: string, data: { name: string; startDate: string; endDate: string; isCurrent?: boolean }) {
  if (data.isCurrent) await prisma.academicYear.updateMany({ where: { schoolId }, data: { isCurrent: false } });
  return prisma.academicYear.create({ data: { ...data, startDate: new Date(data.startDate), endDate: new Date(data.endDate), schoolId } });
}

export async function getAcademicYears(schoolId: string) {
  return prisma.academicYear.findMany({ where: { schoolId }, orderBy: { startDate: 'desc' }, include: { semesters: true } });
}

// Subjects
export async function createSubject(schoolId: string, data: { name: string; code: string; gradeLevel: string; curriculum: string; weeklyHours?: number }) {
  const exists = await prisma.subject.findFirst({ where: { code: data.code, schoolId } });
  if (exists) throw new ConflictError('Kode mata pelajaran sudah digunakan');
  return prisma.subject.create({ data: { ...data, schoolId } });
}

export async function getSubjects(schoolId: string) {
  return prisma.subject.findMany({ where: { schoolId, isActive: true }, orderBy: { name: 'asc' } });
}

// Classes
export async function createClass(schoolId: string, data: { name: string; gradeLevel: string; academicYearId: string; homeroomTeacherId?: string; capacity?: number }) {
  return prisma.schoolClass.create({ data: { ...data, schoolId } });
}

export async function getClasses(schoolId: string, academicYearId?: string) {
  const where: Record<string, unknown> = { schoolId };
  if (academicYearId) where.academicYearId = academicYearId;
  return prisma.schoolClass.findMany({ where, include: { homeroomTeacher: { select: { name: true } }, _count: { select: { enrollments: true } } } });
}

// Enrollments
export async function enrollStudent(schoolId: string, data: { studentId: string; classId: string; academicYearId: string }) {
  const exists = await prisma.enrollment.findFirst({ where: { studentId: data.studentId, academicYearId: data.academicYearId } });
  if (exists) throw new ConflictError('Siswa sudah terdaftar di tahun ajaran ini');
  return prisma.enrollment.create({ data: { ...data, schoolId, status: 'ACTIVE' } });
}

// Grades
export async function inputGrade(schoolId: string, data: { studentId: string; subjectId: string; semesterId: string; type: string; score: number; maxScore?: number; description?: string; teacherId: string }) {
  return prisma.grade.upsert({
    where: { studentId_subjectId_semesterId_type: { studentId: data.studentId, subjectId: data.subjectId, semesterId: data.semesterId, type: data.type } },
    create: { ...data, schoolId, maxScore: data.maxScore || 100 },
    update: { score: data.score, description: data.description },
  });
}

export async function getStudentGrades(schoolId: string, studentId: string, semesterId: string) {
  return prisma.grade.findMany({
    where: { studentId, semesterId, schoolId },
    include: { subject: { select: { name: true, code: true } } },
    orderBy: { subject: { name: 'asc' } },
  });
}

// Report Card
export async function generateReportCard(schoolId: string, studentId: string, semesterId: string) {
  const student = await prisma.student.findFirst({ where: { id: studentId, schoolId }, include: { enrollments: { include: { schoolClass: true } } } });
  if (!student) throw new NotFoundError('Siswa');

  const grades = await prisma.grade.findMany({ where: { studentId, semesterId, schoolId }, include: { subject: true } });
  const avg = grades.length ? grades.reduce((s, g) => s + g.score, 0) / grades.length : 0;

  const reportCard = await prisma.reportCard.upsert({
    where: { studentId_semesterId: { studentId, semesterId } },
    create: { studentId, semesterId, schoolId, averageScore: avg, status: 'DRAFT' },
    update: { averageScore: avg },
  });
  return { reportCard, grades };
}

// Schedule
export async function createSchedule(schoolId: string, data: { classId: string; subjectId: string; teacherId: string; dayOfWeek: number; startTime: string; endTime: string; room?: string; academicYearId: string }) {
  return prisma.schedule.create({ data: { ...data, schoolId } });
}

export async function getSchedules(schoolId: string, query: { classId?: string; teacherId?: string; academicYearId?: string }) {
  const where: Record<string, unknown> = { schoolId };
  if (query.classId) where.classId = query.classId;
  if (query.teacherId) where.teacherId = query.teacherId;
  if (query.academicYearId) where.academicYearId = query.academicYearId;
  return prisma.schedule.findMany({ where, orderBy: [{ dayOfWeek: 'asc' }, { startTime: 'asc' }] });
}

export async function getTeachers(schoolId: string, query: { page?: number; limit?: number; search?: string }) {
  const { page, limit, skip } = getPaginationParams(query);
  const where: Record<string, unknown> = { schoolId };
  if (query.search) where.name = { contains: query.search, mode: 'insensitive' };
  const [teachers, total] = await Promise.all([
    prisma.teacher.findMany({ where, skip, take: limit, orderBy: { name: 'asc' } }),
    prisma.teacher.count({ where }),
  ]);
  return { teachers, meta: buildPaginationMeta(total, page, limit) };
}

export async function createTeacher(schoolId: string, data: { name: string; nip?: string; email: string; phone?: string; subjects?: string[] }) {
  return prisma.teacher.create({ data: { ...data, schoolId } });
}
