import { prisma } from '../../lib/prisma';
import { NotFoundError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';

export async function getPosts(schoolId: string, query: { category?: string; search?: string; page?: number; limit?: number }) {
  const { page, limit, skip } = getPaginationParams(query);
  const where: Record<string, unknown> = { schoolId, isPublished: true };
  if (query.category) where.category = query.category;
  if (query.search) where.title = { contains: query.search, mode: 'insensitive' };
  const [posts, total] = await Promise.all([
    prisma.forumPost.findMany({ where, skip, take: limit, orderBy: { createdAt: 'desc' }, include: { author: { select: { name: true } }, _count: { select: { comments: true } } } }),
    prisma.forumPost.count({ where }),
  ]);
  return { posts, meta: buildPaginationMeta(total, page, limit) };
}

export async function createPost(schoolId: string, authorId: string, data: { title: string; content: string; category: string; tags?: string[] }) {
  return prisma.forumPost.create({ data: { ...data, schoolId, authorId, isPublished: true, tags: data.tags || [] } });
}

export async function getPost(schoolId: string, id: string) {
  const post = await prisma.forumPost.findFirst({ where: { id, schoolId }, include: { author: { select: { name: true } }, comments: { include: { author: { select: { name: true } } }, orderBy: { createdAt: 'asc' } } } });
  if (!post) throw new NotFoundError('Post');
  await prisma.forumPost.update({ where: { id }, data: { viewCount: { increment: 1 } } });
  return post;
}

export async function addComment(schoolId: string, postId: string, authorId: string, content: string) {
  const post = await prisma.forumPost.findFirst({ where: { id: postId, schoolId } });
  if (!post) throw new NotFoundError('Post');
  return prisma.forumComment.create({ data: { postId, authorId, content, schoolId } });
}

export async function deletePost(schoolId: string, id: string, userId: string) {
  const post = await prisma.forumPost.findFirst({ where: { id, schoolId } });
  if (!post) throw new NotFoundError('Post');
  return prisma.forumPost.delete({ where: { id } });
}
