import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export const budgetRouter = Router();
budgetRouter.use(authenticate);

budgetRouter.post('/plans', authorize('BENDAHARA', 'KEPALA_SEKOLAH'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.createPlan(req.user!.schoolId, req.body)); } catch (err) { next(err); }
});

budgetRouter.get('/plans', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { plans, meta } = await service.listPlans(req.user!.schoolId, req.query as Record<string, string>);
    sendSuccess(res, plans, 200, meta);
  } catch (err) { next(err); }
});

budgetRouter.get('/summary', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getBudgetSummary(req.user!.schoolId, req.query.planId as string)); } catch (err) { next(err); }
});

budgetRouter.get('/plans/:id', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getPlan(req.user!.schoolId, req.params.id as string)); } catch (err) { next(err); }
});

budgetRouter.patch('/plans/:id/approve', authorize('KEPALA_SEKOLAH'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.approvePlan(req.user!.schoolId, req.params.id as string, req.user!.userId)); } catch (err) { next(err); }
});

budgetRouter.post('/plans/:id/lines', authorize('BENDAHARA'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.addLine(req.user!.schoolId, req.params.id as string, req.body)); } catch (err) { next(err); }
});

budgetRouter.post('/realizations', authorize('BENDAHARA'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.addRealization(req.user!.schoolId, { ...req.body, recordedBy: req.user!.userId })); } catch (err) { next(err); }
});

budgetRouter.get('/realizations', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { realizations, meta } = await service.listRealizations(req.user!.schoolId, req.query as Record<string, string>);
    sendSuccess(res, realizations, 200, meta);
  } catch (err) { next(err); }
});
