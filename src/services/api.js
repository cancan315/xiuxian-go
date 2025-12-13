// API服务类，用于与后端进行通信
const API_BASE_URL = '/api';

// 将大驼峰命名转换为小驼峰命名的通用函数
function convertToCamelCase(obj) {
  if (obj === null || typeof obj !== 'object') {
    return obj;
  }

  if (Array.isArray(obj)) {
    return obj.map(item => convertToCamelCase(item));
  }

  const converted = {};
  for (const key in obj) {
    if (obj.hasOwnProperty(key)) {
      // 定义特殊字段映射规则
      const fieldMapping = {
        'ID': 'id',
        'UserID': 'userId',
        'EquipmentID': 'equipmentId',
        'EquipType': 'equipType',
        'EnhanceLevel': 'enhanceLevel',
        'Equipped': 'equipped',
        'ExtraAttributes': 'extraAttributes',
        'RequiredRealm': 'requiredRealm'
      };
      
      // 使用特殊映射规则或默认转换规则
      const camelCaseKey = fieldMapping[key] || (key.charAt(0).toLowerCase() + key.slice(1));
      converted[camelCaseKey] = convertToCamelCase(obj[key]);
    }
  }
  return converted;
}

class APIService {
  // 用户认证相关方法
  
  /**
   * 用户注册
   * @param {string} username - 用户名
   * @param {string} password - 密码
   * @returns {Promise<Object>} 注册结果
   */
  static async register(username, password) {
    const response = await fetch(`${API_BASE_URL}/auth/register`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ username, password })
    });
    
    return response.json();
  }
  
  /**
   * 用户登录
   * @param {string} username - 用户名
   * @param {string} password - 密码
   * @returns {Promise<Object>} 登录结果，包含token等信息
   */
  static async login(username, password) {
    const response = await fetch(`${API_BASE_URL}/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ username, password })
    });
    
    return response.json();
  }
  
  /**
   * 获取用户信息
   * @param {string} token - 认证令牌
   * @returns {Promise<Object>} 用户信息
   */
  static async getUser(token) {
    const response = await fetch(`${API_BASE_URL}/auth/user`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });
    
    return response.json();
  }
  
  // 在线状态相关方法
  
  /**
   * 标记玩家在线
   * @param {string} playerId - 玩家ID
   * @param {string} ip - 玩家IP地址（可选）
   * @returns {Promise<Object>} 在线状态结果
   */
  static async playerOnline(playerId, ip = '') {
    const response = await fetch(`${API_BASE_URL}/online/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ playerId, ip })
    });
    
    return response.json();
  }
  
  
  
  /**
   * 标记玩家离线
   * @param {string} playerId - 玩家ID
   * @returns {Promise<Object>} 离线状态结果
   */
  static async playerOffline(playerId) {
    const response = await fetch(`${API_BASE_URL}/online/logout`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ playerId })
    });
    
    return response.json();
  }
  
  // 玩家数据相关方法
  
  /**
   * 获取玩家完整数据
   * @param {string} token - 认证令牌
   * @returns {Promise<Object>} 玩家数据
   */
  static async getPlayerData(token) {
    console.log('[API Service] 调用获取玩家数据接口: /api/player/data');
    const response = await fetch(`${API_BASE_URL}/player/data`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      }
    });
    
    if (!response.ok) {
      throw new Error('获取玩家数据失败');
    }
    
    const data = await response.json();
    console.log('从后端API获取玩家数据:', data);
    console.log('[API Service] 获取玩家数据接口调用完成: /api/player/data');
    // 统一转换后端返回的数据字段为小驼峰命名法
    return convertToCamelCase(data);
  }
  
  /**
   * 仅获取玩家灵力值
   * @param {string} token - 认证令牌
   * @returns {Promise<Object>} 玩家灵力值信息
   */
  static async getPlayerSpirit(token) {
    const response = await fetch(`${API_BASE_URL}/player/spirit`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      }
    });
    
    if (!response.ok) {
      throw new Error('获取玩家灵力值失败');
    }
    
    return response.json();
  }
  
  /**
   * 增量更新玩家数据
   * @param {string} token - 认证令牌
   * @param {Object} data - 要更新的数据
   * @returns {Promise<Object>} 更新结果
   */
  static async updatePlayerData(token, data) {
    console.log('发送玩家数据增量更新到后端API:', {
      userId: data.user?.id,
      userName: data.user?.playerName,
      itemCount: data.items?.length,
      petCount: data.pets?.length,
      herbCount: data.herbs?.length,
      pillCount: data.pills?.length
    });
    
    const response = await fetch(`${API_BASE_URL}/player/data`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify(data)
    });
    
    const text = await response.text();
    console.log('后端API响应状态:', response.status);
    console.log('后端API响应内容:', text);
    
    try {
      return text ? JSON.parse(text) : { message: '数据增量更新成功' };
    } catch (e) {
      console.error('JSON解析错误:', e);
      console.error('响应内容:', text);
      return { message: '数据增量更新成功' };
    }
  }
  
  /**
   * 删除指定物品
   * @param {string} token - 认证令牌
   * @param {Array<string>} itemIds - 要删除的物品ID列表
   * @returns {Promise<Object>} 删除结果
   */
  static async deleteItems(token, itemIds) {
    const response = await fetch(`${API_BASE_URL}/player/items`, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({ itemIds })
    });
    
    const text = await response.text();
    console.log('后端API响应状态:', response.status);
    console.log('后端API响应内容:', text);
    
    try {
      return text ? JSON.parse(text) : { message: '物品删除成功' };
    } catch (e) {
      console.error('JSON解析错误:', e);
      console.error('响应内容:', text);
      return { message: '物品删除成功' };
    }
  }
  
  /**
   * 删除指定灵宠
   * @param {string} token - 认证令牌
   * @param {Array<string>} petIds - 要删除的灵宠ID列表
   * @returns {Promise<Object>} 删除结果
   */
  static async deletePets(token, petIds) {
    const response = await fetch(`${API_BASE_URL}/player/pets`, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({ petIds })
    });
    
    const text = await response.text();
    console.log('后端API响应状态:', response.status);
    console.log('后端API响应内容:', text);
    
    try {
      return text ? JSON.parse(text) : { message: '灵宠删除成功' };
    } catch (e) {
      console.error('JSON解析错误:', e);
      console.error('响应内容:', text);
      return { message: '灵宠删除成功' };
    }
  }
  
  /**
   * 增量更新灵力值
   * @param {string} token - 认证令牌
   * @param {number} spirit - 灵力值
   * @returns {Promise<Object>} 更新结果
   */
  static async updateSpirit(token, spirit) {
    const response = await fetch(`${API_BASE_URL}/player/spirit`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({ spirit })
    });
    
    if (!response.ok) {
      throw new Error('更新灵力值失败');
    }
    
    return response.json();
  }
  
  /**
   * 标记玩家为离线状态（旧版接口，为了兼容性保留）
   * @param {string} token - 认证令牌
   * @returns {Promise<Object>} 离线状态结果
   */
  static async setPlayerOffline(token) {
    const response = await fetch(`${API_BASE_URL}/player/offline`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      }
    });
    
    if (!response.ok) {
      throw new Error('设置玩家离线状态失败');
    }
    
    return response.json();
  }
  
  /**
   * 获取玩家物品详情
   * @param {string} token - 认证令牌
   * @param {string} itemId - 物品ID
   * @returns {Promise<Object>} 物品详情
   */
  static async getItemDetails(token, itemId) {
    const response = await fetch(`${API_BASE_URL}/player/inventory/item/${itemId}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      }
    });
    
    if (!response.ok) {
      throw new Error('获取物品详情失败');
    }
    
    return response.json();
  }

  /**
   * 获取玩家装备库存
   * @param {string} token - 认证令牌
   * @param {Object} params - 查询参数
   * @returns {Promise<Object>} 装备库存数据
   */
  static async getPlayerInventory(token, params = {}) {
    // 修复URL构造错误，使用相对路径而不是构造完整的URL对象
    let url = `${API_BASE_URL}/player/inventory`;
    
    // 添加查询参数
    const searchParams = new URLSearchParams();
    Object.keys(params).forEach(key => {
      if (params[key] !== undefined && params[key] !== null) {
        searchParams.append(key, params[key]);
      }
    });
    
    if (searchParams.toString()) {
      url += `?${searchParams.toString()}`;
    }

    const response = await fetch(url, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      }
    });

    if (!response.ok) {
      throw new Error('获取玩家装备数据失败');
    }

    return response.json();
  }
  
  // 装备系统相关方法
  
  /**
   * 获取玩家装备列表
   * @param {string} token - 认证令牌
   * @param {Object} params - 查询参数
   * @returns {Promise<Object>} 装备列表
   */
  static async getEquipmentList(token, params = {}) {
    let url = `${API_BASE_URL}/player/equipment`;
    
    // 如果有userId参数，则添加到URL路径中
    if (params.userId) {
      url += `/${params.userId}`;
      delete params.userId;
    }
    
    // 添加其他查询参数，确保参数名与后端一致
    const searchParams = new URLSearchParams();
    Object.keys(params).forEach(key => {
      if (params[key] !== undefined && params[key] !== null) {
        // 不再进行任何转换，直接使用原始键名
        searchParams.append(key, params[key]);
      }
    });

    if (searchParams.toString()) {
      url += `?${searchParams.toString()}`;
    }

    console.log(`[API Service] 调用获取玩家装备列表接口: ${url}`);
    const response = await fetch(url, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      }
    });

    if (!response.ok) {
      throw new Error('获取玩家装备列表失败');
    }

    const data = await response.json();
    console.log(`[API Service] 获取玩家装备列表接口调用完成: ${url}`, data);
    
    // 统一转换后端返回的数据字段为小驼峰命名法
    return convertToCamelCase(data);
  }
  
  /**
   * 获取装备详情
   * @param {string} token - 认证令牌
   * @param {string} equipmentId - 装备ID
   * @returns {Promise<Object>} 装备详情
   */
  static async getEquipmentDetails(token, equipmentId) {
    const response = await fetch(`${API_BASE_URL}/player/equipment/details/${equipmentId}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      }
    });

    if (!response.ok) {
      throw new Error('获取装备详情失败');
    }

    const data = await response.json();
    // 统一转换后端返回的数据字段为小驼峰命名法
    return convertToCamelCase(data);
  }
  
  /**
   * 装备强化
   * @param {string} token - 认证令牌
   * @param {string} equipmentId - 装备ID
   * @param {number} reinforceStones - 当前拥有的强化石数量
   * @returns {Promise<Object>} 强化结果
   */
  static async enhanceEquipment(token, equipmentId, reinforceStones) {
    const response = await fetch(`${API_BASE_URL}/player/equipment/${equipmentId}/enhance`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({ reinforceStones })
    });

    if (!response.ok) {
      throw new Error('装备强化失败');
    }

    const data = await response.json();
    // 统一转换后端返回的数据字段为小驼峰命名法
    return convertToCamelCase(data);
  }
  
  /**
   * 装备洗练
   * @param {string} token - 认证令牌
   * @param {string} equipmentId - 装备ID
   * @param {number} refinementStones - 当前拥有的洗练石数量
   * @returns {Promise<Object>} 洗练结果
   */
  static async reforgeEquipment(token, equipmentId, refinementStones) {
    const response = await fetch(`${API_BASE_URL}/player/equipment/${equipmentId}/reforge`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({ refinementStones })
    });

    if (!response.ok) {
      throw new Error('装备洗练失败');
    }

    const data = await response.json();
    // 统一转换后端返回的数据字段为小驼峰命名法
    return convertToCamelCase(data);
  }
  
  /**
   * 确认洗练结果
   * @param {string} token - 认证令牌
   * @param {string} equipmentId - 装备ID
   * @param {boolean} confirmed - 是否确认新属性
   * @param {Object} newStats - 新的属性值（仅在confirmed为true时需要）
   * @returns {Promise<Object>} 确认结果
   */
  static async confirmReforge(token, equipmentId, confirmed, newStats = null) {
    const response = await fetch(`${API_BASE_URL}/player/equipment/${equipmentId}/reforge-confirm`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({ confirmed, newStats })
    });

    if (!response.ok) {
      throw new Error('确认洗练结果失败');
    }

    const data = await response.json();
    // 统一转换后端返回的数据字段为小驼峰命名法
    return convertToCamelCase(data);
  }
  
  /**
   * 装备穿戴
   * @param {string} token - 认证令牌
   * @param {string} equipmentId - 装备ID
   * @param {string} slot - 装备槽位
   * @returns {Promise<Object>} 穿戴结果
   */
  static async equipEquipment(token, equipmentId, slot) {
    console.log('[API Service] 调用装备穿戴接口前检查参数:', { 
      equipmentId, 
      slot, 
      tokenAvailable: !!token,
      tokenLength: token ? token.length : 0
    });
    
    const response = await fetch(`${API_BASE_URL}/player/equipment/${equipmentId}/equip`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({ slot })
    });

    console.log('[API Service] 装备穿戴接口响应详情:', {
      status: response.status,
      statusText: response.statusText,
      url: response.url,
      headers: [...response.headers.entries()]
    });

    // 检查响应状态并在出错时提供更多详细信息
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      const errorMessage = errorData.message || `装备穿戴失败 (${response.status} ${response.statusText})`;
      console.error('[API Service] 装备穿戴接口调用失败:', {
        status: response.status,
        statusText: response.statusText,
        errorData,
        errorMessage
      });
      throw new Error(errorMessage);
    }

    const data = await response.json();
    console.log('[API Service] 装备穿戴接口返回结果:', data);
    // 统一转换后端返回的数据字段为小驼峰命名法
    return convertToCamelCase(data);
  }
  
  /**
   * 装备卸下
   * @param {string} token - 认证令牌
   * @param {string} equipmentId - 装备ID
   * @returns {Promise<Object>} 卸下结果
   */
  static async unequipEquipment(token, equipmentId) {
    console.log(`[API Service] 调用装备卸下接口前检查参数:`, { 
      equipmentId, 
      tokenAvailable: !!token,
      tokenLength: token ? token.length : 0
    });
    
    const response = await fetch(`${API_BASE_URL}/player/equipment/${equipmentId}/unequip`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      }
    });

    console.log('[API Service] 装备卸下接口响应详情:', {
      status: response.status,
      statusText: response.statusText,
      url: response.url,
      headers: [...response.headers.entries()]
    });

    // 检查响应状态并在出错时提供更多详细信息
    if (!response.ok) {
      const errorText = await response.text();
      console.error(`装备卸下失败 (${response.status}):`, errorText);
      throw new Error(`装备卸下失败: ${response.status} ${response.statusText}`);
    }

    const result = await response.json();
    console.log('[API Service] 装备卸下接口返回结果:', result);
    // 统一转换后端返回的数据字段为小驼峰命名法
    return convertToCamelCase(result);
  }
  
  /**
   * 出售装备
   * @param {string} token - 认证令牌
   * @param {string} equipmentId - 装备ID
   * @returns {Promise<Object>} 出售结果
   */
  static async sellEquipment(token, equipmentId) {
    const response = await fetch(`${API_BASE_URL}/player/equipment/${equipmentId}`, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      }
    });

    if (!response.ok) {
      throw new Error('装备出售失败');
    }

    const data = await response.json();
    // 统一转换后端返回的数据字段为小驼峰命名法
    return convertToCamelCase(data);
  }
  
  /**
   * 批量出售装备
   * @param {string} token - 认证令牌
   * @param {Object} params - 过滤参数
   * @returns {Promise<Object>} 出售结果
   */
  static async batchSellEquipment(token, params = {}) {
    const response = await fetch(`${API_BASE_URL}/player/equipment/batch-sell`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify(params)
    });

    if (!response.ok) {
      throw new Error('批量出售装备失败');
    }

    const data = await response.json();
    // 统一转换后端返回的数据字段为小驼峰命名法
    return convertToCamelCase(data);
  }

  // 灵宠系统相关方法
  
  /**
   * 出战灵宠
   * @param {string} token - 认证令牌
   * @param {string} petId - 灵宠ID
   * @returns {Promise<Object>} 出战结果
   */
  static async deployPet(token, petId) {
    const response = await fetch(`${API_BASE_URL}/player/pets/${petId}/deploy`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      }
    });
    
    if (!response.ok) {
      throw new Error('出战灵宠失败');
    }
    
    const data = await response.json();
    // 统一转换后端返回的数据字段为小驼峰命名法
    return convertToCamelCase(data);
  }
  
  /**
   * 召回灵宠
   * @param {string} token - 认证令牌
   * @param {string} petId - 灵宠ID
   * @returns {Promise<Object>} 召回结果
   */
  static async recallPet(token, petId) {
    const response = await fetch(`${API_BASE_URL}/player/pets/${petId}/recall`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      }
    });
    
    if (!response.ok) {
      throw new Error('召回灵宠失败');
    }
    
    const data = await response.json();
    // 统一转换后端返回的数据字段为小驼峰命名法
    return convertToCamelCase(data);
  }
  
  /**
   * 升级灵宠
   * @param {string} token - 认证令牌
   * @param {string} petId - 灵宠ID
   * @param {number} essenceCount - 消耗的灵宠精华数量
   * @returns {Promise<Object>} 升级结果
   */
  static async upgradePet(token, petId, essenceCount) {
    const response = await fetch(`${API_BASE_URL}/player/pets/${petId}/upgrade`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({ essenceCount })
    });
    
    if (!response.ok) {
      throw new Error('升级灵宠失败');
    }
    
    const data = await response.json();
    // 统一转换后端返回的数据字段为小驼峰命名法
    return convertToCamelCase(data);
  }
  
  /**
   * 升星灵宠
   * @param {string} token - 认证令牌
   * @param {string} petId - 目标灵宠ID
   * @param {string} foodPetId - 作为材料的灵宠ID
   * @returns {Promise<Object>} 升星结果
   */
  static async evolvePet(token, petId, foodPetId) {
    const response = await fetch(`${API_BASE_URL}/player/pets/${petId}/evolve`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({ foodPetId })
    });
    
    if (!response.ok) {
      throw new Error('升星灵宠失败');
    }
    
    const data = await response.json();
    // 统一转换后端返回的数据字段为小驼峰命名法
    return convertToCamelCase(data);
  }
  
  /**
   * 批量放生灵宠
   * @param {string} token - 认证令牌
   * @param {Object} params - 过滤参数
   * @returns {Promise<Object>} 放生结果
   */
  static async batchReleasePets(token, params = {}) {
    const response = await fetch(`${API_BASE_URL}/player/pets/batch-release`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify(params)
    });
    
    if (!response.ok) {
      throw new Error('批量放生灵宠失败');
    }
    
    const data = await response.json();
    // 统一转换后端返回的数据字段为小驼峰命名法
    return convertToCamelCase(data);
  }
  
  // 排行榜相关方法
  
  /**
   * 获取排行榜数据
   * @returns {Promise<Object>} 排行榜数据
   */
  static async getLeaderboard() {
    const response = await fetch(`${API_BASE_URL}/player/leaderboard`);
    const data = await response.json();
    // 统一转换后端返回的数据字段为小驼峰命名法
    return convertToCamelCase(data);
  }
  
  // 抽奖系统相关方法
  
  /**
   * 执行抽奖
   * @param {string} token - 认证令牌
   * @param {Object} params - 抽奖参数
   * @returns {Promise<Object>} 抽奖结果
   */
  static async drawGacha(token, params) {
    const response = await fetch(`${API_BASE_URL}/gacha/draw`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify(params)
    });
    
    if (!response.ok) {
      throw new Error('抽奖失败');
    }
    
    const data = await response.json();
    // 统一转换后端返回的数据字段为小驼峰命名法
    return convertToCamelCase(data);
  }
  
  /**
   * 执行自动处理操作
   * @param {string} token - 认证令牌
   * @param {Object} params - 自动处理参数
   * @returns {Promise<Object>} 处理结果
   */
  static async processAutoActions(token, params) {
    const response = await fetch(`${API_BASE_URL}/gacha/auto-actions`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify(params)
    });
    
    if (!response.ok) {
      throw new Error('自动处理失败');
    }
    
    const data = await response.json();
    // 统一转换后端返回的数据字段为小驼峰命名法
    return convertToCamelCase(data);
  }
}

export default APIService;