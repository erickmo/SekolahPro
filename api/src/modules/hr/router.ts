import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export const hrRouter = Router();
hrRouter.use(authenticate);

hrRouter.get('/teachers', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { teachers, meta } = await service.getTeachers(req.user!.schoolId, req.query as Record<string, string>);
    sendSuccess(res, teachers, 200, meta);
  } catch (err) { next(err); }
});
hrRouter.post('/attendance', authorize('ADMIN_SEKOLAH', 'TATA_USAHA'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.recordTeacherAttendance(req.user!.schoolId, req.body.records)); } catch (err) { next(err); }
});
hrRouter.get('/attendance/:teacherId', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getTeacherAttendance(req.user!.schoolId, req.params.teacherId as string, req.query as { startDate?: string; endDate?: string })); } catch (err) { next(err); }
});
hrRouter.get('/attendance-summary', authorize('ADMIN_SEKOLAH', 'KEPALA_SEKOLAH'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getAttendanceSummary(req.user!.schoolId, Number(req.query.month) || new Date().getMonth() + 1, Number(req.query.year) || new Date().getFullYear())); } catch (err) { next(err); }
});
