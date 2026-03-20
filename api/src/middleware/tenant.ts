import { Response, NextFunction } from 'express';
import { AuthRequest } from '../shared/types';
import { prisma } from '../lib/prisma';
import { ForbiddenError } from '../shared/errors';

export function requireSchool(req: AuthRequest, _res: Response, next: NextFunction): void {
  if (!req.user?.schoolId) return next(new ForbiddenError('School context diperlukan'));
  next();
}

export async function validateTenant(req: AuthRequest, _res: Response, next: NextFunction): Promise<void> {
  try {
    if (!req.user?.schoolId) return next();
    const school = await prisma.school.findUnique({ where: { id: req.user.schoolId }, select: { id: true } });
    if (!school) return next(new ForbiddenError('Sekolah tidak ditemukan'));
    next();
  } catch (err) {
    next(err);
  }
}
