import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import * as ctrl from './controller';

export const examsRouter = Router();
examsRouter.use(authenticate);

examsRouter.post('/generate-questions', authorize('GURU', 'KEPALA_KURIKULUM', 'ADMIN_SEKOLAH'), ctrl.generateQuestionsH);
examsRouter.get('/questions', ctrl.getQuestionsH);
examsRouter.post('/', authorize('GURU', 'KEPALA_KURIKULUM'), ctrl.createExamH);
examsRouter.get('/', ctrl.getExamsH);
examsRouter.put('/:id/publish', authorize('GURU', 'KEPALA_KURIKULUM'), ctrl.publishExamH);
examsRouter.post('/:id/submit', ctrl.submitResultH);
examsRouter.get('/:id/results', authorize('GURU', 'KEPALA_KURIKULUM', 'KEPALA_SEKOLAH', 'WALI_KELAS'), ctrl.getResultsH);
examsRouter.get('/:id/analysis', authorize('GURU', 'KEPALA_KURIKULUM', 'KEPALA_SEKOLAH'), ctrl.getAnalysisH);
