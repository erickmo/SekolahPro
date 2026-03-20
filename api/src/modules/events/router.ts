import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export const eventsRouter = Router();
eventsRouter.use(authenticate);

eventsRouter.post('/', authorize('TATA_USAHA', 'ADMIN_SEKOLAH', 'KEPALA_SEKOLAH'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.createEvent(req.user!.schoolId, req.body)); } catch (err) { next(err); }
});

eventsRouter.get('/', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { events, meta } = await service.listEvents(req.user!.schoolId, req.query as Record<string, string>);
    sendSuccess(res, events, 200, meta);
  } catch (err) { next(err); }
});

eventsRouter.get('/:id', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getEvent(req.user!.schoolId, req.params.id as string)); } catch (err) { next(err); }
});

eventsRouter.patch('/:id', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.updateEvent(req.user!.schoolId, req.params.id as string, req.body)); } catch (err) { next(err); }
});

eventsRouter.post('/:id/tasks', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.addTask(req.user!.schoolId, req.params.id as string, req.body)); } catch (err) { next(err); }
});

eventsRouter.patch('/:id/tasks/:taskId', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.updateTask(req.user!.schoolId, req.params.id as string, req.params.taskId as string, req.body)); } catch (err) { next(err); }
});

eventsRouter.post('/:id/budget', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.addBudgetItem(req.user!.schoolId, req.params.id as string, req.body)); } catch (err) { next(err); }
});

eventsRouter.post('/:id/committee', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.addCommitteeMember(req.user!.schoolId, req.params.id as string, req.body)); } catch (err) { next(err); }
});
