import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export async function sendNotifH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendCreated(res, await service.sendNotification({ ...req.body, schoolId: req.user!.schoolId })); } catch (err) { next(err); }
}
export async function sendBulkH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendCreated(res, await service.sendBulk(req.user!.schoolId, req.body)); } catch (err) { next(err); }
}
export async function getNotifH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try {
    const { notifications, meta } = await service.getNotifications(req.user!.schoolId, req.user!.userId, req.query as Record<string, string>);
    sendSuccess(res, notifications, 200, meta);
  } catch (err) { next(err); }
}
export async function markReadH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendSuccess(res, await service.markAsRead(req.user!.schoolId, req.params.id as string, req.user!.userId)); } catch (err) { next(err); }
}
export async function getTemplatesH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendSuccess(res, await service.getTemplates(req.user!.schoolId)); } catch (err) { next(err); }
}
export async function createTemplateH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendCreated(res, await service.createTemplate(req.user!.schoolId, req.body)); } catch (err) { next(err); }
}
