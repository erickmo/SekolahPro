import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth';
import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import { callClaude } from '../../lib/claude';
import { prisma } from '../../lib/prisma';

export const sentimentRouter = Router();
sentimentRouter.use(authenticate);

sentimentRouter.post('/analyze', async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const { text, source } = req.body;
    const result = await callClaude(
      'Kamu adalah analis sentimen untuk komunitas sekolah. Analisis sentimen teks dalam JSON.',
      `Analisis sentimen berikut: "${text}"\n\nJSON format: {"sentiment": "POSITIF|NEGATIF|NETRAL", "score": -1.0 to 1.0, "topics": string[], "summary": string}`,
      500,
    );
    const match = result.match(/\{[\s\S]*\}/);
    const analysis = match ? JSON.parse(match[0]) : { sentiment: 'NETRAL', score: 0, topics: [], summary: 'Analisis tidak tersedia' };
    sendCreated(res, { ...analysis, source, analyzedAt: new Date() });
  } catch (err) { next(err); }
});

sentimentRouter.get('/dashboard', authorize('ADMIN_SEKOLAH', 'KEPALA_SEKOLAH'), async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    const interactions = await prisma.aIInteraction.findMany({ where: { schoolId: req.user!.schoolId, type: 'SENTIMENT' }, take: 50, orderBy: { createdAt: 'desc' } });
    sendSuccess(res, { total: interactions.length, interactions, satisfactionIndex: 75, message: 'Dashboard sentimen sekolah' });
  } catch (err) { next(err); }
});
