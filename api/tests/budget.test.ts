import request from 'supertest';
import { app } from '../src/index';
import { prisma } from '../src/lib/prisma';

const SCHOOL_ID = 'test-school-bg-001';
let bendaharaToken: string;
let kepalaSkolahToken: string;
let userId: string;
let planId: string;
let lineId: string;

beforeAll(async () => {
  await prisma.school.upsert({
    where: { subdomain: 'bg-test' },
    create: {
      id: SCHOOL_ID,
      name: 'Budget Test School',
      npsn: `BG${Date.now()}`,
      subdomain: 'bg-test',
      address: 'Test Address',
      config: {},
    },
    update: {},
  });

  const bendaharaEmail = `bendahara_${Date.now()}@school.test`;
  const bendaharaRes = await request(app)
    .post('/api/v1/auth/register')
    .send({ name: 'Bendahara Test', email: bendaharaEmail, password: 'Password123!', role: 'BENDAHARA', schoolId: SCHOOL_ID });
  bendaharaToken = bendaharaRes.body.data.accessToken;
  userId = bendaharaRes.body.data.user.id;

  const kepalaEmail = `kepala_${Date.now()}@school.test`;
  const kepalaRes = await request(app)
    .post('/api/v1/auth/register')
    .send({ name: 'Kepala Sekolah', email: kepalaEmail, password: 'Password123!', role: 'KEPALA_SEKOLAH', schoolId: SCHOOL_ID });
  kepalaSkolahToken = kepalaRes.body.data.accessToken;
});

afterAll(async () => {
  await prisma.budgetRealization.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.budgetLine.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.budgetPlan.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.user.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.school.deleteMany({ where: { id: SCHOOL_ID } }).catch(() => null);
  await prisma.$disconnect();
});

describe('Budget Plans (M51 Anggaran/RKAS)', () => {
  it('POST /api/v1/budget/plans - should create budget plan', async () => {
    const res = await request(app)
      .post('/api/v1/budget/plans')
      .set('Authorization', `Bearer ${bendaharaToken}`)
      .send({
        fiscalYear: '2025/2026',
        fundSource: 'BOS',
        totalAmount: 500000000,
      });
    expect(res.status).toBe(201);
    expect(res.body.data.fiscalYear).toBe('2025/2026');
    planId = res.body.data.id;
  });

  it('GET /api/v1/budget/plans - should list budget plans', async () => {
    const res = await request(app)
      .get('/api/v1/budget/plans')
      .set('Authorization', `Bearer ${bendaharaToken}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
  });

  it('GET /api/v1/budget/plans/:id - should get budget plan with lines', async () => {
    if (!planId) return;
    const res = await request(app)
      .get(`/api/v1/budget/plans/${planId}`)
      .set('Authorization', `Bearer ${bendaharaToken}`);
    expect(res.status).toBe(200);
    expect(res.body.data.id).toBe(planId);
  });

  it('PATCH /api/v1/budget/plans/:id/approve - should approve plan (KEPALA_SEKOLAH)', async () => {
    if (!planId) return;
    const res = await request(app)
      .patch(`/api/v1/budget/plans/${planId}/approve`)
      .set('Authorization', `Bearer ${kepalaSkolahToken}`);
    expect(res.status).toBe(200);
  });
});

describe('Budget Lines & Realizations', () => {
  it('POST /api/v1/budget/plans/:id/lines - should add budget line', async () => {
    if (!planId) return;
    const res = await request(app)
      .post(`/api/v1/budget/plans/${planId}/lines`)
      .set('Authorization', `Bearer ${bendaharaToken}`)
      .send({
        category: 'OPERASIONAL',
        description: 'Pembelian ATK',
        plannedAmount: 5000000,
      });
    expect(res.status).toBe(201);
    expect(res.body.data.description).toBe('Pembelian ATK');
    lineId = res.body.data.id;
  });

  it('POST /api/v1/budget/realizations - should add realization', async () => {
    if (!lineId) return;
    const res = await request(app)
      .post('/api/v1/budget/realizations')
      .set('Authorization', `Bearer ${bendaharaToken}`)
      .send({
        lineId,
        amount: 3000000,
        date: new Date().toISOString(),
        description: 'Beli ATK',
        recordedBy: userId,
      });
    expect(res.status).toBe(201);
    expect(Number(res.body.data.amount)).toBe(3000000);
  });

  it('GET /api/v1/budget/summary - should get budget summary', async () => {
    const res = await request(app)
      .get('/api/v1/budget/summary')
      .set('Authorization', `Bearer ${bendaharaToken}`);
    expect(res.status).toBe(200);
  });
});
