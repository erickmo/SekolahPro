import { prisma } from '../../lib/prisma';
import { NotFoundError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';

export async function createPlan(schoolId: string, data: {
  fiscalYear: string;
  fundSource: string;
  totalAmount: number;
}) {
  return prisma.budgetPlan.create({
    data: { schoolId, fiscalYear: data.fiscalYear, fundSource: data.fundSource, totalAmount: data.totalAmount, status: 'DRAFT' },
  });
}

export async function listPlans(schoolId: string, query: { fiscalYear?: string; status?: string; page?: string; limit?: string }) {
  const { page, limit, skip } = getPaginationParams({ page: Number(query.page) || undefined, limit: Number(query.limit) || undefined });
  const where: Record<string, unknown> = { schoolId };
  if (query.fiscalYear) where.fiscalYear = query.fiscalYear;
  if (query.status) where.status = query.status;
  const [plans, total] = await Promise.all([
    prisma.budgetPlan.findMany({ where, skip, take: limit, orderBy: { createdAt: 'desc' } }),
    prisma.budgetPlan.count({ where }),
  ]);
  return { plans, meta: buildPaginationMeta(total, page, limit) };
}

export async function getPlan(schoolId: string, id: string) {
  const plan = await prisma.budgetPlan.findFirst({ where: { id, schoolId }, include: { lines: true } });
  if (!plan) throw new NotFoundError('Rencana anggaran');
  return plan;
}

export async function approvePlan(schoolId: string, id: string, approvedBy: string) {
  const plan = await prisma.budgetPlan.findFirst({ where: { id, schoolId } });
  if (!plan) throw new NotFoundError('Rencana anggaran');
  return prisma.budgetPlan.update({ where: { id }, data: { status: 'APPROVED', approvedBy, approvedAt: new Date() } });
}

export async function addLine(schoolId: string, planId: string, data: {
  category: string;
  description: string;
  plannedAmount: number;
}) {
  const plan = await prisma.budgetPlan.findFirst({ where: { id: planId, schoolId } });
  if (!plan) throw new NotFoundError('Rencana anggaran');
  return prisma.budgetLine.create({
    data: { budgetPlanId: planId, schoolId, category: data.category, description: data.description, plannedAmount: data.plannedAmount },
  });
}

export async function addRealization(schoolId: string, data: {
  lineId: string;
  amount: number;
  date: string;
  description: string;
  receiptUrl?: string;
  recordedBy: string;
}) {
  const line = await prisma.budgetLine.findFirst({ where: { id: data.lineId, schoolId } });
  if (!line) throw new NotFoundError('Baris anggaran');
  const realization = await prisma.budgetRealization.create({
    data: { lineId: data.lineId, schoolId, amount: data.amount, date: new Date(data.date), description: data.description, receiptUrl: data.receiptUrl, recordedBy: data.recordedBy },
  });
  return realization;
}

export async function listRealizations(schoolId: string, query: { lineId?: string; page?: string; limit?: string }) {
  const { page, limit, skip } = getPaginationParams({ page: Number(query.page) || undefined, limit: Number(query.limit) || undefined });
  const where: Record<string, unknown> = { schoolId };
  if (query.lineId) where.lineId = query.lineId;
  const [realizations, total] = await Promise.all([
    prisma.budgetRealization.findMany({ where, skip, take: limit, orderBy: { date: 'desc' } }),
    prisma.budgetRealization.count({ where }),
  ]);
  return { realizations, meta: buildPaginationMeta(total, page, limit) };
}

export async function getBudgetSummary(schoolId: string, planId?: string) {
  const where: Record<string, unknown> = { schoolId };
  if (planId) where.id = planId;
  const plans = await prisma.budgetPlan.findMany({ where, include: { lines: { include: { realizations: true } } } });
  return plans.map(plan => {
    const totalPlanned = plan.lines.reduce((sum, l) => sum + Number(l.plannedAmount), 0);
    const totalRealized = plan.lines.reduce((sum, l) => sum + Number(l.realizedAmount), 0);
    return {
      planId: plan.id,
      fiscalYear: plan.fiscalYear,
      fundSource: plan.fundSource,
      status: plan.status,
      totalPlanned,
      totalRealized,
      utilizationRate: totalPlanned > 0 ? Math.round((totalRealized / totalPlanned) * 100) : 0,
    };
  });
}
