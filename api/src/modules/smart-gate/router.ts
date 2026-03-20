import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import { prisma } from '../../lib/prisma';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';

export const smartGateRouter = Router();
smartGateRouter.use(authenticate);

smartGateRouter.post('/checkin', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { studentId, method, deviceId } = req.body;
    const student = await prisma.student.findFirst({ where: { id: studentId, schoolId: req.user!.schoolId } });
    if (!student) { res.status(404).json({ success: false, error: { code: 'NOT_FOUND', message: 'Siswa tidak ditemukan' } }); return; }
    const today = new Date(); today.setHours(0,0,0,0);
    const tomorrow = new Date(today); tomorrow.setDate(tomorrow.getDate() + 1);
    const existing = await prisma.studentAttendance.findFirst({ where: { studentId, date: { gte: today, lt: tomorrow } } });
    if (!existing) {
      await prisma.studentAttendance.create({ data: { studentId, schoolId: req.user!.schoolId, date: new Date(), status: 'HADIR', method: method || 'SMART_GATE' } });
    }
    sendCreated(res, { student: { name: student.name, nisn: student.nisn }, checkinTime: new Date(), method: method || 'SMART_GATE' });
  } catch (err) { next(err); }
});

smartGateRouter.get('/today', authorize('ADMIN_SEKOLAH', 'TATA_USAHA', 'GURU'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const today = new Date(); today.setHours(0,0,0,0);
    const tomorrow = new Date(today); tomorrow.setDate(tomorrow.getDate() + 1);
    const { page, limit, skip } = getPaginationParams({ page: Number(req.query.page), limit: Number(req.query.limit) });
    const [records, total] = await Promise.all([
      prisma.studentAttendance.findMany({ where: { schoolId: req.user!.schoolId, date: { gte: today, lt: tomorrow } }, skip, take: limit, include: { student: { select: { name: true, nisn: true } } }, orderBy: { date: 'desc' } }),
      prisma.studentAttendance.count({ where: { schoolId: req.user!.schoolId, date: { gte: today, lt: tomorrow } } }),
    ]);
    sendSuccess(res, records, 200, buildPaginationMeta(total, page, limit));
  } catch (err) { next(err); }
});

smartGateRouter.post('/visitor', authenticate, async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { name, purpose, hostId } = req.body;
    sendCreated(res, { visitorId: `VIS-${Date.now()}`, name, purpose, checkIn: new Date(), badgeCode: `BADGE-${Date.now()}` });
  } catch (err) { next(err); }
});
