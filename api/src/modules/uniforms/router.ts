import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export const uniformsRouter = Router();
uniformsRouter.use(authenticate);

// POST /api/v1/uniforms/items
uniformsRouter.post(
  '/items',
  authorize('TATA_USAHA', 'ADMIN_SEKOLAH'),
  async (req: AuthRequest, res: Response, next: NextFunction) => {
    try {
      sendCreated(res, await service.createItem(req.user!.schoolId, req.body));
    } catch (err) { next(err); }
  }
);

// GET /api/v1/uniforms/items
uniformsRouter.get(
  '/items',
  async (req: AuthRequest, res: Response, next: NextFunction) => {
    try {
      const { items, meta } = await service.listItems(req.user!.schoolId, req.query as Record<string, string>);
      sendSuccess(res, items, 200, meta);
    } catch (err) { next(err); }
  }
);

// PATCH /api/v1/uniforms/items/:id/stock
uniformsRouter.patch(
  '/items/:id/stock',
  authorize('TATA_USAHA'),
  async (req: AuthRequest, res: Response, next: NextFunction) => {
    try {
      sendSuccess(res, await service.updateStock(req.user!.schoolId, req.params.id as string, req.body.stock));
    } catch (err) { next(err); }
  }
);

// POST /api/v1/uniforms/orders
uniformsRouter.post(
  '/orders',
  async (req: AuthRequest, res: Response, next: NextFunction) => {
    try {
      sendCreated(res, await service.createOrder(req.user!.schoolId, req.body));
    } catch (err) { next(err); }
  }
);

// GET /api/v1/uniforms/orders
uniformsRouter.get(
  '/orders',
  async (req: AuthRequest, res: Response, next: NextFunction) => {
    try {
      const { orders, meta } = await service.listOrders(req.user!.schoolId, req.query as Record<string, string>);
      sendSuccess(res, orders, 200, meta);
    } catch (err) { next(err); }
  }
);

// GET /api/v1/uniforms/orders/:id
uniformsRouter.get(
  '/orders/:id',
  async (req: AuthRequest, res: Response, next: NextFunction) => {
    try {
      sendSuccess(res, await service.getOrder(req.user!.schoolId, req.params.id as string));
    } catch (err) { next(err); }
  }
);

// PATCH /api/v1/uniforms/orders/:id/status
uniformsRouter.patch(
  '/orders/:id/status',
  authorize('TATA_USAHA'),
  async (req: AuthRequest, res: Response, next: NextFunction) => {
    try {
      sendSuccess(res, await service.updateOrderStatus(req.user!.schoolId, req.params.id as string, req.body.status));
    } catch (err) { next(err); }
  }
);
