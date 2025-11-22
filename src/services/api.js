// API service to communicate with backend
const API_BASE_URL = 'http://localhost:3001/api';

class APIService {
  // Auth methods
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
  
  static async getUser(token) {
    const response = await fetch(`${API_BASE_URL}/auth/user`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });
    
    return response.json();
  }
  
  // Player data methods
  static async getPlayerData(token) {
    const response = await fetch(`${API_BASE_URL}/player/data`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });
    
    return response.json();
  }
  
  static async savePlayerData(token, data) {
    const response = await fetch(`${API_BASE_URL}/player/data`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify(data)
    });
    
    return response.json();
  }
  
  // Leaderboard methods
  static async getLeaderboard() {
    const response = await fetch(`${API_BASE_URL}/player/leaderboard`);
    return response.json();
  }
}

export default APIService;