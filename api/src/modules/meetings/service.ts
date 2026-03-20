import { prisma } from '../../lib/prisma';
import { NotFoundError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';

export async function createMeeting(schoolId: string, organizerId: string, data: { title: string; description?: string; type: string; scheduledAt: string; duration: number; location?: string; meetingUrl?: string; participantIds: string[] }) {
  const { participantIds, ...meetingData } = data;
  return prisma.meeting.create({
    data: { ...meetingData, schoolId, organizerId, scheduledAt: new Date(data.scheduledAt), status: 'SCHEDULED', participants: { create: participantIds.map((id) => ({ userId: id })) } },
    include: { participants: true },
  });
}

export async function getMeetings(schoolId: string, query: { type?: string; status?: string; userId?: string; page?: number; limit?: number }) {
  const { page, limit, skip } = getPaginationParams(query);
  const where: Record<string, unknown> = { schoolId };
  if (query.type) where.type = query.type;
  if (query.status) where.status = query.status;
  if (query.userId) where.participants = { some: { userId: query.userId } };
  const [meetings, total] = await Promise.all([
    prisma.meeting.findMany({ where, skip, take: limit, orderBy: { scheduledAt: 'desc' }, include: { organizer: { select: { name: true } }, _count: { select: { participants: true } } } }),
    prisma.meeting.count({ where }),
  ]);
  return { meetings, meta: buildPaginationMeta(total, page, limit) };
}

export async function updateMeeting(schoolId: string, id: string, data: { status?: string; notes?: string; recordingUrl?: string }) {
  const meeting = await prisma.meeting.findFirst({ where: { id, schoolId } });
  if (!meeting) throw new NotFoundError('Rapat');
  return prisma.meeting.update({ where: { id }, data });
}
