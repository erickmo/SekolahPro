import { prisma } from '../../lib/prisma';
import { NotFoundError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';
import { IoTDeviceType } from '@prisma/client';

export async function registerDevice(schoolId: string, data: {
  name: string;
  type: string;
  location: string;
  mqttTopic: string;
  metadata?: object;
}) {
  return prisma.ioTDevice.create({
    data: { schoolId, name: data.name, type: data.type as IoTDeviceType, location: data.location, mqttTopic: data.mqttTopic, isActive: true, metadata: data.metadata as unknown as never },
  });
}

export async function listDevices(schoolId: string, query: { type?: string; isActive?: string }) {
  const where: Record<string, unknown> = { schoolId };
  if (query.type) where.type = query.type;
  if (query.isActive !== undefined) where.isActive = query.isActive === 'true';
  return prisma.ioTDevice.findMany({ where, orderBy: { createdAt: 'desc' } });
}

export async function getDevice(schoolId: string, id: string) {
  const device = await prisma.ioTDevice.findFirst({ where: { id, schoolId } });
  if (!device) throw new NotFoundError('Perangkat IoT');
  return device;
}

export async function updateDeviceStatus(schoolId: string, id: string, isActive: boolean) {
  const device = await prisma.ioTDevice.findFirst({ where: { id, schoolId } });
  if (!device) throw new NotFoundError('Perangkat IoT');
  return prisma.ioTDevice.update({ where: { id }, data: { isActive, lastPing: new Date() } });
}

export async function recordReading(data: {
  deviceId: string;
  value: number;
  unit: string;
  metadata?: object;
}) {
  const reading = await prisma.ioTReading.create({
    data: { deviceId: data.deviceId, value: data.value, unit: data.unit, metadata: data.metadata as unknown as never },
  });
  await prisma.ioTDevice.update({ where: { id: data.deviceId }, data: { lastPing: new Date() } }).catch(() => null);
  return reading;
}

export async function listReadings(schoolId: string, query: { deviceId?: string; startDate?: string; endDate?: string; page?: string; limit?: string }) {
  const { page, limit, skip } = getPaginationParams({ page: Number(query.page) || undefined, limit: Number(query.limit) || undefined });
  const where: Record<string, unknown> = { device: { schoolId } };
  if (query.deviceId) where.deviceId = query.deviceId;
  if (query.startDate || query.endDate) {
    const recordedAt: Record<string, unknown> = {};
    if (query.startDate) recordedAt.gte = new Date(query.startDate);
    if (query.endDate) recordedAt.lte = new Date(query.endDate);
    where.recordedAt = recordedAt;
  }
  const [readings, total] = await Promise.all([
    prisma.ioTReading.findMany({ where, skip, take: limit, orderBy: { recordedAt: 'desc' } }),
    prisma.ioTReading.count({ where }),
  ]);
  return { readings, meta: buildPaginationMeta(total, page, limit) };
}

export async function getLatestReading(schoolId: string, deviceId: string) {
  const device = await prisma.ioTDevice.findFirst({ where: { id: deviceId, schoolId } });
  if (!device) throw new NotFoundError('Perangkat IoT');
  return prisma.ioTReading.findFirst({ where: { deviceId }, orderBy: { recordedAt: 'desc' } });
}
