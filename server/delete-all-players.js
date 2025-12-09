// Script to delete all player data
const sequelize = require('./models/database');
const User = require('./models/User');
const Item = require('./models/Item');
const Pet = require('./models/Pet');
const Herb = require('./models/Herb');
const Pill = require('./models/Pill');

const deleteAllPlayers = async () => {
  try {
    // Authenticate database connection
    await sequelize.authenticate();
    console.log('Connected to database successfully.');

    // Delete all data from all tables in the correct order (due to foreign key constraints)
    await Pill.destroy({ where: {} });
    console.log('Deleted all pills data.');

    await Herb.destroy({ where: {} });
    console.log('Deleted all herbs data.');

    await Pet.destroy({ where: {} });
    console.log('Deleted all pets data.');

    await Item.destroy({ where: {} });
    console.log('Deleted all items data.');

    await User.destroy({ where: {} });
    console.log('Deleted all users data.');

    console.log('All player data has been successfully deleted.');
  } catch (error) {
    console.error('Error deleting player data:', error);
  } finally {
    await sequelize.close();
  }
};

deleteAllPlayers();