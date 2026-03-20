import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export const roomBookingRouter = Router();
roomBookingRouter.use(authenticate);

roomBookingRouter.get('/rooms', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getRooms(req.user!.schoolId, req.query as Record<string, string>)); } catch (err) { next(err); }
});
roomBookingRouter.post('/rooms', authorize('ADMIN_SEKOLAH', 'TATA_USAHA'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.createRoom(req.user!.schoolId, req.body)); } catch (err) { next(err); }
});
roomBookingRouter.post('/bookings', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.bookRoom(req.user!.schoolId, { ...req.body, requesterId: req.user!.userId })); } catch (err) { next(err); }
});
roomBookingRouter.get('/bookings', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { bookings, meta } = await service.getBookings(req.user!.schoolId, req.query as Record<string, string>);
    sendSuccess(res, bookings, 200, meta);
  } catch (err) { next(err); }
});
roomBookingRouter.put('/bookings/:id/approve', authorize('ADMIN_SEKOLAH', 'TATA_USAHA', 'KEPALA_SEKOLAH'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.approveBooking(req.user!.schoolId, req.params.id as string, req.user!.userId, req.body.approved)); } catch (err) { next(err); }
});
