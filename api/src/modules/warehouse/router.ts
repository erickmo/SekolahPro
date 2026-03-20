import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess } from '../../shared/response';
import * as service from './service';

export const warehouseRouter = Router();
warehouseRouter.use(authenticate);

warehouseRouter.get(
  '/summary',
  authorize('EDS_SUPERADMIN', 'KEPALA_SEKOLAH'),
  async (req: AuthRequest, res: Response, next: NextFunction) => {
    try {
      const schoolId = req.user!.schoolId;
      const summary = await service.getDataSummary(schoolId);
      sendSuccess(res, summary);
    } catch (err) {
      next(err);
    }
  },
);

warehouseRouter.get(
  '/academic-performance',
  authorize('EDS_SUPERADMIN', 'KEPALA_SEKOLAH', 'KEPALA_KURIKULUM', 'ADMIN_SEKOLAH'),
  async (req: AuthRequest, res: Response, next: NextFunction) => {
    try {
      const schoolId = req.user!.schoolId;
      const { academicYearId, subjectId } = req.query as {
        academicYearId?: string;
        subjectId?: string;
      };
      const result = await service.getAcademicPerformance(schoolId, { academicYearId, subjectId });
      sendSuccess(res, result);
    } catch (err) {
      next(err);
    }
  },
);

warehouseRouter.get(
  '/financial-report',
  authorize('EDS_SUPERADMIN', 'KEPALA_SEKOLAH', 'BENDAHARA', 'ADMIN_SEKOLAH'),
  async (req: AuthRequest, res: Response, next: NextFunction) => {
    try {
      const schoolId = req.user!.schoolId;
      const { startDate, endDate } = req.query as {
        startDate?: string;
        endDate?: string;
      };
      const result = await service.getFinancialReport(schoolId, { startDate, endDate });
      sendSuccess(res, result);
    } catch (err) {
      next(err);
    }
  },
);

warehouseRouter.get(
  '/health-stats',
  authorize('EDS_SUPERADMIN', 'KEPALA_SEKOLAH', 'ADMIN_SEKOLAH'),
  async (req: AuthRequest, res: Response, next: NextFunction) => {
    try {
      const schoolId = req.user!.schoolId;
      const { startDate, endDate } = req.query as {
        startDate?: string;
        endDate?: string;
      };
      const result = await service.getHealthStats(schoolId, { startDate, endDate });
      sendSuccess(res, result);
    } catch (err) {
      next(err);
    }
  },
);

warehouseRouter.get(
  '/export',
  authorize('EDS_SUPERADMIN'),
  async (req: AuthRequest, res: Response, next: NextFunction) => {
    try {
      const schoolId = req.user!.schoolId;
      const { entity, format } = req.query as { entity?: string; format?: string };
      const result = await service.exportData(schoolId, entity, format || 'json');
      sendSuccess(res, result);
    } catch (err) {
      next(err);
    }
  },
);
