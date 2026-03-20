import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export const dapodikRouter = Router();
dapodikRouter.use(authenticate, authorize('ADMIN_SEKOLAH', 'KEPALA_SEKOLAH', 'TATA_USAHA', 'SUPERADMIN'));

dapodikRouter.post('/sync', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.syncDapodik(req.user!.schoolId)); } catch (err) { next(err); }
});
dapodikRouter.get('/history', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getSyncHistory(req.user!.schoolId)); } catch (err) { next(err); }
});
dapodikRouter.get('/export/students', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.exportStudentData(req.user!.schoolId)); } catch (err) { next(err); }
});
