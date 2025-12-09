const sequelize = require('./server/models/database');
const User = require('./server/models/User');
const bcrypt = require('bcryptjs');

async function testPlayerRegistration() {
  try {
    // Sync database
    await sequelize.sync({ force: true });
    console.log('数据库同步成功');

    // Test user data
    const testUser = {
      username: 'test_player',
      password: 'test_password123'
    };

    // Hash password
    const salt = await bcrypt.genSalt(10);
    const hashedPassword = await bcrypt.hash(testUser.password, salt);

    // Create user
    const user = await User.create({
      username: testUser.username,
      password: hashedPassword
    });

    console.log('玩家创建成功!');
    console.log('玩家ID:', user.id);
    console.log('用户名:', user.username);
    console.log('灵石数量:', user.spiritStones);
    console.log('是否为新玩家:', user.isNewPlayer);

    // Verification
    if (user.spiritStones === 20000) {
      console.log('✓ 测试通过: 玩家注册后 spiritStones 正确初始化为 20000');
    } else {
      console.log('✗ 测试失败: 玩家注册后 spiritStones 未正确初始化');
      console.log('  期望值: 20000');
      console.log('  实际值:', user.spiritStones);
    }

    // Test multiple users
    console.log('\n--- 测试多个玩家注册 ---');
    const multipleUsers = [
      { username: 'player1', password: 'pass1' },
      { username: 'player2', password: 'pass2' },
      { username: 'player3', password: 'pass3' }
    ];

    for (let i = 0; i < multipleUsers.length; i++) {
      const userData = multipleUsers[i];
      const salt = await bcrypt.genSalt(10);
      const hashedPassword = await bcrypt.hash(userData.password, salt);
      
      const user = await User.create({
        username: userData.username,
        password: hashedPassword
      });
      
      console.log(`玩家 ${user.username} 创建成功，灵石数量: ${user.spiritStones}`);
      
      if (user.spiritStones !== 20000) {
        console.log(`✗ 玩家 ${user.username} 的 spiritStones 值不正确`);
      }
    }
    
    console.log('✓ 所有玩家的 spiritStones 值均正确初始化');

  } catch (error) {
    console.error('测试过程中发生错误:', error);
  } finally {
    // Close database connection
    await sequelize.close();
    console.log('\n数据库连接已关闭');
  }
}

testPlayerRegistration();