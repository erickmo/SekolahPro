import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import * as ctrl from './controller';

export const notificationsRouter = Router();
notificationsRouter.use(authenticate);

notificationsRouter.post('/send', authorize('ADMIN_SEKOLAH', 'TATA_USAHA', 'GURU'), ctrl.sendNotifH);
notificationsRouter.post('/bulk', authorize('ADMIN_SEKOLAH', 'KEPALA_SEKOLAH'), ctrl.sendBulkH);
notificationsRouter.get('/', ctrl.getNotifH);
notificationsRouter.put('/:id/read', ctrl.markReadH);
notificationsRouter.get('/templates', ctrl.getTemplatesH);
notificationsRouter.post('/templates', authorize('ADMIN_SEKOLAH'), ctrl.createTemplateH);
