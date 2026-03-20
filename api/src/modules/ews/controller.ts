import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export async function analyzeStudentH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendSuccess(res, await service.analyzeStudent(req.user!.schoolId, req.params.studentId as string)); } catch (err) { next(err); }
}
export async function getAlertsH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try {
    const { alerts, meta } = await service.getAlerts(req.user!.schoolId, req.query as Record<string, string>);
    sendSuccess(res, alerts, 200, meta);
  } catch (err) { next(err); }
}
export async function resolveAlertH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendSuccess(res, await service.resolveAlert(req.params.id as string, req.user!.userId, req.body.note)); } catch (err) { next(err); }
}
export async function highRiskH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendSuccess(res, await service.getHighRiskStudents(req.user!.schoolId)); } catch (err) { next(err); }
}
export async function batchAnalyzeH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendCreated(res, await service.batchAnalyze(req.user!.schoolId)); } catch (err) { next(err); }
}
