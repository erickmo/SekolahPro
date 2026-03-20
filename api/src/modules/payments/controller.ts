import { Request, Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export async function generateInvoicesH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendCreated(res, await service.generateMonthlyInvoices(req.user!.schoolId, req.body.semesterId, req.body.amount, req.body.dueDate)); } catch (err) { next(err); }
}
export async function getInvoicesH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try {
    const { invoices, meta } = await service.getInvoices(req.user!.schoolId, req.query as Record<string, string>);
    sendSuccess(res, invoices, 200, meta);
  } catch (err) { next(err); }
}
export async function createPaymentTokenH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendCreated(res, await service.createPaymentToken(req.user!.schoolId, req.params.invoiceId as string)); } catch (err) { next(err); }
}
export async function paymentCallbackH(req: Request, res: Response, next: NextFunction): Promise<void> {
  try { await service.handlePaymentCallback(req.body); sendSuccess(res, { ok: true }); } catch (err) { next(err); }
}
export async function financialSummaryH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendSuccess(res, await service.getFinancialSummary(req.user!.schoolId, req.query as { startDate?: string; endDate?: string })); } catch (err) { next(err); }
}
