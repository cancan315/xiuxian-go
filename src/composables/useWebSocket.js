// src/composables/useWebSocket.js
// Vue 3 组合式API - WebSocket管理

import { ref, onMounted, onUnmounted } from 'vue';
import { wsManager, subscribeSpiritGrowth, subscribeDungeonEvent, subscribeLeaderboardUpdate, subscribeExplorationEvent } from '@/services/websocket';

/**
 * WebSocket连接和状态管理
 */
export function useWebSocket() {
  const isConnected = ref(false);
  const spiritData = ref(null);
  const dungeonData = ref(null);
  const leaderboardData = ref(null);
  const explorationData = ref(null);
  const connectionStatus = ref('disconnected'); // disconnected, connecting, connected, error

  /**
   * 初始化WebSocket连接
   */
  async function initWebSocket(token, userId) {
    try {
      connectionStatus.value = 'connecting';
      await wsManager.connect(token, userId);
      isConnected.value = wsManager.isConnected;
      connectionStatus.value = 'connected';
      console.log('WebSocket连接成功');
    } catch (error) {
      connectionStatus.value = 'error';
      console.error('WebSocket连接失败:', error);
      throw error;
    }
  }

  /**
   * 订阅灵力增长事件
   */
  function subscribeSpiritGrowthData(callback) {
    return subscribeSpiritGrowth((data, meta) => {
      spiritData.value = data;
      if (callback) callback(data, meta);
    });
  }

  /**
   * 订阅战斗事件
   */
  function subscribeDungeonEventData(callback) {
    return subscribeDungeonEvent((data, meta) => {
      dungeonData.value = data;
      if (callback) callback(data, meta);
    });
  }

  /**
   * 订阅排行榜更新
   */
  function subscribeLeaderboardUpdateData(callback) {
    return subscribeLeaderboardUpdate((data, meta) => {
      leaderboardData.value = data;
      if (callback) callback(data, meta);
    });
  }

  /**
   * 订阅探索事件
   */
  function subscribeExplorationEventData(callback) {
    return subscribeExplorationEvent((data, meta) => {
      explorationData.value = data;
      if (callback) callback(data, meta);
    });
  }

  /**
   * 监听连接状态变化
   */
  function watchConnectionStatus(callback) {
    wsManager.on('connection:open', () => {
      isConnected.value = true;
      connectionStatus.value = 'connected';
      callback?.('open');
    });

    wsManager.on('connection:close', () => {
      isConnected.value = false;
      connectionStatus.value = 'disconnected';
      callback?.('close');
    });

    wsManager.on('connection:error', (error) => {
      connectionStatus.value = 'error';
      callback?.('error', error);
    });
  }

  /**
   * 断开连接
   */
  function disconnect() {
    wsManager.disconnect();
    isConnected.value = false;
    connectionStatus.value = 'disconnected';
  }

  /**
   * 获取连接状态
   */
  function getStatus() {
    return {
      isConnected: wsManager.isConnected,
      connectionStatus: connectionStatus.value,
      spiritData: spiritData.value,
      dungeonData: dungeonData.value,
      leaderboardData: leaderboardData.value,
      explorationData: explorationData.value
    };
  }

  return {
    isConnected,
    connectionStatus,
    spiritData,
    dungeonData,
    leaderboardData,
    explorationData,
    initWebSocket,
    subscribeSpiritGrowthData,
    subscribeDungeonEventData,
    subscribeLeaderboardUpdateData,
    subscribeExplorationEventData,
    watchConnectionStatus,
    disconnect,
    getStatus
  };
}

/**
 * 灵力增长管理
 */
export function useSpiritGrowth() {
  const spiritGrowthEvents = ref([]);
  const currentSpirit = ref(0);
  const totalGainedSpirit = ref(0);

  function handleSpiritGrowth(data) {
    spiritGrowthEvents.value.push({
      ...data,
      receivedAt: new Date()
    });

    // 保留最后100条记录
    if (spiritGrowthEvents.value.length > 100) {
      spiritGrowthEvents.value.shift();
    }

    currentSpirit.value = data.newSpirit;
    totalGainedSpirit.value += data.gainAmount;

    console.log(`灵力增长: +${data.gainAmount.toFixed(2)}, 当前灵力: ${data.newSpirit.toFixed(2)}`);
  }

  function clearHistory() {
    spiritGrowthEvents.value = [];
    totalGainedSpirit.value = 0;
  }

  return {
    spiritGrowthEvents,
    currentSpirit,
    totalGainedSpirit,
    handleSpiritGrowth,
    clearHistory
  };
}

/**
 * 战斗管理
 */
export function useDungeonCombat() {
  const combatLog = ref([]);
  const currentDungeon = ref(null);
  const combatState = ref('idle'); // idle, fighting, victory, defeat

  function handleDungeonEvent(data) {
    combatLog.value.push({
      ...data,
      receivedAt: new Date()
    });

    // 保留最后50条记录
    if (combatLog.value.length > 50) {
      combatLog.value.shift();
    }

    // 更新战斗状态
    switch (data.eventType) {
      case 'start':
        currentDungeon.value = data.dungeon;
        combatState.value = 'fighting';
        break;
      case 'victory':
        combatState.value = 'victory';
        break;
      case 'defeat':
        combatState.value = 'defeat';
        break;
      case 'combat_round':
        combatState.value = 'fighting';
        break;
    }

    console.log(`战斗事件: ${data.eventType} - ${data.message}`);
  }

  function clearLog() {
    combatLog.value = [];
    currentDungeon.value = null;
    combatState.value = 'idle';
  }

  return {
    combatLog,
    currentDungeon,
    combatState,
    handleDungeonEvent,
    clearLog
  };
}

/**
 * 排行榜管理
 */
export function useLeaderboard() {
  const leaderboards = ref({
    spirit: { top10: [], userRank: null },
    power: { top10: [], userRank: null },
    level: { top10: [], userRank: null }
  });

  function handleLeaderboardUpdate(data) {
    const category = data.category;
    
    if (data.type === 'full_refresh') {
      leaderboards.value[category] = {
        top10: data.top10 || [],
        userRank: data.userRank
      };
    } else if (data.type === 'update') {
      leaderboards.value[category].userRank = data.userRank;
    }

    console.log(`排行榜更新 (${category}):`, leaderboards.value[category].userRank);
  }

  function getLeaderboard(category) {
    return leaderboards.value[category] || null;
  }

  return {
    leaderboards,
    handleLeaderboardUpdate,
    getLeaderboard
  };
}

/**
 * 探索管理
 */
export function useExploration() {
  const explorationLog = ref([]);
  const currentExploration = ref(null);
  const explorationProgress = ref(0);

  function handleExplorationEvent(data) {
    explorationLog.value.push({
      ...data,
      receivedAt: new Date()
    });

    // 保留最后50条记录
    if (explorationLog.value.length > 50) {
      explorationLog.value.shift();
    }

    // 更新探索状态
    switch (data.eventType) {
      case 'start':
        currentExploration.value = data.exploreName;
        explorationProgress.value = 0;
        break;
      case 'progress':
        explorationProgress.value = data.progress;
        break;
      case 'complete':
      case 'failure':
        currentExploration.value = null;
        explorationProgress.value = 0;
        break;
    }

    console.log(`探索事件: ${data.eventType} - ${data.message}`);
  }

  function clearLog() {
    explorationLog.value = [];
    currentExploration.value = null;
    explorationProgress.value = 0;
  }

  return {
    explorationLog,
    currentExploration,
    explorationProgress,
    handleExplorationEvent,
    clearLog
  };
}
