<template>
  <n-config-provider :theme="playerInfoStore.isDarkMode ? darkTheme : null">
    <n-message-provider>
      <n-dialog-provider>
        <n-spin :show="isLoading" description="正在加载游戏数据...">
          <n-layout v-if="isAuthenticated && !isLoggingOut">
            <n-layout-header bordered>
              <div class="header-content">
                <n-page-header>
                  <template #title>我的小小修仙界</template>
                  <template #extra>
                    <n-space>
                      <n-button @click="logout">退出游戏</n-button>
                      <n-button quaternary circle @click="toggleDarkMode">
                        <template #icon>
                          <n-icon>
                            <Sunny v-if="playerInfoStore.isDarkMode" />
                            <Moon v-else />
                          </n-icon>
                        </template>
                      </n-button>
                    </n-space>
                  </template>
                </n-page-header>
                <div class="menu-container">
                  <n-scrollbar x-scrollable>
                    <n-menu mode="horizontal" :options="menuOptions" :value="currentView"
                      @update:value="handleMenuClick" :default-value="currentView" />
                  </n-scrollbar>
                </div>
              </div>
            </n-layout-header>
            <n-layout-content>
              <div class="content-wrapper">
                <n-card>
                  <n-space vertical>
                    <n-descriptions bordered>
                      <n-descriptions-item label="道号">
                        {{ playerInfoStore.name }}
                      </n-descriptions-item>
                      <n-descriptions-item label="境界">
                        {{ getRealmName(playerInfoStore.level).name }}
                      </n-descriptions-item>
                      <n-descriptions-item label="修为">
                        {{ playerInfoStore.cultivation }} / {{ playerInfoStore.maxCultivation }}
                      </n-descriptions-item>
                      <n-descriptions-item label="灵力">
                        {{ playerInfoStore.spirit.toFixed(2) }}
                      </n-descriptions-item>
                      <n-descriptions-item label="灵石">
                        {{ playerInfoStore.spiritStones }}
                      </n-descriptions-item>
                      <n-descriptions-item label="强化石">
                        {{ playerInfoStore.reinforceStones }}
                      </n-descriptions-item>
                    </n-descriptions>
                    <n-collapse>
                      <n-collapse-item title="详细信息" name="1">
                        <n-divider>基础属性</n-divider>
                        <n-descriptions bordered :column="2">
                          <n-descriptions-item label="生命值">
                            {{ (playerInfoStore.baseAttributes?.health || 0).toFixed(0) }}
                          </n-descriptions-item>
                          <n-descriptions-item label="攻击力">
                            {{ (playerInfoStore.baseAttributes?.attack || 0).toFixed(0) }}
                          </n-descriptions-item>
                          <n-descriptions-item label="防御力">
                            {{ (playerInfoStore.baseAttributes?.defense || 0).toFixed(0) }}
                          </n-descriptions-item>
                          <n-descriptions-item label="速度">
                            {{ (playerInfoStore.baseAttributes?.speed || 0).toFixed(0) }}
                          </n-descriptions-item>
                        </n-descriptions>
                        <n-divider>战斗属性</n-divider>
                        <n-descriptions bordered :column="3">
                          <n-descriptions-item label="暴击率">
                            {{ (playerInfoStore.combatAttributes?.critRate * 100 || 0).toFixed(1) }}%
                          </n-descriptions-item>
                          <n-descriptions-item label="连击率">
                            {{ (playerInfoStore.combatAttributes?.comboRate * 100 || 0).toFixed(1) }}%
                          </n-descriptions-item>
                          <n-descriptions-item label="反击率">
                            {{ (playerInfoStore.combatAttributes?.counterRate * 100 || 0).toFixed(1) }}%
                          </n-descriptions-item>
                          <n-descriptions-item label="眩晕率">
                            {{ (playerInfoStore.combatAttributes?.stunRate * 100 || 0).toFixed(1) }}%
                          </n-descriptions-item>
                          <n-descriptions-item label="闪避率">
                            {{ (playerInfoStore.combatAttributes?.dodgeRate * 100 || 0).toFixed(1) }}%
                          </n-descriptions-item>
                          <n-descriptions-item label="吸血率">
                            {{ (playerInfoStore.combatAttributes?.vampireRate * 100 || 0).toFixed(1) }}%
                          </n-descriptions-item>
                        </n-descriptions>
                        <n-divider>战斗抗性</n-divider>
                        <n-descriptions bordered :column="3">
                          <n-descriptions-item label="抗暴击">
                            {{ (playerInfoStore.combatResistance?.critResist * 100 || 0).toFixed(1) }}%
                          </n-descriptions-item>
                          <n-descriptions-item label="抗连击">
                            {{ (playerInfoStore.combatResistance?.comboResist * 100 || 0).toFixed(1) }}%
                          </n-descriptions-item>
                          <n-descriptions-item label="抗反击">
                            {{ (playerInfoStore.combatResistance?.counterResist * 100 || 0).toFixed(1) }}%
                          </n-descriptions-item>
                          <n-descriptions-item label="抗眩晕">
                            {{ (playerInfoStore.combatResistance?.stunResist * 100 || 0).toFixed(1) }}%
                          </n-descriptions-item>
                          <n-descriptions-item label="抗闪避">
                            {{ (playerInfoStore.combatResistance?.dodgeResist * 100 || 0).toFixed(1) }}%
                          </n-descriptions-item>
                          <n-descriptions-item label="抗吸血">
                            {{ (playerInfoStore.combatResistance?.vampireResist * 100 || 0).toFixed(1) }}%
                          </n-descriptions-item>
                        </n-descriptions>
                        <n-divider>特殊属性</n-divider>
                        <n-descriptions bordered :column="4">
                          <n-descriptions-item label="强化治疗">
                            {{ (playerInfoStore.specialAttributes?.healBoost * 100 || 0).toFixed(1) }}%
                          </n-descriptions-item>
                          <n-descriptions-item label="强化爆伤">
                            {{ (playerInfoStore.specialAttributes?.critDamageBoost * 100 || 0).toFixed(1) }}%
                          </n-descriptions-item>
                          <n-descriptions-item label="弱化爆伤">
                            {{ (playerInfoStore.specialAttributes?.critDamageReduce * 100 || 0).toFixed(1) }}%
                          </n-descriptions-item>
                          <n-descriptions-item label="最终增伤">
                            {{ (playerInfoStore.specialAttributes?.finalDamageBoost * 100 || 0).toFixed(1) }}%
                          </n-descriptions-item>
                          <n-descriptions-item label="最终减伤">
                            {{ (playerInfoStore.specialAttributes?.finalDamageReduce * 100 || 0).toFixed(1) }}%
                          </n-descriptions-item>
                          <n-descriptions-item label="战斗属性提升">
                            {{ (playerInfoStore.specialAttributes?.combatBoost * 100 || 0).toFixed(1) }}%
                          </n-descriptions-item>
                          <n-descriptions-item label="战斗抗性提升">
                            {{ (playerInfoStore.specialAttributes?.resistanceBoost * 100 || 0).toFixed(1) }}%
                          </n-descriptions-item>
                        </n-descriptions>
                      </n-collapse-item>
                    </n-collapse>
                    <n-progress type="line"
                      :percentage="Number(((playerInfoStore.cultivation / playerInfoStore.maxCultivation) * 100).toFixed(2))"
                      indicator-text-color="rgba(255, 255, 255, 0.82)" rail-color="rgba(32, 128, 240, 0.2)"
                      color="#2080f0" :show-indicator="true" indicator-placement="inside" processing />
                  </n-space>
                </n-card>

                <!-- 在此处显示选中的视图 -->
                <component :is="currentViewComponent" v-if="currentViewComponent" :key="currentView" />
              </div>
            </n-layout-content>
          </n-layout>
          <div v-else>
            <router-view />
          </div>
        </n-spin>
      </n-dialog-provider>
    </n-message-provider>
  </n-config-provider>
</template>

<script setup>
import { useRouter, useRoute } from 'vue-router'
// 修改为使用模块化store
import { usePlayerInfoStore } from './stores/playerInfo'
import { createDiscreteApi } from 'naive-ui'
import { h, onMounted, onUnmounted, ref, computed, watch } from 'vue'
import { NIcon, darkTheme } from 'naive-ui'
import {
  BookOutline,
  FlaskOutline,
  CompassOutline,
  TrophyOutline,
  SettingsOutline,
  GiftOutline,
  HomeOutline,
  HappyOutline
} from '@vicons/ionicons5'
import { Moon, Sunny, Flash } from '@vicons/ionicons5'
import { getRealmName } from './plugins/realm'
import { getAuthToken, clearAuthToken } from './stores/db'
import APIService from './services/api'
import { useWebSocket, useSpiritGrowth } from './composables/useWebSocket'

// 导入各视图组件
import Cultivation from './views/Cultivation.vue'
import Inventory from './views/Inventory.vue'
import Exploration from './views/Exploration.vue'
import Settings from './views/Settings.vue'
import Alchemy from './views/Alchemy.vue'
import Dungeon from './views/Dungeon.vue'
import Gacha from './views/Gacha.vue'
import Leaderboard from './views/Leaderboard.vue'

const { message } = createDiscreteApi(['message'])
const router = useRouter()
const route = useRoute()
// 使用模块化store替代原来的usePlayerStore
const playerInfoStore = usePlayerInfoStore()

const menuOptions = ref([])
const isLoading = ref(true) // 添加加载状态
const isLoggingOut = ref(false) // 添加登出状态
const currentView = ref('cultivation') // 默认显示修炼页面

// Check if user is authenticated
const isAuthenticated = computed(() => {
  return !!getAuthToken()
})

// 计算当前应该显示的组件
const currentViewComponent = computed(() => {
  switch (currentView.value) {
    case 'cultivation':
      return Cultivation
    case 'inventory':
      return Inventory
    case 'exploration':
      return Exploration
    case 'settings':
      return Settings
    case 'alchemy':
      return Alchemy
    case 'dungeon':
      return Dungeon
    case 'gacha':
      return Gacha
    case 'leaderboard':
      return Leaderboard
    default:
      return Cultivation
  }
})

// 图标
const renderIcon = icon => {
  return () => h(NIcon, null, { default: () => h(icon) })
}

const menuItems = [
  { label: '修炼', key: 'cultivation', icon: BookOutline },
  { label: '背包', key: 'inventory', icon: FlaskOutline },
  { label: '抽奖', key: 'gacha', icon: GiftOutline },
  { label: '炼丹', key: 'alchemy', icon: FlaskOutline },
  { label: '探索', key: 'exploration', icon: CompassOutline },
  { label: '秘境', key: 'dungeon', icon: Flash },
  { label: '排行榜', key: 'leaderboard', icon: TrophyOutline },
  { label: '设置', key: 'settings', icon: SettingsOutline }
]

const getMenuOptions = () => {
  menuOptions.value = menuItems.map(item => ({
    ...item,
    icon: renderIcon(item.icon)
  }))
}

// 初始化数据加载
const getPlayerData = async () => {
  const token = getAuthToken()
  if (token) {
    try {
      const data = await APIService.getPlayerData(token)

      // Load user data
      if (data.user) {
        playerInfoStore.$patch(data.user)
      }

      // Load inventory items
      if (data.items) {
        playerInfoStore.items = data.items
      }

      // Load pets
      if (data.pets) {
        playerInfoStore.pets = data.pets
      }

      // Load herbs
      if (data.herbs) {
        playerInfoStore.herbs = data.herbs
      }

      // Load pills
      if (data.pills) {
        playerInfoStore.pills = data.pills
      }

      // Load inventory data (including spirit stones)
      if (data.user) {
        console.log('[App.vue] 加载玩家资源数据:', {
          灵石: data.user.spiritStones,
          强化石: data.user.reinforceStones,
          洗炼石: data.user.refinementStones,
          灵兽精华: data.user.petEssence
        });
        playerInfoStore.spiritStones = data.user.spiritStones
        playerInfoStore.reinforceStones = data.user.reinforceStones
        playerInfoStore.refinementStones = data.user.refinementStones
        playerInfoStore.petEssence = data.user.petEssence
      }

      isLoading.value = false
      getMenuOptions()
    } catch (error) {
      console.error('初始化玩家数据失败:', error)
      isLoading.value = false
      getMenuOptions()
    }
  } else {
    isLoading.value = false
    getMenuOptions()
  }
}

// 灵力获取相关配置
const baseGainRate = 1 // 基础灵力获取率
const ws = useWebSocket()
const spirit = useSpiritGrowth()
let handleBeforeUnload = null  // ⋆⋆ 先定义引用以便卸载时离检

onMounted(async () => {
  // 获取菜单选项和玩家数据
  getMenuOptions()
  getPlayerData()

  // 监听页面刷新/关闭事件，执行退出游戏操作
  handleBeforeUnload = (event) => {
    // 执行退出游戏操作
    logout()
  }

  // 添加事件监听器
  window.addEventListener('beforeunload', handleBeforeUnload)

})

// 监听 playerInfoStore.id 的变化，当玩家登录成功后初始化 WebSocket
let wsInitialized = false
let previousId = 0  // 追踪前一个id，判断是ID变更还是退出
let heartbeatTimer = null  // 心跳定时器

watch(() => playerInfoStore.id, async (newId) => {
  // 判断是ID从有值变成0 (退出场景)
  if (previousId > 0 && newId === 0) {
    console.log('[App.vue] 检测到玩家登出, playerInfoStore.id置为0')
    wsInitialized = false
    // 退出宜在logout中已经调用了ws.disconnect()，此处不需要重复
    previousId = newId
    return
  }

  // ID变更场景 (例如切换账户)
  if (previousId > 0 && newId !== previousId && newId > 0) {
    console.log('[App.vue] 检测到玩家ID变更', { previousId, newId })
    // 先断开旧的连接
    ws.disconnect()
    wsInitialized = false
  }

  previousId = newId

  // 不处理ID为0的场景 (退出场景)
  if (newId && !wsInitialized) {
    wsInitialized = true
    const token = getAuthToken()
    console.log('[App.vue] 监测到playerInfoStore.id变化，WebSocket初始化检查', { token: !!token, playerId: newId })
    if (token && newId) {
      try {
        console.log('[App.vue] 开始连接WebSocket', { userId: newId })
        await ws.initWebSocket(token, newId)
        console.log('[App.vue] WebSocket连接成功', { wsInstance: !!ws })
        // 重新订阅灵力增长事件
        reSubscribeSpiritGrowth()

        // 启动心跳定时器（每3秒发送一次心跳）
        startHeartbeatTimer(newId, token)
      } catch (error) {
        console.error('[App.vue] WebSocket初始化失败:', error)
        wsInitialized = false
      }
    }
  }
})

// 注册卸载钩子在顶级作用域
onUnmounted(() => {
  // 清理事件监听器
  if (handleBeforeUnload) {
    window.removeEventListener('beforeunload', handleBeforeUnload)
  }
  // 断开WebSocket连接
  ws.disconnect()
  // 停止心跳定时器
  stopHeartbeatTimer()
})

// 菜单 key 到提示文本的映射
const menuKeyToMessage = {
  cultivation: '打开修炼',
  inventory: '打开背包',
  gacha: '打开抽奖',
  alchemy: '打开炼丹',
  exploration: '打开探索',
  dungeon: '进入秘境',
  leaderboard: '查看排行榜',
  settings: '打开设置'
}

// 菜单点击事件
const handleMenuClick = (key) => {
  currentView.value = key
  const msg = menuKeyToMessage[key]
  if (msg) {
    message.success(msg)
  }
  console.log('[App.vue] 菜单点击：', key)

  currentView.value = key
  console.log('[App.vue] currentView 已更新为：', currentView.value)
}

// 监听菜单切换
watch(() => currentView.value, (newView) => {
  console.log('[App.vue] currentView 已改变为：', newView, '对应的组件：', currentViewComponent.value?.name || 'Unknown')
})

// 重新订阅灵力增长事件
let spiritUnsubscribe = null
const reSubscribeSpiritGrowth = () => {
  // 如果已有订阅，先取消
  if (spiritUnsubscribe) {
    spiritUnsubscribe()
  }

  // 重新订阅
  console.log('[App.vue] 重新订阅灵力增长事件')
  spiritUnsubscribe = ws.subscribeSpiritGrowthData((data) => {
    //onsole.log('[App.vue] 收到灵力增长消息', {
    // gainAmount: data.gainAmount?.toFixed(2),
    // newSpirit: data.newSpirit?.toFixed(2),
    // oldSpirit: data.oldSpirit?.toFixed(2)
    //)
    //spirit.handleSpiritGrowth(data)
    // 更新playerInfoStore中的灵力值
    playerInfoStore.spiritGainRate = data.gainAmount
    playerInfoStore.spirit = data.newSpirit
    
    //console.log(`灵力自动增长: +${data.gainAmount.toFixed(2)}, 当前灵力: ${data.newSpirit.toFixed(2)}`)
  })
  
  console.log('[App.vue] 灵力增长事件重新订阅完成')
}

// 启动心跳定时器
const startHeartbeatTimer = (playerId, token) => {
  console.log('[App.vue] 启动心跳定时器', { playerId, tokenAvailable: !!token });
  // 先清除已有的定时器
  if (heartbeatTimer) {
    clearInterval(heartbeatTimer)
  }

  // 每3秒发送一次心跳
  heartbeatTimer = setInterval(async () => {
    try {
      //  console.log('[App.vue] 发送心跳', { playerId });
      await APIService.playerHeartbeat(playerId, token)
      // console.log('[App.vue] 心跳发送成功')
      
    } catch (error) {
      console.error('[App.vue] 心跳发送失败:', error)
    }
  }, 3000)
}

// 停止心跳定时器
const stopHeartbeatTimer = () => {
  if (heartbeatTimer) {
    clearInterval(heartbeatTimer)
    heartbeatTimer = null
  }
}

// 退出游戏
const logout = async () => {
  // 设置登出状态
  isLoggingOut.value = true

  // ✅ 需要在接下来的操作之前，先保存玩家ID
  const playerId = playerInfoStore.id

  // ✅ 需要立即断开WebSocket连接，避免watch监听到playerInfoStore.id变化后进行重连
  ws.disconnect()
  // 停止心跳定时器
  stopHeartbeatTimer()

  // 通知后端玩家已离线
  try {
    if (playerId) {
      // 调用API通知后端玩家离线
      await APIService.playerOffline(String(playerId))  // ✅ 转换为字符串，与后端一致
    }
  } catch (error) {
    console.error('通知后端玩家离线失败:', error)
  }

  // ✅ 清除认证令牌
  clearAuthToken()
  // 重置玩家状态
  playerInfoStore.$reset()
  // 跳转到登录页面
  router.push('/login')
}



// 切换暗黑模式
const toggleDarkMode = () => {
  playerInfoStore.toggleDarkMode()
}
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

:root {
  --n-color: rgb(16, 16, 20);
  --n-text-color: rgba(255, 255, 255, 0.82);
}

html.dark {
  background-color: var(--n-color);
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans',
    'Helvetica Neue', sans-serif;
}

.n-config-provider,
.n-layout {
  height: 100%;
  min-height: 100vh;
}

.header-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 16px;
}

.menu-container {
  width: 100%;
  overflow: hidden;
}

.menu-container .n-scrollbar {
  overflow: hidden;
}

.menu-container .n-scrollbar>.n-scrollbar-container {
  overflow-x: auto !important;
  overflow-y: hidden;
}

.content-wrapper {
  max-width: 1200px;
  margin: 0 auto;
  padding: 16px;
}

.n-card {
  margin-bottom: 16px;
}

.footer-content {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 12px;
}

.n-page-header__title {
  padding: 16px 0;
  margin: 0 16px;
}

::-webkit-scrollbar {
  width: 12px;
  height: 12px;
}

::-webkit-scrollbar-track {
  background-color: rgba(0, 0, 0, 0.03);
}

::-webkit-scrollbar-thumb {
  background-color: rgba(0, 0, 0, 0.2);
  border-radius: 6px;
  border: 3px solid transparent;
  background-clip: padding-box;
}

::-webkit-scrollbar-thumb:hover {
  background-color: rgba(0, 0, 0, 0.3);
}

html.dark ::-webkit-scrollbar-track {
  background-color: rgba(255, 255, 255, 0.03);
}

html.dark ::-webkit-scrollbar-thumb {
  background-color: rgba(255, 255, 255, 0.2);
}

html.dark ::-webkit-scrollbar-thumb:hover {
  background-color: rgba(255, 255, 255, 0.3);
}
</style>