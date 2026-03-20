import request from 'supertest';
import { app } from '../src/index';
import { prisma } from '../src/lib/prisma';
import bcrypt from 'bcryptjs';

const TEST_SCHOOL_ID = 'test-school-auth-001';
const TEST_USER_EMAIL = `auth_test_${Date.now()}@school.test`;

beforeAll(async () => {
  // Ensure test school exists
  await prisma.school.upsert({
    where: { subdomain: 'auth-test' },
    create: {
      id: TEST_SCHOOL_ID,
      name: 'Auth Test School',
      npsn: `AUTH${Date.now()}`,
      subdomain: 'auth-test',
      address: 'Test Address',
      config: {},
    },
    update: {},
  });
});

afterAll(async () => {
  await prisma.user.deleteMany({ where: { email: TEST_USER_EMAIL } });
  await prisma.$disconnect();
});

describe('POST /api/v1/auth/register', () => {
  it('should register a new user', async () => {
    const res = await request(app)
      .post('/api/v1/auth/register')
      .send({
        name: 'Test Admin',
        email: TEST_USER_EMAIL,
        password: 'Password123!',
        role: 'ADMIN_SEKOLAH',
        schoolId: TEST_SCHOOL_ID,
      });
    expect(res.status).toBe(201);
    expect(res.body.success).toBe(true);
    expect(res.body.data).toHaveProperty('accessToken');
    expect(res.body.data).toHaveProperty('refreshToken');
    expect(res.body.data.user.email).toBe(TEST_USER_EMAIL);
  });

  it('should reject duplicate email', async () => {
    const res = await request(app)
      .post('/api/v1/auth/register')
      .send({
        name: 'Another Admin',
        email: TEST_USER_EMAIL,
        password: 'Password123!',
        role: 'ADMIN_SEKOLAH',
        schoolId: TEST_SCHOOL_ID,
      });
    expect(res.status).toBe(409);
    expect(res.body.success).toBe(false);
  });
});

describe('POST /api/v1/auth/login', () => {
  it('should login with correct credentials', async () => {
    const res = await request(app)
      .post('/api/v1/auth/login')
      .send({ email: TEST_USER_EMAIL, password: 'Password123!' });
    expect(res.status).toBe(200);
    expect(res.body.success).toBe(true);
    expect(res.body.data).toHaveProperty('accessToken');
  });

  it('should reject wrong password', async () => {
    const res = await request(app)
      .post('/api/v1/auth/login')
      .send({ email: TEST_USER_EMAIL, password: 'WrongPassword' });
    expect(res.status).toBe(401);
    expect(res.body.success).toBe(false);
  });

  it('should reject unknown email', async () => {
    const res = await request(app)
      .post('/api/v1/auth/login')
      .send({ email: 'nobody@nowhere.com', password: 'Password123!' });
    expect(res.status).toBe(401);
  });
});

describe('GET /api/v1/auth/me', () => {
  let token: string;

  beforeAll(async () => {
    const res = await request(app)
      .post('/api/v1/auth/login')
      .send({ email: TEST_USER_EMAIL, password: 'Password123!' });
    token = res.body.data.accessToken;
  });

  it('should return current user when authenticated', async () => {
    const res = await request(app)
      .get('/api/v1/auth/me')
      .set('Authorization', `Bearer ${token}`);
    expect(res.status).toBe(200);
    expect(res.body.data.email).toBe(TEST_USER_EMAIL);
  });

  it('should reject unauthenticated requests', async () => {
    const res = await request(app).get('/api/v1/auth/me');
    expect(res.status).toBe(401);
  });
});
