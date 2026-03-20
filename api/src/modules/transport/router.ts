import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export const transportRouter = Router();
transportRouter.use(authenticate);

transportRouter.get('/buses', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getBuses(req.user!.schoolId)); } catch (err) { next(err); }
});
transportRouter.post('/buses', authorize('ADMIN_SEKOLAH', 'TATA_USAHA'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.createBus(req.user!.schoolId, req.body)); } catch (err) { next(err); }
});
transportRouter.get('/routes', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getRoutes(req.user!.schoolId)); } catch (err) { next(err); }
});
transportRouter.post('/routes', authorize('ADMIN_SEKOLAH', 'TATA_USAHA'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.createRoute(req.user!.schoolId, req.body)); } catch (err) { next(err); }
});
transportRouter.post('/rides', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.recordRide(req.user!.schoolId, req.body)); } catch (err) { next(err); }
});
transportRouter.get('/rides', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { logs, meta } = await service.getRideLogs(req.user!.schoolId, req.query as Record<string, string>);
    sendSuccess(res, logs, 200, meta);
  } catch (err) { next(err); }
});
