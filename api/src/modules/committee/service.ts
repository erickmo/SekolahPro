import { prisma } from '../../lib/prisma';
import { NotFoundError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';

export async function createBoard(schoolId: string, data: { periodStart: string; periodEnd: string }) {
  return prisma.committeeBoard.create({
    data: { schoolId, periodStart: new Date(data.periodStart), periodEnd: new Date(data.periodEnd) },
  });
}

export async function listBoards(schoolId: string) {
  return prisma.committeeBoard.findMany({ where: { schoolId }, include: { members: true }, orderBy: { periodStart: 'desc' } });
}

export async function getBoard(schoolId: string, id: string) {
  const board = await prisma.committeeBoard.findFirst({ where: { id, schoolId }, include: { members: true, meetings: { orderBy: { scheduledAt: 'desc' } } } });
  if (!board) throw new NotFoundError('Komite');
  return board;
}

export async function addMember(schoolId: string, boardId: string, data: { userId?: string; name: string; role: string; phone?: string }) {
  const board = await prisma.committeeBoard.findFirst({ where: { id: boardId, schoolId } });
  if (!board) throw new NotFoundError('Komite');
  return prisma.committeeMember.create({
    data: { boardId, userId: data.userId, name: data.name, role: data.role, phone: data.phone },
  });
}

export async function createMeeting(schoolId: string, data: { boardId: string; scheduledAt: string; agenda: string[]; videoCallUrl?: string }) {
  const board = await prisma.committeeBoard.findFirst({ where: { id: data.boardId, schoolId } });
  if (!board) throw new NotFoundError('Komite');
  return prisma.committeeMeeting.create({
    data: { boardId: data.boardId, schoolId, scheduledAt: new Date(data.scheduledAt), agenda: data.agenda, videoCallUrl: data.videoCallUrl, status: 'SCHEDULED' },
  });
}

export async function listMeetings(schoolId: string, query: { boardId?: string; status?: string; page?: string; limit?: string }) {
  const { page, limit, skip } = getPaginationParams({ page: Number(query.page) || undefined, limit: Number(query.limit) || undefined });
  const where: Record<string, unknown> = { schoolId };
  if (query.boardId) where.boardId = query.boardId;
  if (query.status) where.status = query.status;
  const [meetings, total] = await Promise.all([
    prisma.committeeMeeting.findMany({ where, skip, take: limit, orderBy: { scheduledAt: 'desc' } }),
    prisma.committeeMeeting.count({ where }),
  ]);
  return { meetings, meta: buildPaginationMeta(total, page, limit) };
}

export async function addDecision(schoolId: string, meetingId: string, data: {
  boardId: string;
  title: string;
  description: string;
  votesFor?: number;
  votesAgainst?: number;
  votesAbstain?: number;
  result: string;
  isBinding?: boolean;
}) {
  const meeting = await prisma.committeeMeeting.findFirst({ where: { id: meetingId, schoolId } });
  if (!meeting) throw new NotFoundError('Rapat komite');
  return prisma.committeeDecision.create({
    data: { meetingId, boardId: data.boardId, title: data.title, description: data.description, votesFor: data.votesFor || 0, votesAgainst: data.votesAgainst || 0, votesAbstain: data.votesAbstain || 0, result: data.result, signedByIds: [], isBinding: data.isBinding || false },
  });
}

export async function listDecisions(schoolId: string, meetingId: string) {
  const meeting = await prisma.committeeMeeting.findFirst({ where: { id: meetingId, schoolId } });
  if (!meeting) throw new NotFoundError('Rapat komite');
  return prisma.committeeDecision.findMany({ where: { meetingId }, orderBy: { createdAt: 'desc' } });
}
