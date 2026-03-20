import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export const listingRouter = Router();

// Public
listingRouter.get('/active', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getActivePlacements(req.query.schoolId as string | undefined)); } catch (err) { next(err); }
});

listingRouter.use(authenticate);

listingRouter.post('/vendors', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.registerVendor(req.body)); } catch (err) { next(err); }
});

listingRouter.get('/vendors', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { vendors, meta } = await service.listVendors(req.query as Record<string, string>);
    sendSuccess(res, vendors, 200, meta);
  } catch (err) { next(err); }
});

listingRouter.get('/vendors/:id', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getVendor(req.params.id as string)); } catch (err) { next(err); }
});

listingRouter.patch('/vendors/:id/approve', authorize('EDS_SUPERADMIN', 'LISTING_MANAGER'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.approveVendor(req.params.id as string, req.user!.userId)); } catch (err) { next(err); }
});

listingRouter.post('/placements', authorize('LISTING_VENDOR', 'EDS_SUPERADMIN'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.createPlacement(req.body.vendorId, req.body)); } catch (err) { next(err); }
});

listingRouter.get('/placements', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { placements, meta } = await service.listPlacements(req.query as Record<string, string>);
    sendSuccess(res, placements, 200, meta);
  } catch (err) { next(err); }
});

listingRouter.post('/placements/:id/review', authorize('ADMIN_SEKOLAH', 'KEPALA_SEKOLAH'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    sendCreated(res, await service.addReview(req.params.id as string, req.user!.schoolId, req.user!.userId, req.body));
  } catch (err) { next(err); }
});
