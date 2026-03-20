import { prisma } from '../../lib/prisma';
import { NotFoundError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';

export async function createHealthRecord(schoolId: string, data: { studentId: string; bloodType?: string; allergies?: string; medicalConditions?: string; emergencyContact?: string }) {
  const { studentId, bloodType, allergies, medicalConditions, emergencyContact } = data;
  const recordData = { studentId, bloodType, allergies, medicalConditions, emergencyContact };
  return prisma.healthRecord.upsert({
    where: { studentId },
    create: { ...recordData, schoolId },
    update: recordData,
  });
}

export async function getHealthRecord(schoolId: string, studentId: string) {
  const record = await prisma.healthRecord.findFirst({ where: { studentId, schoolId }, include: { student: { select: { name: true, nisn: true } } } });
  if (!record) throw new NotFoundError('Rekam medis');
  return record;
}

export async function recordUKSVisit(schoolId: string, data: { studentId: string; complaint: string; action: string; medications?: string; staffId: string; referral?: string }) {
  const { studentId, complaint, action, medications, staffId, referral } = data;
  return prisma.uKSVisit.create({ data: { studentId, complaint, action, medications, staffId, referral, schoolId, visitDate: new Date() } });
}

export async function getUKSVisits(schoolId: string, query: { studentId?: string; startDate?: string; endDate?: string; page?: number; limit?: number }) {
  const { page, limit, skip } = getPaginationParams(query);
  const where: Record<string, unknown> = { schoolId };
  if (query.studentId) where.studentId = query.studentId;
  if (query.startDate || query.endDate) {
    where.visitDate = {};
    if (query.startDate) (where.visitDate as Record<string, unknown>).gte = new Date(query.startDate);
    if (query.endDate) (where.visitDate as Record<string, unknown>).lte = new Date(query.endDate);
  }
  const [visits, total] = await Promise.all([
    prisma.uKSVisit.findMany({ where, skip, take: limit, orderBy: { visitDate: 'desc' }, include: { student: { select: { name: true } } } }),
    prisma.uKSVisit.count({ where }),
  ]);
  return { visits, meta: buildPaginationMeta(total, page, limit) };
}

export async function recordNutrition(schoolId: string, data: { studentId: string; weight: number; height: number; measuredAt?: string }) {
  const bmi = data.weight / Math.pow(data.height / 100, 2);
  let status = 'NORMAL';
  if (bmi < 18.5) status = 'KURANG';
  else if (bmi >= 25 && bmi < 30) status = 'LEBIH';
  else if (bmi >= 30) status = 'OBESITAS';

  return prisma.nutritionRecord.create({ data: { studentId: data.studentId, weight: data.weight, height: data.height, schoolId, measuredAt: data.measuredAt ? new Date(data.measuredAt) : new Date(), bmi, nutritionStatus: status } });
}

export async function getNutritionRecords(schoolId: string, studentId: string) {
  return prisma.nutritionRecord.findMany({ where: { studentId, schoolId }, orderBy: { measuredAt: 'desc' } });
}
