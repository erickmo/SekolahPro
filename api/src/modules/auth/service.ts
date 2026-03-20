import bcrypt from 'bcryptjs';
import jwt from 'jsonwebtoken';
import { prisma } from '../../lib/prisma';
import { redis } from '../../lib/redis';
import { config } from '../../config';
import { AuthPayload } from '../../shared/types';
import { UnauthorizedError, ConflictError, NotFoundError, BadRequestError } from '../../shared/errors';
import { LoginInput, RegisterInput } from './dto';
import { Role } from '@prisma/client';

function generateTokens(payload: AuthPayload) {
  const accessToken = jwt.sign(payload as object, config.jwt.secret, { expiresIn: config.jwt.expiresIn as never });
  const refreshToken = jwt.sign(payload as object, config.jwt.refreshSecret, { expiresIn: config.jwt.refreshExpiresIn as never });
  return { accessToken, refreshToken };
}

export async function login(input: LoginInput) {
  const user = await prisma.user.findUnique({
    where: { email: input.email },
    include: { school: { select: { id: true, name: true, subdomain: true } } },
  });
  if (!user || !user.isActive) throw new UnauthorizedError('Email atau password salah');

  const match = await bcrypt.compare(input.password, user.password);
  if (!match) throw new UnauthorizedError('Email atau password salah');

  const payload: AuthPayload = {
    userId: user.id,
    schoolId: user.schoolId || '',
    role: user.role,
    email: user.email,
  };

  const tokens = generateTokens(payload);
  await redis.setex(`refresh:${user.id}`, 7 * 24 * 3600, tokens.refreshToken);
  await prisma.user.update({ where: { id: user.id }, data: { lastLoginAt: new Date() } });

  return { user: { id: user.id, name: user.name, email: user.email, role: user.role, school: user.school }, ...tokens };
}

export async function register(input: RegisterInput) {
  const exists = await prisma.user.findUnique({ where: { email: input.email } });
  if (exists) throw new ConflictError('Email sudah terdaftar');

  const hashed = await bcrypt.hash(input.password, config.bcrypt.rounds);
  const user = await prisma.user.create({
    data: {
      email: input.email,
      password: hashed,
      name: input.name,
      role: (input.role as Role) || Role.OPERATOR_SIMS,
      schoolId: input.schoolId || null,
    },
  });

  const payload: AuthPayload = { userId: user.id, schoolId: user.schoolId || '', role: user.role, email: user.email };
  const tokens = generateTokens(payload);
  return { user: { id: user.id, name: user.name, email: user.email, role: user.role }, ...tokens };
}

export async function refreshToken(token: string) {
  let payload: AuthPayload;
  try {
    payload = jwt.verify(token, config.jwt.refreshSecret) as AuthPayload;
  } catch {
    throw new UnauthorizedError('Refresh token tidak valid');
  }

  const stored = await redis.get(`refresh:${payload.userId}`);
  if (stored !== token) throw new UnauthorizedError('Refresh token sudah tidak valid');

  const user = await prisma.user.findUnique({ where: { id: payload.userId } });
  if (!user || !user.isActive) throw new UnauthorizedError('User tidak aktif');

  const newPayload: AuthPayload = { userId: user.id, schoolId: user.schoolId || '', role: user.role, email: user.email };
  const tokens = generateTokens(newPayload);
  await redis.setex(`refresh:${user.id}`, 7 * 24 * 3600, tokens.refreshToken);
  return tokens;
}

export async function logout(userId: string) {
  await redis.del(`refresh:${userId}`);
}

export async function changePassword(userId: string, currentPassword: string, newPassword: string) {
  const user = await prisma.user.findUnique({ where: { id: userId } });
  if (!user) throw new NotFoundError('User');

  const match = await bcrypt.compare(currentPassword, user.password);
  if (!match) throw new BadRequestError('AUTH_003', 'Password lama tidak sesuai');

  const hashed = await bcrypt.hash(newPassword, config.bcrypt.rounds);
  await prisma.user.update({ where: { id: userId }, data: { password: hashed } });
  await redis.del(`refresh:${userId}`);
}

export async function getMe(userId: string) {
  const user = await prisma.user.findUnique({
    where: { id: userId },
    select: { id: true, name: true, email: true, role: true, photoUrl: true, schoolId: true, school: { select: { id: true, name: true, subdomain: true, logoUrl: true } } },
  });
  if (!user) throw new NotFoundError('User');
  return user;
}
