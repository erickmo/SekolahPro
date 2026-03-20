import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export const committeeRouter = Router();
committeeRouter.use(authenticate);

committeeRouter.post('/boards', authorize('KEPALA_SEKOLAH', 'ADMIN_SEKOLAH'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.createBoard(req.user!.schoolId, req.body)); } catch (err) { next(err); }
});

committeeRouter.get('/boards', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.listBoards(req.user!.schoolId)); } catch (err) { next(err); }
});

committeeRouter.get('/boards/:id', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getBoard(req.user!.schoolId, req.params.id as string)); } catch (err) { next(err); }
});

committeeRouter.post('/boards/:id/members', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.addMember(req.user!.schoolId, req.params.id as string, req.body)); } catch (err) { next(err); }
});

committeeRouter.post('/meetings', authorize('KOMITE_SEKOLAH', 'KEPALA_SEKOLAH'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.createMeeting(req.user!.schoolId, req.body)); } catch (err) { next(err); }
});

committeeRouter.get('/meetings', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { meetings, meta } = await service.listMeetings(req.user!.schoolId, req.query as Record<string, string>);
    sendSuccess(res, meetings, 200, meta);
  } catch (err) { next(err); }
});

committeeRouter.post('/meetings/:id/decisions', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.addDecision(req.user!.schoolId, req.params.id as string, req.body)); } catch (err) { next(err); }
});

committeeRouter.get('/meetings/:id/decisions', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.listDecisions(req.user!.schoolId, req.params.id as string)); } catch (err) { next(err); }
});
