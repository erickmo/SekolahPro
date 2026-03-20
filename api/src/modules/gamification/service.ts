import { prisma } from '../../lib/prisma';
import { NotFoundError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';
import { PointRuleType } from '@prisma/client';

export async function createRule(schoolId: string, data: {
  type: string;
  points: number;
  description: string;
  conditions?: object;
}) {
  return prisma.pointRule.create({
    data: { schoolId, type: data.type as PointRuleType, points: data.points, description: data.description, conditions: data.conditions as unknown as never, isActive: true },
  });
}

export async function listRules(schoolId: string) {
  return prisma.pointRule.findMany({ where: { schoolId, isActive: true }, orderBy: { createdAt: 'desc' } });
}

export async function awardPoints(schoolId: string, data: { studentId: string; action: string; points: number; reason: string; ruleId?: string }) {
  const account = await prisma.studentPoints.upsert({
    where: { studentId: data.studentId },
    create: { studentId: data.studentId, schoolId, totalPoints: 0, streak: 0 },
    update: {},
  });
  await prisma.pointTransaction.create({
    data: { accountId: account.id, schoolId, type: data.action, points: data.points, reason: data.reason, ruleId: data.ruleId },
  });
  return prisma.studentPoints.update({
    where: { id: account.id },
    data: { totalPoints: { increment: data.points }, lastActivity: new Date() },
  });
}

export async function getStudentPoints(schoolId: string, studentId: string) {
  const account = await prisma.studentPoints.findFirst({
    where: { studentId, schoolId },
    include: { transactions: { orderBy: { createdAt: 'desc' }, take: 20 } },
  });
  if (!account) return { studentId, schoolId, totalPoints: 0, streak: 0, transactions: [] };
  return account;
}

export async function getLeaderboard(schoolId: string, limit = 10) {
  return prisma.studentPoints.findMany({
    where: { schoolId },
    orderBy: { totalPoints: 'desc' },
    take: limit,
    include: { student: { select: { name: true, nisn: true } } },
  });
}

export async function createBadge(schoolId: string, data: {
  name: string;
  description: string;
  iconUrl?: string;
  criteria: object;
}) {
  return prisma.badge.create({
    data: { schoolId, name: data.name, description: data.description, iconUrl: data.iconUrl, criteria: data.criteria as unknown as never },
  });
}

export async function listBadges(schoolId: string) {
  return prisma.badge.findMany({ where: { schoolId }, orderBy: { createdAt: 'desc' } });
}

export async function awardBadge(schoolId: string, data: { studentId: string; badgeId: string }) {
  const badge = await prisma.badge.findFirst({ where: { id: data.badgeId, schoolId } });
  if (!badge) throw new NotFoundError('Badge');
  const account = await prisma.studentPoints.upsert({
    where: { studentId: data.studentId },
    create: { studentId: data.studentId, schoolId, totalPoints: 0, streak: 0 },
    update: {},
  });
  return prisma.studentBadge.create({ data: { studentId: data.studentId, accountId: account.id, badgeId: data.badgeId } });
}

export async function getStudentBadges(schoolId: string, studentId: string) {
  const account = await prisma.studentPoints.findFirst({ where: { studentId, schoolId } });
  if (!account) return [];
  return prisma.studentBadge.findMany({
    where: { accountId: account.id },
    include: { badge: true },
    orderBy: { awardedAt: 'desc' },
  });
}
