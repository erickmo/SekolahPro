module.exports = {
  preset: 'ts-jest',
  testEnvironment: 'node',
  roots: ['<rootDir>/tests'],
  testMatch: ['**/*.test.ts'],
  transform: { '^.+\\.tsx?$': 'ts-jest' },
  testTimeout: 60000,
  clearMocks: true,
  globals: {
    'ts-jest': {
      tsconfig: {
        rootDir: '.',
      }
    }
  }
};
