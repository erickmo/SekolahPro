import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export const eduMarketplaceRouter = Router();

// Public route — no authenticate middleware
eduMarketplaceRouter.get('/content', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { content, meta } = await service.listContent(req.query as Record<string, string>);
    sendSuccess(res, content, 200, meta);
  } catch (err) { next(err); }
});

eduMarketplaceRouter.post('/content', authenticate, authorize('GURU', 'ADMIN_SEKOLAH'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    sendCreated(res, await service.createContent(req.user!.schoolId, req.user!.userId, req.body));
  } catch (err) { next(err); }
});

eduMarketplaceRouter.get('/content/:id', authenticate, async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getContent(req.params.id as string)); } catch (err) { next(err); }
});

eduMarketplaceRouter.patch('/content/:id/publish', authenticate, authorize('GURU', 'KEPALA_KURIKULUM'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.publishContent(req.params.id as string, req.user!.userId)); } catch (err) { next(err); }
});

eduMarketplaceRouter.post('/purchase', authenticate, async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    sendCreated(res, await service.purchaseContent(req.body.contentId, req.user!.schoolId, req.user!.userId));
  } catch (err) { next(err); }
});

eduMarketplaceRouter.get('/my-purchases', authenticate, async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getMyPurchases(req.user!.schoolId, req.user!.userId)); } catch (err) { next(err); }
});

eduMarketplaceRouter.get('/my-content', authenticate, async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getMyContent(req.user!.userId)); } catch (err) { next(err); }
});
