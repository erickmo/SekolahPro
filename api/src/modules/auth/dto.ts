import { z } from 'zod';

export const LoginDto = z.object({
  email: z.string().email(),
  password: z.string().min(6),
});

export const RegisterDto = z.object({
  email: z.string().email(),
  password: z.string().min(8),
  name: z.string().min(2),
  schoolId: z.string().optional(),
  role: z.string().optional(),
});

export const RefreshTokenDto = z.object({
  refreshToken: z.string(),
});

export const ChangePasswordDto = z.object({
  currentPassword: z.string(),
  newPassword: z.string().min(8),
});

export const ForgotPasswordDto = z.object({
  email: z.string().email(),
});

export const ResetPasswordDto = z.object({
  token: z.string(),
  newPassword: z.string().min(8),
});

export type LoginInput = z.infer<typeof LoginDto>;
export type RegisterInput = z.infer<typeof RegisterDto>;
export type RefreshTokenInput = z.infer<typeof RefreshTokenDto>;
export type ChangePasswordInput = z.infer<typeof ChangePasswordDto>;
