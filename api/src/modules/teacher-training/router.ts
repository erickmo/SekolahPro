import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export const teacherTrainingRouter = Router();
teacherTrainingRouter.use(authenticate);

teacherTrainingRouter.get('/programs', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getPrograms(req.user!.schoolId)); } catch (err) { next(err); }
});
teacherTrainingRouter.post('/programs', authorize('ADMIN_SEKOLAH', 'KEPALA_SEKOLAH'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.createProgram(req.user!.schoolId, req.body)); } catch (err) { next(err); }
});
teacherTrainingRouter.post('/enroll', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.enroll(req.user!.schoolId, req.body.programId, req.body.teacherId || req.user!.userId)); } catch (err) { next(err); }
});
teacherTrainingRouter.get('/enrollments', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { enrollments, meta } = await service.getEnrollments(req.user!.schoolId, req.query as Record<string, string>);
    sendSuccess(res, enrollments, 200, meta);
  } catch (err) { next(err); }
});
teacherTrainingRouter.put('/enrollments/:id/complete', authorize('ADMIN_SEKOLAH', 'KEPALA_SEKOLAH'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.completTraining(req.user!.schoolId, req.params.id as string, req.body)); } catch (err) { next(err); }
});
