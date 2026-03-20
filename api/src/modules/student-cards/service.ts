import QRCode from 'qrcode';
import { prisma } from '../../lib/prisma';
import { NotFoundError, ConflictError } from '../../shared/errors';

export async function generateStudentCard(schoolId: string, studentId: string) {
  const student = await prisma.student.findFirst({ where: { id: studentId, schoolId }, include: { school: true } });
  if (!student) throw new NotFoundError('Siswa');

  const exists = await prisma.studentCard.findUnique({ where: { studentId } });
  if (exists && exists.isActive) throw new ConflictError('Kartu siswa sudah aktif');

  const qrData = JSON.stringify({ studentId, nisn: student.nisn, schoolId, ts: Date.now() });
  const qrCode = await QRCode.toDataURL(qrData);

  const validFrom = new Date();
  const validUntil = new Date();
  validUntil.setFullYear(validUntil.getFullYear() + 1);

  if (exists) {
    return prisma.studentCard.update({ where: { studentId }, data: { qrCode, validFrom, validUntil, isActive: true } });
  }
  return prisma.studentCard.create({ data: { studentId, qrCode, validFrom, validUntil, isActive: true } });
}

export async function getStudentCard(schoolId: string, studentId: string) {
  const student = await prisma.student.findFirst({ where: { id: studentId, schoolId } });
  if (!student) throw new NotFoundError('Siswa');
  const card = await prisma.studentCard.findUnique({ where: { studentId }, include: { student: { include: { school: true } } } });
  if (!card) throw new NotFoundError('Kartu siswa');
  return card;
}

export async function verifyQrCode(qrData: string) {
  const data = JSON.parse(qrData) as { studentId: string; nisn: string; schoolId: string };
  const card = await prisma.studentCard.findUnique({
    where: { studentId: data.studentId },
    include: { student: { select: { name: true, nisn: true, photoUrl: true } } },
  });
  if (!card || !card.isActive) return { valid: false, message: 'Kartu tidak valid' };
  if (new Date() > card.validUntil) return { valid: false, message: 'Kartu sudah kadaluarsa' };
  return { valid: true, student: card.student };
}

export async function deactivateCard(schoolId: string, studentId: string) {
  const student = await prisma.student.findFirst({ where: { id: studentId, schoolId } });
  if (!student) throw new NotFoundError('Siswa');
  return prisma.studentCard.update({ where: { studentId }, data: { isActive: false } });
}
