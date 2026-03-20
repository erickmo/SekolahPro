import request from 'supertest';
import { app } from '../src/index';
import { prisma } from '../src/lib/prisma';

const SCHOOL_ID = 'test-school-cm-001';
let kepalaToken: string;
let boardId: string;
let meetingId: string;

beforeAll(async () => {
  await prisma.school.upsert({
    where: { subdomain: 'cm-test' },
    create: {
      id: SCHOOL_ID,
      name: 'Committee Test School',
      npsn: `CM${Date.now()}`,
      subdomain: 'cm-test',
      address: 'Test Address',
      config: {},
    },
    update: {},
  });

  const kepalaEmail = `kepala_${Date.now()}@school.test`;
  const kepalaRes = await request(app)
    .post('/api/v1/auth/register')
    .send({ name: 'Kepala Sekolah', email: kepalaEmail, password: 'Password123!', role: 'KEPALA_SEKOLAH', schoolId: SCHOOL_ID });
  kepalaToken = kepalaRes.body.data.accessToken;
});

afterAll(async () => {
  await prisma.committeeDecision.deleteMany({ where: { meeting: { schoolId: SCHOOL_ID } } }).catch(() => null);
  await prisma.committeeMeeting.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.committeeMember.deleteMany({ where: { board: { schoolId: SCHOOL_ID } } }).catch(() => null);
  await prisma.committeeBoard.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.user.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.school.deleteMany({ where: { id: SCHOOL_ID } }).catch(() => null);
  await prisma.$disconnect();
});

describe('Committee Boards (M53 Komite Sekolah)', () => {
  it('POST /api/v1/committee/boards - should create board', async () => {
    const res = await request(app)
      .post('/api/v1/committee/boards')
      .set('Authorization', `Bearer ${kepalaToken}`)
      .send({
        periodStart: '2025-07-01',
        periodEnd: '2026-06-30',
      });
    expect(res.status).toBe(201);
    expect(res.body.data.periodStart).toBeDefined();
    boardId = res.body.data.id;
  });

  it('GET /api/v1/committee/boards - should list boards', async () => {
    const res = await request(app)
      .get('/api/v1/committee/boards')
      .set('Authorization', `Bearer ${kepalaToken}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
  });

  it('POST /api/v1/committee/boards/:id/members - should add member', async () => {
    if (!boardId) return;
    const res = await request(app)
      .post(`/api/v1/committee/boards/${boardId}/members`)
      .set('Authorization', `Bearer ${kepalaToken}`)
      .send({
        name: 'Bapak Budi',
        role: 'Ketua',
        phone: '08123456789',
      });
    expect(res.status).toBe(201);
    expect(res.body.data.name).toBe('Bapak Budi');
  });

  it('GET /api/v1/committee/boards/:id - should get board with members', async () => {
    if (!boardId) return;
    const res = await request(app)
      .get(`/api/v1/committee/boards/${boardId}`)
      .set('Authorization', `Bearer ${kepalaToken}`);
    expect(res.status).toBe(200);
    expect(res.body.data.id).toBe(boardId);
  });
});

describe('Committee Meetings & Decisions', () => {
  it('POST /api/v1/committee/meetings - should create meeting', async () => {
    if (!boardId) return;
    const res = await request(app)
      .post('/api/v1/committee/meetings')
      .set('Authorization', `Bearer ${kepalaToken}`)
      .send({
        boardId,
        scheduledAt: new Date(Date.now() + 7 * 86400000).toISOString(),
        agenda: ['Pembahasan RKAS', 'Evaluasi Program'],
      });
    expect(res.status).toBe(201);
    expect(Array.isArray(res.body.data.agenda)).toBe(true);
    meetingId = res.body.data.id;
  });

  it('GET /api/v1/committee/meetings - should list meetings', async () => {
    const res = await request(app)
      .get('/api/v1/committee/meetings')
      .set('Authorization', `Bearer ${kepalaToken}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
  });

  it('POST /api/v1/committee/meetings/:id/decisions - should add decision', async () => {
    if (!meetingId || !boardId) return;
    const res = await request(app)
      .post(`/api/v1/committee/meetings/${meetingId}/decisions`)
      .set('Authorization', `Bearer ${kepalaToken}`)
      .send({
        boardId,
        title: 'Persetujuan RKAS',
        description: 'RKAS 2025/2026 disetujui',
        votesFor: 10,
        votesAgainst: 0,
        votesAbstain: 0,
        result: 'APPROVED',
      });
    expect(res.status).toBe(201);
    expect(res.body.data.result).toBe('APPROVED');
  });

  it('GET /api/v1/committee/meetings/:id/decisions - should list decisions', async () => {
    if (!meetingId) return;
    const res = await request(app)
      .get(`/api/v1/committee/meetings/${meetingId}/decisions`)
      .set('Authorization', `Bearer ${kepalaToken}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
  });
});
