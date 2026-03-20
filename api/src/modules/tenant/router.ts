import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export const tenantRouter = Router();
tenantRouter.use(authenticate);

tenantRouter.get('/my', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getMyTenant(req.user!.schoolId)); } catch (err) { next(err); }
});

tenantRouter.post('/', authorize('EDS_SUPERADMIN', 'EDS_SALES'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.createTenant(req.body)); } catch (err) { next(err); }
});

tenantRouter.get('/', authorize('EDS_SUPERADMIN', 'EDS_SUPPORT'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { tenants, meta } = await service.listTenants(req.query as Record<string, string>);
    sendSuccess(res, tenants, 200, meta);
  } catch (err) { next(err); }
});

tenantRouter.get('/:id', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getTenant(req.params.id as string)); } catch (err) { next(err); }
});

tenantRouter.patch('/:id/plan', authorize('EDS_SUPERADMIN'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.upgradePlan(req.params.id as string, req.body.tier)); } catch (err) { next(err); }
});

tenantRouter.patch('/:id/status', authorize('EDS_SUPERADMIN', 'EDS_SUPPORT'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.updateStatus(req.params.id as string, req.body.status)); } catch (err) { next(err); }
});

tenantRouter.get('/:id/features', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getFeatures(req.params.id as string)); } catch (err) { next(err); }
});

tenantRouter.patch('/:id/features', authorize('EDS_SUPERADMIN'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.toggleFeature(req.params.id as string, req.body.moduleId, req.body.enabled, req.user!.userId, req.body.config)); } catch (err) { next(err); }
});

tenantRouter.post('/:id/addons', authorize('EDS_SUPERADMIN', 'EDS_SALES'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.addAddon(req.params.id as string, req.body)); } catch (err) { next(err); }
});
