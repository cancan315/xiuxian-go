import APIService from '../services/api';

// Mock storage for token
let authToken = null;

// Set auth token
export const setAuthToken = (token) => {
  console.log('[DB Store] 设置认证令牌:', { tokenLength: token ? token.length : 0 });
  authToken = token;
};

// Clear auth token
export const clearAuthToken = () => {
  console.log('[DB Store] 清除认证令牌');
  authToken = null;
};

// Get auth token
export const getAuthToken = () => {
  console.log('[DB Store] 获取认证令牌:', { tokenAvailable: !!authToken, tokenLength: authToken ? authToken.length : 0 });
  return authToken;
};

export class GameDB {
  static async getData(key) {
    // If not authenticated, return null
    if (!authToken) {
      console.log('[GameDB] 未认证，无法获取数据:', key);
      return null;
    }
    
    try {
      // For player data, fetch from backend
      if (key === 'playerData') {
        const response = await APIService.getPlayerData(authToken);
        return response;
      }
      
      return null;
    } catch (error) {
      console.error('Error fetching data:', error);
      return null;
    }
  }

  static async setData(key, value) {
    // If not authenticated, do nothing
    if (!authToken) {
      console.log('[GameDB] 未认证，无法保存数据:', key);
      return;
    }
    
    try {
      // For player data, save to backend
      if (key === 'playerData') {
        await APIService.savePlayerData(authToken, value);
      }
    } catch (error) {
      console.error('Error saving data:', error);
    }
  }
}