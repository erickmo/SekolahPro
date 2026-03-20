import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

export async function createAccountHandler(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendCreated(res, await service.createAccount(req.user!.schoolId, req.params.studentId as string)); } catch (err) { next(err); }
}

export async function getAccountHandler(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendSuccess(res, await service.getAccount(req.user!.schoolId, req.params.studentId as string)); } catch (err) { next(err); }
}

export async function depositHandler(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendSuccess(res, await service.deposit(req.user!.schoolId, req.params.studentId as string, req.body.amount, req.body.note, req.user!.userId)); } catch (err) { next(err); }
}

export async function withdrawHandler(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendSuccess(res, await service.withdraw(req.user!.schoolId, req.params.studentId as string, req.body.amount, req.body.note, req.user!.userId)); } catch (err) { next(err); }
}

export async function getTransactionsHandler(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try {
    const { transactions, meta } = await service.getTransactions(req.user!.schoolId, req.params.studentId as string, req.query as Record<string, string>);
    sendSuccess(res, transactions, 200, meta);
  } catch (err) { next(err); }
}

export async function getProductsHandler(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendSuccess(res, await service.getProducts(req.user!.schoolId)); } catch (err) { next(err); }
}

export async function createProductHandler(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendCreated(res, await service.createProduct(req.user!.schoolId, req.body)); } catch (err) { next(err); }
}

export async function getDailyReportHandler(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendSuccess(res, await service.getDailyReport(req.user!.schoolId, req.query.date as string || new Date().toISOString().split('T')[0])); } catch (err) { next(err); }
}
