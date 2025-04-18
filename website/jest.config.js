module.exports = {
  testTimeout: 30000, // Increase timeout to 30 seconds
  moduleNameMapper: {
    '\\.(css|less)$': 'identity-obj-proxy',
    '^react-router-dom$': require.resolve('react-router-dom'),
  },
  testEnvironment: 'jsdom',
  setupFilesAfterEnv: ['<rootDir>/src/setupTests.js'],
  transformIgnorePatterns: [
    '/node_modules/(?!react-markdown|remark-parse|unified|bail|is-plain-obj|trough|mdast-util-from-markdown).+\\.js$', // Add packages that use ES Modules
  ],
};
