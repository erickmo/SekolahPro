import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { Request } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import { prisma } from '../../lib/prisma';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';

export const schoolBlogRouter = Router();

// Public
schoolBlogRouter.get('/public', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const { page, limit, skip } = getPaginationParams({ page: Number(req.query.page), limit: Number(req.query.limit) });
    const where = { schoolId: req.query.schoolId as string, isPublished: true, status: 'PUBLISHED' as const };
    const [posts, total] = await Promise.all([
      prisma.blogPost.findMany({ where, skip, take: limit, orderBy: { publishedAt: 'desc' }, include: { author: { select: { name: true } } } }),
      prisma.blogPost.count({ where }),
    ]);
    sendSuccess(res, posts, 200, buildPaginationMeta(total, page, limit));
  } catch (err) { next(err); }
});

// Protected
schoolBlogRouter.use(authenticate);
schoolBlogRouter.post('/', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const post = await prisma.blogPost.create({
      data: { ...req.body, schoolId: req.user!.schoolId, authorId: req.user!.userId, status: 'DRAFT', isPublished: false },
    });
    sendCreated(res, post);
  } catch (err) { next(err); }
});
schoolBlogRouter.get('/', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { page, limit, skip } = getPaginationParams({ page: Number(req.query.page), limit: Number(req.query.limit) });
    const where = { schoolId: req.user!.schoolId };
    const [posts, total] = await Promise.all([
      prisma.blogPost.findMany({ where, skip, take: limit, orderBy: { createdAt: 'desc' }, include: { author: { select: { name: true } } } }),
      prisma.blogPost.count({ where }),
    ]);
    sendSuccess(res, posts, 200, buildPaginationMeta(total, page, limit));
  } catch (err) { next(err); }
});
schoolBlogRouter.put('/:id/publish', authorize('GURU', 'ADMIN_SEKOLAH', 'TATA_USAHA'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const post = await prisma.blogPost.update({ where: { id: req.params.id as string }, data: { status: 'PUBLISHED' as never, isPublished: true, publishedAt: new Date() } });
    sendSuccess(res, post);
  } catch (err) { next(err); }
});
