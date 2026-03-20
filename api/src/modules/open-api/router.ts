import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated, sendNoContent } from '../../shared/response';
import * as service from './service';

export const openApiRouter = Router();
openApiRouter.use(authenticate);

openApiRouter.post('/keys', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.createApiKey(req.user!.schoolId, { ...req.body, userId: req.user!.userId })); } catch (err) { next(err); }
});
openApiRouter.get('/keys', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getApiKeys(req.user!.schoolId)); } catch (err) { next(err); }
});
openApiRouter.delete('/keys/:id', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { await service.revokeApiKey(req.user!.schoolId, req.params.id as string); sendNoContent(res); } catch (err) { next(err); }
});
openApiRouter.get('/audit-logs', authorize('ADMIN_SEKOLAH', 'KEPALA_SEKOLAH', 'SUPERADMIN'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { logs, meta } = await service.getAuditLogs(req.user!.schoolId, req.query as Record<string, string>);
    sendSuccess(res, logs, 200, meta);
  } catch (err) { next(err); }
});
