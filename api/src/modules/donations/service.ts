import { prisma } from '../../lib/prisma';
import { NotFoundError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';

export async function createDonation(schoolId: string, data: {
  donorId?: string;
  donorName: string;
  amount: number;
  purpose: string;
  eventId?: string;
  isAnonymous?: boolean;
  receiptUrl?: string;
}) {
  return prisma.donationRecord.create({
    data: { schoolId, donorId: data.donorId, donorName: data.donorName, amount: data.amount, purpose: data.purpose, eventId: data.eventId, isAnonymous: data.isAnonymous || false, receiptUrl: data.receiptUrl },
  });
}

export async function listDonations(schoolId: string, query: { purpose?: string; page?: string; limit?: string }) {
  const { page, limit, skip } = getPaginationParams({ page: Number(query.page) || undefined, limit: Number(query.limit) || undefined });
  const where: Record<string, unknown> = { schoolId };
  if (query.purpose) where.purpose = query.purpose;
  const [donations, total] = await Promise.all([
    prisma.donationRecord.findMany({ where, skip, take: limit, orderBy: { createdAt: 'desc' } }),
    prisma.donationRecord.count({ where }),
  ]);
  return { donations, meta: buildPaginationMeta(total, page, limit) };
}

export async function getDonation(schoolId: string, id: string) {
  const donation = await prisma.donationRecord.findFirst({ where: { id, schoolId } });
  if (!donation) throw new NotFoundError('Donasi');
  return donation;
}

export async function getDonationSummary(schoolId: string) {
  const [totalAmount, count] = await Promise.all([
    prisma.donationRecord.aggregate({ where: { schoolId }, _sum: { amount: true }, _count: true }),
    prisma.donationRecord.groupBy({ by: ['purpose'], where: { schoolId }, _sum: { amount: true }, _count: true }),
  ]);
  return {
    totalAmount: Number(totalAmount._sum.amount || 0),
    totalCount: totalAmount._count,
    byPurpose: count.map(c => ({ purpose: c.purpose, amount: Number(c._sum.amount || 0), count: c._count })),
  };
}
