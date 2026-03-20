import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess } from '../../shared/response';
import * as service from './service';

export const securityRouter = Router();
securityRouter.use(authenticate);

securityRouter.get('/dashboard', authorize('EDS_SUPERADMIN', 'ADMIN_SEKOLAH'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const isEdsAdmin = req.user!.role === ('EDS_SUPERADMIN' as never);
    const schoolId = isEdsAdmin ? undefined : req.user!.schoolId;
    sendSuccess(res, await service.getSecurityDashboard(schoolId));
  } catch (err) { next(err); }
});

securityRouter.get('/suspicious', authorize('EDS_SUPERADMIN', 'ADMIN_SEKOLAH'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const isEdsAdmin = req.user!.role === ('EDS_SUPERADMIN' as never);
    const schoolId = isEdsAdmin ? undefined : req.user!.schoolId;
    const { logs, meta } = await service.getSuspiciousActivity(schoolId, req.query as Record<string, string>);
    sendSuccess(res, logs, 200, meta);
  } catch (err) { next(err); }
});

securityRouter.post('/mfa/enable', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.enableMFA(req.user!.userId)); } catch (err) { next(err); }
});

securityRouter.post('/mfa/disable', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.disableMFA(req.user!.userId)); } catch (err) { next(err); }
});

securityRouter.get('/active-sessions', authorize('EDS_SUPERADMIN'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { users, meta } = await service.getActiveSessions(req.query.schoolId as string | undefined, req.query as Record<string, string>);
    sendSuccess(res, users, 200, meta);
  } catch (err) { next(err); }
});
