import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export const gamificationRouter = Router();
gamificationRouter.use(authenticate);

gamificationRouter.post(
  '/rules',
  authorize('ADMIN_SEKOLAH', 'KEPALA_KURIKULUM'),
  async (req: AuthRequest, res: Response, next: NextFunction) => {
    try {
      sendCreated(res, await service.createRule(req.user!.schoolId, req.body));
    } catch (err) { next(err); }
  }
);

gamificationRouter.get('/rules', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    sendSuccess(res, await service.listRules(req.user!.schoolId));
  } catch (err) { next(err); }
});

gamificationRouter.post(
  '/points',
  authorize('GURU', 'ADMIN_SEKOLAH'),
  async (req: AuthRequest, res: Response, next: NextFunction) => {
    try {
      sendCreated(res, await service.awardPoints(req.user!.schoolId, req.body));
    } catch (err) { next(err); }
  }
);

gamificationRouter.get('/points/:studentId', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    sendSuccess(res, await service.getStudentPoints(req.user!.schoolId, req.params.studentId as string));
  } catch (err) { next(err); }
});

gamificationRouter.get('/leaderboard', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const limit = Number(req.query.limit) || 10;
    sendSuccess(res, await service.getLeaderboard(req.user!.schoolId, limit));
  } catch (err) { next(err); }
});

gamificationRouter.post(
  '/badges',
  authorize('ADMIN_SEKOLAH'),
  async (req: AuthRequest, res: Response, next: NextFunction) => {
    try {
      sendCreated(res, await service.createBadge(req.user!.schoolId, req.body));
    } catch (err) { next(err); }
  }
);

gamificationRouter.get('/badges', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    sendSuccess(res, await service.listBadges(req.user!.schoolId));
  } catch (err) { next(err); }
});

gamificationRouter.post(
  '/badges/award',
  authorize('GURU', 'ADMIN_SEKOLAH'),
  async (req: AuthRequest, res: Response, next: NextFunction) => {
    try {
      sendCreated(res, await service.awardBadge(req.user!.schoolId, req.body));
    } catch (err) { next(err); }
  }
);

gamificationRouter.get('/badges/student/:studentId', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    sendSuccess(res, await service.getStudentBadges(req.user!.schoolId, req.params.studentId as string));
  } catch (err) { next(err); }
});
