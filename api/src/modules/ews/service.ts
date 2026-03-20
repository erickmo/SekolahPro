import { prisma } from '../../lib/prisma';
import { callClaude } from '../../lib/claude';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';
import { NotFoundError } from '../../shared/errors';

interface RiskIndicators {
  attendanceRate: number;
  gradeDropPercent: number;
  failedSubjectsCount: number;
  disciplinaryRecords: number;
  consecutiveAbsences: number;
}

async function calculateRiskScore(indicators: RiskIndicators): Promise<number> {
  let score = 0;
  if (indicators.attendanceRate < 75) score += 30;
  else if (indicators.attendanceRate < 85) score += 15;
  if (indicators.gradeDropPercent > 20) score += 25;
  else if (indicators.gradeDropPercent > 10) score += 10;
  if (indicators.failedSubjectsCount >= 3) score += 20;
  else if (indicators.failedSubjectsCount >= 1) score += 10;
  if (indicators.disciplinaryRecords >= 3) score += 15;
  if (indicators.consecutiveAbsences >= 5) score += 20;
  else if (indicators.consecutiveAbsences >= 3) score += 10;
  return Math.min(100, score);
}

export async function analyzeStudent(schoolId: string, studentId: string) {
  const student = await prisma.student.findFirst({ where: { id: studentId, schoolId }, include: { guardians: true } });
  if (!student) throw new NotFoundError('Siswa');

  const thirtyDaysAgo = new Date();
  thirtyDaysAgo.setDate(thirtyDaysAgo.getDate() - 30);

  const attendances = await prisma.studentAttendance.findMany({ where: { studentId, date: { gte: thirtyDaysAgo } } });
  const grades = await prisma.grade.findMany({ where: { studentId, schoolId } });

  const totalDays = attendances.length || 1;
  const presentDays = attendances.filter((a) => a.status === 'HADIR').length;
  const attendanceRate = (presentDays / totalDays) * 100;

  let consecutiveAbsences = 0;
  let maxConsecutive = 0;
  for (const a of attendances.sort((x, y) => x.date.getTime() - y.date.getTime())) {
    if (a.status !== 'HADIR') { consecutiveAbsences++; maxConsecutive = Math.max(maxConsecutive, consecutiveAbsences); }
    else consecutiveAbsences = 0;
  }

  const failedSubjects = grades.filter((g) => g.score < 70).length;
  const indicators: RiskIndicators = { attendanceRate, gradeDropPercent: 0, failedSubjectsCount: failedSubjects, disciplinaryRecords: 0, consecutiveAbsences: maxConsecutive };
  const riskScore = await calculateRiskScore(indicators);

  await prisma.student.update({ where: { id: studentId }, data: { riskScore } });

  if (riskScore >= 50) {
    const recommendation = await getAIRecommendation(student.name, indicators, riskScore);
    const alert = await prisma.riskAlert.create({
      data: { studentId, schoolId, riskScore, indicators: indicators as unknown as never, recommendation, status: 'OPEN' },
    });
    return { student, riskScore, indicators, alert, recommendation };
  }

  return { student, riskScore, indicators };
}

async function getAIRecommendation(studentName: string, indicators: RiskIndicators, riskScore: number): Promise<string> {
  try {
    return await callClaude(
      'Kamu adalah konselor sekolah yang berpengalaman. Berikan rekomendasi intervensi yang konkret dan empati.',
      `Siswa "${studentName}" memiliki skor risiko ${riskScore}/100. Indikator: Kehadiran ${indicators.attendanceRate.toFixed(0)}%, ${indicators.failedSubjectsCount} mapel di bawah KKM, ${indicators.consecutiveAbsences} hari absen berturut-turut. Berikan 3 rekomendasi tindakan spesifik.`,
    );
  } catch {
    return 'Jadwalkan pertemuan dengan wali murid dan konselor BK sesegera mungkin.';
  }
}

export async function getAlerts(schoolId: string, query: { status?: string; page?: number; limit?: number }) {
  const { page, limit, skip } = getPaginationParams(query);
  const where: Record<string, unknown> = { student: { schoolId } };
  if (query.status) where.status = query.status;
  const [alerts, total] = await Promise.all([
    prisma.riskAlert.findMany({ where, skip, take: limit, orderBy: { riskScore: 'desc' }, include: { student: { select: { name: true, nisn: true } } } }),
    prisma.riskAlert.count({ where }),
  ]);
  return { alerts, meta: buildPaginationMeta(total, page, limit) };
}

export async function resolveAlert(alertId: string, assignedTo: string, note?: string) {
  return prisma.riskAlert.update({ where: { id: alertId }, data: { status: 'RESOLVED', assignedTo, resolvedAt: new Date() } });
}

export async function getHighRiskStudents(schoolId: string) {
  return prisma.student.findMany({ where: { schoolId, riskScore: { gte: 50 } }, orderBy: { riskScore: 'desc' }, select: { id: true, name: true, nisn: true, riskScore: true } });
}

export async function batchAnalyze(schoolId: string) {
  const students = await prisma.student.findMany({ where: { schoolId }, select: { id: true } });
  const results = await Promise.allSettled(students.map((s) => analyzeStudent(schoolId, s.id)));
  return { analyzed: students.length, alerts: results.filter((r) => r.status === 'fulfilled' && (r.value as { alert?: unknown }).alert).length };
}
