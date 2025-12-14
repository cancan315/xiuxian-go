// 诊断脚本：检查Redis中的在线状态
const http = require('http');

/**
 * HTTP请求辅助函数
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

async function diagnose() {
  console.log('========== 在线玩家诊断 ==========\n');

  try {
    // 1. 获取所有在线玩家
    console.log('[诊断1] 查询在线玩家列表');
    const onlineResp = await httpRequest('GET', '/api/online/players');
    const onlinePlayerIds = onlineResp.data.players.map(p => p.playerId);
    console.log(`✓ 在线玩家数: ${onlineResp.data.players.length}`);
    console.log(`在线玩家ID列表: ${onlinePlayerIds.join(', ')}\n`);

    // 2. 检查userID 8的具体状态
    console.log('[诊断2] 检查userID 8的在线状态');
    const player8StatusResp = await httpRequest('GET', '/api/online/player/8');
    
    if (player8StatusResp.status === 404) {
      console.log('✗ userID 8 不在线');
      console.log('原因: userID 8没有调用/api/online/login来标记上线\n');
      
      // 3. 尝试将userID 8标记为在线
      console.log('[修复] 标记userID 8为在线');
      const markOnlineResp = await httpRequest('POST', '/api/online/login', {
        playerId: '8',
        ip: '127.0.0.1'
      });
      
      if (markOnlineResp.status === 200) {
        console.log('✓ userID 8已标记为在线');
        console.log(`  登录时间: ${new Date(markOnlineResp.data.loginTime).toLocaleString()}\n`);
        
        // 4. 验证现在的数据
        console.log('[诊断3] 重新检查在线玩家列表');
        const newOnlineResp = await httpRequest('GET', '/api/online/players');
        const newOnlinePlayerIds = newOnlineResp.data.players.map(p => p.playerId);
        console.log(`✓ 在线玩家数: ${newOnlineResp.data.players.length}`);
        console.log(`在线玩家ID列表: ${newOnlinePlayerIds.join(', ')}\n`);
      } else {
        console.log('✗ 标记失败:', markOnlineResp.data);
      }
    } else {
      console.log('✓ userID 8已在线');
      console.log('  状态:', player8StatusResp.data.status);
      console.log(`  登录时间: ${new Date(player8StatusResp.data.loginTime).toLocaleString()}`);
      console.log(`  最后心跳: ${new Date(player8StatusResp.data.lastHeartbeat).toLocaleString()}\n`);
    }

    // 5. 显示所有在线玩家的详细信息
    if (onlineResp.data.players.length > 0) {
      console.log('[诊断4] 所有在线玩家的详细信息');
      for (const player of onlineResp.data.players) {
        console.log(`玩家ID ${player.playerId}:`);
        console.log(`  登录时间: ${new Date(player.loginTime).toLocaleString()}`);
        console.log(`  最后心跳: ${new Date(player.lastHeartbeat).toLocaleString()}`);
        console.log(`  IP: ${player.ip}`);
        console.log('');
      }
    }

    console.log('========== 诊断完成 ==========');

  } catch (error) {
    console.error('诊断过程中发生错误:', error.message);
  }
}

diagnose().catch(console.error);
