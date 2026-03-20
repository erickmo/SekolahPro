import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { optionalAuth } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { Request } from 'express';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export const ppdbRouter = Router();

// Public routes
ppdbRouter.get('/periods/public', async (req: Request, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getPeriods(req.query.schoolId as string)); } catch (err) { next(err); }
});
ppdbRouter.post('/apply', async (req: Request, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.applyPPDB(req.body)); } catch (err) { next(err); }
});

// Protected routes
ppdbRouter.use(authenticate);
ppdbRouter.get('/periods', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getPeriods(req.user!.schoolId)); } catch (err) { next(err); }
});
ppdbRouter.post('/periods', authorize('ADMIN_SEKOLAH', 'OPERATOR_SIMS', 'KEPALA_SEKOLAH'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.createPeriod(req.user!.schoolId, req.body)); } catch (err) { next(err); }
});
ppdbRouter.get('/applications', authorize('ADMIN_SEKOLAH', 'OPERATOR_SIMS'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { applications, meta } = await service.getApplications(req.user!.schoolId, req.query as Record<string, string>);
    sendSuccess(res, applications, 200, meta);
  } catch (err) { next(err); }
});
ppdbRouter.put('/applications/:id', authorize('ADMIN_SEKOLAH', 'OPERATOR_SIMS'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.updateApplicationStatus(req.params.id as string, req.body.status, req.body.notes)); } catch (err) { next(err); }
});
ppdbRouter.get('/dashboard', authorize('ADMIN_SEKOLAH', 'KEPALA_SEKOLAH', 'OPERATOR_SIMS'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getDashboard(req.user!.schoolId, req.query.periodId as string)); } catch (err) { next(err); }
});
