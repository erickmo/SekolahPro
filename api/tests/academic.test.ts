import request from 'supertest';
import { app } from '../src/index';
import { prisma } from '../src/lib/prisma';

const SCHOOL_ID = 'test-school-academic-001';
let adminToken: string;
let academicYearId: string;
let subjectId: string;
let classId: string;
let teacherId: string;

beforeAll(async () => {
  await prisma.school.upsert({
    where: { subdomain: 'academic-test' },
    create: {
      id: SCHOOL_ID,
      name: 'Academic Test School',
      npsn: `ACA${Date.now()}`,
      subdomain: 'academic-test',
      address: 'Test Address',
      config: {},
    },
    update: {},
  });

  // Clean up existing data from previous runs
  await prisma.schoolClass.deleteMany({ where: { academicYear: { schoolId: SCHOOL_ID } } }).catch(() => null);
  await prisma.academicYear.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.subject.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.teacher.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);

  const email = `admin_aca_${Date.now()}@school.test`;
  const regRes = await request(app)
    .post('/api/v1/auth/register')
    .send({ name: 'Admin Academic', email, password: 'Password123!', role: 'ADMIN_SEKOLAH', schoolId: SCHOOL_ID });
  adminToken = regRes.body.data.accessToken;
});

afterAll(async () => {
  if (classId) await prisma.schoolClass.deleteMany({ where: { academicYearId } }).catch(() => null);
  if (academicYearId) await prisma.academicYear.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  if (subjectId) await prisma.subject.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  if (teacherId) await prisma.teacher.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.$disconnect();
});

describe('Academic Years', () => {
  it('POST /api/v1/academic/academic-years - should create academic year', async () => {
    const res = await request(app)
      .post('/api/v1/academic/academic-years')
      .set('Authorization', `Bearer ${adminToken}`)
      .send({
        name: '2024/2025',
        startDate: '2024-07-15',
        endDate: '2025-06-30',
        isCurrent: true,
      });
    expect(res.status).toBe(201);
    expect(res.body.data.name).toBe('2024/2025');
    academicYearId = res.body.data.id;
  });

  it('GET /api/v1/academic/academic-years - should list academic years', async () => {
    const res = await request(app)
      .get('/api/v1/academic/academic-years')
      .set('Authorization', `Bearer ${adminToken}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
  });
});

describe('Subjects', () => {
  it('POST /api/v1/academic/subjects - should create subject', async () => {
    const res = await request(app)
      .post('/api/v1/academic/subjects')
      .set('Authorization', `Bearer ${adminToken}`)
      .send({
        name: 'Matematika',
        code: `MTK${Date.now()}`,
        gradeLevel: 'X',
        curriculum: 'MERDEKA',
        weeklyHours: 4,
      });
    expect(res.status).toBe(201);
    expect(res.body.data.name).toBe('Matematika');
    subjectId = res.body.data.id;
  });

  it('GET /api/v1/academic/subjects - should list subjects', async () => {
    const res = await request(app)
      .get('/api/v1/academic/subjects')
      .set('Authorization', `Bearer ${adminToken}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
  });
});

describe('Teachers', () => {
  it('POST /api/v1/academic/teachers - should create teacher', async () => {
    const res = await request(app)
      .post('/api/v1/academic/teachers')
      .set('Authorization', `Bearer ${adminToken}`)
      .send({
        name: 'Budi Santoso',
        email: `budi_${Date.now()}@school.test`,
        phone: '081234567890',
      });
    expect(res.status).toBe(201);
    expect(res.body.data.name).toBe('Budi Santoso');
    teacherId = res.body.data.id;
  });

  it('GET /api/v1/academic/teachers - should list teachers with pagination', async () => {
    const res = await request(app)
      .get('/api/v1/academic/teachers?page=1&limit=10')
      .set('Authorization', `Bearer ${adminToken}`);
    expect(res.status).toBe(200);
    expect(res.body.meta).toHaveProperty('total');
  });
});

describe('Classes', () => {
  it('POST /api/v1/academic/classes - should create class', async () => {
    if (!academicYearId) return;
    const res = await request(app)
      .post('/api/v1/academic/classes')
      .set('Authorization', `Bearer ${adminToken}`)
      .send({
        name: 'X-A',
        gradeLevel: 'X',
        academicYearId,
        capacity: 32,
      });
    expect(res.status).toBe(201);
    expect(res.body.data.name).toBe('X-A');
    classId = res.body.data.id;
  });

  it('GET /api/v1/academic/classes - should list classes', async () => {
    const res = await request(app)
      .get('/api/v1/academic/classes')
      .set('Authorization', `Bearer ${adminToken}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
  });
});
