import { Router } from 'express';
import { authenticate, authorize, optionalAuth } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { Request } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export const websiteRouter = Router();

// Public
websiteRouter.get('/articles', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const { articles, meta } = await service.getArticles(req.query.schoolId as string, req.query as Record<string, string>);
    sendSuccess(res, articles, 200, meta);
  } catch (err) { next(err); }
});
websiteRouter.get('/articles/:slug', async (req: Request, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getArticle(req.query.schoolId as string, req.params.slug as string)); } catch (err) { next(err); }
});
websiteRouter.get('/announcements', async (req: Request, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getAnnouncements(req.query.schoolId as string)); } catch (err) { next(err); }
});
websiteRouter.get('/gallery', async (req: Request, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getGallery(req.query.schoolId as string)); } catch (err) { next(err); }
});

// Protected
websiteRouter.use(authenticate);
websiteRouter.post('/articles', authorize('ADMIN_SEKOLAH', 'TATA_USAHA', 'GURU'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.createArticle(req.user!.schoolId, req.user!.userId, req.body)); } catch (err) { next(err); }
});
websiteRouter.put('/articles/:id', authorize('ADMIN_SEKOLAH', 'TATA_USAHA'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.updateArticle(req.user!.schoolId, req.params.id as string, req.body)); } catch (err) { next(err); }
});
websiteRouter.post('/announcements', authorize('ADMIN_SEKOLAH', 'KEPALA_SEKOLAH', 'TATA_USAHA'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.createAnnouncement(req.user!.schoolId, { ...req.body, createdBy: req.user!.userId })); } catch (err) { next(err); }
});
websiteRouter.post('/gallery', authorize('ADMIN_SEKOLAH', 'TATA_USAHA'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.addGalleryItem(req.user!.schoolId, req.body)); } catch (err) { next(err); }
});
