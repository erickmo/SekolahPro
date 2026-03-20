import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export const lessonPlansRouter = Router();
lessonPlansRouter.use(authenticate);

lessonPlansRouter.post('/', authorize('GURU', 'KEPALA_KURIKULUM'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.createLessonPlan(req.user!.schoolId, req.body)); } catch (err) { next(err); }
});

lessonPlansRouter.get('/', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { plans, meta } = await service.listLessonPlans(req.user!.schoolId, req.query as Record<string, string>);
    sendSuccess(res, plans, 200, meta);
  } catch (err) { next(err); }
});

lessonPlansRouter.get('/:id/versions', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getVersions(req.user!.schoolId, req.params.id as string)); } catch (err) { next(err); }
});

lessonPlansRouter.get('/:id', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getLessonPlan(req.user!.schoolId, req.params.id as string)); } catch (err) { next(err); }
});

lessonPlansRouter.patch('/:id', authorize('GURU', 'KEPALA_KURIKULUM'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.updateLessonPlan(req.user!.schoolId, req.params.id as string, req.body)); } catch (err) { next(err); }
});

lessonPlansRouter.post('/:id/publish', authorize('KEPALA_KURIKULUM'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.publishLessonPlan(req.user!.schoolId, req.params.id as string, req.user!.userId)); } catch (err) { next(err); }
});
