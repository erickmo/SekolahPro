import nodemailer from 'nodemailer';
import { prisma } from '../../lib/prisma';
import { config } from '../../config';
import { logger } from '../../lib/logger';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';

const transporter = nodemailer.createTransport({
  host: config.smtp.host,
  port: config.smtp.port,
  auth: { user: config.smtp.user, pass: config.smtp.pass },
});

export async function sendNotification(data: {
  schoolId: string;
  recipientId: string;
  type: string;
  channel: string;
  title: string;
  body: string;
  metadata?: Record<string, unknown>;
}) {
  const notif = await prisma.notification.create({
    data: {
      schoolId: data.schoolId,
      recipientId: data.recipientId,
      type: data.type as never,
      channel: data.channel as never,
      title: data.title,
      body: data.body,
      status: 'PENDING',
    },
  });

  try {
    if (data.channel === 'EMAIL') {
      await sendEmail(data.recipientId, data.title, data.body);
    } else if (data.channel === 'WHATSAPP') {
      await sendWhatsApp(data.recipientId, data.body);
    }
    await prisma.notification.update({ where: { id: notif.id }, data: { status: 'SENT', sentAt: new Date() } });
  } catch (err) {
    logger.error('Failed to send notification', { id: notif.id, error: (err as Error).message });
    await prisma.notification.update({ where: { id: notif.id }, data: { status: 'FAILED' } });
  }

  return notif;
}

async function sendEmail(recipientId: string, subject: string, body: string) {
  const user = await prisma.user.findUnique({ where: { id: recipientId }, select: { email: true } });
  if (!user?.email) return;
  await transporter.sendMail({ from: config.smtp.user, to: user.email, subject, html: `<p>${body}</p>` });
}

async function sendWhatsApp(recipientId: string, message: string) {
  if (!config.whatsapp.apiUrl || config.whatsapp.apiUrl.includes('placeholder')) {
    logger.info('[WhatsApp Mock]', { recipientId, message });
    return;
  }
  const user = await prisma.user.findUnique({ where: { id: recipientId }, select: { phone: true } });
  if (!user?.phone) return;
  await fetch(config.whatsapp.apiUrl, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${config.whatsapp.apiKey}` },
    body: JSON.stringify({ phone: user.phone, message }),
  });
}

export async function sendBulk(schoolId: string, data: { recipientIds: string[]; type: string; title: string; body: string; channel?: string }) {
  const results = await Promise.allSettled(
    data.recipientIds.map((id) =>
      sendNotification({ schoolId, recipientId: id, type: data.type, channel: data.channel || 'PUSH_NOTIFICATION', title: data.title, body: data.body }),
    ),
  );
  return { sent: results.filter((r) => r.status === 'fulfilled').length, failed: results.filter((r) => r.status === 'rejected').length };
}

export async function getNotifications(schoolId: string, recipientId: string, query: { page?: number; limit?: number }) {
  const { page, limit, skip } = getPaginationParams(query);
  const [notifications, total] = await Promise.all([
    prisma.notification.findMany({ where: { schoolId, recipientId }, skip, take: limit, orderBy: { createdAt: 'desc' } }),
    prisma.notification.count({ where: { schoolId, recipientId } }),
  ]);
  return { notifications, meta: buildPaginationMeta(total, page, limit) };
}

export async function markAsRead(schoolId: string, notifId: string, recipientId: string) {
  return prisma.notification.updateMany({ where: { id: notifId, schoolId, recipientId }, data: { status: 'READ' } });
}

export async function getTemplates(schoolId: string) {
  return prisma.notificationTemplate.findMany({ where: { schoolId } });
}

export async function createTemplate(schoolId: string, data: { name: string; type: string; channel: string; titleTemplate: string; bodyTemplate: string }) {
  return prisma.notificationTemplate.create({ data: { ...data, schoolId } });
}
