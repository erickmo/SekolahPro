import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import * as ctrl from './controller';

export const ewsRouter = Router();
ewsRouter.use(authenticate);

ewsRouter.get('/high-risk', authorize('GURU_BK', 'KEPALA_SEKOLAH', 'ADMIN_SEKOLAH', 'WALI_KELAS'), ctrl.highRiskH);
ewsRouter.get('/alerts', authorize('GURU_BK', 'KEPALA_SEKOLAH', 'WALI_KELAS'), ctrl.getAlertsH);
ewsRouter.post('/analyze/:studentId', authorize('GURU_BK', 'ADMIN_SEKOLAH'), ctrl.analyzeStudentH);
ewsRouter.post('/batch-analyze', authorize('ADMIN_SEKOLAH', 'KEPALA_SEKOLAH'), ctrl.batchAnalyzeH);
ewsRouter.put('/alerts/:id/resolve', authorize('GURU_BK', 'WALI_KELAS'), ctrl.resolveAlertH);
