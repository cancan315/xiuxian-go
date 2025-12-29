// API服务类，用于与后端进行通信
const API_BASE_URL = '/api';

// 将大驼峰命名转换为小驼峰命名的通用函数
function convertToCamelCase(obj) {
  if (obj === null || obj === undefined) {
    return obj;
  }
  
  // 处理 JSON 字符串，自动解析成对象
  if (typeof obj === 'string') {
    // 排除正常的文本字符串，仅处理 JSON 格式的字符串
    if ((obj.startsWith('{') && obj.endsWith('}')) || (obj.startsWith('[') && obj.endsWith(']'))) {
      try {
        const parsed = JSON.parse(obj);
        return convertToCamelCase(parsed);
      } catch (e) {
        // 不是有效 JSON，作为正常字符串返回
        return obj;
      }
    }
    return obj;
  }
  
  if (typeof obj !== 'object') {
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
        'RequiredRealm': 'requiredRealm',
        'Quality': 'quality',
        'Rarity': 'rarity',
        'PlayerName': 'playerName',
        'PetID': 'petId',
        'SpiritStones': 'spiritStones',
        'Name': 'name',
        'Type': 'type',
        'Slot': 'slot',
        'Details': 'details',
        'Stats': 'stats',
        'Description': 'description',
        'Level': 'level',
        'CombatAttributes': 'combatAttributes',
        'AttackBonus': 'attackBonus',
        'DefenseBonus': 'defenseBonus',
        'HealthBonus': 'healthBonus',
        'Experience': 'experience',
        'MaxExperience': 'maxExperience',
        'IsActive': 'isActive',
        'Star': 'star'
      };
      
      // 使用特殊映射规则或默认转换规则（小驼峰字段直接返回）
      const camelCaseKey = fieldMapping[key] || key;
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
   * 获取玩家累积的灵力增长量
   * @param {string} token - 认证令牌
   * @returns {Promise<Object>} 灵力增长量信息
   */
  static async getPlayerSpiritGain(token) {
    const response = await fetch(`${API_BASE_URL}/player/spirit/gain`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      }
    });
    
    if (!response.ok) {
      throw new Error('获取灵力增长量失败');
    }
    
    return response.json();
  }
  
  /**
   * 应用灵力增长（写入数据库并清空缓存）
   * @param {string} token - 认证令牌
   * @param {number} spiritGain - 灵力增长量
   * @returns {Promise<Object>} 应用结果
   */
  static async applySpiritGain(token, spiritGain) {
    const response = await fetch(`${API_BASE_URL}/player/spirit/apply-gain`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({ spiritGain })
    });
    
    if (!response.ok) {
      throw new Error('应用灵力增长失败');
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
  /**
   * 获取排行榜数据
   * @param {string} type - 排行榜类型: realm(境界), spiritStones(灵石), equipment(装备), pets(灵宠)
   * @returns {Promise<Object>} 排行榜数据
   */
  static async getLeaderboard(type = 'realm') {
    let endpoint = '/player/leaderboard'
    if (type && type !== 'realm') {
      endpoint = `/player/leaderboard/${type}`
    }
    const response = await fetch(`${API_BASE_URL}${endpoint}`);
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
  
  /**
   * 修改玩家道号
   * @param {string} token - 认证令牌
   * @param {string} newName - 新的道号
   * @returns {Promise<Object>} 修改结果
   */
  static async changePlayerName(token, newName) {
    const response = await fetch(`${API_BASE_URL}/player/change-name`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({ newName })
    });
    
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.message || `修改道号失败 (${response.status})`);
    }
    
    const data = await response.json();
    // 统一转换后端返回的数据字段为小驼峰命名法
    return convertToCamelCase(data);
  }
  

  /**
   * 发送POST请求
   * @param {string} url - API端点
   * @param {Object} data - 请求体数据
   * @param {string} token - 认证令牌（可选）
   * @returns {Promise<Object>} 响应数据
   */
  static async post(url, data = {}, token = null) {
    const headers = {
      'Content-Type': 'application/json'
    };
    
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    } else {
      // 尝试从存储中获取token
      const storedToken = localStorage.getItem('authToken');
      if (storedToken) {
        headers['Authorization'] = `Bearer ${storedToken}`;
      }
    }
    
    const response = await fetch(`${API_BASE_URL}${url}`, {
      method: 'POST',
      headers,
      body: JSON.stringify(data)
    });
    
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.message || `请求失败 (${response.status})`);
    }
    
    const responseData = await response.json();
    return convertToCamelCase(responseData);
  }
  
  /**
   * 发送GET请求
   * @param {string} url - API端点
   * @param {Object} params - 查询参数
   * @param {string} token - 认证令牌（可选）
   * @returns {Promise<Object>} 响应数据
   */
  static async get(url, params = {}, token = null) {
    let fetchUrl = `${API_BASE_URL}${url}`;
    
    if (Object.keys(params).length > 0) {
      const searchParams = new URLSearchParams();
      Object.entries(params).forEach(([key, value]) => {
        searchParams.append(key, value);
      });
      fetchUrl += `?${searchParams.toString()}`;
    }
    
    const headers = {
      'Content-Type': 'application/json'
    };
    
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    } else {
      const storedToken = localStorage.getItem('authToken');
      if (storedToken) {
        headers['Authorization'] = `Bearer ${storedToken}`;
      }
    }
    
    const response = await fetch(fetchUrl, {
      method: 'GET',
      headers
    });
    
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.message || `请求失败 (${response.status})`);
    }
    
    const responseData = await response.json();
    return convertToCamelCase(responseData);
  }
  
  /**
   * 发送PUT请求
   * @param {string} url - API端点
   * @param {Object} data - 请求体数据
   * @param {string} token - 认证令牌（可选）
   * @returns {Promise<Object>} 响应数据
   */
  static async put(url, data = {}, token = null) {
    const headers = {
      'Content-Type': 'application/json'
    };
    
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    } else {
      const storedToken = localStorage.getItem('authToken');
      if (storedToken) {
        headers['Authorization'] = `Bearer ${storedToken}`;
      }
    }
    
    const response = await fetch(`${API_BASE_URL}${url}`, {
      method: 'PUT',
      headers,
      body: JSON.stringify(data)
    });
    
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.message || `请求失败 (${response.status})`);
    }
    
    const responseData = await response.json();
    return convertToCamelCase(responseData);
  }
  
  /**
   * 发送DELETE请求
   * @param {string} url - API端点
   * @param {Object} params - 查询参数
   * @param {string} token - 认证令牌（可选）
   * @returns {Promise<Object>} 响应数据
   */
  static async delete(url, params = {}, token = null) {
    let fetchUrl = `${API_BASE_URL}${url}`;
    
    if (Object.keys(params).length > 0) {
      const searchParams = new URLSearchParams();
      Object.entries(params).forEach(([key, value]) => {
        searchParams.append(key, value);
      });
      fetchUrl += `?${searchParams.toString()}`;
    }
    
    const headers = {
      'Content-Type': 'application/json'
    };
    
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    } else {
      const storedToken = localStorage.getItem('authToken');
      if (storedToken) {
        headers['Authorization'] = `Bearer ${storedToken}`;
      }
    }
    
    const response = await fetch(fetchUrl, {
      method: 'DELETE',
      headers
    });
    
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.message || `请求失败 (${response.status})`);
    }
    
    const responseData = await response.json();
    return convertToCamelCase(responseData);
  }
  
  /**
   * 获取修炼数据
   * @param {string} token - 认证令牌
   * @returns {Promise<Object>} 修炼数据
   */
  static async getCultivationData(token) {
    const response = await fetch(`${API_BASE_URL}/cultivation/data`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      }
    });
    
    if (!response.ok) {
      throw new Error('获取修炼数据失败');
    }
    
    const data = await response.json();
    return convertToCamelCase(data);
  }
  
  /**
   * 发送玩家心跳
   * @param {string} playerId - 玩家ID
   * @param {string} token - 认证令牌
   * @returns {Promise<Object>} 心跳响应
   */
  static async playerHeartbeat(playerId, token) {
  //  console.log('[API Service] 发送心跳请求', { playerId, tokenAvailable: !!token, tokenLength: token ? token.length : 0 });
    
    // 确保playerId是字符串类型
    const requestData = { playerId: String(playerId) };
  //  console.log('[API Service] 心跳请求数据', requestData);
    
    const response = await fetch(`${API_BASE_URL}/online/heartbeat`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify(requestData)
    });
    
  //  console.log('[API Service] 心跳响应状态', { status: response.status, statusText: response.statusText });
    
    if (!response.ok) {
      const errorText = await response.text();
      console.error('[API Service] 心跳发送失败详情', { 
        status: response.status, 
        statusText: response.statusText,
        errorText,
        url: response.url,
        headers: [...response.headers.entries()]
      });
      throw new Error(`心跳发送失败: ${response.status} ${response.statusText} - ${errorText}`);
    }
    
    const data = await response.json();
  //  console.log('[API Service] 心跳响应数据', data);
    return convertToCamelCase(data);
  }

  // 探索系统相关方法
  
  /**
   * 开始探索（单次触发）
   * @param {string} token - 认证令符
   * @returns {Promise<Object>} 探索结果
   */
  static async startExploration(token) {
    const response = await fetch(`${API_BASE_URL}/exploration/start`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({})
    });
      
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      // 优先返回后端的 error 字段（业务错误），否则返回 message 字段
      throw new Error(errorData.error || errorData.message || '探索失败');
    }
      
    const data = await response.json();
    return convertToCamelCase(data);
  }
  
  /**
   * 处理探索事件选择
   * @param {string} token - 认证令牌
   * @param {string} eventType - 事件类型
   * @param {Object} choice - 选择对象
   * @returns {Promise<Object>} 处理结果
   */
  static async handleExplorationEventChoice(token, eventType, choice) {
    const response = await fetch(`${API_BASE_URL}/exploration/event-choice`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({ eventType, choice })
    });
    
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.message || '处理事件失败');
    }
    
    const data = await response.json();
    return convertToCamelCase(data);
  }
  // ... 现有所有方法保持不变 ...

  // ======================================
  // 斗法系统相关方法
  // ======================================

  /**
   * 获取斗法双方的完整战斗属性数据
   * @param {string} token - 认证令牌
   * @param {number} playerID - 玩家ID
   * @param {number} opponentID - 对手ID
   * @returns {Promise<Object>} 双方属性数据
   */
  static async getBattleAttributes(token, playerID, opponentID) {
    try {
      console.log('[API Service] 调用获取斗法战斗属性API');
      
      const requestBody = {
        playerID,
        opponentID
      };
      
      console.log('[API Service] 发送的请求体:', JSON.stringify(requestBody));
      console.log('[API Service] playerID 类型:', typeof playerID, '值:', playerID);
      console.log('[API Service] opponentID 类型:', typeof opponentID, '值:', opponentID);
      
      const response = await fetch(`${API_BASE_URL}/duel/battle-attributes`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify(requestBody)
      });
      
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        console.error('[API Service] 获取斗法属性失败:', response.status, errorData);
        return {
          success: false,
          message: errorData.message || '获取斗法属性失败'
        };
      }
      
      const data = await response.json();
      console.log('[API Service] 从后端API获取斗法属性:', data);
      // 统一转换后端返回的数据字段为小驼峰命名法
      return convertToCamelCase(data);
    } catch (error) {
      console.error('[API Service] 获取斗法战斗属性异常:', error);
      return {
        success: false,
        message: '获取斗法战斗属性失败'
      };
    }
  }

  /**
   * 获取斗法道友列表
   * @param {string} token - 认证令牌
   * @param {number} page - 页码（可选，默认为1）
   * @param {number} pageSize - 每页数量（可选，默认为10）
   * @returns {Promise<Object>} 道友列表
   */
  static async getDuelOpponents(token, page = 1, pageSize = 10) {
    try {
      console.log('[API Service] 获取斗法道友列表');
      
      // 调用Go后端API获取对手列表
      const response = await fetch(`${API_BASE_URL}/duel/opponents?page=${page}&pageSize=${pageSize}`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        }
      });
      
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || '获取对手列表失败');
      }
      
      const data = await response.json();
      return convertToCamelCase(data);
    } catch (error) {
      console.error('获取道友列表失败:', error);
      return {
        success: false,
        message: '获取道友列表失败'
      };
    }
  }
  
  /**
   * 获取姐兽列表
   * @param {string} token - 认证令牌
   * @returns {Promise<Object>} 姐兽列表
   */
  static async getMonsters(token) {
    try {
      console.log('[API Service] 获取妖兽列表');
      
      // 返回模拟数据
      return {
        success: true,
        data: {
          monsters: [
            {
              id: 1,
              name: '赤焰虎',
              difficulty: 'normal',
              level: 1,
              health: 150,
              attack: 25,
              defense: 10,
              speed: 15,
              critRate: 0.1,
              dodgeRate: 0.05,
              rewards: '修为150，灵石20，可能掉落虎骨',
              description: '生活在火焰山脉的猛虎，浑身赤红如火'
            },
            {
              id: 2,
              name: '黑水玄蛇',
              difficulty: 'normal',
              level: 2,
              health: 200,
              attack: 30,
              defense: 15,
              speed: 18,
              critRate: 0.15,
              stunRate: 0.1,
              rewards: '修为200，灵石30，可能掉落蛇胆',
              description: '潜伏在深潭中的巨蛇，毒性猛烈'
            },
            {
              id: 3,
              name: '金翅大鹏',
              difficulty: 'hard',
              level: 3,
              health: 300,
              attack: 45,
              defense: 20,
              speed: 30,
              critRate: 0.2,
              dodgeRate: 0.15,
              rewards: '修为300，灵石50，可能掉落鹏羽',
              description: '翱翔天际的神鸟，速度极快'
            }
          ]
        }
      };
    } catch (error) {
      console.error('获取妖兽列表失败:', error);
      return {
        success: false,
        message: '获取妖兽列表失败'
      };
    }
  }

  /**
   * 获取玩家战斗数据
   * @param {string|number} playerId - 玩家ID
   * @param {string} token - 认证令牌
   * @returns {Promise<Object>} 玩家战斗数据
   */
  static async getPlayerBattleData(playerId, token) {
    try {
      console.log('[API Service] 获取玩家战斗数据:', playerId);
      
      // 调用Go后端API获取玩家战斗数据
      const response = await fetch(`${API_BASE_URL}/duel/player/${playerId}/battle-data`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        }
      });
      
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || '获取玩家数据失败');
      }
      
      const data = await response.json();
      return convertToCamelCase(data);
    } catch (error) {
      console.error('获取玩家战斗数据失败:', error);
      return {
        success: false,
        message: '获取玩家数据失败'
      };
    }
  }

  /**
   * 记录战斗结果
   * @param {string} token - 认证令牌
   * @param {Object} battleData - 战斗数据
   * @returns {Promise<Object>} 记录结果
   */
  static async recordBattleResult(token, battleData) {
    try {
      console.log('[API Service] 记录战斗结果:', battleData);
      
      // 调用Go后端API记录战斗结果
      const response = await fetch(`${API_BASE_URL}/duel/record-result`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify(battleData)
      });
      
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || '记录战斗结果失败');
      }
      
      const data = await response.json();
      return convertToCamelCase(data);
    } catch (error) {
      console.error('记录战斗结果失败:', error);
      return {
        success: false,
        message: '记录战斗结果失败'
      };
    }
  }

  /**
   * 获取战斗记录
   * @param {string} token - 认证令牌
   * @param {number} page - 页码（可选，默认为1）
   * @param {number} pageSize - 每页数量（可选，默认为20）
   * @returns {Promise<Object>} 战斗记录
   */
  static async getBattleRecords(token, page = 1, pageSize = 20) {
    try {
      console.log('[API Service] 获取战斗记录');
      
      // 调用Go后端API获取战斗记录
      const response = await fetch(`${API_BASE_URL}/duel/records?page=${page}&pageSize=${pageSize}`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        }
      });
      
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || '获取战斗记录失败');
      }
      
      const data = await response.json();
      return convertToCamelCase(data);
    } catch (error) {
      console.error('获取战斗记录失败:', error);
      return {
        success: false,
        message: '获取战斗记录失败'
      };
    }
  }

  /**
   * 领取战斗奖励
   * @param {string} token - 认证令牌
   * @param {Array} rewards - 奖励列表
   * @returns {Promise<Object>} 领取结果
   */
  static async claimBattleRewards(token, rewards) {
    try {
      console.log('[API Service] 领取战斗奖励:', rewards);
      
      // 调用Go后端API领取奖励
      const response = await fetch(`${API_BASE_URL}/duel/claim-rewards`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({ rewards })
      });
      
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || '领取奖励失败');
      }
      
      const data = await response.json();
      return convertToCamelCase(data);
    } catch (error) {
      console.error('领取奖励失败:', error);
      return {
        success: false,
        message: '领取奖励失败'
      };
    }
  }

  /**
   * 开始PvP战斗
   * @param {string} token - 认证令牌
   * @param {number} opponentId - 对手ID
   * @param {Object} playerData - 玩家战斗数据
   * @param {Object} opponentData - 对手战斗数据
   * @returns {Promise<Object>} 战斗初始化结果
   */
  static async startPvPBattle(token, opponentId, playerData, opponentData) {
    try {
      console.log('[API Service] 开始PvP战斗');
      
      const response = await fetch(`${API_BASE_URL}/duel/start-pvp`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({
          opponentId,
          playerData,
          opponentData
        })
      });
      
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || '开始战斗失败');
      }
      
      const data = await response.json();
      return convertToCamelCase(data);
    } catch (error) {
      console.error('开始PvP战斗失败:', error);
      return {
        success: false,
        message: error.message || '开始战斗失败'
      };
    }
  }

  /**
   * 执行PvP战斗回合
   * @param {string} token - 认证令牌
   * @param {number} opponentId - 对手ID
   * @returns {Promise<Object>} 回合数据
   */
  static async executePvPRound(token, opponentId) {
    try {
      console.log('[API Service] 执行PvP战斗回合');
      
      const response = await fetch(`${API_BASE_URL}/duel/execute-pvp-round`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({ opponentId })
      });
      
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || '执行回合失败');
      }
      
      const data = await response.json();
      return convertToCamelCase(data);
    } catch (error) {
      console.error('执行PvP回合失败:', error);
      return {
        success: false,
        message: '执行回合失败'
      };
    }
  }

  /**
   * 结束PvP战斗
   * @param {string} token - 认证令牌
   * @param {number} opponentId - 对手ID
   * @returns {Promise<Object>} 结束结果
   */
  static async endPvPBattle(token, opponentId) {
    try {
      console.log('[API Service] 结束PvP战斗');
      
      const response = await fetch(`${API_BASE_URL}/duel/end-pvp`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({ opponentId })
      });
      
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || '结束战斗失败');
      }
      
      const data = await response.json();
      return convertToCamelCase(data);
    } catch (error) {
      console.error('结束PvP战斗失败:', error);
      return {
        success: false,
        message: '结束战斗失败'
      };
    }
  }
  
  // 获取默认妖兽数据（开发用）
  static getDefaultMonsters() {
    return [
      {
        id: 1,
        name: '赤焰虎',
        difficulty: 'normal',
        level: 1,
        health: 150,
        attack: 25,
        defense: 10,
        speed: 15,
        critRate: 0.1,
        dodgeRate: 0.05,
        rewards: '修为150，灵石20，可能掉落虎骨',
        description: '生活在火焰山脉的猛虎，浑身赤红如火'
      },
      {
        id: 2,
        name: '黑水玄蛇',
        difficulty: 'normal',
        level: 2,
        health: 200,
        attack: 30,
        defense: 15,
        speed: 18,
        critRate: 0.15,
        stunRate: 0.1,
        rewards: '修为200，灵石30，可能掉落蛇胆',
        description: '潜伏在深潭中的巨蛇，毒性猛烈'
      },
      {
        id: 3,
        name: '金翅大鹏',
        difficulty: 'hard',
        level: 3,
        health: 300,
        attack: 45,
        defense: 20,
        speed: 30,
        critRate: 0.2,
        dodgeRate: 0.15,
        rewards: '修为300，灵石50，可能掉落鹏羽',
        description: '翱翔天际的神鸟，速度极快'
      },
      {
        id: 4,
        name: '九头狮子',
        difficulty: 'hard',
        level: 4,
        health: 400,
        attack: 55,
        defense: 30,
        speed: 20,
        critRate: 0.15,
        counterRate: 0.1,
        rewards: '修为400，灵石70，可能掉落狮心',
        description: '拥有九颗头颅的妖狮，凶猛异常'
      },
      {
        id: 5,
        name: '吞天蟒',
        difficulty: 'nightmare',
        level: 5,
        health: 600,
        attack: 70,
        defense: 40,
        speed: 25,
        critRate: 0.25,
        vampireRate: 0.2,
        rewards: '修为600，灵石100，必定掉落蟒皮',
        description: '传说能吞噬天地的上古妖兽'
      }
    ]
  }
  
  // 获取模拟战斗记录（开发用）
  static getMockBattleRecords() {
    return [
      {
        id: 1,
        battleType: 'pve',
        opponent: '赤焰虎',
        result: '胜利',
        rewards: '修为150，灵石20',
        time: '2024-01-15 10:30:15'
      },
      {
        id: 2,
        battleType: 'pve',
        opponent: '黑水玄蛇',
        result: '失败',
        rewards: '无',
        time: '2024-01-15 11:20:45'
      },
      {
        id: 3,
        battleType: 'pvp',
        opponent: '张三丰',
        result: '胜利',
        rewards: '声望25，灵石15',
        time: '2024-01-14 09:15:30'
      },
      {
        id: 4,
        battleType: 'pvp',
        opponent: '李逍遥',
        result: '失败',
        rewards: '无',
        time: '2024-01-13 14:40:22'
      },
      {
        id: 5,
        battleType: 'pve',
        opponent: '金翅大鹏',
        result: '胜利',
        rewards: '修为300，灵石50，鹏羽1',
        time: '2024-01-12 16:25:18'
      }
    ]
  }
  
  // 获取模拟战斗统计（开发用）
  static getMockBattleStats() {
    return {
      totalBattles: 15,
      wins: 9,
      losses: 6,
      winRate: 60,
      currentWinStreak: 3,
      maxWinStreak: 5
    }
  }
}
export default APIService;