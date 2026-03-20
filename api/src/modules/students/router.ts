import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { validate } from '../../middleware/validate';
import { CreateStudentDto, UpdateStudentDto, ImportStudentsDto } from './dto';
import * as ctrl from './controller';

export const studentsRouter = Router();

studentsRouter.use(authenticate);
studentsRouter.get('/', ctrl.getStudentsHandler);
studentsRouter.post('/', authorize('ADMIN_SEKOLAH', 'OPERATOR_SIMS', 'TATA_USAHA'), validate(CreateStudentDto), ctrl.createStudentHandler);
studentsRouter.post('/import', authorize('ADMIN_SEKOLAH', 'OPERATOR_SIMS'), validate(ImportStudentsDto), ctrl.importStudentsHandler);
studentsRouter.post('/attendance', authorize('GURU', 'WALI_KELAS', 'OPERATOR_SIMS'), ctrl.recordAttendanceHandler);
studentsRouter.get('/:id', ctrl.getStudentHandler);
studentsRouter.put('/:id', authorize('ADMIN_SEKOLAH', 'OPERATOR_SIMS'), validate(UpdateStudentDto), ctrl.updateStudentHandler);
studentsRouter.delete('/:id', authorize('ADMIN_SEKOLAH', 'OPERATOR_SIMS'), ctrl.deleteStudentHandler);
studentsRouter.get('/:id/attendance', ctrl.getAttendanceHandler);
