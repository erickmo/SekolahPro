import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export const specialNeedsRouter = Router();
specialNeedsRouter.use(authenticate);

specialNeedsRouter.post('/profiles', authorize('GPK', 'ADMIN_SEKOLAH'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.createProfile(req.user!.schoolId, req.body)); } catch (err) { next(err); }
});

specialNeedsRouter.get('/profiles', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { profiles, meta } = await service.listProfiles(req.user!.schoolId, req.query as Record<string, string>);
    sendSuccess(res, profiles, 200, meta);
  } catch (err) { next(err); }
});

specialNeedsRouter.get('/profiles/:studentId', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getProfile(req.user!.schoolId, req.params.studentId as string)); } catch (err) { next(err); }
});

specialNeedsRouter.post('/iep', authorize('GPK'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.createIEP(req.user!.schoolId, req.user!.userId, req.body)); } catch (err) { next(err); }
});

specialNeedsRouter.get('/iep/:studentId', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getStudentIEPs(req.user!.schoolId, req.params.studentId as string)); } catch (err) { next(err); }
});

specialNeedsRouter.post('/progress', authorize('GPK'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.addProgressReport(req.user!.schoolId, req.user!.userId, req.body)); } catch (err) { next(err); }
});

specialNeedsRouter.get('/progress/:studentId', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getProgressReports(req.user!.schoolId, req.params.studentId as string)); } catch (err) { next(err); }
});
