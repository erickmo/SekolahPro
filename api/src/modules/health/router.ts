import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export const healthRouter = Router();
healthRouter.use(authenticate);

healthRouter.post('/records', authorize('PETUGAS_UKS', 'ADMIN_SEKOLAH'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.createHealthRecord(req.user!.schoolId, req.body)); } catch (err) { next(err); }
});
healthRouter.get('/records/:studentId', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getHealthRecord(req.user!.schoolId, req.params.studentId as string)); } catch (err) { next(err); }
});
healthRouter.post('/uks-visits', authorize('PETUGAS_UKS'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.recordUKSVisit(req.user!.schoolId, { ...req.body, staffId: req.user!.userId })); } catch (err) { next(err); }
});
healthRouter.get('/uks-visits', authorize('PETUGAS_UKS', 'ADMIN_SEKOLAH', 'KEPALA_SEKOLAH'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { visits, meta } = await service.getUKSVisits(req.user!.schoolId, req.query as Record<string, string>);
    sendSuccess(res, visits, 200, meta);
  } catch (err) { next(err); }
});
healthRouter.post('/nutrition', authorize('PETUGAS_UKS', 'ADMIN_SEKOLAH'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.recordNutrition(req.user!.schoolId, req.body)); } catch (err) { next(err); }
});
healthRouter.get('/nutrition/:studentId', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getNutritionRecords(req.user!.schoolId, req.params.studentId as string)); } catch (err) { next(err); }
});
