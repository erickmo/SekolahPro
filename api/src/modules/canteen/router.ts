import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export const canteenRouter = Router();
canteenRouter.use(authenticate);

canteenRouter.get('/menu', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getMenuItems(req.user!.schoolId, req.query.date as string)); } catch (err) { next(err); }
});
canteenRouter.post('/menu', authorize('ADMIN_SEKOLAH', 'KASIR_KOPERASI'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.addMenuItem(req.user!.schoolId, req.body)); } catch (err) { next(err); }
});
canteenRouter.post('/orders', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.placeOrder(req.user!.schoolId, req.body)); } catch (err) { next(err); }
});
canteenRouter.get('/orders', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { orders, meta } = await service.getOrders(req.user!.schoolId, req.query as Record<string, string>);
    sendSuccess(res, orders, 200, meta);
  } catch (err) { next(err); }
});
canteenRouter.post('/wallet/topup', authorize('KASIR_KOPERASI', 'BENDAHARA', 'ADMIN_SEKOLAH'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.topUpWallet(req.user!.schoolId, req.body.ownerId, req.body.amount, req.user!.userId)); } catch (err) { next(err); }
});
canteenRouter.get('/wallet/:ownerId', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getWallet(req.user!.schoolId, req.params.ownerId as string)); } catch (err) { next(err); }
});
