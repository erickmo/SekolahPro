import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess } from '../../shared/response';
import { prisma } from '../../lib/prisma';

export const parentPortalRouter = Router();
parentPortalRouter.use(authenticate);

parentPortalRouter.get('/children', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const guardians = await prisma.guardian.findMany({ where: { email: req.user!.email }, include: { student: { include: { enrollments: { include: { schoolClass: true } }, savingsAccount: true } } } });
    sendSuccess(res, guardians.map((g) => g.student));
  } catch (err) { next(err); }
});

parentPortalRouter.get('/children/:studentId/grades', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const grades = await prisma.grade.findMany({ where: { studentId: req.params.studentId as string }, include: { subject: { select: { name: true } } }, orderBy: { subject: { name: 'asc' } } });
    sendSuccess(res, grades);
  } catch (err) { next(err); }
});

parentPortalRouter.get('/children/:studentId/attendance', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const attendances = await prisma.studentAttendance.findMany({ where: { studentId: req.params.studentId as string }, orderBy: { date: 'desc' }, take: 30 });
    sendSuccess(res, attendances);
  } catch (err) { next(err); }
});

parentPortalRouter.get('/children/:studentId/invoices', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const invoices = await prisma.invoice.findMany({ where: { studentId: req.params.studentId as string }, orderBy: { createdAt: 'desc' } });
    sendSuccess(res, invoices);
  } catch (err) { next(err); }
});

parentPortalRouter.get('/children/:studentId/savings', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const account = await prisma.savingsAccount.findUnique({ where: { studentId: req.params.studentId as string }, include: { transactions: { orderBy: { createdAt: 'desc' }, take: 10 } } });
    sendSuccess(res, account);
  } catch (err) { next(err); }
});
