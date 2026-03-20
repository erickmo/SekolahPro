import { Router } from 'express';
import { validate } from '../../middleware/validate';
import { authenticate } from '../../middleware/auth';
import { LoginDto, RegisterDto, RefreshTokenDto, ChangePasswordDto } from './dto';
import * as ctrl from './controller';

export const authRouter = Router();

authRouter.post('/login', validate(LoginDto), ctrl.loginHandler);
authRouter.post('/register', validate(RegisterDto), ctrl.registerHandler);
authRouter.post('/refresh', validate(RefreshTokenDto), ctrl.refreshHandler);
authRouter.post('/logout', authenticate, ctrl.logoutHandler);
authRouter.post('/change-password', authenticate, validate(ChangePasswordDto), ctrl.changePasswordHandler);
authRouter.get('/me', authenticate, ctrl.getMeHandler);
