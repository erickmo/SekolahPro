import request from 'supertest';
import { app } from '../src/index';
import { prisma } from '../src/lib/prisma';

const SCHOOL_ID = 'test-school-vol-001';
let token: string;
let profileId: string;
let activityId: string;

beforeAll(async () => {
  await prisma.school.upsert({
    where: { subdomain: 'vol-test' },
    create: {
      id: SCHOOL_ID,
      name: 'Volunteer Test School',
      npsn: `VOL${Date.now()}`,
      subdomain: 'vol-test',
      address: 'Test Address',
      config: {},
    },
    update: {},
  });

  const email = `volunteer_admin_${Date.now()}@school.test`;
  const regRes = await request(app)
    .post('/api/v1/auth/register')
    .send({ name: 'Admin Volunteer', email, password: 'Password123!', role: 'ADMIN_SEKOLAH', schoolId: SCHOOL_ID });
  token = regRes.body.data.accessToken;
});

afterAll(async () => {
  await prisma.volunteerCertificate.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.volunteerActivity.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.volunteerProfile.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.school.deleteMany({ where: { id: SCHOOL_ID } }).catch(() => null);
  await prisma.$disconnect();
});

describe('Volunteer Module', () => {
  it('POST /api/v1/volunteer/profiles - should create volunteer profile', async () => {
    const res = await request(app)
      .post('/api/v1/volunteer/profiles')
      .set('Authorization', `Bearer ${token}`)
      .send({
        name: 'Ibu Sari',
        relation: 'PARENT',
        phone: '08111222333',
        skills: ['Memasak', 'Pertolongan Pertama'],
      });
    expect(res.status).toBe(201);
    expect(res.body.data.name).toBe('Ibu Sari');
    expect(res.body.data.relation).toBe('PARENT');
    profileId = res.body.data.id;
  });

  it('GET /api/v1/volunteer/profiles - should list profiles', async () => {
    const res = await request(app)
      .get('/api/v1/volunteer/profiles')
      .set('Authorization', `Bearer ${token}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
  });

  it('GET /api/v1/volunteer/profiles/:id - should get profile by id', async () => {
    if (!profileId) return;
    const res = await request(app)
      .get(`/api/v1/volunteer/profiles/${profileId}`)
      .set('Authorization', `Bearer ${token}`);
    expect(res.status).toBe(200);
    expect(res.body.data.id).toBe(profileId);
  });

  it('POST /api/v1/volunteer/activities - should create activity', async () => {
    if (!profileId) return;
    const res = await request(app)
      .post('/api/v1/volunteer/activities')
      .set('Authorization', `Bearer ${token}`)
      .send({
        volunteerId: profileId,
        title: 'Gotong Royong Sekolah',
        date: new Date().toISOString(),
        hours: 4,
      });
    expect(res.status).toBe(201);
    expect(res.body.data.title).toBe('Gotong Royong Sekolah');
    expect(res.body.data.hours).toBe(4);
    activityId = res.body.data.id;
  });

  it('GET /api/v1/volunteer/activities - should list activities', async () => {
    const res = await request(app)
      .get('/api/v1/volunteer/activities')
      .set('Authorization', `Bearer ${token}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
  });

  it('PATCH /api/v1/volunteer/activities/:id - should update activity notes', async () => {
    if (!activityId) return;
    const res = await request(app)
      .patch(`/api/v1/volunteer/activities/${activityId}`)
      .set('Authorization', `Bearer ${token}`)
      .send({ notes: 'Kegiatan berjalan lancar' });
    expect(res.status).toBe(200);
    expect(res.body.data.notes).toBe('Kegiatan berjalan lancar');
  });

  it('POST /api/v1/volunteer/certificates - should issue certificate', async () => {
    if (!profileId) return;
    const res = await request(app)
      .post('/api/v1/volunteer/certificates')
      .set('Authorization', `Bearer ${token}`)
      .send({
        volunteerId: profileId,
        title: 'Sertifikat Relawan 2025',
        issuedAt: new Date().toISOString(),
      });
    expect(res.status).toBe(201);
    expect(res.body.data.title).toBe('Sertifikat Relawan 2025');
    expect(res.body.data.volunteerId).toBe(profileId);
  });
});
