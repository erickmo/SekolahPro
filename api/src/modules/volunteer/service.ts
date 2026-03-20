import { prisma } from '../../lib/prisma';
import { NotFoundError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';
import { VolunteerRelation } from '@prisma/client';

export async function createProfile(schoolId: string, data: {
  name: string;
  email?: string;
  phone?: string;
  relation: string;
  skills?: string[];
  availability?: object;
}) {
  return prisma.volunteerProfile.create({
    data: { schoolId, name: data.name, email: data.email, phone: data.phone, relation: data.relation as VolunteerRelation, skills: data.skills || [], availability: data.availability as unknown as never },
  });
}

export async function listProfiles(schoolId: string, query: { status?: string; relation?: string; page?: string; limit?: string }) {
  const { page, limit, skip } = getPaginationParams({ page: Number(query.page) || undefined, limit: Number(query.limit) || undefined });
  const where: Record<string, unknown> = { schoolId };
  if (query.relation) where.relation = query.relation;
  const [profiles, total] = await Promise.all([
    prisma.volunteerProfile.findMany({ where, skip, take: limit, orderBy: { createdAt: 'desc' } }),
    prisma.volunteerProfile.count({ where }),
  ]);
  return { profiles, meta: buildPaginationMeta(total, page, limit) };
}

export async function getProfile(schoolId: string, id: string) {
  const profile = await prisma.volunteerProfile.findFirst({
    where: { id, schoolId },
    include: { activities: { orderBy: { date: 'desc' } }, certificates: true },
  });
  if (!profile) throw new NotFoundError('Profil relawan');
  return profile;
}

export async function createActivity(schoolId: string, data: {
  volunteerId: string;
  title: string;
  date: string;
  hours: number;
  notes?: string;
}) {
  const volunteer = await prisma.volunteerProfile.findFirst({ where: { id: data.volunteerId, schoolId } });
  if (!volunteer) throw new NotFoundError('Profil relawan');
  const activity = await prisma.volunteerActivity.create({
    data: { volunteerId: data.volunteerId, schoolId, title: data.title, date: new Date(data.date), hours: data.hours, notes: data.notes },
  });
  // Update totalHours
  await prisma.volunteerProfile.update({ where: { id: data.volunteerId }, data: { totalHours: { increment: data.hours } } });
  return activity;
}

export async function listActivities(schoolId: string, query: { volunteerId?: string; page?: string; limit?: string }) {
  const { page, limit, skip } = getPaginationParams({ page: Number(query.page) || undefined, limit: Number(query.limit) || undefined });
  const where: Record<string, unknown> = { schoolId };
  if (query.volunteerId) where.volunteerId = query.volunteerId;
  const [activities, total] = await Promise.all([
    prisma.volunteerActivity.findMany({ where, skip, take: limit, orderBy: { date: 'desc' } }),
    prisma.volunteerActivity.count({ where }),
  ]);
  return { activities, meta: buildPaginationMeta(total, page, limit) };
}

export async function updateActivity(schoolId: string, id: string, data: { notes?: string }) {
  const activity = await prisma.volunteerActivity.findFirst({ where: { id, schoolId } });
  if (!activity) throw new NotFoundError('Aktivitas relawan');
  return prisma.volunteerActivity.update({ where: { id }, data });
}

export async function issueCertificate(schoolId: string, data: {
  volunteerId: string;
  title: string;
  issuedAt: string;
  pdfUrl?: string;
}) {
  const volunteer = await prisma.volunteerProfile.findFirst({ where: { id: data.volunteerId, schoolId } });
  if (!volunteer) throw new NotFoundError('Profil relawan');
  return prisma.volunteerCertificate.create({
    data: { volunteerId: data.volunteerId, schoolId, title: data.title, issuedAt: new Date(data.issuedAt), pdfUrl: data.pdfUrl },
  });
}
