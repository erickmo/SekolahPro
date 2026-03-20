import { Request, Response, NextFunction } from 'express';
import { sendSuccess, sendCreated, sendNoContent } from '../../shared/response';
import * as service from './service';

export async function createSchoolHandler(req: Request, res: Response, next: NextFunction): Promise<void> {
  try { sendCreated(res, await service.createSchool(req.body)); } catch (err) { next(err); }
}

export async function getSchoolsHandler(req: Request, res: Response, next: NextFunction): Promise<void> {
  try {
    const { schools, meta } = await service.getSchools(req.query as { page?: number; limit?: number; search?: string });
    sendSuccess(res, schools, 200, meta);
  } catch (err) { next(err); }
}

export async function getSchoolHandler(req: Request, res: Response, next: NextFunction): Promise<void> {
  try { sendSuccess(res, await service.getSchool(req.params.id as string)); } catch (err) { next(err); }
}

export async function updateSchoolHandler(req: Request, res: Response, next: NextFunction): Promise<void> {
  try { sendSuccess(res, await service.updateSchool(req.params.id as string, req.body)); } catch (err) { next(err); }
}

export async function deleteSchoolHandler(req: Request, res: Response, next: NextFunction): Promise<void> {
  try { await service.deleteSchool(req.params.id as string); sendNoContent(res); } catch (err) { next(err); }
}
