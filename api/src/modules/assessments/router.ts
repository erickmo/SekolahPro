import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export const assessmentsRouter = Router();
assessmentsRouter.use(authenticate);

assessmentsRouter.post('/rubrics', authorize('GURU', 'KEPALA_KURIKULUM'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.createRubric(req.user!.schoolId, req.body)); } catch (err) { next(err); }
});

assessmentsRouter.get('/rubrics', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { rubrics, meta } = await service.listRubrics(req.user!.schoolId, req.query as Record<string, string>);
    sendSuccess(res, rubrics, 200, meta);
  } catch (err) { next(err); }
});

assessmentsRouter.get('/rubrics/:id', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getRubric(req.user!.schoolId, req.params.id as string)); } catch (err) { next(err); }
});

assessmentsRouter.get('/projects/student/:studentId', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getStudentAssessments(req.user!.schoolId, req.params.studentId as string)); } catch (err) { next(err); }
});

assessmentsRouter.post('/projects', authorize('GURU'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.createProjectAssessment(req.user!.schoolId, req.user!.userId, req.body)); } catch (err) { next(err); }
});

assessmentsRouter.get('/projects', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { assessments, meta } = await service.listProjectAssessments(req.user!.schoolId, req.query as Record<string, string>);
    sendSuccess(res, assessments, 200, meta);
  } catch (err) { next(err); }
});
