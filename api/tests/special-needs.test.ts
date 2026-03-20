import request from 'supertest';
import { app } from '../src/index';
import { prisma } from '../src/lib/prisma';

const SCHOOL_ID = 'test-school-abk-001';
let gpkToken: string;
let adminToken: string;
let studentId: string;
let profileId: string;
let iepId: string;

beforeAll(async () => {
  await prisma.school.upsert({
    where: { subdomain: 'abk-test' },
    create: {
      id: SCHOOL_ID,
      name: 'ABK Test School',
      npsn: `ABK${Date.now()}`,
      subdomain: 'abk-test',
      address: 'Test Address',
      config: {},
    },
    update: {},
  });

  const gpkEmail = `gpk_${Date.now()}@school.test`;
  const gpkRes = await request(app)
    .post('/api/v1/auth/register')
    .send({ name: 'Guru Pendamping Khusus', email: gpkEmail, password: 'Password123!', role: 'GPK', schoolId: SCHOOL_ID });
  gpkToken = gpkRes.body.data.accessToken;

  const adminEmail = `abk_admin_${Date.now()}@school.test`;
  const adminRes = await request(app)
    .post('/api/v1/auth/register')
    .send({ name: 'Admin ABK', email: adminEmail, password: 'Password123!', role: 'ADMIN_SEKOLAH', schoolId: SCHOOL_ID });
  adminToken = adminRes.body.data.accessToken;

  const student = await prisma.student.create({
    data: {
      nisn: `ABK${Date.now()}`.slice(-10),
      name: 'Siswa ABK',
      birthDate: new Date('2010-04-15'),
      gender: 'MALE',
      birthPlace: 'Jakarta',
      religion: 'ISLAM',
      address: 'Test Address',
      schoolId: SCHOOL_ID,
    },
  });
  studentId = student.id;
});

afterAll(async () => {
  await prisma.aBKProgressReport.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.iEP.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.specialNeedsProfile.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.student.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.school.deleteMany({ where: { id: SCHOOL_ID } }).catch(() => null);
  await prisma.$disconnect();
});

describe('Special Needs (ABK) Module', () => {
  it('POST /api/v1/special-needs/profiles - should create ABK profile', async () => {
    const res = await request(app)
      .post('/api/v1/special-needs/profiles')
      .set('Authorization', `Bearer ${gpkToken}`)
      .send({
        studentId,
        encryptedData: 'encrypted-medical-data',
        requiresExtraTime: true,
        requiresAssistant: false,
        requiresLargeFont: false,
      });
    expect(res.status).toBe(201);
    expect(res.body.data.studentId).toBe(studentId);
    expect(res.body.data.requiresExtraTime).toBe(true);
    profileId = res.body.data.id;
  });

  it('GET /api/v1/special-needs/profiles - should list profiles', async () => {
    const res = await request(app)
      .get('/api/v1/special-needs/profiles')
      .set('Authorization', `Bearer ${adminToken}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
  });

  it('GET /api/v1/special-needs/profiles/:studentId - should get profile by studentId', async () => {
    const res = await request(app)
      .get(`/api/v1/special-needs/profiles/${studentId}`)
      .set('Authorization', `Bearer ${gpkToken}`);
    expect(res.status).toBe(200);
    expect(res.body.data.studentId).toBe(studentId);
  });

  it('POST /api/v1/special-needs/iep - should create IEP', async () => {
    if (!profileId) return;
    const res = await request(app)
      .post('/api/v1/special-needs/iep')
      .set('Authorization', `Bearer ${gpkToken}`)
      .send({
        profileId,
        academicYear: '2025/2026',
        semester: 1,
        goals: { reading: 'Mampu membaca dengan lancar' },
      });
    expect(res.status).toBe(201);
    expect(res.body.data.profileId).toBe(profileId);
    expect(res.body.data.academicYear).toBe('2025/2026');
    iepId = res.body.data.id;
  });

  it('GET /api/v1/special-needs/iep/:studentId - should get IEPs by studentId', async () => {
    const res = await request(app)
      .get(`/api/v1/special-needs/iep/${studentId}`)
      .set('Authorization', `Bearer ${gpkToken}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
  });

  it('POST /api/v1/special-needs/progress - should add progress report', async () => {
    if (!profileId) return;
    const res = await request(app)
      .post('/api/v1/special-needs/progress')
      .set('Authorization', `Bearer ${gpkToken}`)
      .send({
        profileId,
        period: 'Semester 1 2025',
        content: { progress: 'Baik', notes: 'Perlu perhatian lebih' },
      });
    expect(res.status).toBe(201);
    expect(res.body.data.profileId).toBe(profileId);
    expect(res.body.data.period).toBe('Semester 1 2025');
  });

  it('GET /api/v1/special-needs/progress/:studentId - should get progress reports', async () => {
    const res = await request(app)
      .get(`/api/v1/special-needs/progress/${studentId}`)
      .set('Authorization', `Bearer ${gpkToken}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
  });
});
