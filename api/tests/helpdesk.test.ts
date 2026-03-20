import request from 'supertest';
import { app } from '../src/index';
import { prisma } from '../src/lib/prisma';

const SCHOOL_ID = 'test-school-hd-001';
let userToken: string;
let ticketId: string;

beforeAll(async () => {
  await prisma.school.upsert({
    where: { subdomain: 'hd-test' },
    create: {
      id: SCHOOL_ID,
      name: 'Helpdesk Test School',
      npsn: `HDS${Date.now()}`,
      subdomain: 'hd-test',
      address: 'Test Address',
      config: {},
    },
    update: {},
  });

  const email = `user_hd_${Date.now()}@school.test`;
  const regRes = await request(app)
    .post('/api/v1/auth/register')
    .send({ name: 'User HD', email, password: 'Password123!', role: 'GURU', schoolId: SCHOOL_ID });
  userToken = regRes.body.data.accessToken;
});

afterAll(async () => {
  await prisma.$disconnect();
});

describe('Helpdesk Tickets', () => {
  it('POST /api/v1/helpdesk - should create ticket', async () => {
    const res = await request(app)
      .post('/api/v1/helpdesk')
      .set('Authorization', `Bearer ${userToken}`)
      .send({
        title: 'AC Rusak di Ruang X-A',
        description: 'AC di ruang kelas X-A tidak berfungsi sejak pagi',
        category: 'FASILITAS',
        location: 'Ruang X-A',
        priority: 'MEDIUM',
      });
    expect(res.status).toBe(201);
    expect(res.body.data.title).toBe('AC Rusak di Ruang X-A');
    expect(res.body.data.status).toBe('OPEN');
    ticketId = res.body.data.id;
  });

  it('GET /api/v1/helpdesk - should list tickets', async () => {
    const res = await request(app)
      .get('/api/v1/helpdesk')
      .set('Authorization', `Bearer ${userToken}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
  });

  it('PUT /api/v1/helpdesk/:id - should update ticket status', async () => {
    if (!ticketId) return;
    const res = await request(app)
      .put(`/api/v1/helpdesk/${ticketId}`)
      .set('Authorization', `Bearer ${userToken}`)
      .send({ status: 'IN_PROGRESS', assignedTo: 'teknisi-001' });
    expect(res.status).toBe(200);
  });

  it('GET /api/v1/helpdesk/stats - should get stats', async () => {
    const res = await request(app)
      .get('/api/v1/helpdesk/stats')
      .set('Authorization', `Bearer ${userToken}`);
    expect(res.status).toBe(200);
    expect(res.body.data).toHaveProperty('open');
    expect(res.body.data).toHaveProperty('total');
  });
});
