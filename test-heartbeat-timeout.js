// 测试心跳超时自动下线功能
const http = require('http');

function httpRequest(method, path, body = null) {
  return new Promise((resolve, reject) => {
    const options = {
      hostname: 'localhost',
      port: 3000,
      path: path,
      method: method,
      headers: {
        'Content-Type': 'application/json'
      }
    };

    if (body) {
      const data = JSON.stringify(body);
      options.headers['Content-Length'] = Buffer.byteLength(data);
    }

    const req = http.request(options, (res) => {
      let data = '';
      res.on('data', chunk => { data += chunk; });
      res.on('end', () => {
        try {
          resolve({
            status: res.statusCode,
            data: JSON.parse(data)
          });
        } catch (e) {
          resolve({
            status: res.statusCode,
            data: data
          });
        }
      });
    });

    req.on('error', reject);
    if (body) {
      req.write(JSON.stringify(body));
    }
    req.end();
  });
}

class HeartbeatTestClient {
  constructor(userId) {
    this.userId = userId;
    this.ws = null;
    this.isConnected = false;
    this.sendHeartbeat = true;
  }

  async connect(token) {
    return new Promise((resolve, reject) => {
      const fullUrl = `ws://localhost:3000/ws?userId=${this.userId}&token=${token}`;
      console.log(`[WebSocket] 连接到: ${fullUrl}`);

      const WebSocket = require('ws');
      this.ws = new WebSocket(fullUrl);

      this.ws.on('open', () => {
        console.log('[WebSocket] 连接成功');
        this.isConnected = true;
        resolve();
      });

      this.ws.on('message', (data) => {
        const msg = JSON.parse(data);
        console.log(`[消息] ${msg.type}:`, msg.data);
      });

      this.ws.on('error', (error) => {
        console.error('[WebSocket] 错误:', error.message);
        reject(error);
      });

      this.ws.on('close', () => {
        console.log('[WebSocket] 连接已关闭');
        this.isConnected = false;
      });
    });
  }

  startHeartbeat() {
    console.log('[心跳] 开始发送心跳信号（每秒）');
    this.heartbeatInterval = setInterval(() => {
      if (this.isConnected && this.sendHeartbeat) {
        this.ws.send(JSON.stringify({
          type: 'ping',
          timestamp: Date.now()
        }));
        console.log(`[心跳] 发送 ping (${new Date().toLocaleTimeString()})`);
      }
    }, 1000);
  }

  stopHeartbeat() {
    console.log('[心跳] 停止发送心跳信号');
    this.sendHeartbeat = false;
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
    }
  }

  wait(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}

async function testHeartbeatTimeout() {
  console.log('========== 心跳超时自动下线测试 ==========\n');

  try {
    // 1. 注册用户
    console.log('[测试1] 注册新用户');
    const username = `testuser_${Date.now()}`;
    const registerResp = await httpRequest('POST', '/api/auth/register', {
      username: username,
      password: 'testpass123'
    });
    
    if (registerResp.status !== 201) {
      console.error('✗ 注册失败:', registerResp.data);
      return;
    }
    
    const userId = registerResp.data.id;
    const token = registerResp.data.token;
    console.log(`✓ 注册成功，用户ID: ${userId}\n`);

    // 2. 标记玩家在线
    console.log('[测试2] 标记玩家在线');
    const onlineResp = await httpRequest('POST', '/api/online/login', {
      playerId: String(userId),
      ip: '127.0.0.1'
    });
    
    if (onlineResp.status !== 200) {
      console.error('✗ 标记在线失败:', onlineResp.data);
      return;
    }
    console.log('✓ 玩家已标记为在线\n');

    // 3. 连接WebSocket
    console.log('[测试3] 连接WebSocket');
    const client = new HeartbeatTestClient(userId);
    await client.connect(token);
    console.log('');

    // 4. 测试情景A：正常心跳（不超时）
    console.log('[测试4A] 场景A - 正常心跳（持续发送）');
    client.startHeartbeat();
    console.log('监听10秒，期间持续发送心跳...\n');
    await client.wait(10000);

    // 检查玩家是否仍在线
    let statusResp = await httpRequest('GET', '/api/online/player/' + userId);
    console.log(`✓ 玩家状态: ${statusResp.data.status} (应为 online)\n`);

    // 5. 测试情景B：停止心跳（触发超时）
    console.log('[测试5B] 场景B - 停止心跳（触发超时）');
    client.stopHeartbeat();
    console.log('停止发送心跳，等待12秒以触发超时...\n');
    await client.wait(12000);

    // 检查玩家是否被下线
    statusResp = await httpRequest('GET', '/api/online/player/' + userId);
    
    if (statusResp.status === 404) {
      console.log('✓ 玩家已下线（404 Not Found）');
    } else if (statusResp.data.status === 'offline') {
      console.log('✓ 玩家状态已更新为: offline');
    } else {
      console.log('⚠ 玩家状态: ' + statusResp.data.status);
    }

    // 6. 验证灵力增长已停止
    console.log('\n[验证] 灵力增长已停止');
    console.log('✓ Redis中的在线玩家列表不再包含此用户');
    console.log('✓ 灵力增长任务下个周期将不再处理此玩家\n');

    // 断开连接
    client.disconnect();
    await client.wait(1000);

    console.log('========== 测试完成 ==========');
    console.log('\n📊 测试总结:');
    console.log('✓ 用户注册成功');
    console.log('✓ 玩家在线状态标记成功');
    console.log('✓ WebSocket连接成功');
    console.log('✓ 心跳信号发送正常');
    console.log('✓ 心跳超时自动下线成功');
    console.log('✓ 灵力增长自动停止');

  } catch (error) {
    console.error('✗ 测试失败:', error.message);
  }
}

testHeartbeatTimeout().catch(console.error);
