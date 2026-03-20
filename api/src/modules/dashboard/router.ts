import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess } from '../../shared/response';
import * as service from './service';

export const dashboardRouter = Router();
dashboardRouter.use(authenticate);

dashboardRouter.get('/', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getSchoolDashboard(req.user!.schoolId)); } catch (err) { next(err); }
});
dashboardRouter.get('/insight', authorize('KEPALA_SEKOLAH', 'ADMIN_SEKOLAH'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getWeeklyInsight(req.user!.schoolId)); } catch (err) { next(err); }
});
dashboardRouter.get('/attendance-trend', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getAttendanceTrend(req.user!.schoolId, Number(req.query.days) || 30)); } catch (err) { next(err); }
});
dashboardRouter.get('/foundation', authorize('ADMIN_YAYASAN', 'SUPERADMIN'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getFoundationDashboard(req.user!.userId)); } catch (err) { next(err); }
});
