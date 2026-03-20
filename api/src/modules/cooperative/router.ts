import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import * as ctrl from './controller';

export const cooperativeRouter = Router();
cooperativeRouter.use(authenticate);

cooperativeRouter.get('/products', ctrl.getProductsHandler);
cooperativeRouter.post('/products', authorize('KASIR_KOPERASI', 'BENDAHARA', 'ADMIN_SEKOLAH'), ctrl.createProductHandler);
cooperativeRouter.get('/reports/daily', authorize('BENDAHARA', 'ADMIN_SEKOLAH', 'KEPALA_SEKOLAH'), ctrl.getDailyReportHandler);
cooperativeRouter.post('/:studentId/account', authorize('KASIR_KOPERASI', 'BENDAHARA', 'ADMIN_SEKOLAH'), ctrl.createAccountHandler);
cooperativeRouter.get('/:studentId/account', ctrl.getAccountHandler);
cooperativeRouter.post('/:studentId/deposit', authorize('KASIR_KOPERASI', 'BENDAHARA'), ctrl.depositHandler);
cooperativeRouter.post('/:studentId/withdraw', authorize('KASIR_KOPERASI', 'BENDAHARA'), ctrl.withdrawHandler);
cooperativeRouter.get('/:studentId/transactions', ctrl.getTransactionsHandler);
