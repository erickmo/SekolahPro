import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { validate } from '../../middleware/validate';
import { CreateSchoolDto, UpdateSchoolDto } from './dto';
import * as ctrl from './controller';

export const schoolsRouter = Router();

schoolsRouter.use(authenticate);
schoolsRouter.get('/', authorize('SUPERADMIN'), ctrl.getSchoolsHandler);
schoolsRouter.post('/', authorize('SUPERADMIN'), validate(CreateSchoolDto), ctrl.createSchoolHandler);
schoolsRouter.get('/:id', ctrl.getSchoolHandler);
schoolsRouter.put('/:id', authorize('SUPERADMIN', 'ADMIN_SEKOLAH'), validate(UpdateSchoolDto), ctrl.updateSchoolHandler);
schoolsRouter.delete('/:id', authorize('SUPERADMIN'), ctrl.deleteSchoolHandler);
