import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess } from '../../shared/response';
import * as dashboardService from '../dashboard/service';

export const foundationRouter = Router();
foundationRouter.use(authenticate, authorize('ADMIN_YAYASAN', 'SUPERADMIN'));

foundationRouter.get('/dashboard', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await dashboardService.getFoundationDashboard(req.user!.userId)); } catch (err) { next(err); }
});
