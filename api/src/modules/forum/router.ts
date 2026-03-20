import { Router } from 'express';
import { authenticate } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated, sendNoContent } from '../../shared/response';
import * as service from './service';

export const forumRouter = Router();
forumRouter.use(authenticate);

forumRouter.get('/', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { posts, meta } = await service.getPosts(req.user!.schoolId, req.query as Record<string, string>);
    sendSuccess(res, posts, 200, meta);
  } catch (err) { next(err); }
});
forumRouter.post('/', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.createPost(req.user!.schoolId, req.user!.userId, req.body)); } catch (err) { next(err); }
});
forumRouter.get('/:id', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getPost(req.user!.schoolId, req.params.id as string)); } catch (err) { next(err); }
});
forumRouter.post('/:id/comments', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.addComment(req.user!.schoolId, req.params.id as string, req.user!.userId, req.body.content)); } catch (err) { next(err); }
});
forumRouter.delete('/:id', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { await service.deletePost(req.user!.schoolId, req.params.id as string, req.user!.userId); sendNoContent(res); } catch (err) { next(err); }
});
