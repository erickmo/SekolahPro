import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export const remediationRouter = Router();
remediationRouter.use(authenticate);

remediationRouter.post('/programs', authorize('GURU', 'KEPALA_KURIKULUM'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.createProgram(req.user!.schoolId, req.body)); } catch (err) { next(err); }
});

remediationRouter.get('/programs', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { programs, meta } = await service.listPrograms(req.user!.schoolId, req.query as Record<string, string>);
    sendSuccess(res, programs, 200, meta);
  } catch (err) { next(err); }
});

remediationRouter.post('/programs/:id/sessions', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.addSession(req.user!.schoolId, req.params.id as string, req.body)); } catch (err) { next(err); }
});

remediationRouter.get('/programs/:id/sessions', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.listSessions(req.user!.schoolId, req.params.id as string)); } catch (err) { next(err); }
});

remediationRouter.patch('/sessions/:id', authorize('GURU'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.updateSession(req.user!.schoolId, req.params.id as string, req.body)); } catch (err) { next(err); }
});

remediationRouter.get('/students/:studentId', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getStudentRemediations(req.user!.schoolId, req.params.studentId as string)); } catch (err) { next(err); }
});
