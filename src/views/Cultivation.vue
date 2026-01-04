<template>
  <n-card title="修炼">
    <n-tabs type="line" animated>
      <!-- 修炼分页 -->
      <n-tab-pane name="cultivation" tab="修炼">
        <n-space vertical>
          <n-alert type="info" show-icon>
            <template #icon>
              <n-icon>
                <BookOutline />
              </n-icon>
            </template>
            通过打坐修炼来提升修为，获得修为，是修炼效率的随机倍数，积累足够的修为后可以尝试突破境界，境界越高吸纳灵力速度越快。
          </n-alert>
          <n-space vertical>
            <n-button type="primary" size="large" block @click="cultivate" :disabled="playerInfoStore.spirit < cultivationCost">
              打坐修炼 (消耗 {{ cultivationCost }} 灵力)
            </n-button>
            <n-button :type="isAutoCultivating ? 'warning' : 'success'" size="large" block @click="toggleAutoCultivation">
              {{ isAutoCultivating ? '停止自动修炼' : '开始自动修炼' }}
            </n-button>
          </n-space>
          <n-divider>修炼详情</n-divider>
          <n-descriptions bordered>
            <n-descriptions-item label="灵力获取速率">{{ playerInfoStore.spiritRate }} / 秒</n-descriptions-item>
            <n-descriptions-item label="修炼效率">{{ playerInfoStore.cultivationGain }} 修为 / 次</n-descriptions-item>
            <n-descriptions-item label="突破所需修为">
              {{ playerInfoStore.maxCultivation }}
            </n-descriptions-item>
          </n-descriptions>
          <log-panel ref="logRef" title="修炼日志" />
        </n-space>
      </n-tab-pane>

      <!-- 聚灵阵分页 -->
      <n-tab-pane name="formation" tab="聚灵阵">
        <n-space vertical>
          <n-alert type="info" show-icon>
            <template #icon>
              <n-icon>
                <Sparkles />
              </n-icon>
            </template>
            每日前10次降低10倍消耗。使用聚灵阵可以加速修炼，消耗灵石获得修为，是聚灵效率的随机倍数。
          </n-alert>
          <n-space vertical>
            <n-button type="primary" size="large" block @click="useFormation" :disabled="playerInfoStore.spiritStones < formationCost">
              启动聚灵阵 (消耗 {{ formationCost }} 灵石)
            </n-button>
            <n-button :type="isAutoFormation ? 'warning' : 'success'" size="large" block @click="toggleAutoFormation">
              {{ isAutoFormation ? '停止自动聚灵' : '开始自动聚灵' }}
            </n-button>
          </n-space>
          <n-divider>聚灵阵详情</n-divider>
          <n-descriptions bordered>
            <n-descriptions-item label="聚灵阵等级">{{ playerInfoStore.formationLevel }}</n-descriptions-item>
            <n-descriptions-item label="聚灵效率">{{ playerInfoStore.formationGain }} 修为 / 次</n-descriptions-item>
            <n-descriptions-item label="聚灵消耗">{{ formationCost }} 灵石 / 次</n-descriptions-item>
          </n-descriptions>
          <log-panel ref="formationLogRef" title="聚火日志" />
        </n-space>
      </n-tab-pane>
      
      <!-- 渡劫破镜分页 -->
      <n-tab-pane name="breakthrough" tab="渡劫">
        <n-space vertical>
          <n-alert v-if="playerInfoStore.level === 27 && playerInfoStore.cultivation >= playerInfoStore.maxCultivation" type="warning" show-icon>
            <template #icon>
              <n-icon>
                <Sparkles />
              </n-icon>
            </template>
            道友困于瓶颈百年，已将境界打磨圆满，理应借助雷劫，打碎桎梏。渡劫会消耗装备和灵宠。服用渡劫丹可增加渡劫成功率。
          </n-alert>
          <n-alert v-else type="info" show-icon>
            <template #icon>
              <n-icon>
                <BookOutline />
              </n-icon>
            </template>
            修仙本是逆天而行，道友境界不足，无需渡劫
          </n-alert>
          <n-descriptions v-if="playerInfoStore.level === 27" bordered>
            <n-descriptions-item label="当前修为">{{ playerInfoStore.cultivation.toFixed(1) }} / {{ playerInfoStore.maxCultivation }}</n-descriptions-item>
            <n-descriptions-item label="渡劫成功率">{{ (playerInfoStore.duJieRate * 100).toFixed(1) }}%</n-descriptions-item>
          </n-descriptions>
          <n-button 
            v-if="playerInfoStore.level === 27 && playerInfoStore.cultivation >= playerInfoStore.maxCultivation"
            type="success" 
            size="large" 
            block 
            @click="breakthroughJieYing"
            :loading="isBreakthroughLoading"
          >
            渡劫 (成功率: {{ (playerInfoStore.duJieRate * 100).toFixed(1) }}%)
          </n-button>
          <n-button 
            v-else
            type="primary" 
            size="large" 
            block
            disabled
          >
            不满足突破条件
          </n-button>
          <log-panel ref="breakthroughLogRef" title="突破日志" />
        </n-space>
      </n-tab-pane>
    </n-tabs>
  </n-card>
</template>

<script setup>
  // 修改为使用模块化store
  import { usePlayerInfoStore } from '../stores/playerInfo'
  import { ref, computed, onMounted, onUnmounted } from 'vue'
  import { useMessage, NCard, NTabs, NTabPane, NAlert, NSpace, NButton, NDivider, NDescriptions, NDescriptionsItem, NIcon } from 'naive-ui'
  import { BookOutline, Sparkles } from '@vicons/ionicons5' // 导入图标组件
  import LogPanel from '../components/LogPanel.vue'
  import APIService from '../services/api'
  import { getAuthToken } from '../stores/db'

  const playerInfoStore = usePlayerInfoStore()
  
  const message = useMessage()
  const isAutoCultivating = ref(false)
  const isAutoFormation = ref(false)
  const isBreakthroughLoading = ref(false)
  const logRef = ref(null)
  const formationLogRef = ref(null)
  const breakthroughLogRef = ref(null)

  // 修炼消耗和获得（从后端获取准确数据）
  const cultivationCost = computed(() => {
    // 从后端获取准确的修炼消耗数据
    return playerInfoStore.cultivationCost || 1
  })

  // 聚灵阵消耗（固定值或从后端获取）
  const formationCost = computed(() => {
    // 聚灵阵每次使用的消耗，可以从后端获取或者根据等级计算
    return playerInfoStore.formationCost || 10
  })

  // 修炼获得的修为
  const cultivationGain = computed(() => {
    // 从后端获取准确的修炼获得数据
    return playerInfoStore.cultivationGain || 1
  })

  // 基础灵力获取速率
  const baseGainRate = 1

  // 计算突破所需成本
  const calculateBreakthroughCost = () => {
    return playerInfoStore.maxCultivation - playerInfoStore.cultivation
  }

  // 打坐修练方法
  const cultivate = async () => {
    try {
      // ✅ 使用正确的URL（不需要前面的 /api）
      const token = getAuthToken();
      const response = await APIService.post('/cultivation/single', {}, token)
      
      if (response.success) {
        // 后端返回的数据字段名
        playerInfoStore.spirit -= response.spiritCost
        playerInfoStore.cultivation = response.currentCultivation
        
        // 记录日志
        if (logRef.value) {
          logRef.value.addLog(`修炼获得 ${response.cultivationGain.toFixed(1)} 点修为`)
        }
        
        // 检查是否有突破
        if (response.breakthrough) {
          const bt = response.breakthrough
          playerInfoStore.level = bt.newLevel
          playerInfoStore.realm = bt.newRealm
          playerInfoStore.maxCultivation = bt.newMaxCultivation
          playerInfoStore.spirit += bt.spiritReward
          playerInfoStore.spiritRate = bt.newSpiritRate
          playerInfoStore.breakthroughCount += 1
          message.success(bt.message)
          if (logRef.value) {
            logRef.value.addLog(bt.message)
          }
        }
        message.success('修炼成功，获得 ' + response.cultivationGain.toFixed(1) + ' 点修为')
        // 增加修炼时间统计
        playerInfoStore.totalCultivationTime += 1
        return true
      } else {
        message.warning(response.error || '修炼失败')
        return false
      }
    } catch (error) {
      message.error('修炼请求失败：' + error.message)
      if (logRef.value) {
        logRef.value.addLog(`修炼失败：${error.message}`)
      }
      return false
    }
  }

  // 使用聚灵阵方法
  const useFormation = async () => {
    try {
      const token = getAuthToken();
      const response = await APIService.post('/cultivation/formation', {}, token)
      
      if (response.success) {
        // 后端返回的数据字段名
        playerInfoStore.spiritStones -= response.stoneCost
        playerInfoStore.cultivation = response.currentCultivation
        
        // ✅ 高等级聚灵阵逻辑（等级>27）：增加修为上限
        if (response.maxCultivationGain > 0) {
          playerInfoStore.maxCultivation = response.newMaxCultivation
          
          // 记录日志
          if (formationLogRef.value) {
            formationLogRef.value.addLog(`聚灵阵威能增幅！修为上限+${response.maxCultivationGain.toFixed(1)} (${response.maxCultivationRate.toFixed(1)}%)，消耗 ${response.stoneCost} 灵石`)
          }
          message.success(`聚灵阵威能增幅！修为上限+${response.maxCultivationGain.toFixed(1)} (${response.maxCultivationRate.toFixed(1)}%)`)
        } else {
          // ✅ 普通等级聚灵阵逻辑（等级<=27）：增加修为
          // 记录日志
          if (formationLogRef.value) {
            formationLogRef.value.addLog(`聚灵阵获得 ${response.cultivationGain.toFixed(1)} 点修为，消耗 ${response.stoneCost} 灵石`)
          }
          message.success('聚灵阵使用成功，获得 ' + response.cultivationGain.toFixed(1) + ' 点修为')
        }
        
        // 检查是否有突破
        if (response.breakthrough) {
          const bt = response.breakthrough
          playerInfoStore.level = bt.newLevel
          playerInfoStore.realm = bt.newRealm
          playerInfoStore.maxCultivation = bt.newMaxCultivation
          playerInfoStore.spirit += bt.spiritReward
          playerInfoStore.spiritRate = bt.newSpiritRate
          playerInfoStore.breakthroughCount += 1
          message.success(bt.message)
          if (formationLogRef.value) {
            formationLogRef.value.addLog(bt.message)
          }
        }
        return true
      } else {
        message.warning(response.error || '聚灵阵使用失败')
        return false
      }
    } catch (error) {
      message.error('聚灵阵请求失败：' + error.message)
      if (formationLogRef.value) {
        formationLogRef.value.addLog(`聚灵阵使用失败：${error.message}`)
      }
      return false
    }
  }

  // 结婴突破方法
  const breakthroughJieYing = async () => {
    try {
      isBreakthroughLoading.value = true
      const token = getAuthToken();
      const response = await APIService.post('/cultivation/breakthrough-jieying', {}, token)
      
      if (response.success) {
        // 结婴成功
        playerInfoStore.level = response.newLevel
        playerInfoStore.realm = response.newRealm
        playerInfoStore.maxCultivation = response.newMaxCultivation || 10000
        playerInfoStore.cultivation = 0.0
        playerInfoStore.duJieRate = 0.05
        
        message.success(response.message)
        if (breakthroughLogRef.value) {
          breakthroughLogRef.value.addLog(response.message)
          // 显示消耗提示消息
          if (response.consumeMessage) {
            breakthroughLogRef.value.addLog(response.consumeMessage)
          }
        }
        return true
      } else {
        // 结婴失败
        playerInfoStore.cultivation = 0.0
        playerInfoStore.duJieRate = 0.0
        
        message.error(response.message)
        if (breakthroughLogRef.value) {
          breakthroughLogRef.value.addLog(response.message)
          // 显示消耗提示消息
          if (response.consumeMessage) {
            breakthroughLogRef.value.addLog(response.consumeMessage)
          }
        }
        return false
      }
    } catch (error) {
      message.error('渡劫请求失败：' + error.message)
      if (breakthroughLogRef.value) {
        breakthroughLogRef.value.addLog(`渡劫失败：${error.message}`)
      }
      return false
    } finally {
      isBreakthroughLoading.value = false
    }
  }

  // 切换自动修炼
  const toggleAutoCultivation = async () => {
    if (isAutoCultivating.value) {
      // 停止自动修炼
      isAutoCultivating.value = false
      message.info('停止自动修炼')
      return
    }
    
    // 检查灵力是否满足一次修炼消耗
    if (playerInfoStore.spirit < cultivationCost.value) {
      message.error('灵力不足，修炼失败')
      return
    }
    
    // 开始自动修练
    isAutoCultivating.value = true
    message.success('开始自动修练')
    
    // 循环调用单次修炼直到灵力不足或用户停止
    while (isAutoCultivating.value) {
      // 检查灵力是否满足一次修炼消耗
      if (playerInfoStore.spirit < cultivationCost.value) {
        isAutoCultivating.value = false
        message.warning('灵力不足，修炼停止')
        if (logRef.value) {
          logRef.value.addLog('灵力不足，修炼停止')
        }
        break
      }
      
      // 调用单次修炼
      const success = await cultivate()
      if (!success) {
        isAutoCultivating.value = false
        break
      }
      
      // 短暂延迟，避免过快调用
      await new Promise(resolve => setTimeout(resolve, 1000))
    }
  }

  // 切换自动聚灵
  const toggleAutoFormation = async () => {
    if (isAutoFormation.value) {
      // 停止自动聚灵
      isAutoFormation.value = false
      message.info('停止自动聚灵')
      return
    }
    
    // 检查灵石是否满足一次聚灵消耗
    if (playerInfoStore.spiritStones < formationCost.value) {
      message.error('灵石不足，聚灵失败')
      return
    }
    
    // 开始自动聚灵
    isAutoFormation.value = true
    message.success('开始自动聚灵')
    
    // 循环调用聚灵阵直到灵石不足或用户停止
    while (isAutoFormation.value) {
      // 检查灵石是否满足一次聚灵消耗
      if (playerInfoStore.spiritStones < formationCost.value) {
        isAutoFormation.value = false
        message.warning('灵石不足，聚灵停止')
        if (formationLogRef.value) {
          formationLogRef.value.addLog('灵石不足，聚灵停止')
        }
        break
      }
      
      // 调用聚灵阵
      const success = await useFormation()
      if (!success) {
        isAutoFormation.value = false
        break
      }
      
      // 短暂延迟，避免过快调用
      await new Promise(resolve => setTimeout(resolve, 3000))
    }
  }

  // 从后端获取最新修炼数据
  const syncCultivationData = async () => {
    try {
      const token = getAuthToken();
      
      // 获取修炼消耗和获得数据
      const response = await APIService.getCultivationData(token)
      if (response.success) {
      //  console.log('[Cultivation] 修炼消耗:', response.data.spiritCost, '获得:', response.data.cultivationGain)
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
        // ✅ 新增：洗练石和灵宠精华
        playerInfoStore.refinementStones = response.data.refinementStones // 洗练石数量
        playerInfoStore.petEssence = response.data.petEssence // 灵宠精华数量
        // 新增：聚灵阵相关数据
        playerInfoStore.formationLevel = response.data.baseAttributes?.formationLevel ?? 1
        playerInfoStore.formationGain = response.data.baseAttributes?.formationGain ?? 5
        playerInfoStore.formationCost = response.data.baseAttributes?.formationCost ?? 10
        // ✅ 新增：渡劫成功率（从 baseAttributes 中读取）
        const duJieRate = response.data.baseAttributes?.duJieRate || 0.05
        playerInfoStore.duJieRate = duJieRate
        
      }
    } catch (error) {
      console.error('同步修为数据失败:', error)
    }
  }

  // 绑定自动修炼停止快捷键
  const handleKeyboard = (event) => {
    if (event.code === 'Space') {
      event.preventDefault()
      cultivate()
    }
  }

  onMounted(() => {
    // 页面加载时同步修炼数据
    syncCultivationData()
    // 绑定键盘快捷键
    window.addEventListener('keydown', handleKeyboard)
  })

  onUnmounted(() => {
    window.removeEventListener('keydown', handleKeyboard)
  })
</script>

<style scoped>
  .n-space {
    width: 100%;
  }

  .n-button {
    margin-bottom: 12px;
  }

  .n-collapse {
    margin-top: 12px;
  }
  
  .n-tabs {
    margin-top: 16px;
  }
</style>