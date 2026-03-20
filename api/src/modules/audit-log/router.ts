import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess } from '../../shared/response';
import * as service from './service';

export { logAudit } from './service';

export const auditLogRouter = Router();
auditLogRouter.use(authenticate);

auditLogRouter.get('/', authorize('EDS_SUPERADMIN', 'ADMIN_SEKOLAH', 'KEPALA_SEKOLAH'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const isEdsAdmin = req.user!.role === ('EDS_SUPERADMIN' as never);
    const schoolId = isEdsAdmin ? (req.query.schoolId as string | undefined) : req.user!.schoolId;
    const { logs, meta } = await service.listAuditLogs(schoolId, req.query as Record<string, string>);
    sendSuccess(res, logs, 200, meta);
  } catch (err) { next(err); }
});

auditLogRouter.get('/:id', authorize('EDS_SUPERADMIN', 'ADMIN_SEKOLAH'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getAuditLog(req.params.id as string)); } catch (err) { next(err); }
});
