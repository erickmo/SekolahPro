import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

const sid = (req: AuthRequest) => req.user!.schoolId;

export async function createAcademicYearH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendCreated(res, await service.createAcademicYear(sid(req), req.body)); } catch (err) { next(err); }
}
export async function getAcademicYearsH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendSuccess(res, await service.getAcademicYears(sid(req))); } catch (err) { next(err); }
}
export async function createSubjectH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendCreated(res, await service.createSubject(sid(req), req.body)); } catch (err) { next(err); }
}
export async function getSubjectsH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendSuccess(res, await service.getSubjects(sid(req))); } catch (err) { next(err); }
}
export async function createClassH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendCreated(res, await service.createClass(sid(req), req.body)); } catch (err) { next(err); }
}
export async function getClassesH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendSuccess(res, await service.getClasses(sid(req), req.query.academicYearId as string)); } catch (err) { next(err); }
}
export async function enrollStudentH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendCreated(res, await service.enrollStudent(sid(req), req.body)); } catch (err) { next(err); }
}
export async function inputGradeH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendCreated(res, await service.inputGrade(sid(req), { ...req.body, teacherId: req.user!.userId })); } catch (err) { next(err); }
}
export async function getStudentGradesH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendSuccess(res, await service.getStudentGrades(sid(req), req.params.studentId as string, req.query.semesterId as string)); } catch (err) { next(err); }
}
export async function generateReportCardH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendCreated(res, await service.generateReportCard(sid(req), req.params.studentId as string, req.body.semesterId)); } catch (err) { next(err); }
}
export async function createScheduleH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendCreated(res, await service.createSchedule(sid(req), req.body)); } catch (err) { next(err); }
}
export async function getSchedulesH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendSuccess(res, await service.getSchedules(sid(req), { classId: req.query.classId as string, teacherId: req.query.teacherId as string, academicYearId: req.query.academicYearId as string })); } catch (err) { next(err); }
}
export async function getTeachersH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try {
    const { teachers, meta } = await service.getTeachers(sid(req), req.query as Record<string, string>);
    sendSuccess(res, teachers, 200, meta);
  } catch (err) { next(err); }
}
export async function createTeacherH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendCreated(res, await service.createTeacher(sid(req), req.body)); } catch (err) { next(err); }
}
