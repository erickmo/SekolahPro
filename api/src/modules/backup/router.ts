import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export const backupRouter = Router();
backupRouter.use(authenticate);

backupRouter.post(
  '/create',
  authorize('EDS_SUPERADMIN', 'ADMIN_SEKOLAH'),
  async (req: AuthRequest, res: Response, next: NextFunction) => {
    try {
      const schoolId = req.user!.schoolId;
      const backup = await service.createBackup(schoolId);
      sendCreated(res, backup);
    } catch (err) {
      next(err);
    }
  },
);

backupRouter.get(
  '/list',
  authorize('EDS_SUPERADMIN', 'ADMIN_SEKOLAH'),
  async (req: AuthRequest, res: Response, next: NextFunction) => {
    try {
      const schoolId = req.user!.schoolId;
      const backups = await service.listBackups(schoolId);
      sendSuccess(res, backups);
    } catch (err) {
      next(err);
    }
  },
);

backupRouter.post(
  '/restore',
  authorize('EDS_SUPERADMIN'),
  async (req: AuthRequest, res: Response, next: NextFunction) => {
    try {
      const { backupId } = req.body as { backupId: string };
      const result = await service.restoreBackup(backupId);
      sendSuccess(res, result);
    } catch (err) {
      next(err);
    }
  },
);

backupRouter.get(
  '/status',
  authorize('EDS_SUPERADMIN', 'ADMIN_SEKOLAH'),
  async (req: AuthRequest, res: Response, next: NextFunction) => {
    try {
      const schoolId = req.user!.schoolId;
      const status = await service.getBackupStatus(schoolId);
      sendSuccess(res, status);
    } catch (err) {
      next(err);
    }
  },
);
