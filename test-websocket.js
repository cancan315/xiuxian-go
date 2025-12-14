// test-websocket.js
// WebSocket功能测试脚本

const http = require('http');

/**
 * 简单的WebSocket测试客户端
 */
class WebSocketTestClient {
  constructor(url) {
    this.url = url;
    this.ws = null;
    this.messageCount = 0;
    this.lastMessageTime = null;
  }

  async connect(userId, token) {
    return new Promise((resolve, reject) => {
      const fullUrl = `ws://localhost:3000/ws?userId=${userId}&token=${token}`;
      console.log(`[测试] 连接到: ${fullUrl}`);

      const WebSocket = require('ws');
      this.ws = new WebSocket(fullUrl);

      this.ws.on('open', () => {
        console.log('[测试] WebSocket连接成功');
        resolve();
      });

      this.ws.on('message', (data) => {
        this.messageCount++;
        this.lastMessageTime = Date.now();
        const msg = JSON.parse(data);
        console.log(`[消息#${this.messageCount}] 类型: ${msg.type}`, msg.data);
      });

      this.ws.on('error', (error) => {
        console.error('[测试] 连接错误:', error);
        reject(error);
      });

      this.ws.on('close', () => {
        console.log('[测试] 连接已关闭');
      });
    });
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      console.log('[测试] 已断开连接');
    }
  }

  getStats() {
    return {
      messageCount: this.messageCount,
      lastMessageTime: this.lastMessageTime,
      isConnected: this.ws && this.ws.readyState === 1
    };
  }

  wait(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}

/**
 * 标记玩家在线
 */
async function markPlayerOnline(userId) {
  return new Promise((resolve, reject) => {
    const postData = JSON.stringify({ playerId: String(userId), ip: '127.0.0.1' });

    const options = {
      hostname: 'localhost',
      port: 3000,
      path: '/api/online/login',
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Content-Length': postData.length
      }
    };

    const req = http.request(options, (res) => {
      let data = '';
      res.on('data', chunk => { data += chunk; });
      res.on('end', () => {
        console.log(`[后端响应] Status: ${res.statusCode}`);
        if (res.statusCode === 200) {
          console.log(`[后端响应] ${data}`);
          resolve();
        } else {
          reject(new Error(`标记玩家在线失败: ${data}`));
        }
      });
    });

    req.on('error', reject);
    req.write(postData);
    req.end();
  });
}

/**
 * 标记玩家离线
 */
async function markPlayerOffline(userId) {
  return new Promise((resolve, reject) => {
    const postData = JSON.stringify({ playerId: String(userId) });

    const options = {
      hostname: 'localhost',
      port: 3000,
      path: '/api/online/logout',
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Content-Length': postData.length
      }
    };

    const req = http.request(options, (res) => {
      let data = '';
      res.on('data', chunk => { data += chunk; });
      res.on('end', () => {
        if (res.statusCode === 200) {
          resolve();
        } else {
          reject(new Error(`标记玩家离线失败: ${data}`));
        }
      });
    });

    req.on('error', reject);
    req.write(postData);
    req.end();
  });
}

/**
 * 运行测试
 */
async function runTests() {
  console.log('========== WebSocket 功能测试 ==========\n');

  const client = new WebSocketTestClient();
  const userId = 1;
  const token = 'test_token_12345';

  try {
    // 测试0: 标记玩家在线
    console.log('测试0: 标记玩家在线');
    await markPlayerOnline(userId);
    console.log('✓ 玩家已标记为在线\n');

    // 等待灵力增长任务启动
    console.log('等待后端灵力增长任务启动（10秒）...');
    await new Promise(resolve => setTimeout(resolve, 10000));

    // 测试1: 基本连接
    console.log('测试1: 基本连接');
    await client.connect(userId, token);
    console.log('✓ 连接成功\n');

    // 测试2: 接收消息（等待30秒，观察灵力增长消息）
    console.log('测试2: 等待灵力增长消息（30秒）');
    await client.wait(30000);

    const stats = client.getStats();
    console.log(`✓ 接收到 ${stats.messageCount} 条消息\n`);

    if (stats.messageCount === 0) {
      console.log('⚠ 警告: 没有接收到任何消息，请检查后端服务\n');
    }

    // 测试3: 标记玩家离线
    console.log('\n测试3: 标记玩家离线');
    await markPlayerOffline(userId);
    console.log('✓ 玩家已标记为离线\n');

    // 测试4: 断开连接
    console.log('测试4: 断开连接');
    client.disconnect();
    await client.wait(1000);
    console.log('✓ 连接已断开\n');

    // 总结
    console.log('========== 测试总结 ==========');
    console.log(`总消息数: ${stats.messageCount}`);
    console.log(`测试结果: ${stats.messageCount > 0 ? '✓ 成功' : '✗ 失败'}`);

  } catch (error) {
    console.error('✗ 测试失败:', error.message);
  }
}

// 命令行执行
if (require.main === module) {
  runTests().catch(console.error);
}

module.exports = WebSocketTestClient;
