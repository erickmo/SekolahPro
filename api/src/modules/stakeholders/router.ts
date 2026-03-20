import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';

export const stakeholdersRouter = Router();
stakeholdersRouter.use(authenticate);

stakeholdersRouter.get('/partners', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    sendSuccess(res, { partners: [], message: 'Daftar mitra sekolah (belum ada data)' });
  } catch (err) { next(err); }
});

stakeholdersRouter.post('/partners', authorize('ADMIN_SEKOLAH', 'KEPALA_SEKOLAH'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    sendCreated(res, { ...req.body, id: `PARTNER-${Date.now()}`, schoolId: req.user!.schoolId });
  } catch (err) { next(err); }
});
