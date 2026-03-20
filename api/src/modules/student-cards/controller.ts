import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export async function generateCardHandler(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendCreated(res, await service.generateStudentCard(req.user!.schoolId, req.params.studentId as string)); } catch (err) { next(err); }
}

export async function getCardHandler(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendSuccess(res, await service.getStudentCard(req.user!.schoolId, req.params.studentId as string)); } catch (err) { next(err); }
}

export async function verifyQrHandler(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendSuccess(res, await service.verifyQrCode(req.body.qrData)); } catch (err) { next(err); }
}

export async function deactivateCardHandler(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendSuccess(res, await service.deactivateCard(req.user!.schoolId, req.params.studentId as string)); } catch (err) { next(err); }
}
