const express = require('express');
const cors = require('cors');
const sequelize = require('./models/database');
const { connectRedis } = require('./utils/redis');

// Load environment variables
require('dotenv').config();

// Import routes
const authRoutes = require('./routes/authRoutes');
const playerRoutes = require('./routes/playerRoutes');

// Create Express app
const app = express();
const PORT = process.env.PORT || 3000;

// Middleware
app.use(cors());
app.use(express.json());

// Routes
app.use('/api/auth', authRoutes);
app.use('/api/player', playerRoutes);

app.get('/', (req, res) => {
  res.json({ message: 'Vue Idle Xiuxian Backend API' });
});

// Test database connection
const connectDB = async () => {
  try {
    await sequelize.authenticate();
    console.log('Database connection has been established successfully.');
  } catch (error) {
    console.error('Unable to connect to the database:', error);
    process.exit(1);
  }
};

// Sync database models
const syncDatabase = async () => {
  try {
    await sequelize.sync({ alter: true });
    console.log('Database synced');
  } catch (error) {
    console.error('Error syncing database:', error);
  }
};

// Start server
const startServer = async () => {
  await connectDB();
  await syncDatabase();
  await connectRedis();
  
  app.listen(PORT, () => {
    console.log(`Server is running on port ${PORT}`);
  });
};

startServer();