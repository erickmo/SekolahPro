import request from 'supertest';
import { app } from '../src/index';
import { prisma } from '../src/lib/prisma';

const SCHOOL_ID = 'test-school-lp-001';
let guruToken: string;
let kurikulumToken: string;
let userId: string;
let planId: string;

beforeAll(async () => {
  await prisma.school.upsert({
    where: { subdomain: 'lp-test' },
    create: {
      id: SCHOOL_ID,
      name: 'Lesson Plan Test School',
      npsn: `LP${Date.now()}`,
      subdomain: 'lp-test',
      address: 'Test Address',
      config: {},
    },
    update: {},
  });

  const guruEmail = `guru_${Date.now()}@school.test`;
  const guruRes = await request(app)
    .post('/api/v1/auth/register')
    .send({ name: 'Guru Test', email: guruEmail, password: 'Password123!', role: 'GURU', schoolId: SCHOOL_ID });
  guruToken = guruRes.body.data.accessToken;
  userId = guruRes.body.data.user.id;

  const kurikulumEmail = `kurikulum_${Date.now()}@school.test`;
  const kurikulumRes = await request(app)
    .post('/api/v1/auth/register')
    .send({ name: 'Kepala Kurikulum', email: kurikulumEmail, password: 'Password123!', role: 'KEPALA_KURIKULUM', schoolId: SCHOOL_ID });
  kurikulumToken = kurikulumRes.body.data.accessToken;
});

afterAll(async () => {
  await prisma.lessonPlanVersion.deleteMany({ where: { lessonPlan: { schoolId: SCHOOL_ID } } }).catch(() => null);
  await prisma.lessonPlan.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.user.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.school.deleteMany({ where: { id: SCHOOL_ID } }).catch(() => null);
  await prisma.$disconnect();
});

describe('Lesson Plans (M48 RPP Digital)', () => {
  it('POST /api/v1/lesson-plans - should create RPP', async () => {
    const res = await request(app)
      .post('/api/v1/lesson-plans')
      .set('Authorization', `Bearer ${guruToken}`)
      .send({
        teacherId: userId,
        subject: 'Matematika',
        gradeLevel: '10',
        topic: 'Persamaan Linear',
        curriculum: 'MERDEKA',
        durationMins: 90,
        content: { objectives: [], activities: [] },
      });
    expect(res.status).toBe(201);
    expect(res.body.data.topic).toBe('Persamaan Linear');
    planId = res.body.data.id;
  });

  it('GET /api/v1/lesson-plans - should list RPPs', async () => {
    const res = await request(app)
      .get('/api/v1/lesson-plans')
      .set('Authorization', `Bearer ${guruToken}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
  });

  it('GET /api/v1/lesson-plans/:id - should get single RPP', async () => {
    if (!planId) return;
    const res = await request(app)
      .get(`/api/v1/lesson-plans/${planId}`)
      .set('Authorization', `Bearer ${guruToken}`);
    expect(res.status).toBe(200);
    expect(res.body.data.id).toBe(planId);
  });

  it('PATCH /api/v1/lesson-plans/:id - should update RPP', async () => {
    if (!planId) return;
    const res = await request(app)
      .patch(`/api/v1/lesson-plans/${planId}`)
      .set('Authorization', `Bearer ${guruToken}`)
      .send({ topic: 'Persamaan Linear & Kuadrat', durationMins: 120 });
    expect(res.status).toBe(200);
    expect(res.body.data.durationMins).toBe(120);
  });

  it('POST /api/v1/lesson-plans/:id/publish - should publish RPP (KEPALA_KURIKULUM)', async () => {
    if (!planId) return;
    const res = await request(app)
      .post(`/api/v1/lesson-plans/${planId}/publish`)
      .set('Authorization', `Bearer ${kurikulumToken}`);
    expect(res.status).toBe(200);
  });
});
