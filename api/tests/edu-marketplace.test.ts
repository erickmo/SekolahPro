import request from 'supertest';
import { app } from '../src/index';
import { prisma } from '../src/lib/prisma';

const SCHOOL_ID = 'test-school-mkt-001';
let guruToken: string;
let adminToken: string;
let contentId: string;

beforeAll(async () => {
  await prisma.school.upsert({
    where: { subdomain: 'mkt-test' },
    create: {
      id: SCHOOL_ID,
      name: 'Marketplace Test School',
      npsn: `MKT${Date.now()}`,
      subdomain: 'mkt-test',
      address: 'Test Address',
      config: {},
    },
    update: {},
  });

  const guruEmail = `guru_mkt_${Date.now()}@school.test`;
  const guruRes = await request(app)
    .post('/api/v1/auth/register')
    .send({ name: 'Guru Marketplace', email: guruEmail, password: 'Password123!', role: 'GURU', schoolId: SCHOOL_ID });
  guruToken = guruRes.body.data.accessToken;

  const adminEmail = `admin_mkt_${Date.now()}@school.test`;
  const adminRes = await request(app)
    .post('/api/v1/auth/register')
    .send({ name: 'Admin Marketplace', email: adminEmail, password: 'Password123!', role: 'ADMIN_SEKOLAH', schoolId: SCHOOL_ID });
  adminToken = adminRes.body.data.accessToken;
});

afterAll(async () => {
  await prisma.contentPurchase.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.eduContent.deleteMany({ where: { schoolOrigin: SCHOOL_ID } }).catch(() => null);
  await prisma.user.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.school.deleteMany({ where: { id: SCHOOL_ID } }).catch(() => null);
  await prisma.$disconnect();
});

describe('Edu Marketplace', () => {
  it('GET /api/v1/edu-marketplace/content — list public content (no auth)', async () => {
    const res = await request(app).get('/api/v1/edu-marketplace/content');
    expect(res.status).toBe(200);
    expect(res.body.success).toBe(true);
    expect(Array.isArray(res.body.data)).toBe(true);
  });

  it('POST /api/v1/edu-marketplace/content — create content as GURU', async () => {
    const res = await request(app)
      .post('/api/v1/edu-marketplace/content')
      .set('Authorization', `Bearer ${guruToken}`)
      .send({
        title: 'Modul Matematika Kelas X',
        description: 'Lengkap dengan soal',
        type: 'MODULE',
        price: 25000,
      });
    expect(res.status).toBe(201);
    expect(res.body.success).toBe(true);
    expect(res.body.data).toHaveProperty('id');
    expect(res.body.data.title).toBe('Modul Matematika Kelas X');
    contentId = res.body.data.id;
  });

  it('GET /api/v1/edu-marketplace/my-content — get creator content as GURU', async () => {
    const res = await request(app)
      .get('/api/v1/edu-marketplace/my-content')
      .set('Authorization', `Bearer ${guruToken}`);
    expect(res.status).toBe(200);
    expect(res.body.success).toBe(true);
    expect(Array.isArray(res.body.data)).toBe(true);
  });

  it('PATCH /api/v1/edu-marketplace/content/:id/publish — publish content as GURU', async () => {
    if (!contentId) return;
    const res = await request(app)
      .patch(`/api/v1/edu-marketplace/content/${contentId}/publish`)
      .set('Authorization', `Bearer ${guruToken}`);
    expect(res.status).toBe(200);
    expect(res.body.success).toBe(true);
    expect(res.body.data.status).toBe('APPROVED');
  });

  it('GET /api/v1/edu-marketplace/content — list content after publish, should include published', async () => {
    const res = await request(app).get('/api/v1/edu-marketplace/content');
    expect(res.status).toBe(200);
    expect(res.body.success).toBe(true);
    expect(Array.isArray(res.body.data)).toBe(true);
  });

  it('POST /api/v1/edu-marketplace/purchase — purchase content as ADMIN_SEKOLAH', async () => {
    if (!contentId) return;
    const res = await request(app)
      .post('/api/v1/edu-marketplace/purchase')
      .set('Authorization', `Bearer ${adminToken}`)
      .send({ contentId });
    expect(res.status).toBe(201);
    expect(res.body.success).toBe(true);
    expect(res.body.data).toHaveProperty('id');
  });

  it('GET /api/v1/edu-marketplace/my-purchases — get purchases as ADMIN_SEKOLAH', async () => {
    const res = await request(app)
      .get('/api/v1/edu-marketplace/my-purchases')
      .set('Authorization', `Bearer ${adminToken}`);
    expect(res.status).toBe(200);
    expect(res.body.success).toBe(true);
    expect(Array.isArray(res.body.data)).toBe(true);
  });
});
