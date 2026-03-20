import request from 'supertest';
import { app } from '../src/index';
import { prisma } from '../src/lib/prisma';

const SCHOOL_ID = 'test-school-notif-001';
let adminToken: string;
let userId: string;

beforeAll(async () => {
  await prisma.school.upsert({
    where: { subdomain: 'notif-test' },
    create: {
      id: SCHOOL_ID,
      name: 'Notif Test School',
      npsn: `NOT${Date.now()}`,
      subdomain: 'notif-test',
      address: 'Test Address',
      config: {},
    },
    update: {},
  });

  const email = `admin_notif_${Date.now()}@school.test`;
  const regRes = await request(app)
    .post('/api/v1/auth/register')
    .send({ name: 'Admin Notif', email, password: 'Password123!', role: 'ADMIN_SEKOLAH', schoolId: SCHOOL_ID });
  adminToken = regRes.body.data.accessToken;
  userId = regRes.body.data.user.id;
});

afterAll(async () => {
  await prisma.$disconnect();
});

describe('Notifications', () => {
  it('POST /api/v1/notifications/send - should create notification', async () => {
    const res = await request(app)
      .post('/api/v1/notifications/send')
      .set('Authorization', `Bearer ${adminToken}`)
      .send({
        recipientId: userId,
        type: 'ANNOUNCEMENT',
        channel: 'EMAIL',
        title: 'Test Notification',
        body: 'This is a test notification',
      });
    expect(res.status).toBe(201);
    expect(res.body.data.title).toBe('Test Notification');
  });

  it('GET /api/v1/notifications - should list notifications', async () => {
    const res = await request(app)
      .get('/api/v1/notifications')
      .set('Authorization', `Bearer ${adminToken}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
  });
});

describe('Notification Templates', () => {
  it('POST /api/v1/notifications/templates - should create template', async () => {
    const res = await request(app)
      .post('/api/v1/notifications/templates')
      .set('Authorization', `Bearer ${adminToken}`)
      .send({
        name: `Absensi Template ${Date.now()}`,
        type: 'ATTENDANCE',
        channel: 'WHATSAPP',
        titleTemplate: 'Absensi {{nama_siswa}}',
        bodyTemplate: 'Siswa {{nama_siswa}} tidak hadir pada tanggal {{tanggal}}',
      });
    expect(res.status).toBe(201);
    expect(res.body.data.name).toContain('Absensi Template');
  });

  it('GET /api/v1/notifications/templates - should list templates', async () => {
    const res = await request(app)
      .get('/api/v1/notifications/templates')
      .set('Authorization', `Bearer ${adminToken}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
  });
});
