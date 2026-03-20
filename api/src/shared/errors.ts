export class AppError extends Error {
  constructor(
    public statusCode: number,
    public code: string,
    message: string,
  ) {
    super(message);
    this.name = 'AppError';
  }
}

export class NotFoundError extends AppError {
  constructor(resource = 'Resource') {
    super(404, 'NOT_FOUND', `${resource} tidak ditemukan`);
  }
}

export class UnauthorizedError extends AppError {
  constructor(message = 'Token tidak valid atau sudah kadaluarsa') {
    super(401, 'AUTH_001', message);
  }
}

export class ForbiddenError extends AppError {
  constructor(message = 'Anda tidak memiliki akses') {
    super(403, 'AUTH_002', message);
  }
}

export class ValidationError extends AppError {
  constructor(message: string) {
    super(400, 'VALIDATION_ERROR', message);
  }
}

export class ConflictError extends AppError {
  constructor(message: string) {
    super(409, 'CONFLICT', message);
  }
}

export class BadRequestError extends AppError {
  constructor(code: string, message: string) {
    super(400, code, message);
  }
}
