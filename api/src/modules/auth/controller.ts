import { Request, Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export async function loginHandler(req: Request, res: Response, next: NextFunction): Promise<void> {
  try {
    const result = await service.login(req.body);
    sendSuccess(res, result);
  } catch (err) { next(err); }
}

export async function registerHandler(req: Request, res: Response, next: NextFunction): Promise<void> {
  try {
    const result = await service.register(req.body);
    sendCreated(res, result);
  } catch (err) { next(err); }
}

export async function refreshHandler(req: Request, res: Response, next: NextFunction): Promise<void> {
  try {
    const result = await service.refreshToken(req.body.refreshToken);
    sendSuccess(res, result);
  } catch (err) { next(err); }
}

export async function logoutHandler(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try {
    await service.logout(req.user!.userId);
    sendSuccess(res, { message: 'Berhasil logout' });
  } catch (err) { next(err); }
}

export async function changePasswordHandler(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try {
    await service.changePassword(req.user!.userId, req.body.currentPassword, req.body.newPassword);
    sendSuccess(res, { message: 'Password berhasil diubah' });
  } catch (err) { next(err); }
}

export async function getMeHandler(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try {
    const user = await service.getMe(req.user!.userId);
    sendSuccess(res, user);
  } catch (err) { next(err); }
}
