<template>
  <n-config-provider :theme="playerInfoStore.isDarkMode ? darkTheme : null">
    <n-message-provider>
      <n-dialog-provider>
        <n-spin :show="isLoading" description="正在加载游戏数据...">
          <!-- ✅ 将 style 移到这个 div 上，而不是组件上 -->
          <div style="min-height: 100vh; width: 100%; display: flex; flex-direction: column;">
            <n-layout v-if="isAuthenticated && !isLoggingOut">
              <n-layout-header bordered>
                <div class="header-content">
                  <n-page-header>
                    <template #title>我的小小修仙界   欢迎道友，请入Q群: 755301571 恳请道友邀请好友一起修仙</template>
                    <template #extra>
                      <n-space align="center">
                        <n-popover trigger="click" placement="bottom">
                          <template #trigger>
                            <n-tag type="success" style="cursor: pointer;">
                              在线: {{ onlineCount }} 人
                            </n-tag>
                          </template>
                          <div style="max-height: 300px; overflow-y: auto;">
                            <n-list bordered size="small" style="min-width: 150px;">
                              <n-list-item v-for="player in onlinePlayers" :key="player.playerId">
                                {{ player.name }}
                              </n-list-item>
                              <n-list-item v-if="onlinePlayers.length === 0">
                                <n-text depth="3">暂无在线道友</n-text>
                              </n-list-item>
                            </n-list>
                          </div>
                        </n-popover>
                        <n-button type="warning" @click="openCheckInModal">
                          🎁 签到
                        </n-button>
                        <n-button @click="logout">退出游戏</n-button>
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
                          {{ playerInfoStore.playerName }}
                        </n-descriptions-item>
                        <n-descriptions-item label="境界">
                          {{ getRealmName(playerInfoStore.level).name }}
                        </n-descriptions-item>
                        <n-descriptions-item label="修为">
                          {{ playerInfoStore.cultivation }} / {{ playerInfoStore.maxCultivation }}
                        </n-descriptions-item>
                        <n-descriptions-item label="灵根">
                          {{ getSpiritRootName(playerInfoStore.spiritRate) }}
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
                        <n-descriptions-item label="洗炼石">
                          {{ playerInfoStore.refinementStones }}
                        </n-descriptions-item>
                        <n-descriptions-item label="灵兽精华">
                          {{ playerInfoStore.petEssence }}
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
            <div v-else style="min-height: 100vh; display: flex; flex-direction: column;">
              <router-view style="flex: 1;" />
            </div>
          </div>
        </n-spin>

        <!-- 签到弹窗 -->
        <n-modal v-model:show="showCheckInModal" preset="card" title="🎁 每日签到" style="width: 600px;">
          <n-space vertical>
            <n-alert type="info" :bordered="false">
              连续签到7天可获得丰厚奖励，中断需从第1天重新开始
            </n-alert>
            
            <div style="display: flex; gap: 8px; justify-content: space-between;">
              <div 
                v-for="(reward, index) in checkInStatus.rewards" 
                :key="index"
                :style="{
                  flex: '1',
                  textAlign: 'center',
                  padding: '12px 8px',
                  borderRadius: '8px',
                  border: '1px solid',
                  backgroundColor: index < checkInStatus.checkInDay ? '#18a05822' : (index === checkInStatus.checkInDay && !checkInStatus.hasCheckedInToday ? '#f0a02022' : 'transparent'),
                  borderColor: index < checkInStatus.checkInDay ? '#18a058' : (index === checkInStatus.checkInDay && !checkInStatus.hasCheckedInToday ? '#f0a020' : '#555')
                }"
              >
                <div style="font-size: 12px; color: #999;">第{{ index + 1 }}天</div>
                <div style="font-size: 16px; font-weight: bold; color: #f0a020; margin: 4px 0;">{{ reward }}</div>
                <div style="font-size: 10px; color: #888;">灵石</div>
                <div style="height: 20px; margin-top: 4px;">
                  <n-tag v-if="index < checkInStatus.checkInDay" type="success" size="tiny">✓</n-tag>
                  <n-tag v-else-if="index === checkInStatus.checkInDay && checkInStatus.hasCheckedInToday" type="success" size="tiny">今日</n-tag>
                </div>
              </div>
            </div>
            
            <n-divider style="margin: 12px 0;" />
            
            <div style="text-align: center;">
              <div v-if="checkInStatus.hasCheckedInToday" style="color: #18a058; margin-bottom: 12px;">
                ✅ 今日已签到（连续第 {{ checkInStatus.checkInDay }} 天）
              </div>
              <div v-else style="color: #666; margin-bottom: 12px;">
                今日签到可获得 <span style="color: #f0a020; font-weight: bold;">{{ checkInStatus.nextReward }}</span> 灵石
              </div>
              <n-button 
                type="warning" 
                size="large" 
                block
                :disabled="checkInStatus.hasCheckedInToday"
                :loading="isCheckingIn"
                @click="doCheckIn"
              >
                {{ checkInStatus.hasCheckedInToday ? '明日再来' : '立即签到' }}
              </n-button>
            </div>
          </n-space>
        </n-modal>
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
  BriefcaseOutline,
  FlaskOutline,
  CompassOutline,
  TrophyOutline,
  SettingsOutline,
  GiftOutline,
  HomeOutline,
  HappyOutline,
  FlashOutline  // 新增斗法图标
} from '@vicons/ionicons5'
import { Moon, Sunny } from '@vicons/ionicons5'
import { getRealmName } from './plugins/realm'
import { getAuthToken, clearAuthToken } from './stores/db'
import APIService from './services/api'

// 导入各视图组件
import Cultivation from './views/Cultivation.vue'
import Inventory from './views/Inventory.vue'
import Exploration from './views/Exploration.vue'
import Settings from './views/Settings.vue'
import Alchemy from './views/Alchemy.vue'
// Dungeon component removed - dungeon feature deleted
import Gacha from './views/Gacha.vue'
import Leaderboard from './views/Leaderboard.vue'
import Duel from './views/Duel.vue'  // 新增斗法页面组件

// ================================
// 🌙 跟随系统暗黑模式
// ================================

let mediaQuery = null

const applySystemTheme = (isDark) => {
  playerInfoStore.isDarkMode = isDark

  // 同步 html class（用于你 CSS 里的 html.dark）
  const html = document.documentElement
  if (isDark) {
    html.classList.add('dark')
  } else {
    html.classList.remove('dark')
  }
}

onMounted(() => {
  // 监听系统主题
  mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')

  // 初始化时立即应用一次
  applySystemTheme(mediaQuery.matches)

  // 监听系统主题变化
  mediaQuery.addEventListener('change', (e) => {
    applySystemTheme(e.matches)
  })
})

onUnmounted(() => {
  if (mediaQuery) {
    mediaQuery.removeEventListener('change', () => {})
    mediaQuery = null
  }
})


const { message } = createDiscreteApi(['message'])
const router = useRouter()
const route = useRoute()
// 使用模块化store替代原来的usePlayerStore
const playerInfoStore = usePlayerInfoStore()

const menuOptions = ref([])
const isLoading = ref(true) // 添加加载状态
const isLoggingOut = ref(false) // 添加登出状态
const currentView = ref('cultivation') // 默认显示修炼页面

// 在线玩家状态
const onlinePlayers = ref([])
const onlineCount = ref(0)
let onlinePlayersTimer = null

// 签到系统状态
const showCheckInModal = ref(false)
const checkInStatus = ref({
  checkInDay: 0,
  hasCheckedInToday: false,
  nextReward: 1000,
  rewards: [1000, 2000, 3000, 4000, 5000, 6000, 10000]
})
const isCheckingIn = ref(false)

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

    case 'gacha':
      return Gacha
    case 'leaderboard':
      return Leaderboard
    case 'duel':  // 新增斗法页面
      return Duel
    default:
      return Cultivation
  }
})

// 图标
const renderIcon = icon => {
  return () => h(NIcon, null, { default: () => h(icon) })
}

const getSpiritRootName = (value) => {
  const map = {
    1: '五灵根',
    2: '四灵根',
    3: '三灵根',
    4: '二灵根',
    5: '异灵根',
    6: '天灵根'
  }
  return map[value] || '未知灵根'
}

const menuItems = [
  { label: '修炼', key: 'cultivation', icon: BookOutline },
  { label: '背包', key: 'inventory', icon: BriefcaseOutline },
  { label: '抽奖', key: 'gacha', icon: GiftOutline },
  { label: '探索', key: 'exploration', icon: CompassOutline },
  { label: '斗法', key: 'duel', icon: FlashOutline },
  { label: '炼丹', key: 'alchemy', icon: FlaskOutline },
  { label: '排行榜', key: 'leaderboard', icon: TrophyOutline },
  { label: '设置', key: 'settings', icon: SettingsOutline }
]

const getMenuOptions = () => {
  menuOptions.value = menuItems.map(item => ({
    ...item,
    icon: renderIcon(item.icon)
  }))
}

const syncCultivationData = async () => {
    try {
      const token = getAuthToken();
      
      // 获取修炎消耗和获得数据
      const response = await APIService.getCultivationData(token)
      if (response.success) {
      //  console.log('[Cultivation] 修炎消耗:', response.data.spiritCost, '获得:', response.data.cultivationGain)
        playerInfoStore.level = response.data.level // 境界等级
        playerInfoStore.realm = response.data.realm // 境界
        playerInfoStore.cultivation = response.data.cultivation // 当前修为
        playerInfoStore.maxCultivation = response.data.maxCultivation // 最大修为
        // playerInfoStore.spirit = response.data.spirit // 当前灵力
        playerInfoStore.cultivationCost = response.data.spiritCost       // 修炎消耗灵力
        playerInfoStore.cultivationGain = response.data.cultivationGain // 修炎获得修为
        playerInfoStore.spiritRate = response.data.spiritRate // 灵力获取倍率
        playerInfoStore.spiritStones = response.data.spiritStones // 灵石数量
        playerInfoStore.reinforceStones = response.data.reinforceStones // 强化石数量
        // ✅ 新增：洗练石和灵宠精华
        playerInfoStore.refinementStones = response.data.refinementStones // 洗练石数量
        playerInfoStore.petEssence = response.data.petEssence // 灵宠精华数量
        
        // ✅ 新增：同步属性数据
        if (response.data.baseAttributes) {
          playerInfoStore.baseAttributes = response.data.baseAttributes
        }
        if (response.data.combatAttributes) {
          playerInfoStore.combatAttributes = response.data.combatAttributes
        }
        if (response.data.combatResistance) {
          playerInfoStore.combatResistance = response.data.combatResistance
        }
        if (response.data.specialAttributes) {
          playerInfoStore.specialAttributes = response.data.specialAttributes
        }
      }
    } catch (error) {
      console.error('同步修为数据失败:', error)
    }
  }

// 初始化数据加载
const getPlayerData = async () => {
  const token = getAuthToken()
  if (token) {
    try {
      const response = await APIService.getCultivationData(token)
      if (response.success) {
      //  console.log('[Cultivation] 修炼消耗:', response.data.spiritCost, '获得:', response.data.cultivationGain)
        
        // ✅ 设置用户ID并标记上线（页面刷新后重新标记）
        if (response.data.id) {
          playerInfoStore.id = response.data.id
          // 重新标记玩家上线，创建Redis记录
          try {
            await APIService.playerOnline(String(response.data.id))
            console.log('[App.vue] 页面刷新后重新标记玩家上线:', response.data.id)
            // ✅ 登录后立即更新在线玩家数量
            await fetchOnlinePlayers()
          } catch (onlineErr) {
            console.warn('[App.vue] 标记玩家上线失败:', onlineErr)
          }
        }
        
        playerInfoStore.playerName = response.data.playerName // 玩家名称
        playerInfoStore.level = response.data.level // 境界等级
        playerInfoStore.realm = response.data.realm // 境界
        playerInfoStore.cultivation = response.data.cultivation // 当前修为
        playerInfoStore.maxCultivation = response.data.maxCultivation // 最大修为
        playerInfoStore.spirit = response.data.spirit // 当前灵力
        playerInfoStore.cultivationCost = response.data.spiritCost       // 修炼消耗灵力
        playerInfoStore.cultivationGain = response.data.cultivationGain // 修炼获得修为
        playerInfoStore.spiritRate = response.data.spiritRate // 灵力获取倍率
        playerInfoStore.spiritStones = response.data.spiritStones // 灵石数量
        playerInfoStore.reinforceStones = response.data.reinforceStones // 强化石数量
        
        // ✅ 新增：初始化时同步属性数据
        if (response.data.baseAttributes) {
          playerInfoStore.baseAttributes = response.data.baseAttributes
          console.log('[App.vue] 已加载基础属性:', response.data.baseAttributes)
        }
        if (response.data.combatAttributes) {
          playerInfoStore.combatAttributes = response.data.combatAttributes
          console.log('[App.vue] 已加载战斗属性:', response.data.combatAttributes)
        }
        if (response.data.combatResistance) {
          playerInfoStore.combatResistance = response.data.combatResistance
          console.log('[App.vue] 已加载战斗抗性:', response.data.combatResistance)
        }
        if (response.data.specialAttributes) {
          playerInfoStore.specialAttributes = response.data.specialAttributes
          console.log('[App.vue] 已加载特殊属性:', response.data.specialAttributes)
        }
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

// ==================== 签到系统 ====================

// 打开签到弹窗
const openCheckInModal = async () => {
  const token = getAuthToken()
  if (!token) return
  
  const result = await APIService.getCheckInStatus(token)
  if (result.success) {
    checkInStatus.value = {
      checkInDay: result.checkInDay,
      hasCheckedInToday: result.hasCheckedInToday,
      nextReward: result.nextReward,
      rewards: result.rewards
    }
  }
  showCheckInModal.value = true
}

// 执行签到
const doCheckIn = async () => {
  if (isCheckingIn.value || checkInStatus.value.hasCheckedInToday) return
  
  const token = getAuthToken()
  if (!token) return
  
  isCheckingIn.value = true
  const result = await APIService.doCheckIn(token)
  isCheckingIn.value = false
  
  if (result.success) {
    message.success(`签到成功！获得 ${result.reward} 灵石`)
    checkInStatus.value.hasCheckedInToday = true
    checkInStatus.value.checkInDay = result.checkInDay
    playerInfoStore.spiritStones = result.spiritStones
  } else {
    message.error(result.message || '签到失败')
  }
}

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

// 心跳定时器和灵力定期同步
let heartbeatTimer = null  // 心跳定时器
let spiritSyncTimer = null  // 灵力同步定时器

watch(() => playerInfoStore.id, async (newId) => {
  if (newId > 0) {
    const token = getAuthToken()
    if (token) {
      // 启动心跳和灵力同步定时器
      startHeartbeatTimer(newId, token)
      startSpiritSyncTimer(token)
      // 启动在线玩家刷新定时器
      startOnlinePlayersTimer()
    }
  }
})

// 注册卸载钩子在顶级作用域
onUnmounted(() => {
  // 清理事件监听器
  if (handleBeforeUnload) {
    window.removeEventListener('beforeunload', handleBeforeUnload)
  }
  // 停止定时器
  stopHeartbeatTimer()
  stopSpiritSyncTimer()
  stopOnlinePlayersTimer()
})

// ========== 在线玩家相关函数 ==========

// 获取在线玩家列表
const fetchOnlinePlayers = async () => {
  try {
    const response = await fetch(`${import.meta.env.VITE_API_BASE_URL || ''}/api/online/players`)
    if (response.ok) {
      const data = await response.json()
      onlinePlayers.value = data.players || []
      onlineCount.value = data.count || 0
    }
  } catch (error) {
    console.error('获取在线玩家失败:', error)
  }
}

// 启动在线玩家刷新定时器（30秒刷新一次）
const startOnlinePlayersTimer = () => {
  // 立即获取一次
  fetchOnlinePlayers()
  // 每30秒刷新
  onlinePlayersTimer = setInterval(fetchOnlinePlayers, 30000)
}

// 停止在线玩家刷新定时器
const stopOnlinePlayersTimer = () => {
  if (onlinePlayersTimer) {
    clearInterval(onlinePlayersTimer)
    onlinePlayersTimer = null
  }
}

// 菜单 key 到提示文本的映射
const menuKeyToMessage = {
  cultivation: '进入修炼',
  inventory: '打开背包',
  gacha: '打开抽奖',
  alchemy: '进入丹房',
  exploration: '打开探索',
  duel: '进入斗法',
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

// 灵力定期同步
const startSpiritSyncTimer = (token) => {
  // 清除已有的定时器
  if (spiritSyncTimer) {
    clearInterval(spiritSyncTimer)
  }

  // 每10秒同步一次玩家灵力数据
  spiritSyncTimer = setInterval(async () => {
    try {
      // ✅ 新流程：先获取Redis中累积的灵力增长量
      const gainResponse = await APIService.getPlayerSpiritGain(token)
      console.log('[Spirit] 获取灵力增长响应:', gainResponse)
      
      if (gainResponse.success) {
        const spiritGain = gainResponse.spiritGain || 0
        if (spiritGain > 0) {
          // 将灵力增长量应用到数据库
          const applyResponse = await APIService.applySpiritGain(token, spiritGain)
          if (applyResponse.success) {
            // 更新本地状态
            playerInfoStore.spirit = applyResponse.newSpirit
            console.log('[Spirit] 灵力已更新', {
              gain: spiritGain,
              newSpirit: applyResponse.newSpirit
            })
            syncCultivationData()
            message.success(`五灵之体，自动吸纳天地灵气，增加灵力${spiritGain}点`)
          }
        } else {
          console.log('[Spirit] 当前无灵力增长（spiritGain为0）')
        }
      } else {
        console.warn('[Spirit] 获取灵力增长失败:', gainResponse.message)
      }
    } catch (error) {
      console.error('[App.vue] 灵力增长同步失败:', error)
    }
  }, 10000)  // 10秒同步一次
}

const stopSpiritSyncTimer = () => {
  if (spiritSyncTimer) {
    clearInterval(spiritSyncTimer)
    spiritSyncTimer = null
  }
}

// 启动心跳定时器
const startHeartbeatTimer = (playerId, token) => {
  console.log('[App.vue] 启动心跳定时器', { playerId, tokenAvailable: !!token });
  // 先清除已有的定时器
  if (heartbeatTimer) {
    clearInterval(heartbeatTimer)
  }

  let heartbeatFailureCount = 0; // 心跳连续失败计数

  // 每60秒发送一次心跳
  heartbeatTimer = setInterval(async () => {
    try {
      //  console.log('[App.vue] 发送心跳', { playerId });
      await APIService.playerHeartbeat(playerId, token)
      // console.log('[App.vue] 心跳发送成功')
      
      // 心跳成功，重置失败计数
      heartbeatFailureCount = 0
      
    } catch (error) {
      console.error('[App.vue] 心跳发送失败:', error)
      
      // 增加失败计数
      heartbeatFailureCount++
      console.warn(`[App.vue] 心跳发送失败，失败次数: ${heartbeatFailureCount}/3`)
      
      // 连续失败3次，调用退出游戏接口
      if (heartbeatFailureCount >= 200) {
        console.error('[App.vue] 心跳连续失败100次，调用退出游戏接口')
        stopHeartbeatTimer()
        
        try {
          // 调用退出游戏接口
          await APIService.playerOffline(String(playerId))
          console.log('[App.vue] 已通知后端玩家离线')
        } catch (offlineError) {
          console.error('[App.vue] 通知后端离线失败:', offlineError)
        }
        
        // 清空玩家数据并返回登录页
        playerInfoStore.$reset()
        router.push('/login')
      }
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

  // 停止定时器
  stopHeartbeatTimer()
  stopSpiritSyncTimer()

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
  // 移除切换暗黑模式功能，改为跟随浏览器主题模式
  // playerInfoStore.toggleDarkMode()
}

// 移除空白行

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

/* ✅ 为最外层容器设置样式 */
#app {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
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