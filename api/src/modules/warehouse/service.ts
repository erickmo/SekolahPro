import { prisma } from '../../lib/prisma';
import { NotFoundError, ValidationError } from '../../shared/errors';

export async function getDataSummary(schoolId: string) {
  const [
    totalStudents,
    totalGrades,
    totalInvoices,
    totalPayments,
    totalUKSVisits,
    totalNutritionRecords,
    totalAlerts,
  ] = await Promise.all([
    prisma.student.count({ where: { schoolId, isActive: true } }),
    prisma.grade.count({ where: { schoolId } }),
    prisma.invoice.count({ where: { schoolId } }),
    prisma.payment.count({ where: { schoolId } }),
    prisma.uKSVisit.count({ where: { schoolId } }),
    prisma.nutritionRecord.count({ where: { schoolId } }),
    prisma.riskAlert.count({ where: { schoolId } }),
  ]);

  return {
    students: totalStudents,
    grades: totalGrades,
    invoices: totalInvoices,
    payments: totalPayments,
    uksVisits: totalUKSVisits,
    nutritionRecords: totalNutritionRecords,
    riskAlerts: totalAlerts,
    generatedAt: new Date().toISOString(),
  };
}

export async function getAcademicPerformance(
  schoolId: string,
  query: { academicYearId?: string; subjectId?: string },
) {
  const where: Record<string, unknown> = { schoolId };
  if (query.subjectId) where.subjectId = query.subjectId;
  if (query.academicYearId) {
    // Filter grades via semester -> academicYear
    where.semester = { academicYearId: query.academicYearId };
  }

  const grades = await prisma.grade.findMany({
    where,
    select: {
      score: true,
      subjectId: true,
      subject: { select: { name: true } },
      student: {
        select: {
          enrollments: {
            select: {
              classId: true,
              schoolClass: { select: { name: true } },
            },
            take: 1,
            orderBy: { enrolledAt: 'desc' },
          },
        },
      },
    },
  });

  // Aggregate avg score per subject
  const subjectMap: Record<string, { subjectId: string; subjectName: string; totalScore: number; count: number }> = {};

  for (const g of grades) {
    if (!subjectMap[g.subjectId]) {
      subjectMap[g.subjectId] = {
        subjectId: g.subjectId,
        subjectName: g.subject.name,
        totalScore: 0,
        count: 0,
      };
    }
    subjectMap[g.subjectId].totalScore += g.score;
    subjectMap[g.subjectId].count += 1;
  }

  const performance = Object.values(subjectMap).map((s) => ({
    subjectId: s.subjectId,
    subjectName: s.subjectName,
    averageScore: s.count > 0 ? Math.round((s.totalScore / s.count) * 100) / 100 : 0,
    totalGrades: s.count,
  }));

  return { performance, generatedAt: new Date().toISOString() };
}

export async function getFinancialReport(
  schoolId: string,
  query: { startDate?: string; endDate?: string },
) {
  const dateFilter: Record<string, Date> = {};
  if (query.startDate) dateFilter.gte = new Date(query.startDate);
  if (query.endDate) dateFilter.lte = new Date(query.endDate);

  const invoiceWhere: Record<string, unknown> = { schoolId };
  const paymentWhere: Record<string, unknown> = { schoolId };

  if (Object.keys(dateFilter).length) {
    invoiceWhere.createdAt = dateFilter;
    paymentWhere.createdAt = dateFilter;
  }

  const [invoiceAggregate, paymentAggregate, invoicesByStatus] = await Promise.all([
    prisma.invoice.aggregate({
      where: invoiceWhere,
      _sum: { amount: true },
      _count: { id: true },
    }),
    prisma.payment.aggregate({
      where: { ...paymentWhere, status: 'SUCCESS' },
      _sum: { amount: true },
      _count: { id: true },
    }),
    prisma.invoice.groupBy({
      by: ['status'],
      where: invoiceWhere,
      _count: { id: true },
      _sum: { amount: true },
    }),
  ]);

  return {
    totalInvoiced: invoiceAggregate._sum.amount ?? 0,
    totalInvoiceCount: invoiceAggregate._count.id,
    totalCollected: paymentAggregate._sum.amount ?? 0,
    totalPaymentCount: paymentAggregate._count.id,
    invoicesByStatus: invoicesByStatus.map((s) => ({
      status: s.status,
      count: s._count.id,
      amount: s._sum.amount ?? 0,
    })),
    generatedAt: new Date().toISOString(),
  };
}

export async function getHealthStats(
  schoolId: string,
  query: { startDate?: string; endDate?: string },
) {
  const dateFilter: Record<string, Date> = {};
  if (query.startDate) dateFilter.gte = new Date(query.startDate);
  if (query.endDate) dateFilter.lte = new Date(query.endDate);

  const uksWhere: Record<string, unknown> = { schoolId };
  const nutritionWhere: Record<string, unknown> = { schoolId };

  if (Object.keys(dateFilter).length) {
    uksWhere.visitDate = dateFilter;
    nutritionWhere.measuredAt = dateFilter;
  }

  const [uksCount, uksComplaints, nutritionAggregate, nutritionByStatus] = await Promise.all([
    prisma.uKSVisit.count({ where: uksWhere }),
    prisma.uKSVisit.groupBy({
      by: ['complaint'],
      where: uksWhere,
      _count: { id: true },
      orderBy: { _count: { id: 'desc' } },
    }),
    prisma.nutritionRecord.aggregate({
      where: nutritionWhere,
      _avg: { bmi: true, weight: true, height: true },
      _count: { id: true },
    }),
    prisma.nutritionRecord.groupBy({
      by: ['nutritionStatus'],
      where: nutritionWhere,
      _count: { id: true },
    }),
  ]);

  return {
    uksVisits: {
      total: uksCount,
      topComplaints: uksComplaints.slice(0, 5).map((c) => ({
        complaint: c.complaint,
        count: c._count.id,
      })),
    },
    nutrition: {
      totalRecords: nutritionAggregate._count.id,
      averageBmi: nutritionAggregate._avg.bmi ?? 0,
      averageWeight: nutritionAggregate._avg.weight ?? 0,
      averageHeight: nutritionAggregate._avg.height ?? 0,
      byStatus: nutritionByStatus.map((s) => ({
        status: s.nutritionStatus,
        count: s._count.id,
      })),
    },
    generatedAt: new Date().toISOString(),
  };
}

const EXPORTABLE_ENTITIES: Record<string, () => Promise<unknown>> = {};

export async function exportData(schoolId: string, entity?: string, format = 'json') {
  if (!entity) throw new ValidationError('Parameter entity wajib diisi');

  let data: unknown;

  switch (entity.toLowerCase()) {
    case 'students':
      data = await prisma.student.findMany({
        where: { schoolId },
        select: { id: true, nisn: true, name: true, gender: true, birthDate: true, isActive: true, createdAt: true },
      });
      break;
    case 'grades':
      data = await prisma.grade.findMany({
        where: { schoolId },
        select: { id: true, studentId: true, subjectId: true, score: true, type: true, createdAt: true },
      });
      break;
    case 'invoices':
      data = await prisma.invoice.findMany({
        where: { schoolId },
        select: { id: true, studentId: true, amount: true, status: true, dueDate: true, paidAt: true, createdAt: true },
      });
      break;
    case 'attendance':
      data = await prisma.studentAttendance.findMany({
        where: { schoolId },
        select: { id: true, studentId: true, date: true, status: true, createdAt: true },
      });
      break;
    case 'nutrition':
      data = await prisma.nutritionRecord.findMany({
        where: { schoolId },
        select: { id: true, studentId: true, weight: true, height: true, bmi: true, nutritionStatus: true, measuredAt: true },
      });
      break;
    case 'uks':
      data = await prisma.uKSVisit.findMany({
        where: { schoolId },
        select: { id: true, studentId: true, complaint: true, action: true, visitDate: true, createdAt: true },
      });
      break;
    default:
      throw new NotFoundError(`Entity '${entity}' tidak didukung. Pilih: students, grades, invoices, attendance, nutrition, uks`);
  }

  return {
    entity,
    format,
    exportedAt: new Date().toISOString(),
    count: Array.isArray(data) ? data.length : 0,
    data,
  };
}
