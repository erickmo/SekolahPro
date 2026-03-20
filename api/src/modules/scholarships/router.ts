import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export const scholarshipsRouter = Router();
scholarshipsRouter.use(authenticate);

scholarshipsRouter.get('/programs', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getPrograms(req.user!.schoolId)); } catch (err) { next(err); }
});
scholarshipsRouter.post('/programs', authorize('BENDAHARA', 'ADMIN_SEKOLAH', 'KEPALA_SEKOLAH'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.createProgram(req.user!.schoolId, req.body)); } catch (err) { next(err); }
});
scholarshipsRouter.post('/apply', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.applyScholarship(req.user!.schoolId, req.body)); } catch (err) { next(err); }
});
scholarshipsRouter.get('/applications', authorize('BENDAHARA', 'ADMIN_SEKOLAH', 'KEPALA_SEKOLAH'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { applications, meta } = await service.getApplications(req.user!.schoolId, req.query as Record<string, string>);
    sendSuccess(res, applications, 200, meta);
  } catch (err) { next(err); }
});
scholarshipsRouter.put('/applications/:id', authorize('BENDAHARA', 'ADMIN_SEKOLAH'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.updateApplicationStatus(req.user!.schoolId, req.params.id as string, req.body.status, req.body.reviewNotes)); } catch (err) { next(err); }
});
