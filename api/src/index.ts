import 'dotenv/config';
import express from 'express';
import cors from 'cors';
import helmet from 'helmet';
import { config } from './config';
import { connectRedis } from './lib/redis';
import { logger } from './lib/logger';
import { errorHandler, notFoundHandler } from './middleware/error';
import { router } from './routes';

const app = express();

app.use(helmet());
app.use(cors({ origin: true, credentials: true }));
app.use(express.json({ limit: '10mb' }));
app.use(express.urlencoded({ extended: true, limit: '10mb' }));

app.get('/health', (_req, res) => res.json({ status: 'ok', timestamp: new Date().toISOString() }));

app.use('/api/v1', router);

app.use(notFoundHandler);
app.use(errorHandler);

async function start() {
  await connectRedis();
  logger.info('Redis connected');

  app.listen(config.port, () => {
    logger.info(`EDS API running on port ${config.port} [${config.env}]`);
  });
}

if (process.env.NODE_ENV !== 'test') {
  start().catch((err) => {
    logger.error('Failed to start server', { error: err.message });
    process.exit(1);
  });
}

export { app };
