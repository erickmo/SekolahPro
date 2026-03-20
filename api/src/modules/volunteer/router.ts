import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export const volunteerRouter = Router();
volunteerRouter.use(authenticate);

volunteerRouter.post('/profiles', authorize('ADMIN_SEKOLAH', 'TATA_USAHA'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.createProfile(req.user!.schoolId, req.body)); } catch (err) { next(err); }
});

volunteerRouter.get('/profiles', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { profiles, meta } = await service.listProfiles(req.user!.schoolId, req.query as Record<string, string>);
    sendSuccess(res, profiles, 200, meta);
  } catch (err) { next(err); }
});

volunteerRouter.get('/profiles/:id', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getProfile(req.user!.schoolId, req.params.id as string)); } catch (err) { next(err); }
});

volunteerRouter.post('/activities', authorize('ADMIN_SEKOLAH'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.createActivity(req.user!.schoolId, req.body)); } catch (err) { next(err); }
});

volunteerRouter.get('/activities', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { activities, meta } = await service.listActivities(req.user!.schoolId, req.query as Record<string, string>);
    sendSuccess(res, activities, 200, meta);
  } catch (err) { next(err); }
});

volunteerRouter.patch('/activities/:id', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.updateActivity(req.user!.schoolId, req.params.id as string, req.body)); } catch (err) { next(err); }
});

volunteerRouter.post('/certificates', authorize('ADMIN_SEKOLAH'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.issueCertificate(req.user!.schoolId, req.body)); } catch (err) { next(err); }
});
