import { prisma } from '../../lib/prisma';
import { NotFoundError, ConflictError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';

export async function createPeriod(schoolId: string, data: { name: string; academicYear: string; startDate: string; endDate: string; quota: number; requirements?: string; selectionType: string }) {
  return prisma.pPDBPeriod.create({ data: { ...data, schoolId, startDate: new Date(data.startDate), endDate: new Date(data.endDate), isOpen: true } });
}

export async function getPeriods(schoolId: string) {
  return prisma.pPDBPeriod.findMany({ where: { schoolId }, orderBy: { startDate: 'desc' }, include: { _count: { select: { applications: true } } } });
}

export async function applyPPDB(data: { periodId: string; applicantName: string; birthDate: string; gender: string; parentName: string; parentPhone: string; email?: string; address: string; previousSchool?: string; documents?: Record<string, string> }) {
  const period = await prisma.pPDBPeriod.findUnique({ where: { id: data.periodId } });
  if (!period) throw new NotFoundError('Periode PPDB');
  if (!period.isOpen) throw new ConflictError('Periode PPDB sudah ditutup');

  const registrationNumber = `PPDB-${period.schoolId.slice(-4)}-${Date.now()}`;
  return prisma.pPDBApplication.create({ data: { ...data, registrationNumber, birthDate: new Date(data.birthDate), status: 'SUBMITTED', documents: data.documents || {} } });
}

export async function getApplications(schoolId: string, query: { periodId?: string; status?: string; page?: number; limit?: number }) {
  const { page, limit, skip } = getPaginationParams(query);
  const where: Record<string, unknown> = { period: { schoolId } };
  if (query.periodId) where.periodId = query.periodId;
  if (query.status) where.status = query.status;
  const [applications, total] = await Promise.all([
    prisma.pPDBApplication.findMany({ where, skip, take: limit, orderBy: { createdAt: 'desc' }, include: { period: { select: { name: true } } } }),
    prisma.pPDBApplication.count({ where }),
  ]);
  return { applications, meta: buildPaginationMeta(total, page, limit) };
}

export async function updateApplicationStatus(id: string, status: string, notes?: string) {
  const app = await prisma.pPDBApplication.findUnique({ where: { id } });
  if (!app) throw new NotFoundError('Pendaftaran');
  return prisma.pPDBApplication.update({ where: { id }, data: { status: status as never, verificationNotes: notes } });
}

export async function getDashboard(schoolId: string, periodId?: string) {
  const where: Record<string, unknown> = { period: { schoolId } };
  if (periodId) where.periodId = periodId;
  const [total, submitted, verified, accepted, rejected] = await Promise.all([
    prisma.pPDBApplication.count({ where }),
    prisma.pPDBApplication.count({ where: { ...where, status: 'SUBMITTED' } }),
    prisma.pPDBApplication.count({ where: { ...where, status: 'VERIFIED' } }),
    prisma.pPDBApplication.count({ where: { ...where, status: 'ACCEPTED' } }),
    prisma.pPDBApplication.count({ where: { ...where, status: 'REJECTED' } }),
  ]);
  return { total, submitted, verified, accepted, rejected };
}
