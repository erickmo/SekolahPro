import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export const internshipRouter = Router();
internshipRouter.use(authenticate);

internshipRouter.post('/', authorize('ADMIN_SEKOLAH', 'GURU', 'OPERATOR_SIMS'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.createPlacement(req.user!.schoolId, req.body)); } catch (err) { next(err); }
});
internshipRouter.get('/', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { placements, meta } = await service.getPlacements(req.user!.schoolId, req.query as Record<string, string>);
    sendSuccess(res, placements, 200, meta);
  } catch (err) { next(err); }
});
internshipRouter.post('/:id/journals', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.addJournalEntry(req.user!.schoolId, req.params.id as string, req.body.studentId, req.body)); } catch (err) { next(err); }
});
internshipRouter.get('/:id/journals', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getJournals(req.user!.schoolId, req.params.id as string)); } catch (err) { next(err); }
});
internshipRouter.post('/:id/grade', authorize('GURU', 'ADMIN_SEKOLAH', 'MITRA_DU_DI'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.submitGrade(req.user!.schoolId, req.params.id as string, req.body)); } catch (err) { next(err); }
});
