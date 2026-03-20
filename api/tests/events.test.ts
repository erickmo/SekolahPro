import request from 'supertest';
import { app } from '../src/index';
import { prisma } from '../src/lib/prisma';

const SCHOOL_ID = 'test-school-evt-001';
let token: string;
let eventId: string;
let taskId: string;

beforeAll(async () => {
  await prisma.school.upsert({
    where: { subdomain: 'evt-test' },
    create: {
      id: SCHOOL_ID,
      name: 'Events Test School',
      npsn: `EVT${Date.now()}`,
      subdomain: 'evt-test',
      address: 'Test Address',
      config: {},
    },
    update: {},
  });

  const email = `events_admin_${Date.now()}@school.test`;
  const regRes = await request(app)
    .post('/api/v1/auth/register')
    .send({ name: 'Admin Events', email, password: 'Password123!', role: 'ADMIN_SEKOLAH', schoolId: SCHOOL_ID });
  token = regRes.body.data.accessToken;
});

afterAll(async () => {
  await prisma.eventCommittee.deleteMany({ where: { event: { schoolId: SCHOOL_ID } } }).catch(() => null);
  await prisma.eventBudget.deleteMany({ where: { event: { schoolId: SCHOOL_ID } } }).catch(() => null);
  await prisma.eventTask.deleteMany({ where: { event: { schoolId: SCHOOL_ID } } }).catch(() => null);
  await prisma.schoolEvent.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.school.deleteMany({ where: { id: SCHOOL_ID } }).catch(() => null);
  await prisma.$disconnect();
});

describe('Events Module', () => {
  it('POST /api/v1/events - should create event', async () => {
    const res = await request(app)
      .post('/api/v1/events')
      .set('Authorization', `Bearer ${token}`)
      .send({
        title: 'Pentas Seni 2025',
        description: 'Acara tahunan',
        startDate: new Date(Date.now() + 30 * 86400000).toISOString(),
        endDate: new Date(Date.now() + 31 * 86400000).toISOString(),
        location: 'Aula Sekolah',
        type: 'SENI_BUDAYA',
      });
    expect(res.status).toBe(201);
    expect(res.body.data.title).toBe('Pentas Seni 2025');
    expect(res.body.data.type).toBe('SENI_BUDAYA');
    eventId = res.body.data.id;
  });

  it('GET /api/v1/events - should list events', async () => {
    const res = await request(app)
      .get('/api/v1/events')
      .set('Authorization', `Bearer ${token}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
  });

  it('GET /api/v1/events/:id - should get event by id', async () => {
    if (!eventId) return;
    const res = await request(app)
      .get(`/api/v1/events/${eventId}`)
      .set('Authorization', `Bearer ${token}`);
    expect(res.status).toBe(200);
    expect(res.body.data.id).toBe(eventId);
  });

  it('PATCH /api/v1/events/:id - should update event status', async () => {
    if (!eventId) return;
    const res = await request(app)
      .patch(`/api/v1/events/${eventId}`)
      .set('Authorization', `Bearer ${token}`)
      .send({ status: 'ACTIVE' });
    expect(res.status).toBe(200);
    expect(res.body.data.status).toBe('ACTIVE');
  });

  it('POST /api/v1/events/:id/tasks - should add task to event', async () => {
    if (!eventId) return;
    const res = await request(app)
      .post(`/api/v1/events/${eventId}/tasks`)
      .set('Authorization', `Bearer ${token}`)
      .send({
        title: 'Persiapan Panggung',
        dueDate: new Date(Date.now() + 14 * 86400000).toISOString(),
      });
    expect(res.status).toBe(201);
    expect(res.body.data.title).toBe('Persiapan Panggung');
    expect(res.body.data.eventId).toBe(eventId);
    taskId = res.body.data.id;
  });

  it('POST /api/v1/events/:id/budget - should add budget item to event', async () => {
    if (!eventId) return;
    const res = await request(app)
      .post(`/api/v1/events/${eventId}/budget`)
      .set('Authorization', `Bearer ${token}`)
      .send({
        category: 'DEKORASI',
        planned: 5000000,
      });
    expect(res.status).toBe(201);
    expect(res.body.data.category).toBe('DEKORASI');
    expect(Number(res.body.data.planned)).toBe(5000000);
  });

  it('GET /api/v1/events/:id - should verify tasks and budgets included', async () => {
    if (!eventId) return;
    const res = await request(app)
      .get(`/api/v1/events/${eventId}`)
      .set('Authorization', `Bearer ${token}`);
    expect(res.status).toBe(200);
    expect(res.body.data.tasks.length).toBeGreaterThan(0);
    expect(res.body.data.budgets.length).toBeGreaterThan(0);
  });
});
