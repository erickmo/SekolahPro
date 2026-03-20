import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import * as ctrl from './controller';

export const academicRouter = Router();
academicRouter.use(authenticate);

academicRouter.get('/academic-years', ctrl.getAcademicYearsH);
academicRouter.post('/academic-years', authorize('ADMIN_SEKOLAH', 'KEPALA_SEKOLAH'), ctrl.createAcademicYearH);
academicRouter.get('/subjects', ctrl.getSubjectsH);
academicRouter.post('/subjects', authorize('ADMIN_SEKOLAH', 'KEPALA_KURIKULUM'), ctrl.createSubjectH);
academicRouter.get('/classes', ctrl.getClassesH);
academicRouter.post('/classes', authorize('ADMIN_SEKOLAH', 'TATA_USAHA'), ctrl.createClassH);
academicRouter.post('/enroll', authorize('ADMIN_SEKOLAH', 'OPERATOR_SIMS'), ctrl.enrollStudentH);
academicRouter.post('/grades', authorize('GURU', 'WALI_KELAS'), ctrl.inputGradeH);
academicRouter.get('/grades/:studentId', ctrl.getStudentGradesH);
academicRouter.post('/report-cards/:studentId', authorize('WALI_KELAS', 'ADMIN_SEKOLAH', 'KEPALA_KURIKULUM'), ctrl.generateReportCardH);
academicRouter.get('/schedules', ctrl.getSchedulesH);
academicRouter.post('/schedules', authorize('ADMIN_SEKOLAH', 'KEPALA_KURIKULUM', 'TATA_USAHA'), ctrl.createScheduleH);
academicRouter.get('/teachers', ctrl.getTeachersH);
academicRouter.post('/teachers', authorize('ADMIN_SEKOLAH', 'TATA_USAHA'), ctrl.createTeacherH);
