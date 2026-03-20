import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess } from '../../shared/response';
import * as service from './service';

export const analyticsRouter = Router();
analyticsRouter.use(authenticate);

analyticsRouter.get(
  '/risk-alerts',
  authorize('ADMIN_SEKOLAH', 'KEPALA_SEKOLAH', 'WALI_KELAS'),
  async (req: AuthRequest, res: Response, next: NextFunction) => {
    try {
      const schoolId = req.user!.schoolId;
      const { severity, resolved, page, limit } = req.query as {
        severity?: string;
        resolved?: string;
        page?: string;
        limit?: string;
      };
      const result = await service.listRiskAlerts(schoolId, {
        severity,
        resolved: resolved !== undefined ? resolved === 'true' : undefined,
        page: page ? Number(page) : undefined,
        limit: limit ? Number(limit) : undefined,
      });
      sendSuccess(res, result.alerts, 200, result.meta);
    } catch (err) {
      next(err);
    }
  },
);

analyticsRouter.get(
  '/risk-alerts/:studentId',
  authorize('ADMIN_SEKOLAH', 'KEPALA_SEKOLAH', 'WALI_KELAS'),
  async (req: AuthRequest, res: Response, next: NextFunction) => {
    try {
      const schoolId = req.user!.schoolId;
      const studentId = req.params.studentId as string;
      const alerts = await service.getStudentRiskAlerts(schoolId, studentId);
      sendSuccess(res, alerts);
    } catch (err) {
      next(err);
    }
  },
);

analyticsRouter.patch(
  '/risk-alerts/:id/resolve',
  authorize('GURU', 'WALI_KELAS'),
  async (req: AuthRequest, res: Response, next: NextFunction) => {
    try {
      const id = req.params.id as string;
      const resolvedBy = req.user!.userId;
      const alert = await service.resolveAlert(id, resolvedBy);
      sendSuccess(res, alert);
    } catch (err) {
      next(err);
    }
  },
);

analyticsRouter.get(
  '/overview',
  authorize('KEPALA_SEKOLAH', 'ADMIN_SEKOLAH'),
  async (req: AuthRequest, res: Response, next: NextFunction) => {
    try {
      const schoolId = req.user!.schoolId;
      const overview = await service.getSchoolOverview(schoolId);
      sendSuccess(res, overview);
    } catch (err) {
      next(err);
    }
  },
);

analyticsRouter.get(
  '/attendance-trend',
  authorize('ADMIN_SEKOLAH', 'KEPALA_SEKOLAH', 'WALI_KELAS'),
  async (req: AuthRequest, res: Response, next: NextFunction) => {
    try {
      const schoolId = req.user!.schoolId;
      const { startDate, endDate, classId } = req.query as {
        startDate?: string;
        endDate?: string;
        classId?: string;
      };
      const trend = await service.getAttendanceTrend(schoolId, { startDate, endDate, classId });
      sendSuccess(res, trend);
    } catch (err) {
      next(err);
    }
  },
);
