import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess } from '../../shared/response';
import { prisma } from '../../lib/prisma';

export const governmentReportsRouter = Router();
governmentReportsRouter.use(authenticate, authorize('ADMIN_SEKOLAH', 'KEPALA_SEKOLAH', 'TATA_USAHA', 'BENDAHARA'));

governmentReportsRouter.get('/student-stats', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const schoolId = req.user!.schoolId;
    const [total, male, female] = await Promise.all([
      prisma.student.count({ where: { schoolId } }),
      prisma.student.count({ where: { schoolId, gender: 'MALE' } }),
      prisma.student.count({ where: { schoolId, gender: 'FEMALE' } }),
    ]);
    const byClass = await prisma.enrollment.groupBy({ by: ['classId'], where: { schoolId, status: 'ACTIVE' }, _count: { classId: true } });
    sendSuccess(res, { total, male, female, byClass });
  } catch (err) { next(err); }
});

governmentReportsRouter.get('/attendance-report', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { startDate, endDate } = req.query as { startDate?: string; endDate?: string };
    const schoolId = req.user!.schoolId;
    const where: Record<string, unknown> = { schoolId };
    if (startDate || endDate) {
      where.date = {};
      if (startDate) (where.date as Record<string, unknown>).gte = new Date(startDate);
      if (endDate) (where.date as Record<string, unknown>).lte = new Date(endDate);
    }
    const attendance = await prisma.studentAttendance.groupBy({ by: ['status'], where, _count: { status: true } });
    sendSuccess(res, { period: { startDate, endDate }, attendance });
  } catch (err) { next(err); }
});

governmentReportsRouter.get('/financial-report', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const schoolId = req.user!.schoolId;
    const [collected, unpaid, totalStudents] = await Promise.all([
      prisma.invoice.aggregate({ where: { schoolId, status: 'PAID' }, _sum: { amount: true }, _count: true }),
      prisma.invoice.aggregate({ where: { schoolId, status: 'UNPAID' }, _sum: { amount: true }, _count: true }),
      prisma.student.count({ where: { schoolId } }),
    ]);
    sendSuccess(res, { collected: { total: collected._sum.amount || 0, count: collected._count }, unpaid: { total: unpaid._sum.amount || 0, count: unpaid._count }, totalStudents });
  } catch (err) { next(err); }
});
