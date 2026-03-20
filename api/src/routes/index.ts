import { Router } from 'express';
import { authRouter } from '../modules/auth/router';
import { schoolsRouter } from '../modules/schools/router';
import { studentsRouter } from '../modules/students/router';
import { studentCardsRouter } from '../modules/student-cards/router';
import { cooperativeRouter } from '../modules/cooperative/router';
import { academicRouter } from '../modules/academic/router';
import { examsRouter } from '../modules/exams/router';
import { extracurricularRouter } from '../modules/extracurricular/router';
import { websiteRouter } from '../modules/website/router';
import { ppdbRouter } from '../modules/ppdb/router';
import { governmentReportsRouter } from '../modules/government-reports/router';
import { stakeholdersRouter } from '../modules/stakeholders/router';
import { tutoringRouter } from '../modules/tutoring/router';
import { dashboardRouter } from '../modules/dashboard/router';
import { parentPortalRouter } from '../modules/parent-portal/router';
import { notificationsRouter } from '../modules/notifications/router';
import { forumRouter } from '../modules/forum/router';
import { meetingsRouter } from '../modules/meetings/router';
import { schoolBlogRouter } from '../modules/school-blog/router';
import { paymentsRouter } from '../modules/payments/router';
import { canteenRouter } from '../modules/canteen/router';
import { assetsRouter } from '../modules/assets/router';
import { libraryRouter } from '../modules/library/router';
import { scholarshipsRouter } from '../modules/scholarships/router';
import { ewsRouter } from '../modules/ews/router';
import { learningPathRouter } from '../modules/learning-path/router';
import { aiTutorRouter } from '../modules/ai-tutor/router';
import { sentimentRouter } from '../modules/sentiment/router';
import { healthRouter } from '../modules/health/router';
import { counselingRouter } from '../modules/counseling/router';
import { hrRouter } from '../modules/hr/router';
import { smartGateRouter } from '../modules/smart-gate/router';
import { roomBookingRouter } from '../modules/room-booking/router';
import { transportRouter } from '../modules/transport/router';
import { helpdeskRouter } from '../modules/helpdesk/router';
import { alumniRouter } from '../modules/alumni/router';
import { internshipRouter } from '../modules/internship/router';
import { teacherTrainingRouter } from '../modules/teacher-training/router';
import { dapodikRouter } from '../modules/dapodik/router';
import { foundationRouter } from '../modules/foundation/router';
import { openApiRouter } from '../modules/open-api/router';
// v3.0 new modules
import { lessonPlansRouter } from '../modules/lesson-plans/router';
import { assessmentsRouter } from '../modules/assessments/router';
import { remediationRouter } from '../modules/remediation/router';
import { budgetRouter } from '../modules/budget/router';
import { donationsRouter } from '../modules/donations/router';
import { committeeRouter } from '../modules/committee/router';
import { volunteerRouter } from '../modules/volunteer/router';
import { specialNeedsRouter } from '../modules/special-needs/router';
import { uniformsRouter } from '../modules/uniforms/router';
import { eventsRouter } from '../modules/events/router';
import { iotRouter } from '../modules/iot/router';
import { gamificationRouter } from '../modules/gamification/router';
import { eduMarketplaceRouter } from '../modules/edu-marketplace/router';
import { tenantRouter } from '../modules/tenant/router';
import { listingRouter } from '../modules/listing/router';
import { auditLogRouter } from '../modules/audit-log/router';
import { securityRouter } from '../modules/security/router';
import { analyticsRouter } from '../modules/analytics/router';
import { lmsRouter } from '../modules/lms/router';
import { backupRouter } from '../modules/backup/router';
import { warehouseRouter } from '../modules/warehouse/router';

export const router = Router();

// Auth & Core
router.use('/auth', authRouter);
router.use('/schools', schoolsRouter);
router.use('/dashboard', dashboardRouter);

// Module Group A — Core
router.use('/students', studentsRouter);
router.use('/student-cards', studentCardsRouter);
router.use('/cooperative', cooperativeRouter);
router.use('/academic', academicRouter);
router.use('/exams', examsRouter);
router.use('/extracurricular', extracurricularRouter);
router.use('/website', websiteRouter);
router.use('/ppdb', ppdbRouter);
router.use('/reports/government', governmentReportsRouter);
router.use('/stakeholders', stakeholdersRouter);
router.use('/tutoring', tutoringRouter);

// Module Group B — Communication
router.use('/parent-portal', parentPortalRouter);
router.use('/notifications', notificationsRouter);
router.use('/forum', forumRouter);
router.use('/meetings', meetingsRouter);
router.use('/blog', schoolBlogRouter);

// Module Group C — Finance
router.use('/payments', paymentsRouter);
router.use('/canteen', canteenRouter);
router.use('/assets', assetsRouter);
router.use('/library', libraryRouter);
router.use('/scholarships', scholarshipsRouter);
router.use('/budget', budgetRouter);
router.use('/donations', donationsRouter);

// Module Group D — AI
router.use('/ews', ewsRouter);
router.use('/learning-path', learningPathRouter);
router.use('/ai-tutor', aiTutorRouter);
router.use('/sentiment', sentimentRouter);

// Module Group E — Health & Wellbeing
router.use('/health', healthRouter);
router.use('/counseling', counselingRouter);
router.use('/special-needs', specialNeedsRouter);

// Module Group F — Operations
router.use('/hr', hrRouter);
router.use('/smart-gate', smartGateRouter);
router.use('/room-booking', roomBookingRouter);
router.use('/transport', transportRouter);
router.use('/helpdesk', helpdeskRouter);
router.use('/uniforms', uniformsRouter);
router.use('/events', eventsRouter);
router.use('/iot', iotRouter);
router.use('/volunteer', volunteerRouter);
router.use('/committee', committeeRouter);

// Module Group G — Career & Academic
router.use('/alumni', alumniRouter);
router.use('/internship', internshipRouter);
router.use('/teacher-training', teacherTrainingRouter);
router.use('/lesson-plans', lessonPlansRouter);
router.use('/assessments', assessmentsRouter);
router.use('/remediation', remediationRouter);

// Module Group H — Integrations
router.use('/dapodik', dapodikRouter);
router.use('/foundation', foundationRouter);
router.use('/api-management', openApiRouter);

// Module Group I — Gamification & Marketplace
router.use('/gamification', gamificationRouter);
router.use('/edu-marketplace', eduMarketplaceRouter);

// Module Group J — SaaS Platform
router.use('/tenant', tenantRouter);
router.use('/listing', listingRouter);

// Module Group K — Platform Infrastructure
router.use('/audit-logs', auditLogRouter);
router.use('/security', securityRouter);

// Module Group L — Analytics, LMS, Backup, Warehouse
router.use('/analytics', analyticsRouter);
router.use('/lms', lmsRouter);
router.use('/backup', backupRouter);
router.use('/warehouse', warehouseRouter);
