import request from 'supertest';
import { app } from '../src/index';
import { prisma } from '../src/lib/prisma';

const SCHOOL_ID = 'test-school-students-001';
let adminToken: string;
let createdStudentId: string;

beforeAll(async () => {
  await prisma.school.upsert({
    where: { subdomain: 'students-test' },
    create: {
      id: SCHOOL_ID,
      name: 'Students Test School',
      npsn: `STD${Date.now()}`,
      subdomain: 'students-test',
      address: 'Test Address',
      config: {},
    },
    update: {},
  });

  const email = `admin_std_${Date.now()}@school.test`;
  const regRes = await request(app)
    .post('/api/v1/auth/register')
    .send({ name: 'Admin', email, password: 'Password123!', role: 'ADMIN_SEKOLAH', schoolId: SCHOOL_ID });
  adminToken = regRes.body.data.accessToken;
});

afterAll(async () => {
  if (createdStudentId) {
    await prisma.student.deleteMany({ where: { id: createdStudentId } });
  }
  await prisma.$disconnect();
});

describe('POST /api/v1/students', () => {
  it('should create a student', async () => {
    const res = await request(app)
      .post('/api/v1/students')
      .set('Authorization', `Bearer ${adminToken}`)
      .send({
        nisn: `${Date.now()}`.slice(-10),
        name: 'Test Student',
        birthDate: '2008-01-15',
        gender: 'MALE',
        birthPlace: 'Jakarta',
        religion: 'ISLAM',
        address: 'Test Address',
      });
    expect(res.status).toBe(201);
    expect(res.body.success).toBe(true);
    expect(res.body.data.name).toBe('Test Student');
    createdStudentId = res.body.data.id;
  });

  it('should reject duplicate NISN', async () => {
    const nisn = `${Date.now()}`.slice(-10);
    await request(app)
      .post('/api/v1/students')
      .set('Authorization', `Bearer ${adminToken}`)
      .send({ nisn, name: 'Student 1', birthDate: '2008-01-15', gender: 'MALE', birthPlace: 'Jakarta', religion: 'ISLAM', address: 'Addr' });

    const res = await request(app)
      .post('/api/v1/students')
      .set('Authorization', `Bearer ${adminToken}`)
      .send({ nisn, name: 'Student 2', birthDate: '2008-01-15', gender: 'MALE', birthPlace: 'Jakarta', religion: 'ISLAM', address: 'Addr' });
    expect(res.status).toBe(409);
  });
});

describe('GET /api/v1/students', () => {
  it('should return paginated students', async () => {
    const res = await request(app)
      .get('/api/v1/students')
      .set('Authorization', `Bearer ${adminToken}`);
    expect(res.status).toBe(200);
    expect(res.body.success).toBe(true);
    expect(Array.isArray(res.body.data)).toBe(true);
    expect(res.body.meta).toHaveProperty('total');
  });

  it('should filter by search query', async () => {
    const res = await request(app)
      .get('/api/v1/students?search=Test')
      .set('Authorization', `Bearer ${adminToken}`);
    expect(res.status).toBe(200);
    expect(res.body.data.length).toBeGreaterThan(0);
  });
});

describe('GET /api/v1/students/:id', () => {
  it('should return student by id', async () => {
    if (!createdStudentId) return;
    const res = await request(app)
      .get(`/api/v1/students/${createdStudentId}`)
      .set('Authorization', `Bearer ${adminToken}`);
    expect(res.status).toBe(200);
    expect(res.body.data.id).toBe(createdStudentId);
  });

  it('should return 404 for non-existent student', async () => {
    const res = await request(app)
      .get('/api/v1/students/non-existent-id')
      .set('Authorization', `Bearer ${adminToken}`);
    expect(res.status).toBe(404);
  });
});
