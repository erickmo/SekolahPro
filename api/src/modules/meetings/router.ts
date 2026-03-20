import { Router } from 'express';
import { authenticate } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export const meetingsRouter = Router();
meetingsRouter.use(authenticate);

meetingsRouter.post('/', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.createMeeting(req.user!.schoolId, req.user!.userId, req.body)); } catch (err) { next(err); }
});
meetingsRouter.get('/', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { meetings, meta } = await service.getMeetings(req.user!.schoolId, req.query as Record<string, string>);
    sendSuccess(res, meetings, 200, meta);
  } catch (err) { next(err); }
});
meetingsRouter.put('/:id', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.updateMeeting(req.user!.schoolId, req.params.id as string, req.body)); } catch (err) { next(err); }
});
