import { prisma } from '../../lib/prisma';
import { NotFoundError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';

export async function createProfile(schoolId: string, data: {
  studentId: string;
  encryptedData: string;
  requiresExtraTime?: boolean;
  requiresAssistant?: boolean;
  requiresLargeFont?: boolean;
  gpkId?: string;
}) {
  return prisma.specialNeedsProfile.create({
    data: { studentId: data.studentId, schoolId, encryptedData: data.encryptedData, requiresExtraTime: data.requiresExtraTime || false, requiresAssistant: data.requiresAssistant || false, requiresLargeFont: data.requiresLargeFont || false, gpkId: data.gpkId },
  });
}

export async function listProfiles(schoolId: string, query: { gpkId?: string; page?: string; limit?: string }) {
  const { page, limit, skip } = getPaginationParams({ page: Number(query.page) || undefined, limit: Number(query.limit) || undefined });
  const where: Record<string, unknown> = { schoolId };
  if (query.gpkId) where.gpkId = query.gpkId;
  const [profiles, total] = await Promise.all([
    prisma.specialNeedsProfile.findMany({ where, skip, take: limit, orderBy: { createdAt: 'desc' }, include: { student: { select: { name: true, nisn: true } } } }),
    prisma.specialNeedsProfile.count({ where }),
  ]);
  return { profiles, meta: buildPaginationMeta(total, page, limit) };
}

export async function getProfile(schoolId: string, studentId: string) {
  const profile = await prisma.specialNeedsProfile.findFirst({ where: { studentId, schoolId }, include: { student: { select: { name: true } }, ieps: { orderBy: { createdAt: 'desc' } } } });
  if (!profile) throw new NotFoundError('Profil ABK');
  return profile;
}

export async function createIEP(schoolId: string, gpkId: string, data: {
  profileId: string;
  academicYear: string;
  semester: number;
  goals: object;
  approvedBy?: string;
}) {
  const profile = await prisma.specialNeedsProfile.findFirst({ where: { id: data.profileId, schoolId } });
  if (!profile) throw new NotFoundError('Profil ABK');
  return prisma.iEP.create({
    data: { profileId: data.profileId, schoolId, academicYear: data.academicYear, semester: data.semester, goals: data.goals as unknown as never, gpkId, approvedBy: data.approvedBy },
  });
}

export async function getStudentIEPs(schoolId: string, studentId: string) {
  const profile = await prisma.specialNeedsProfile.findFirst({ where: { studentId, schoolId } });
  if (!profile) throw new NotFoundError('Profil ABK');
  return prisma.iEP.findMany({ where: { profileId: profile.id, schoolId }, orderBy: { createdAt: 'desc' } });
}

export async function addProgressReport(schoolId: string, gpkId: string, data: {
  profileId: string;
  period: string;
  content: object;
}) {
  const profile = await prisma.specialNeedsProfile.findFirst({ where: { id: data.profileId, schoolId } });
  if (!profile) throw new NotFoundError('Profil ABK');
  return prisma.aBKProgressReport.create({
    data: { profileId: data.profileId, schoolId, period: data.period, content: data.content as unknown as never, gpkId },
  });
}

export async function getProgressReports(schoolId: string, studentId: string) {
  const profile = await prisma.specialNeedsProfile.findFirst({ where: { studentId, schoolId } });
  if (!profile) throw new NotFoundError('Profil ABK');
  return prisma.aBKProgressReport.findMany({ where: { profileId: profile.id, schoolId }, orderBy: { createdAt: 'desc' } });
}
