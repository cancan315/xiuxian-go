const { Sequelize } = require('sequelize');
const sequelize = require('./models/database');
const Item = require('./models/Item');
const Pet = require('./models/Pet');
const User = require('./models/User');

// Helper function to format values
function formatValue(attr, value) {
  // Format percentage values
  if (typeof value === 'number' && 
      (attr.includes('Rate') || 
       attr.includes('Resist') || 
       attr.includes('Boost') || 
       attr.includes('crit') || 
       attr.includes('combo') || 
       attr.includes('counter') || 
       attr.includes('stun') || 
       attr.includes('dodge') || 
       attr.includes('vampire'))) {
    return `${(value * 100).toFixed(2)}%`;
  }
  return value;
}

async function viewPlayerPets(userId = 42) {
  try {
    // Authenticate and sync database
    await sequelize.authenticate();
    console.log('Connected to the database successfully.');

    // Check if user exists
    const user = await User.findOne({ where: { id: userId } });
    if (!user) {
      console.log(`User with ID ${userId} not found.`);
      return;
    }

    console.log(`\n=== Player Information ===`);
    console.log(`ID: ${user.id}`);
    console.log(`Username: ${user.username}`);
    console.log(`Player Name: ${user.playerName}`);
    console.log(`Level: ${user.level}`);
    console.log(`Realm: ${user.realm}`);

    // Get pets from Items table where type is 'pet'
    console.log(`\n=== Pets in Inventory (from Items table) ===`);
    const petItems = await Item.findAll({
      where: {
        userId: userId,
        type: 'pet'
      }
    });

    if (petItems.length === 0) {
      console.log('No pets found in inventory.');
    } else {
      console.log(`Found ${petItems.length} pet(s) in inventory:`);
      petItems.forEach((item, index) => {
        console.log(`\n--- Pet ${index + 1} ---`);
        console.log(`ID: ${item.id}`);
        console.log(`Item ID: ${item.itemId}`);
        console.log(`Name: ${item.name}`);
        console.log(`Type: ${item.type}`);
        console.log(`Quality: ${item.quality || 'N/A'}`);
        console.log(`Slot: ${item.slot || 'N/A'}`);
        console.log(`Equipped: ${item.equipped ? 'Yes' : 'No'}`);
        console.log(`Active: ${item.isActive ? 'Yes' : 'No'}`);
        
        if (item.details) {
          console.log(`Rarity: ${item.details.rarity || 'N/A'}`);
          console.log(`Level: ${item.details.level || 'N/A'}`);
          console.log(`Star: ${item.details.star || 'N/A'}`);
          console.log(`Description: ${item.details.description || 'N/A'}`);
          
          if (item.details.combatAttributes) {
            console.log(`Combat Attributes:`);
            Object.entries(item.details.combatAttributes).forEach(([attr, value]) => {
              console.log(`  ${attr}: ${formatValue(attr, value)}`);
            });
          }
          
          if (item.details.quality) {
            console.log(`Quality Details:`);
            Object.entries(item.details.quality).forEach(([attr, value]) => {
              console.log(`  ${attr}: ${value}`);
            });
          }
        }
        
        if (item.stats) {
          console.log(`Stats:`);
          Object.entries(item.stats).forEach(([stat, value]) => {
            console.log(`  ${stat}: ${formatValue(stat, value)}`);
          });
        }
      });
    }

    // Get pets from dedicated Pet table
    console.log(`\n=== Pets in Dedicated Pet Table ===`);
    const pets = await Pet.findAll({
      where: {
        userId: userId
      }
    });

    if (pets.length === 0) {
      console.log('No pets found in dedicated pet table.');
    } else {
      console.log(`Found ${pets.length} pet(s) in dedicated pet table:`);
      pets.forEach((pet, index) => {
        console.log(`\n--- Pet ${index + 1} ---`);
        console.log(`ID: ${pet.id}`);
        console.log(`Pet ID: ${pet.petId}`);
        console.log(`Name: ${pet.name}`);
        console.log(`Type: ${pet.type}`);
        console.log(`Rarity: ${pet.rarity}`);
        console.log(`Level: ${pet.level}`);
        console.log(`Star: ${pet.star}`);
        console.log(`Experience: ${pet.experience}/${pet.maxExperience}`);
        console.log(`Description: ${pet.description || 'N/A'}`);
        console.log(`Active: ${pet.isActive ? 'Yes' : 'No'}`);
        console.log(`Power: ${pet.power}`);
        console.log(`Upgrade Items: ${pet.upgradeItems}`);
        
        if (pet.quality) {
          console.log(`Quality:`);
          Object.entries(pet.quality).forEach(([attr, value]) => {
            console.log(`  ${attr}: ${value}`);
          });
        }
        
        if (pet.combatAttributes) {
          console.log(`Combat Attributes:`);
          Object.entries(pet.combatAttributes).forEach(([attr, value]) => {
            console.log(`  ${attr}: ${formatValue(attr, value)}`);
          });
        }
      });
    }

    // Summary
    console.log(`\n=== Summary ===`);
    console.log(`Total pets in inventory: ${petItems.length}`);
    console.log(`Total pets in dedicated table: ${pets.length}`);

  } catch (error) {
    console.error('Error:', error);
  } finally {
    await sequelize.close();
    console.log('\nDatabase connection closed.');
  }
}

// Run the function with default userId 42 or accept command line argument
const userId = process.argv[2] ? parseInt(process.argv[2]) : 42;
viewPlayerPets(userId);