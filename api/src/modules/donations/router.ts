import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export const donationsRouter = Router();
donationsRouter.use(authenticate);

donationsRouter.post('/', authorize('BENDAHARA', 'KOMITE_SEKOLAH'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.createDonation(req.user!.schoolId, req.body)); } catch (err) { next(err); }
});

donationsRouter.get('/summary', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getDonationSummary(req.user!.schoolId)); } catch (err) { next(err); }
});

donationsRouter.get('/', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { donations, meta } = await service.listDonations(req.user!.schoolId, req.query as Record<string, string>);
    sendSuccess(res, donations, 200, meta);
  } catch (err) { next(err); }
});

donationsRouter.get('/:id', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getDonation(req.user!.schoolId, req.params.id as string)); } catch (err) { next(err); }
});
