import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export async function chatH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendSuccess(res, await service.chat(req.user!.schoolId, req.user!.userId, req.body)); } catch (err) { next(err); }
}
export async function plagiarismH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendCreated(res, await service.checkPlagiarism(req.user!.schoolId, req.body)); } catch (err) { next(err); }
}
export async function learningPathH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendSuccess(res, await service.generateLearningPath(req.user!.schoolId, req.params.studentId as string)); } catch (err) { next(err); }
}
export async function historyH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try {
    const { interactions, meta } = await service.getInteractionHistory(req.user!.schoolId, req.user!.userId, req.query as Record<string, string>);
    sendSuccess(res, interactions, 200, meta);
  } catch (err) { next(err); }
}
