import request from 'supertest';
import { app } from '../src/index';
import { prisma } from '../src/lib/prisma';

const SCHOOL_ID = 'test-school-sec-001';
let token: string;

beforeAll(async () => {
  await prisma.school.upsert({
    where: { subdomain: 'sec-test' },
    create: {
      id: SCHOOL_ID,
      name: 'Security Test School',
      npsn: `SEC${Date.now()}`,
      subdomain: 'sec-test',
      address: 'Test Address',
      config: {},
    },
    update: {},
  });

  const email = `admin_sec_${Date.now()}@school.test`;
  const regRes = await request(app)
    .post('/api/v1/auth/register')
    .send({ name: 'Admin Security', email, password: 'Password123!', role: 'ADMIN_SEKOLAH', schoolId: SCHOOL_ID });
  token = regRes.body.data.accessToken;
});

afterAll(async () => {
  await prisma.user.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.school.deleteMany({ where: { id: SCHOOL_ID } }).catch(() => null);
  await prisma.$disconnect();
});

describe('Security Center', () => {
  it('GET /api/v1/security/dashboard — get security dashboard stats', async () => {
    const res = await request(app)
      .get('/api/v1/security/dashboard')
      .set('Authorization', `Bearer ${token}`);
    expect(res.status).toBe(200);
    expect(res.body.success).toBe(true);
    expect(res.body.data).toHaveProperty('failedLogins24h');
    expect(res.body.data).toHaveProperty('totalUsers');
    expect(typeof res.body.data.failedLogins24h).toBe('number');
    expect(typeof res.body.data.totalUsers).toBe('number');
  });

  it('GET /api/v1/security/suspicious — get suspicious activity (may be empty)', async () => {
    const res = await request(app)
      .get('/api/v1/security/suspicious')
      .set('Authorization', `Bearer ${token}`);
    expect(res.status).toBe(200);
    expect(res.body.success).toBe(true);
    expect(Array.isArray(res.body.data)).toBe(true);
  });

  it('POST /api/v1/security/mfa/enable — enable MFA', async () => {
    const res = await request(app)
      .post('/api/v1/security/mfa/enable')
      .set('Authorization', `Bearer ${token}`);
    expect(res.status).toBe(200);
    expect(res.body.success).toBe(true);
    expect(res.body.data.mfaEnabled).toBe(true);
  });

  it('POST /api/v1/security/mfa/disable — disable MFA', async () => {
    const res = await request(app)
      .post('/api/v1/security/mfa/disable')
      .set('Authorization', `Bearer ${token}`);
    expect(res.status).toBe(200);
    expect(res.body.success).toBe(true);
    expect(res.body.data.mfaEnabled).toBe(false);
  });

  it('GET /api/v1/security/dashboard — rejects unauthenticated request', async () => {
    const res = await request(app).get('/api/v1/security/dashboard');
    expect(res.status).toBe(401);
  });
});
