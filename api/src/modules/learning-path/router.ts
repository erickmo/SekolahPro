import { Router } from 'express';
import { authenticate } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess } from '../../shared/response';
import * as aiTutorService from '../ai-tutor/service';

export const learningPathRouter = Router();
learningPathRouter.use(authenticate);

learningPathRouter.get('/:studentId', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    sendSuccess(res, await aiTutorService.generateLearningPath(req.user!.schoolId, req.params.studentId as string));
  } catch (err) { next(err); }
});
