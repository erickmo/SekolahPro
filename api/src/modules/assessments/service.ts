import { prisma } from '../../lib/prisma';
import { NotFoundError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';

export async function createRubric(schoolId: string, data: {
  teacherId: string;
  subject: string;
  gradeLevel: string;
  title: string;
  criteria: object;
}) {
  return prisma.assessmentRubric.create({
    data: { ...data, schoolId, criteria: data.criteria as unknown as never },
  });
}

export async function listRubrics(schoolId: string, query: { page?: string; limit?: string }) {
  const { page, limit, skip } = getPaginationParams({ page: Number(query.page) || undefined, limit: Number(query.limit) || undefined });
  const [rubrics, total] = await Promise.all([
    prisma.assessmentRubric.findMany({ where: { schoolId }, skip, take: limit, orderBy: { createdAt: 'desc' } }),
    prisma.assessmentRubric.count({ where: { schoolId } }),
  ]);
  return { rubrics, meta: buildPaginationMeta(total, page, limit) };
}

export async function getRubric(schoolId: string, id: string) {
  const rubric = await prisma.assessmentRubric.findFirst({ where: { id, schoolId } });
  if (!rubric) throw new NotFoundError('Rubrik');
  return rubric;
}

export async function createProjectAssessment(schoolId: string, assessorId: string, data: {
  rubricId: string;
  studentId: string;
  assessorType: string;
  scores: object;
  totalScore: number;
  feedback?: string;
  evidenceUrls?: string[];
}) {
  const rubric = await prisma.assessmentRubric.findFirst({ where: { id: data.rubricId, schoolId } });
  if (!rubric) throw new NotFoundError('Rubrik');
  return prisma.projectAssessment.create({
    data: {
      schoolId,
      rubricId: data.rubricId,
      studentId: data.studentId,
      assessorId,
      assessorType: data.assessorType || 'TEACHER',
      scores: data.scores as unknown as never,
      totalScore: data.totalScore,
      feedback: data.feedback,
      evidenceUrls: data.evidenceUrls || [],
    },
  });
}

export async function listProjectAssessments(schoolId: string, query: { studentId?: string; rubricId?: string; page?: string; limit?: string }) {
  const { page, limit, skip } = getPaginationParams({ page: Number(query.page) || undefined, limit: Number(query.limit) || undefined });
  const where: Record<string, unknown> = { schoolId };
  if (query.studentId) where.studentId = query.studentId;
  if (query.rubricId) where.rubricId = query.rubricId;
  const [assessments, total] = await Promise.all([
    prisma.projectAssessment.findMany({ where, skip, take: limit, orderBy: { createdAt: 'desc' }, include: { rubric: { select: { title: true } } } }),
    prisma.projectAssessment.count({ where }),
  ]);
  return { assessments, meta: buildPaginationMeta(total, page, limit) };
}

export async function getStudentAssessments(schoolId: string, studentId: string) {
  return prisma.projectAssessment.findMany({
    where: { schoolId, studentId },
    orderBy: { createdAt: 'desc' },
    include: { rubric: { select: { title: true, subject: true } } },
  });
}
