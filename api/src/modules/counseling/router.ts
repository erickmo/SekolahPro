import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export const counselingRouter = Router();
counselingRouter.use(authenticate);

// Counseling sessions
counselingRouter.post('/sessions', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.bookSession(req.user!.schoolId, req.body)); } catch (err) { next(err); }
});
counselingRouter.get('/sessions', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { sessions, meta } = await service.getSessions(req.user!.schoolId, req.query as Record<string, string>);
    sendSuccess(res, sessions, 200, meta);
  } catch (err) { next(err); }
});
counselingRouter.put('/sessions/:id', authorize('GURU_BK'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.updateSession(req.user!.schoolId, req.params.id as string, req.body)); } catch (err) { next(err); }
});

// Bullying reports
counselingRouter.post('/bullying-reports', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.reportBullying(req.user!.schoolId, req.body)); } catch (err) { next(err); }
});
counselingRouter.get('/bullying-reports', authorize('GURU_BK', 'KEPALA_SEKOLAH', 'ADMIN_SEKOLAH'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { reports, meta } = await service.getBullyingReports(req.user!.schoolId, req.query as Record<string, string>);
    sendSuccess(res, reports, 200, meta);
  } catch (err) { next(err); }
});
counselingRouter.put('/bullying-reports/:id', authorize('GURU_BK', 'KEPALA_SEKOLAH'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.updateBullyingReport(req.user!.schoolId, req.params.id as string, req.body)); } catch (err) { next(err); }
});
