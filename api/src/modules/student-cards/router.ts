import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import * as ctrl from './controller';

export const studentCardsRouter = Router();

studentCardsRouter.use(authenticate);
studentCardsRouter.post('/verify', ctrl.verifyQrHandler);
studentCardsRouter.get('/:studentId', ctrl.getCardHandler);
studentCardsRouter.post('/:studentId/generate', authorize('ADMIN_SEKOLAH', 'OPERATOR_SIMS'), ctrl.generateCardHandler);
studentCardsRouter.put('/:studentId/deactivate', authorize('ADMIN_SEKOLAH', 'OPERATOR_SIMS'), ctrl.deactivateCardHandler);
