import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export const alumniRouter = Router();
alumniRouter.use(authenticate);

alumniRouter.post('/profile', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.createAlumniProfile(req.user!.schoolId, { ...req.body, userId: req.user!.userId })); } catch (err) { next(err); }
});
alumniRouter.get('/', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { alumni, meta } = await service.getAlumni(req.user!.schoolId, req.query as Record<string, string>);
    sendSuccess(res, alumni, 200, meta);
  } catch (err) { next(err); }
});
alumniRouter.post('/portfolio', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.createPortfolio(req.user!.schoolId, req.body.studentId, req.body)); } catch (err) { next(err); }
});
alumniRouter.post('/portfolio/:portfolioId/items', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.addPortfolioItem(req.params.portfolioId as string, req.body)); } catch (err) { next(err); }
});
alumniRouter.get('/portfolio/:studentId', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getPortfolio(req.user!.schoolId, req.params.studentId as string)); } catch (err) { next(err); }
});
alumniRouter.post('/tracer-study', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.createTracerStudy(req.user!.schoolId, req.user!.userId, req.body)); } catch (err) { next(err); }
});
