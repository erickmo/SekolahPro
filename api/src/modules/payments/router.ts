import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import * as ctrl from './controller';

export const paymentsRouter = Router();

paymentsRouter.post('/callback', ctrl.paymentCallbackH);
paymentsRouter.use(authenticate);
paymentsRouter.post('/invoices/generate', authorize('BENDAHARA', 'ADMIN_SEKOLAH'), ctrl.generateInvoicesH);
paymentsRouter.get('/invoices', ctrl.getInvoicesH);
paymentsRouter.post('/invoices/:invoiceId/pay', ctrl.createPaymentTokenH);
paymentsRouter.get('/financial-summary', authorize('BENDAHARA', 'KEPALA_SEKOLAH', 'ADMIN_SEKOLAH'), ctrl.financialSummaryH);
