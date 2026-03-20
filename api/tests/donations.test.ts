import request from 'supertest';
import { app } from '../src/index';
import { prisma } from '../src/lib/prisma';

const SCHOOL_ID = 'test-school-dn-001';
let bendaharaToken: string;
let donationId: string;

beforeAll(async () => {
  await prisma.school.upsert({
    where: { subdomain: 'dn-test' },
    create: {
      id: SCHOOL_ID,
      name: 'Donation Test School',
      npsn: `DN${Date.now()}`,
      subdomain: 'dn-test',
      address: 'Test Address',
      config: {},
    },
    update: {},
  });

  const bendaharaEmail = `bendahara_${Date.now()}@school.test`;
  const bendaharaRes = await request(app)
    .post('/api/v1/auth/register')
    .send({ name: 'Bendahara Komite', email: bendaharaEmail, password: 'Password123!', role: 'BENDAHARA', schoolId: SCHOOL_ID });
  bendaharaToken = bendaharaRes.body.data.accessToken;
});

afterAll(async () => {
  await prisma.donationRecord.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.user.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.school.deleteMany({ where: { id: SCHOOL_ID } }).catch(() => null);
  await prisma.$disconnect();
});

describe('Donations (M52 Dana Komite)', () => {
  it('POST /api/v1/donations - should create donation', async () => {
    const res = await request(app)
      .post('/api/v1/donations')
      .set('Authorization', `Bearer ${bendaharaToken}`)
      .send({
        donorName: 'PT Maju Jaya',
        amount: 10000000,
        purpose: 'Renovasi Perpustakaan',
      });
    expect(res.status).toBe(201);
    expect(res.body.data.donorName).toBe('PT Maju Jaya');
    donationId = res.body.data.id;
  });

  it('GET /api/v1/donations - should list donations', async () => {
    const res = await request(app)
      .get('/api/v1/donations')
      .set('Authorization', `Bearer ${bendaharaToken}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
  });

  it('GET /api/v1/donations/:id - should get donation', async () => {
    if (!donationId) return;
    const res = await request(app)
      .get(`/api/v1/donations/${donationId}`)
      .set('Authorization', `Bearer ${bendaharaToken}`);
    expect(res.status).toBe(200);
    expect(res.body.data.id).toBe(donationId);
  });

  it('GET /api/v1/donations/summary - should get donation summary', async () => {
    const res = await request(app)
      .get('/api/v1/donations/summary')
      .set('Authorization', `Bearer ${bendaharaToken}`);
    expect(res.status).toBe(200);
    expect(res.body.data.totalAmount).toBeGreaterThanOrEqual(0);
  });
});
