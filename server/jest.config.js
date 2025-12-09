module.exports = {
  testEnvironment: 'node',
  roots: ['<rootDir>/__tests__'],
  testMatch: ['**/__tests__/**/*.test.js'],
  collectCoverageFrom: [
    'controllers/**/*.js',
    'models/**/*.js',
    '!models/database.js'
  ],
  setupFilesAfterEnv: [],
  moduleDirectories: ['node_modules', '<rootDir>'],
  transform: {},
  verbose: true
};