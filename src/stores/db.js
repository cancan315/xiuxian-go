import APIService from '../services/api';

// Mock storage for token
let authToken = localStorage.getItem('authToken');

// Set auth token
export const setAuthToken = (token) => {
  authToken = token;
  localStorage.setItem('authToken', token);
};

// Clear auth token
export const clearAuthToken = () => {
  authToken = null;
  localStorage.removeItem('authToken');
};

// Get auth token
export const getAuthToken = () => {
  return authToken;
};

export class GameDB {
  static async getData(key) {
    // If not authenticated, return null
    if (!authToken) {
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