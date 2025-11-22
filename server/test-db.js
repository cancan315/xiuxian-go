// Test database connection
const { Sequelize } = require('sequelize');

// Database configuration for testing (using localhost)
const sequelize = new Sequelize(
  'xiuxian_db',
  'xiuxian_user',
  'xiuxian_password',
  {
    host: 'localhost',
    port: 5432,
    dialect: 'postgres',
    logging: false
  }
);

const testConnection = async () => {
  try {
    await sequelize.authenticate();
    console.log('Connection has been established successfully.');
  } catch (error) {
    console.error('Unable to connect to the database:', error);
  } finally {
    await sequelize.close();
  }
};

testConnection();