import { prisma } from '../../lib/prisma';
import { ConflictError, NotFoundError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';
import { CreateSchoolInput, UpdateSchoolInput } from './dto';

export async function createSchool(input: CreateSchoolInput) {
  const exists = await prisma.school.findFirst({ where: { OR: [{ npsn: input.npsn }, { subdomain: input.subdomain }] } });
  if (exists) throw new ConflictError('NPSN atau subdomain sudah digunakan');

  return prisma.school.create({ data: { ...input, config: (input.config || {}) as unknown as never } });
}

export async function getSchools(query: { page?: number; limit?: number; search?: string }) {
  const { page, limit, skip } = getPaginationParams(query);
  const where = query.search ? { name: { contains: query.search, mode: 'insensitive' as const } } : {};
  const [schools, total] = await Promise.all([
    prisma.school.findMany({ where, skip, take: limit, orderBy: { createdAt: 'desc' } }),
    prisma.school.count({ where }),
  ]);
  return { schools, meta: buildPaginationMeta(total, page, limit) };
}

export async function getSchool(id: string) {
  const school = await prisma.school.findUnique({ where: { id }, include: { _count: { select: { students: true, teachers: true } } } });
  if (!school) throw new NotFoundError('Sekolah');
  return school;
}

export async function updateSchool(id: string, input: UpdateSchoolInput) {
  await getSchool(id);
  const { config, ...rest } = input;
  return prisma.school.update({ where: { id }, data: { ...rest, ...(config !== undefined ? { config: config as unknown as never } : {}) } });
}

export async function deleteSchool(id: string) {
  await getSchool(id);
  return prisma.school.delete({ where: { id } });
}

export async function getSchoolBySubdomain(subdomain: string) {
  const school = await prisma.school.findUnique({ where: { subdomain } });
  if (!school) throw new NotFoundError('Sekolah');
  return school;
}
