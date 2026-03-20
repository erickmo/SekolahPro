import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export const assetsRouter = Router();
assetsRouter.use(authenticate);

assetsRouter.get('/', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { assets, meta } = await service.getAssets(req.user!.schoolId, req.query as Record<string, string>);
    sendSuccess(res, assets, 200, meta);
  } catch (err) { next(err); }
});
assetsRouter.post('/', authorize('ADMIN_SEKOLAH', 'TATA_USAHA'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.createAsset(req.user!.schoolId, req.body)); } catch (err) { next(err); }
});
assetsRouter.put('/:id', authorize('ADMIN_SEKOLAH', 'TATA_USAHA'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.updateAsset(req.user!.schoolId, req.params.id as string, req.body)); } catch (err) { next(err); }
});
assetsRouter.post('/maintenance', authorize('ADMIN_SEKOLAH', 'TATA_USAHA'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.scheduleMaintenance(req.user!.schoolId, req.body)); } catch (err) { next(err); }
});
assetsRouter.get('/maintenance', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { maintenances, meta } = await service.getMaintenances(req.user!.schoolId, req.query as Record<string, string>);
    sendSuccess(res, maintenances, 200, meta);
  } catch (err) { next(err); }
});
