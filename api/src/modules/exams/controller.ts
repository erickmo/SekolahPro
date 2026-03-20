import { Response, NextFunction } from 'express';
import { AuthRequest } from '../../shared/types';
import { sendSuccess, sendCreated } from '../../shared/response';
import * as service from './service';

const sid = (req: AuthRequest) => req.user!.schoolId;

export async function generateQuestionsH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try {
    const questions = await service.generateQuestions(sid(req), req.body);
    const saved = await service.saveGeneratedQuestions(sid(req), questions, { subject: req.body.subject, topic: req.body.topic, gradeLevel: req.body.gradeLevel });
    sendCreated(res, { questions, saved });
  } catch (err) { next(err); }
}

export async function getQuestionsH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try {
    const { questions, meta } = await service.getQuestions(sid(req), req.query as Record<string, string>);
    sendSuccess(res, questions, 200, meta);
  } catch (err) { next(err); }
}

export async function createExamH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendCreated(res, await service.createExam(sid(req), { ...req.body, createdBy: req.user!.userId })); } catch (err) { next(err); }
}

export async function getExamsH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try {
    const { exams, meta } = await service.getExams(sid(req), req.query as Record<string, string>);
    sendSuccess(res, exams, 200, meta);
  } catch (err) { next(err); }
}

export async function publishExamH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendSuccess(res, await service.publishExam(sid(req), req.params.id as string)); } catch (err) { next(err); }
}

export async function submitResultH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendCreated(res, await service.submitExamResult(sid(req), req.params.id as string, req.body.studentId, req.body.answers)); } catch (err) { next(err); }
}

export async function getResultsH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendSuccess(res, await service.getExamResults(sid(req), req.params.id as string)); } catch (err) { next(err); }
}

export async function getAnalysisH(req: AuthRequest, res: Response, next: NextFunction): Promise<void> {
  try { sendSuccess(res, await service.getExamAnalysis(sid(req), req.params.id as string)); } catch (err) { next(err); }
}
