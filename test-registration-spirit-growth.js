// 测试注册后灵力增长功能
const http = require('http');

/**
 * 发送HTTP请求
 */
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

/**
 * WebSocket测试客户端
 */
class WebSocketTestClient {
  constructor() {
    this.ws = null;
    this.messageCount = 0;
    this.messages = [];
  }

  async connect(userId, token) {
    return new Promise((resolve, reject) => {
      const fullUrl = `ws://localhost:3000/ws?userId=${userId}&token=${token}`;
      console.log(`[WebSocket] 连接到: ${fullUrl}`);

      const WebSocket = require('ws');
      this.ws = new WebSocket(fullUrl);

      this.ws.on('open', () => {
        console.log('[WebSocket] 连接成功');
        resolve();
      });

      this.ws.on('message', (data) => {
        this.messageCount++;
        const msg = JSON.parse(data);
        this.messages.push(msg);
        console.log(`[消息#${this.messageCount}] ${msg.type}:`, msg.data);
      });

      this.ws.on('error', (error) => {
        console.error('[WebSocket] 错误:', error.message);
        reject(error);
      });

      this.ws.on('close', () => {
        console.log('[WebSocket] 连接已关闭');
      });
    });
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

async function runTest() {
  console.log('========== 注册后灵力增长功能测试 ==========\n');

  try {
    // 生成唯一用户名
    const username = `testuser_${Date.now()}`;
    const password = 'testpass123';
    
    console.log(`[测试] 注册用户: ${username}`);
    
    // 1. 注册新用户
    const registerResp = await httpRequest('POST', '/api/auth/register', {
      username: username,
      password: password
    });
    
    if (registerResp.status !== 201) {
      console.error('✗ 注册失败:', registerResp.data);
      return;
    }
    
    const userId = registerResp.data.id;
    const token = registerResp.data.token;
    console.log(`✓ 注册成功，用户ID: ${userId}\n`);
    
    // 2. 标记玩家在线
    console.log('[测试] 标记玩家在线');
    const onlineResp = await httpRequest('POST', '/api/online/login', {
      playerId: String(userId),
      ip: '127.0.0.1'
    });
    
    if (onlineResp.status !== 200) {
      console.error('✗ 标记在线失败:', onlineResp.data);
      return;
    }
    console.log('✓ 玩家已标记为在线\n');
    
    // 3. 等待灵力增长任务启动
    console.log('[测试] 等待灵力增长任务启动（10秒）...');
    await new Promise(resolve => setTimeout(resolve, 10000));
    
    // 4. 连接WebSocket
    console.log('\n[测试] 连接WebSocket\n');
    const client = new WebSocketTestClient();
    await client.connect(userId, token);
    
    // 5. 等待灵力增长消息
    console.log('[测试] 监听灵力增长消息（25秒）\n');
    await client.wait(25000);
    
    // 6. 断开连接
    console.log('\n[测试] 断开连接');
    client.disconnect();
    await client.wait(1000);
    
    // 7. 输出结果
    console.log('\n========== 测试结果 ==========');
    console.log(`接收到灵力增长消息: ${client.messageCount} 条`);
    
    if (client.messageCount > 0) {
      console.log('\n✓ 测试成功！注册后灵力增长功能正常工作');
      console.log('\n接收到的消息详情:');
      client.messages.forEach((msg, idx) => {
        if (msg.type === 'spirit:grow') {
          console.log(`  消息${idx + 1}: 灵力从 ${msg.data.oldSpirit.toFixed(2)} → ${msg.data.newSpirit.toFixed(2)} (增长: ${msg.data.gainAmount.toFixed(2)})`)
        }
      });
    } else {
      console.log('\n✗ 测试失败！没有接收到灵力增长消息');
    }
    
  } catch (error) {
    console.error('✗ 测试过程中发生错误:', error.message);
  }
}

// 执行测试
runTest().catch(console.error);
