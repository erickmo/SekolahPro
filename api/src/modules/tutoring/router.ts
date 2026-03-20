import { Router } from 'express';
import { authenticate } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import { prisma } from '../../lib/prisma';

export const tutoringRouter = Router();
tutoringRouter.use(authenticate);

tutoringRouter.get('/tutors', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const tutors = await prisma.user.findMany({
      where: { schoolId: req.user!.schoolId, role: 'GURU_LES' },
      select: { id: true, name: true, photoUrl: true },
    });
    sendSuccess(res, tutors);
  } catch (err) { next(err); }
});

tutoringRouter.post('/sessions', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    sendCreated(res, {
      id: `SES-${Date.now()}`,
      tutorId: req.body.tutorId,
      studentId: req.body.studentId,
      subject: req.body.subject,
      scheduledAt: req.body.scheduledAt,
      status: 'PENDING',
      message: 'Sesi les berhasil dibuat. Menunggu konfirmasi dari tutor.',
    });
  } catch (err) { next(err); }
});
