import request from 'supertest';
import { app } from '../src/index';
import { prisma } from '../src/lib/prisma';

const SCHOOL_ID = 'test-school-ten-001';
let superadminToken: string;
let tenantId: string;

beforeAll(async () => {
  await prisma.school.upsert({
    where: { subdomain: 'ten-test' },
    create: {
      id: SCHOOL_ID,
      name: 'Tenant Test School',
      npsn: `TEN${Date.now()}`,
      subdomain: 'ten-test',
      address: 'Test Address',
      config: {},
    },
    update: {},
  });

  const email = `superadmin_ten_${Date.now()}@school.test`;
  const regRes = await request(app)
    .post('/api/v1/auth/register')
    .send({ name: 'EDS Superadmin', email, password: 'Password123!', role: 'EDS_SUPERADMIN', schoolId: SCHOOL_ID });
  superadminToken = regRes.body.data.accessToken;
});

afterAll(async () => {
  if (tenantId) {
    await prisma.tenantAddon.deleteMany({ where: { tenantId } }).catch(() => null);
    await prisma.tenantFeature.deleteMany({ where: { tenantId } }).catch(() => null);
    await prisma.tenantSubscription.deleteMany({ where: { tenantId } }).catch(() => null);
    await prisma.tenant.deleteMany({ where: { id: tenantId } }).catch(() => null);
  }
  await prisma.user.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.school.deleteMany({ where: { id: SCHOOL_ID } }).catch(() => null);
  await prisma.$disconnect();
});

describe('Tenant Management', () => {
  it('POST /api/v1/tenant — create tenant as EDS_SUPERADMIN', async () => {
    const res = await request(app)
      .post('/api/v1/tenant')
      .set('Authorization', `Bearer ${superadminToken}`)
      .send({
        schoolId: SCHOOL_ID,
        subdomain: `tenant-test-${Date.now()}`,
        billingEmail: 'billing@school.test',
        tier: 'BASIC',
      });
    expect(res.status).toBe(201);
    expect(res.body.success).toBe(true);
    expect(res.body.data).toHaveProperty('id');
    expect(res.body.data.tier).toBe('BASIC');
    tenantId = res.body.data.id;
  });

  it('GET /api/v1/tenant/my — get own tenant', async () => {
    const res = await request(app)
      .get('/api/v1/tenant/my')
      .set('Authorization', `Bearer ${superadminToken}`);
    expect(res.status).toBe(200);
    expect(res.body.success).toBe(true);
  });

  it('GET /api/v1/tenant — list tenants as EDS_SUPERADMIN', async () => {
    const res = await request(app)
      .get('/api/v1/tenant')
      .set('Authorization', `Bearer ${superadminToken}`);
    expect(res.status).toBe(200);
    expect(res.body.success).toBe(true);
    expect(Array.isArray(res.body.data)).toBe(true);
  });

  it('GET /api/v1/tenant/:id — get tenant detail with features', async () => {
    if (!tenantId) return;
    const res = await request(app)
      .get(`/api/v1/tenant/${tenantId}`)
      .set('Authorization', `Bearer ${superadminToken}`);
    expect(res.status).toBe(200);
    expect(res.body.success).toBe(true);
    expect(res.body.data).toHaveProperty('id', tenantId);
    expect(res.body.data).toHaveProperty('features');
    expect(Array.isArray(res.body.data.features)).toBe(true);
  });

  it('PATCH /api/v1/tenant/:id/plan — upgrade plan to PRO', async () => {
    if (!tenantId) return;
    const res = await request(app)
      .patch(`/api/v1/tenant/${tenantId}/plan`)
      .set('Authorization', `Bearer ${superadminToken}`)
      .send({ tier: 'PRO' });
    expect(res.status).toBe(200);
    expect(res.body.success).toBe(true);
    expect(res.body.data.tier).toBe('PRO');
  });

  it('PATCH /api/v1/tenant/:id/features — toggle feature', async () => {
    if (!tenantId) return;
    const res = await request(app)
      .patch(`/api/v1/tenant/${tenantId}/features`)
      .set('Authorization', `Bearer ${superadminToken}`)
      .send({ moduleId: 'M04', enabled: true });
    expect(res.status).toBe(200);
    expect(res.body.success).toBe(true);
    expect(res.body.data.moduleId).toBe('M04');
    expect(res.body.data.enabled).toBe(true);
  });

  it('POST /api/v1/tenant/:id/addons — add addon', async () => {
    if (!tenantId) return;
    const res = await request(app)
      .post(`/api/v1/tenant/${tenantId}/addons`)
      .set('Authorization', `Bearer ${superadminToken}`)
      .send({
        addon: 'AI_QUOTA_EXTRA',
        quantity: 5,
        activeFrom: new Date().toISOString(),
        activeTo: new Date(Date.now() + 365 * 86400000).toISOString(),
      });
    expect(res.status).toBe(201);
    expect(res.body.success).toBe(true);
    expect(res.body.data).toHaveProperty('id');
    expect(res.body.data.addon).toBe('AI_QUOTA_EXTRA');
  });
});
