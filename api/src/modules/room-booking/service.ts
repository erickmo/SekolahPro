import { prisma } from '../../lib/prisma';
import { NotFoundError, ConflictError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';

export async function getRooms(schoolId: string, query: { type?: string; available?: string }) {
  const where: Record<string, unknown> = { schoolId };
  if (query.type) where.type = query.type;
  if (query.available === 'true') where.isAvailable = true;
  return prisma.room.findMany({ where, orderBy: { name: 'asc' } });
}

export async function createRoom(schoolId: string, data: { name: string; type: string; capacity: number; facilities?: string[] }) {
  return prisma.room.create({ data: { ...data, schoolId, isAvailable: true, facilities: data.facilities || [] } });
}

export async function bookRoom(schoolId: string, data: { roomId: string; requesterId: string; purpose: string; startTime: string; endTime: string; attendees?: number }) {
  // Check conflict
  const conflict = await prisma.roomBooking.findFirst({
    where: {
      roomId: data.roomId,
      status: { in: ['APPROVED', 'PENDING'] },
      OR: [
        { startTime: { lte: new Date(data.startTime) }, endTime: { gte: new Date(data.startTime) } },
        { startTime: { lte: new Date(data.endTime) }, endTime: { gte: new Date(data.endTime) } },
      ],
    },
  });
  if (conflict) throw new ConflictError('Ruangan sudah dibooking pada waktu tersebut');

  return prisma.roomBooking.create({
    data: { roomId: data.roomId, requesterId: data.requesterId, schoolId, purpose: data.purpose, startTime: new Date(data.startTime), endTime: new Date(data.endTime), attendees: data.attendees, status: 'PENDING' },
  });
}

export async function getBookings(schoolId: string, query: { roomId?: string; status?: string; requesterId?: string; page?: number; limit?: number }) {
  const { page, limit, skip } = getPaginationParams(query);
  const where: Record<string, unknown> = { schoolId };
  if (query.roomId) where.roomId = query.roomId;
  if (query.status) where.status = query.status;
  if (query.requesterId) where.requesterId = query.requesterId;
  const [bookings, total] = await Promise.all([
    prisma.roomBooking.findMany({ where, skip, take: limit, orderBy: { startTime: 'desc' }, include: { room: { select: { name: true, type: true } } } }),
    prisma.roomBooking.count({ where }),
  ]);
  return { bookings, meta: buildPaginationMeta(total, page, limit) };
}

export async function approveBooking(schoolId: string, id: string, approverId: string, approved: boolean) {
  const booking = await prisma.roomBooking.findFirst({ where: { id, schoolId } });
  if (!booking) throw new NotFoundError('Booking');
  return prisma.roomBooking.update({ where: { id }, data: { status: approved ? 'APPROVED' : 'REJECTED', approverId } });
}
