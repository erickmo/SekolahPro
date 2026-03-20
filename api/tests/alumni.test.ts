import request from 'supertest';
import { app } from '../src/index';
import { prisma } from '../src/lib/prisma';

const SCHOOL_ID = 'test-school-alumni-001';
let adminToken: string;
let userId: string;
let studentId: string;

beforeAll(async () => {
  await prisma.school.upsert({
    where: { subdomain: 'alumni-test' },
    create: {
      id: SCHOOL_ID,
      name: 'Alumni Test School',
      npsn: `ALM${Date.now()}`,
      subdomain: 'alumni-test',
      address: 'Test Address',
      config: {},
    },
    update: {},
  });

  const email = `alumni_admin_${Date.now()}@school.test`;
  const regRes = await request(app)
    .post('/api/v1/auth/register')
    .send({ name: 'Admin Alumni', email, password: 'Password123!', role: 'ALUMNI', schoolId: SCHOOL_ID });
  adminToken = regRes.body.data.accessToken;
  userId = regRes.body.data.user.id;

  const student = await prisma.student.create({
    data: {
      nisn: `ALM${Date.now()}`.slice(-10),
      name: 'Alumni Student',
      birthDate: new Date('2003-05-10'),
      gender: 'FEMALE',
      birthPlace: 'Semarang',
      religion: 'KRISTEN',
      address: 'Test',
      schoolId: SCHOOL_ID,
    },
  });
  studentId = student.id;
});

afterAll(async () => {
  await prisma.portfolioItem.deleteMany({ where: { portfolio: { schoolId: SCHOOL_ID } } });
  await prisma.portfolio.deleteMany({ where: { schoolId: SCHOOL_ID } });
  await prisma.alumniProfile.deleteMany({ where: { schoolId: SCHOOL_ID } });
  await prisma.student.deleteMany({ where: { schoolId: SCHOOL_ID } });
  await prisma.$disconnect();
});

describe('Alumni Module', () => {
  it('POST /api/v1/alumni/profile - should create alumni profile', async () => {
    const res = await request(app)
      .post('/api/v1/alumni/profile')
      .set('Authorization', `Bearer ${adminToken}`)
      .send({
        graduationYear: 2022,
        major: 'Teknik Informatika',
        university: 'Universitas Gadjah Mada',
        occupation: 'Software Engineer',
        company: 'PT Teknologi Maju',
        linkedinUrl: 'https://linkedin.com/in/test',
      });
    expect(res.status).toBe(201);
    expect(res.body.data.graduationYear).toBe(2022);
  });

  it('GET /api/v1/alumni - should list alumni', async () => {
    const res = await request(app)
      .get('/api/v1/alumni')
      .set('Authorization', `Bearer ${adminToken}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
    expect(res.body.meta).toHaveProperty('total');
  });

  it('POST /api/v1/alumni/portfolio - should create portfolio', async () => {
    const res = await request(app)
      .post('/api/v1/alumni/portfolio')
      .set('Authorization', `Bearer ${adminToken}`)
      .send({
        studentId,
        title: 'Portofolio Alumni Test',
        description: 'Kumpulan karya selama di sekolah',
        visibility: 'PUBLIC',
      });
    expect(res.status).toBe(201);
    expect(res.body.data.title).toBe('Portofolio Alumni Test');
  });

  it('GET /api/v1/alumni/portfolio/:studentId - should get portfolio', async () => {
    const res = await request(app)
      .get(`/api/v1/alumni/portfolio/${studentId}`)
      .set('Authorization', `Bearer ${adminToken}`);
    expect(res.status).toBe(200);
    expect(res.body.data.title).toBe('Portofolio Alumni Test');
  });
});
