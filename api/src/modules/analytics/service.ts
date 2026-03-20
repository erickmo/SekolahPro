import { prisma } from '../../lib/prisma';
import { NotFoundError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';

export async function listRiskAlerts(
  schoolId: string,
  query: {
    severity?: string;
    resolved?: boolean;
    page?: number;
    limit?: number;
  },
) {
  const { page, limit, skip } = getPaginationParams(query);

  const where: Record<string, unknown> = { schoolId };

  // Map resolved boolean to AlertStatus
  if (query.resolved === true) {
    where.status = 'RESOLVED';
  } else if (query.resolved === false) {
    where.status = { in: ['OPEN', 'IN_PROGRESS'] };
  }

  // severity maps to riskScore ranges: LOW <40, MEDIUM 40-69, HIGH >=70
  if (query.severity) {
    switch (query.severity.toUpperCase()) {
      case 'LOW':
        where.riskScore = { lt: 40 };
        break;
      case 'MEDIUM':
        where.riskScore = { gte: 40, lt: 70 };
        break;
      case 'HIGH':
        where.riskScore = { gte: 70 };
        break;
    }
  }

  const [alerts, total] = await Promise.all([
    prisma.riskAlert.findMany({
      where,
      skip,
      take: limit,
      orderBy: { createdAt: 'desc' },
      include: {
        student: { select: { id: true, name: true, nisn: true } },
      },
    }),
    prisma.riskAlert.count({ where }),
  ]);

  return { alerts, meta: buildPaginationMeta(total, page, limit) };
}

export async function getStudentRiskAlerts(schoolId: string, studentId: string) {
  const student = await prisma.student.findFirst({ where: { id: studentId, schoolId } });
  if (!student) throw new NotFoundError('Siswa');

  return prisma.riskAlert.findMany({
    where: { studentId, schoolId },
    orderBy: { createdAt: 'desc' },
  });
}

export async function resolveAlert(alertId: string, resolvedBy: string) {
  const alert = await prisma.riskAlert.findUnique({ where: { id: alertId } });
  if (!alert) throw new NotFoundError('Alert');

  return prisma.riskAlert.update({
    where: { id: alertId },
    data: {
      status: 'RESOLVED',
      assignedTo: resolvedBy,
      resolvedAt: new Date(),
    },
  });
}

export async function getSchoolOverview(schoolId: string) {
  const [totalStudents, activeAlerts, gradeAggregate] = await Promise.all([
    prisma.student.count({ where: { schoolId, isActive: true } }),
    prisma.riskAlert.count({ where: { schoolId, status: { in: ['OPEN', 'IN_PROGRESS'] } } }),
    prisma.grade.aggregate({
      where: { schoolId },
      _avg: { score: true },
    }),
  ]);

  return {
    totalStudents,
    activeAlerts,
    averageGrade: gradeAggregate._avg.score ?? 0,
  };
}

export async function getAttendanceTrend(
  schoolId: string,
  query: {
    startDate?: string;
    endDate?: string;
    classId?: string;
  },
) {
  const where: Record<string, unknown> = { schoolId };

  const dateFilter: Record<string, Date> = {};
  if (query.startDate) dateFilter.gte = new Date(query.startDate);
  if (query.endDate) dateFilter.lte = new Date(query.endDate);
  if (Object.keys(dateFilter).length) where.date = dateFilter;

  // If classId provided, filter by students enrolled in that class
  if (query.classId) {
    const enrolledStudentIds = await prisma.enrollment
      .findMany({ where: { classId: query.classId, schoolId }, select: { studentId: true } })
      .then((rows) => rows.map((r) => r.studentId));
    where.studentId = { in: enrolledStudentIds };
  }

  const attendances = await prisma.studentAttendance.findMany({
    where,
    select: { date: true, status: true },
    orderBy: { date: 'asc' },
  });

  // Group by date string
  const grouped: Record<string, { date: string; present: number; absent: number; late: number; sick: number; permission: number; total: number }> = {};

  for (const record of attendances) {
    const dateKey = record.date.toISOString().split('T')[0];
    if (!grouped[dateKey]) {
      grouped[dateKey] = { date: dateKey, present: 0, absent: 0, late: 0, sick: 0, permission: 0, total: 0 };
    }
    grouped[dateKey].total += 1;
    const status = record.status.toUpperCase();
    if (status === 'HADIR' || status === 'PRESENT') grouped[dateKey].present += 1;
    else if (status === 'ALPHA' || status === 'ABSENT') grouped[dateKey].absent += 1;
    else if (status === 'TERLAMBAT' || status === 'LATE') grouped[dateKey].late += 1;
    else if (status === 'SAKIT' || status === 'SICK') grouped[dateKey].sick += 1;
    else if (status === 'IZIN' || status === 'PERMISSION') grouped[dateKey].permission += 1;
  }

  return Object.values(grouped);
}
