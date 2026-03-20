import { Router } from 'express';
import { authenticate } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export const helpdeskRouter = Router();
helpdeskRouter.use(authenticate);

helpdeskRouter.post('/', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.createTicket(req.user!.schoolId, { ...req.body, reporterId: req.user!.userId })); } catch (err) { next(err); }
});
helpdeskRouter.get('/', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { tickets, meta } = await service.getTickets(req.user!.schoolId, req.query as Record<string, string>);
    sendSuccess(res, tickets, 200, meta);
  } catch (err) { next(err); }
});
helpdeskRouter.put('/:id', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.updateTicket(req.user!.schoolId, req.params.id as string, req.body)); } catch (err) { next(err); }
});
helpdeskRouter.get('/stats', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getTicketStats(req.user!.schoolId)); } catch (err) { next(err); }
});
