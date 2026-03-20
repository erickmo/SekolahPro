import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export const extracurricularRouter = Router();
extracurricularRouter.use(authenticate);

extracurricularRouter.get('/', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getExtraCurriculars(req.user!.schoolId)); } catch (err) { next(err); }
});
extracurricularRouter.post('/', authorize('ADMIN_SEKOLAH', 'GURU'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.createExtraCurricular(req.user!.schoolId, req.body)); } catch (err) { next(err); }
});
extracurricularRouter.post('/enroll', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.enrollStudent(req.user!.schoolId, req.body)); } catch (err) { next(err); }
});
extracurricularRouter.get('/enrollments', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { enrollments, meta } = await service.getEnrollments(req.user!.schoolId, req.query as Record<string, string>);
    sendSuccess(res, enrollments, 200, meta);
  } catch (err) { next(err); }
});
extracurricularRouter.post('/activities', authorize('GURU', 'ADMIN_SEKOLAH'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.recordActivity(req.user!.schoolId, req.body)); } catch (err) { next(err); }
});
extracurricularRouter.post('/activities/:activityId/attendance', authorize('GURU'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.recordAttendance(req.user!.schoolId, req.params.activityId as string, req.body.records)); } catch (err) { next(err); }
});
