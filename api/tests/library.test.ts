import request from 'supertest';
import { app } from '../src/index';
import { prisma } from '../src/lib/prisma';

const SCHOOL_ID = 'test-school-lib-001';
let libToken: string;
let studentId: string;
let bookId: string;

beforeAll(async () => {
  await prisma.school.upsert({
    where: { subdomain: 'lib-test' },
    create: {
      id: SCHOOL_ID,
      name: 'Library Test School',
      npsn: `LIB${Date.now()}`,
      subdomain: 'lib-test',
      address: 'Test Address',
      config: {},
    },
    update: {},
  });

  const email = `pustakawan_${Date.now()}@school.test`;
  const regRes = await request(app)
    .post('/api/v1/auth/register')
    .send({ name: 'Pustakawan', email, password: 'Password123!', role: 'PUSTAKAWAN', schoolId: SCHOOL_ID });
  libToken = regRes.body.data.accessToken;

  const student = await prisma.student.create({
    data: {
      nisn: `LIB${Date.now()}`.slice(-10),
      name: 'Library Student',
      birthDate: new Date('2008-04-12'),
      gender: 'FEMALE',
      birthPlace: 'Denpasar',
      religion: 'HINDU',
      address: 'Test',
      schoolId: SCHOOL_ID,
    },
  });
  studentId = student.id;
});

afterAll(async () => {
  await prisma.libraryLoan.deleteMany({ where: { schoolId: SCHOOL_ID } });
  await prisma.libraryBook.deleteMany({ where: { schoolId: SCHOOL_ID } });
  await prisma.student.deleteMany({ where: { schoolId: SCHOOL_ID } });
  await prisma.$disconnect();
});

describe('Library Books', () => {
  it('POST /api/v1/library/books - should add book', async () => {
    const res = await request(app)
      .post('/api/v1/library/books')
      .set('Authorization', `Bearer ${libToken}`)
      .send({
        title: 'Pemrograman Web Modern',
        author: 'Budi Rahardjo',
        isbn: `ISBN${Date.now()}`,
        publisher: 'Informatika',
        publishYear: 2022,
        category: 'TEKNOLOGI',
        stock: 5,
      });
    expect(res.status).toBe(201);
    expect(res.body.data.title).toBe('Pemrograman Web Modern');
    bookId = res.body.data.id;
  });

  it('GET /api/v1/library/books - should list books', async () => {
    const res = await request(app)
      .get('/api/v1/library/books')
      .set('Authorization', `Bearer ${libToken}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
  });
});

describe('Library Loans', () => {
  it('POST /api/v1/library/loans - should borrow book', async () => {
    if (!bookId) return;
    const res = await request(app)
      .post('/api/v1/library/loans')
      .set('Authorization', `Bearer ${libToken}`)
      .send({
        borrowerId: studentId,
        bookId,
        dueDate: new Date(Date.now() + 14 * 86400000).toISOString(),
      });
    expect(res.status).toBe(201);
    expect(res.body.data.borrowerId).toBe(studentId);
    expect(res.body.data.status).toBe('BORROWED');
  });

  it('GET /api/v1/library/loans - should list loans', async () => {
    const res = await request(app)
      .get('/api/v1/library/loans')
      .set('Authorization', `Bearer ${libToken}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
    expect(res.body.data.length).toBeGreaterThan(0);
  });
});
