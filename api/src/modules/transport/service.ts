import { prisma } from '../../lib/prisma';
import { NotFoundError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';

export async function getBuses(schoolId: string) {
  return prisma.transportBus.findMany({ where: { schoolId, isActive: true }, include: { routes: true } });
}

export async function createBus(schoolId: string, data: { name: string; plateNumber: string; capacity: number; driverName: string; driverPhone: string }) {
  return prisma.transportBus.create({ data: { ...data, schoolId, isActive: true } });
}

export async function getRoutes(schoolId: string) {
  return prisma.transportRoute.findMany({ where: { schoolId, isActive: true }, include: { bus: { select: { name: true, plateNumber: true } } } });
}

export async function createRoute(schoolId: string, data: { busId: string; name: string; stops: Array<{ name: string; estimatedTime: string }>; departureTime: string; type: string }) {
  return prisma.transportRoute.create({ data: { ...data, schoolId, isActive: true, stops: data.stops as never } });
}

export async function recordRide(schoolId: string, data: { studentId: string; busId: string; routeId: string; type: string; stopName?: string }) {
  return prisma.transportRideLog.create({ data: { ...data, schoolId, timestamp: new Date() } });
}

export async function getRideLogs(schoolId: string, query: { busId?: string; studentId?: string; date?: string; page?: number; limit?: number }) {
  const { page, limit, skip } = getPaginationParams(query);
  const where: Record<string, unknown> = { schoolId };
  if (query.busId) where.busId = query.busId;
  if (query.studentId) where.studentId = query.studentId;
  if (query.date) {
    const d = new Date(query.date);
    const nd = new Date(query.date);
    nd.setDate(nd.getDate() + 1);
    where.timestamp = { gte: d, lt: nd };
  }
  const [logs, total] = await Promise.all([
    prisma.transportRideLog.findMany({ where, skip, take: limit, orderBy: { timestamp: 'desc' }, include: { student: { select: { name: true } }, bus: { select: { name: true } } } }),
    prisma.transportRideLog.count({ where }),
  ]);
  return { logs, meta: buildPaginationMeta(total, page, limit) };
}
