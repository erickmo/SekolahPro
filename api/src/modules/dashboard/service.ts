import { prisma } from '../../lib/prisma';
import { callClaude } from '../../lib/claude';

export async function getSchoolDashboard(schoolId: string) {
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  const tomorrow = new Date(today);
  tomorrow.setDate(tomorrow.getDate() + 1);

  const [totalStudents, presentToday, totalTeachers, unpaidInvoices, openTickets, highRiskStudents] = await Promise.all([
    prisma.student.count({ where: { schoolId } }),
    prisma.studentAttendance.count({ where: { schoolId, date: { gte: today, lt: tomorrow }, status: 'HADIR' } }),
    prisma.teacher.count({ where: { schoolId } }),
    prisma.invoice.count({ where: { schoolId, status: 'UNPAID' } }),
    prisma.helpdeskTicket.count({ where: { schoolId, status: 'OPEN' } }),
    prisma.student.count({ where: { schoolId, riskScore: { gte: 50 } } }),
  ]);

  const attendanceRate = totalStudents > 0 ? Math.round((presentToday / totalStudents) * 100) : 0;

  return { totalStudents, presentToday, attendanceRate, totalTeachers, unpaidInvoices, openTickets, highRiskStudents, generatedAt: new Date().toISOString() };
}

export async function getWeeklyInsight(schoolId: string) {
  const data = await getSchoolDashboard(schoolId);

  try {
    const insight = await callClaude(
      'Kamu adalah asisten kepala sekolah. Berikan rangkuman kondisi sekolah secara singkat dan profesional dalam Bahasa Indonesia.',
      `Data sekolah minggu ini: ${JSON.stringify(data)}. Berikan insight dalam 3-4 kalimat.`,
      1000,
    );
    return { insight, data };
  } catch {
    return { insight: 'Layanan AI sedang tidak tersedia.', data };
  }
}

export async function getAttendanceTrend(schoolId: string, days = 30) {
  const startDate = new Date();
  startDate.setDate(startDate.getDate() - days);

  const attendances = await prisma.studentAttendance.groupBy({
    by: ['date', 'status'],
    where: { schoolId, date: { gte: startDate } },
    _count: { status: true },
    orderBy: { date: 'asc' },
  });

  const trend: Record<string, Record<string, number>> = {};
  for (const a of attendances) {
    const dateKey = a.date.toISOString().split('T')[0];
    if (!trend[dateKey]) trend[dateKey] = {};
    trend[dateKey][a.status] = a._count.status;
  }

  return Object.entries(trend).map(([date, statuses]) => ({ date, ...statuses }));
}

export async function getFoundationDashboard(adminYayasanId: string) {
  const schools = await prisma.school.findMany({ select: { id: true, name: true } });

  const stats = await Promise.all(
    schools.map(async (school) => {
      const [students, teachers, unpaid] = await Promise.all([
        prisma.student.count({ where: { schoolId: school.id } }),
        prisma.teacher.count({ where: { schoolId: school.id } }),
        prisma.invoice.count({ where: { schoolId: school.id, status: 'UNPAID' } }),
      ]);
      return { ...school, students, teachers, unpaidInvoices: unpaid };
    }),
  );

  return { totalSchools: schools.length, schools: stats };
}
