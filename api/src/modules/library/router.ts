import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export const libraryRouter = Router();
libraryRouter.use(authenticate);

libraryRouter.get('/books', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { books, meta } = await service.getBooks(req.user!.schoolId, req.query as Record<string, string>);
    sendSuccess(res, books, 200, meta);
  } catch (err) { next(err); }
});
libraryRouter.post('/books', authorize('PUSTAKAWAN', 'ADMIN_SEKOLAH'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.addBook(req.user!.schoolId, req.body)); } catch (err) { next(err); }
});
libraryRouter.post('/loans', authorize('PUSTAKAWAN'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.borrowBook(req.user!.schoolId, req.body)); } catch (err) { next(err); }
});
libraryRouter.put('/loans/:id/return', authorize('PUSTAKAWAN'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.returnBook(req.user!.schoolId, req.params.id as string)); } catch (err) { next(err); }
});
libraryRouter.get('/loans', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { loans, meta } = await service.getLoans(req.user!.schoolId, req.query as Record<string, string>);
    sendSuccess(res, loans, 200, meta);
  } catch (err) { next(err); }
});
