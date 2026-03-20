import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export const lmsRouter = Router();
lmsRouter.use(authenticate);

lmsRouter.post(
  '/connect',
  authorize('ADMIN_SEKOLAH'),
  async (req: AuthRequest, res: Response, next: NextFunction) => {
    try {
      const schoolId = req.user!.schoolId;
      const userId = req.user!.userId;
      const { provider, apiUrl, apiKey } = req.body as {
        provider: 'MOODLE' | 'GOOGLE_CLASSROOM' | 'CANVAS';
        apiUrl: string;
        apiKey: string;
      };
      const result = await service.configureLMS(schoolId, userId, { provider, apiUrl, apiKey });
      sendCreated(res, result);
    } catch (err) {
      next(err);
    }
  },
);

lmsRouter.get(
  '/status',
  async (req: AuthRequest, res: Response, next: NextFunction) => {
    try {
      const schoolId = req.user!.schoolId;
      const status = await service.getLMSStatus(schoolId);
      sendSuccess(res, status);
    } catch (err) {
      next(err);
    }
  },
);

lmsRouter.post(
  '/sync-courses',
  authorize('ADMIN_SEKOLAH', 'KEPALA_KURIKULUM'),
  async (req: AuthRequest, res: Response, next: NextFunction) => {
    try {
      const schoolId = req.user!.schoolId;
      const result = await service.syncCourses(schoolId);
      sendSuccess(res, result);
    } catch (err) {
      next(err);
    }
  },
);

lmsRouter.post(
  '/sync-grades',
  authorize('GURU'),
  async (req: AuthRequest, res: Response, next: NextFunction) => {
    try {
      const schoolId = req.user!.schoolId;
      const result = await service.syncGrades(schoolId);
      sendSuccess(res, result);
    } catch (err) {
      next(err);
    }
  },
);

lmsRouter.get(
  '/courses',
  async (req: AuthRequest, res: Response, next: NextFunction) => {
    try {
      const schoolId = req.user!.schoolId;
      const courses = await service.listCourses(schoolId);
      sendSuccess(res, courses);
    } catch (err) {
      next(err);
    }
  },
);
