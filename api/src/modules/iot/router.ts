import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export const iotRouter = Router();

iotRouter.post('/devices', authenticate, authorize('ADMIN_SEKOLAH', 'TATA_USAHA'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.registerDevice(req.user!.schoolId, req.body)); } catch (err) { next(err); }
});

iotRouter.get('/devices', authenticate, async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.listDevices(req.user!.schoolId, req.query as Record<string, string>)); } catch (err) { next(err); }
});

iotRouter.get('/devices/:id/latest', authenticate, async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getLatestReading(req.user!.schoolId, req.params.id as string)); } catch (err) { next(err); }
});

iotRouter.get('/devices/:id', authenticate, async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.getDevice(req.user!.schoolId, req.params.id as string)); } catch (err) { next(err); }
});

iotRouter.patch('/devices/:id/status', authenticate, async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendSuccess(res, await service.updateDeviceStatus(req.user!.schoolId, req.params.id as string, req.body.isActive)); } catch (err) { next(err); }
});

// Public — IoT push
iotRouter.post('/readings', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try { sendCreated(res, await service.recordReading(req.body)); } catch (err) { next(err); }
});

iotRouter.get('/readings', authenticate, async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { readings, meta } = await service.listReadings(req.user!.schoolId, req.query as Record<string, string>);
    sendSuccess(res, readings, 200, meta);
  } catch (err) { next(err); }
});
