import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import * as ctrl from './controller';

export const aiTutorRouter = Router();
aiTutorRouter.use(authenticate);

aiTutorRouter.post('/chat', ctrl.chatH);
aiTutorRouter.post('/plagiarism-check', authorize('GURU', 'KEPALA_KURIKULUM'), ctrl.plagiarismH);
aiTutorRouter.get('/learning-path/:studentId', ctrl.learningPathH);
aiTutorRouter.get('/history', ctrl.historyH);
