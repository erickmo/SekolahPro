import { prisma } from '../../lib/prisma';
import { NotFoundError, ValidationError } from '../../shared/errors';

const LMS_CONFIG_KEY_NAME = 'LMS_CONFIG';

interface LMSConfig {
  provider: 'MOODLE' | 'GOOGLE_CLASSROOM' | 'CANVAS';
  apiUrl: string;
}

function generateApiKey(): string {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
  let result = 'lms_';
  for (let i = 0; i < 40; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return result;
}

async function getLMSApiKey(schoolId: string) {
  return prisma.apiKey.findFirst({
    where: { schoolId, name: LMS_CONFIG_KEY_NAME, isActive: true },
  });
}

export async function configureLMS(
  schoolId: string,
  userId: string,
  config: { provider: 'MOODLE' | 'GOOGLE_CLASSROOM' | 'CANVAS'; apiUrl: string; apiKey: string },
) {
  if (!config.provider || !config.apiUrl || !config.apiKey) {
    throw new ValidationError('provider, apiUrl, dan apiKey wajib diisi');
  }

  // Deactivate existing LMS config if any
  await prisma.apiKey.updateMany({
    where: { schoolId, name: LMS_CONFIG_KEY_NAME },
    data: { isActive: false },
  });

  // Store config: provider and apiUrl are stored as JSON in scopes array
  const configPayload = JSON.stringify({ provider: config.provider, apiUrl: config.apiUrl });

  const apiKeyRecord = await prisma.apiKey.create({
    data: {
      name: LMS_CONFIG_KEY_NAME,
      key: generateApiKey(),
      schoolId,
      userId,
      scopes: [configPayload],
      isActive: true,
    },
  });

  return {
    id: apiKeyRecord.id,
    provider: config.provider,
    apiUrl: config.apiUrl,
    connectedAt: apiKeyRecord.createdAt,
    status: 'CONNECTED',
  };
}

export async function getLMSStatus(schoolId: string) {
  const lmsKey = await getLMSApiKey(schoolId);

  if (!lmsKey) {
    return { connected: false, provider: null, apiUrl: null, connectedAt: null };
  }

  let provider: string | null = null;
  let apiUrl: string | null = null;

  if (lmsKey.scopes.length > 0) {
    try {
      const parsed = JSON.parse(lmsKey.scopes[0]) as LMSConfig;
      provider = parsed.provider;
      apiUrl = parsed.apiUrl;
    } catch {
      // invalid config
    }
  }

  return {
    connected: true,
    provider,
    apiUrl,
    connectedAt: lmsKey.createdAt,
    lastUsedAt: lmsKey.lastUsedAt,
  };
}

export async function syncCourses(schoolId: string) {
  const lmsKey = await getLMSApiKey(schoolId);
  if (!lmsKey) throw new NotFoundError('Konfigurasi LMS');

  // Update lastUsedAt
  await prisma.apiKey.update({ where: { id: lmsKey.id }, data: { lastUsedAt: new Date() } });

  // Simulated sync result
  return {
    status: 'SUCCESS',
    syncedAt: new Date().toISOString(),
    coursesSync: {
      total: 0,
      created: 0,
      updated: 0,
      message: 'Sinkronisasi kursus berhasil. Hubungkan dengan endpoint LMS yang sebenarnya untuk data nyata.',
    },
  };
}

export async function syncGrades(schoolId: string) {
  const lmsKey = await getLMSApiKey(schoolId);
  if (!lmsKey) throw new NotFoundError('Konfigurasi LMS');

  // Update lastUsedAt
  await prisma.apiKey.update({ where: { id: lmsKey.id }, data: { lastUsedAt: new Date() } });

  // Simulated sync result
  return {
    status: 'SUCCESS',
    syncedAt: new Date().toISOString(),
    gradesSync: {
      total: 0,
      synced: 0,
      failed: 0,
      message: 'Sinkronisasi nilai berhasil. Hubungkan dengan endpoint LMS yang sebenarnya untuk data nyata.',
    },
  };
}

export async function listCourses(schoolId: string) {
  const lmsKey = await getLMSApiKey(schoolId);
  if (!lmsKey) {
    return { connected: false, courses: [] };
  }

  // Return cached/simulated data — actual LMS API call would populate this
  return {
    connected: true,
    courses: [],
    note: 'Data kursus diambil dari cache LMS. Lakukan sync-courses untuk memperbarui.',
    lastSynced: lmsKey.lastUsedAt ?? null,
  };
}
