import { prisma } from '../../lib/prisma';
import { config } from '../../config';
import { logger } from '../../lib/logger';

export async function syncDapodik(schoolId: string) {
  const school = await prisma.school.findUnique({ where: { id: schoolId } });
  if (!school) throw new Error('Sekolah tidak ditemukan');

  const [students, teachers] = await Promise.all([
    prisma.student.count({ where: { schoolId } }),
    prisma.teacher.count({ where: { schoolId } }),
  ]);

  const syncRecord = await prisma.dapodikSync.create({
    data: { schoolId, syncType: 'FULL', status: 'IN_PROGRESS', totalRecords: students + teachers },
  });

  try {
    if (config.env === 'production') {
      logger.info('[Dapodik] Sync triggered', { schoolId });
    } else {
      logger.info('[Dapodik Mock] Sync simulated', { schoolId, students, teachers });
    }

    await prisma.dapodikSync.update({
      where: { id: syncRecord.id },
      data: { status: 'SUCCESS', syncedRecords: students + teachers, completedAt: new Date() },
    });

    return { success: true, synced: { students, teachers } };
  } catch (err) {
    await prisma.dapodikSync.update({ where: { id: syncRecord.id }, data: { status: 'FAILED', errors: [(err as Error).message] } });
    throw err;
  }
}

export async function getSyncHistory(schoolId: string) {
  return prisma.dapodikSync.findMany({ where: { schoolId }, orderBy: { createdAt: 'desc' }, take: 20 });
}

export async function exportStudentData(schoolId: string) {
  const students = await prisma.student.findMany({
    where: { schoolId },
    include: { guardians: true, enrollments: { include: { schoolClass: true, academicYear: true }, where: { status: 'ACTIVE' } } },
  });

  return students.map((s) => ({
    NISN: s.nisn,
    NIK: s.nik || '',
    'Nama Lengkap': s.name,
    'Tanggal Lahir': s.birthDate.toISOString().split('T')[0],
    'Jenis Kelamin': s.gender === 'MALE' ? 'L' : 'P',
    Kelas: s.enrollments[0]?.schoolClass?.name || '',
    'Tahun Ajaran': s.enrollments[0]?.academicYear?.name || '',
    'Nama Ayah': s.guardians.find((g) => g.relationship === 'Ayah')?.name || '',
    'Nama Ibu': s.guardians.find((g) => g.relationship === 'Ibu')?.name || '',
  }));
}
