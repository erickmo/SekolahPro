import request from 'supertest';
import { app } from '../src/index';
import { prisma } from '../src/lib/prisma';

const SCHOOL_ID = 'test-school-aud-001';
let token: string;

beforeAll(async () => {
  await prisma.school.upsert({
    where: { subdomain: 'aud-test' },
    create: {
      id: SCHOOL_ID,
      name: 'Audit Log Test School',
      npsn: `AUD${Date.now()}`,
      subdomain: 'aud-test',
      address: 'Test Address',
      config: {},
    },
    update: {},
  });

  const email = `admin_aud_${Date.now()}@school.test`;
  const regRes = await request(app)
    .post('/api/v1/auth/register')
    .send({ name: 'Admin Audit', email, password: 'Password123!', role: 'ADMIN_SEKOLAH', schoolId: SCHOOL_ID });
  token = regRes.body.data.accessToken;
});

afterAll(async () => {
  await prisma.auditLog.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.user.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.school.deleteMany({ where: { id: SCHOOL_ID } }).catch(() => null);
  await prisma.$disconnect();
});

describe('Audit Log', () => {
  it('GET /api/v1/audit-logs — list audit logs (may be empty initially)', async () => {
    const res = await request(app)
      .get('/api/v1/audit-logs')
      .set('Authorization', `Bearer ${token}`);
    expect(res.status).toBe(200);
    expect(res.body.success).toBe(true);
    expect(Array.isArray(res.body.data)).toBe(true);
  });

  it('GET /api/v1/audit-logs — with entity filter and pagination', async () => {
    const res = await request(app)
      .get('/api/v1/audit-logs?entity=Student&page=1&limit=10')
      .set('Authorization', `Bearer ${token}`);
    expect(res.status).toBe(200);
    expect(res.body.success).toBe(true);
    expect(Array.isArray(res.body.data)).toBe(true);
    expect(res.body.meta).toHaveProperty('total');
  });

  it('GET /api/v1/audit-logs — with date range filter', async () => {
    const res = await request(app)
      .get('/api/v1/audit-logs?startDate=2024-01-01&endDate=2030-12-31')
      .set('Authorization', `Bearer ${token}`);
    expect(res.status).toBe(200);
    expect(res.body.success).toBe(true);
    expect(Array.isArray(res.body.data)).toBe(true);
  });

  it('GET /api/v1/audit-logs — rejects unauthenticated request', async () => {
    const res = await request(app).get('/api/v1/audit-logs');
    expect(res.status).toBe(401);
  });
});
