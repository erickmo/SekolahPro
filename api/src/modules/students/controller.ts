import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated, sendNoContent } from '../../shared/response';
import * as service from './service';

export async function createStudentHandler(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendCreated(res, await service.createStudent(req.user!.schoolId, req.body)); } catch (err) { next(err); }
}

export async function getStudentsHandler(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try {
    const { students, meta } = await service.getStudents(req.user!.schoolId, req.query as Record<string, string>);
    sendSuccess(res, students, 200, meta);
  } catch (err) { next(err); }
}

export async function getStudentHandler(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendSuccess(res, await service.getStudent(req.user!.schoolId, req.params.id as string)); } catch (err) { next(err); }
}

export async function updateStudentHandler(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendSuccess(res, await service.updateStudent(req.user!.schoolId, req.params.id as string, req.body)); } catch (err) { next(err); }
}

export async function deleteStudentHandler(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { await service.deleteStudent(req.user!.schoolId, req.params.id as string); sendNoContent(res); } catch (err) { next(err); }
}

export async function importStudentsHandler(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendCreated(res, await service.importStudents(req.user!.schoolId, req.body.students)); } catch (err) { next(err); }
}

export async function getAttendanceHandler(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendSuccess(res, await service.getStudentAttendance(req.user!.schoolId, req.params.id as string, req.query as { startDate?: string; endDate?: string })); } catch (err) { next(err); }
}

export async function recordAttendanceHandler(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendCreated(res, await service.recordAttendance(req.user!.schoolId, req.body.records)); } catch (err) { next(err); }
}
