import { Request, Response, NextFunction } from 'express';
import { ZodError } from 'zod';
import { AppError } from '../shared/errors';
import { sendError } from '../shared/response';
import { logger } from '../lib/logger';

export function errorHandler(err: Error, req: Request, res: Response, _next: NextFunction): void {
  if (err instanceof AppError) {
    sendError(res, err.statusCode, err.code, err.message);
    return;
  }

  if (err instanceof ZodError) {
    const message = err.errors.map((e) => `${e.path.join('.')}: ${e.message}`).join(', ');
    sendError(res, 400, 'VALIDATION_ERROR', message);
    return;
  }

  logger.error('Unhandled error', { error: err.message, stack: err.stack, url: req.url });
  sendError(res, 500, 'INTERNAL_ERROR', 'Terjadi kesalahan internal server');
}

export function notFoundHandler(req: Request, res: Response): void {
  sendError(res, 404, 'NOT_FOUND', `Route ${req.method} ${req.path} tidak ditemukan`);
}
