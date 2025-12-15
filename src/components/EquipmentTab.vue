<template>
  <div>
    <n-grid :cols="2" :x-gap="12" :y-gap="8">
      <n-grid-item v-for="(type, index) in Object.keys(equipmentTypes)" :key="index">
        <n-card hoverable>
          <template #header>
            <n-space justify="space-between">
              <span>{{ equipmentTypes[type] }}</span>
              <n-button size="small" @click="() => showEquipmentList(type)">
                更多
              </n-button>
            </n-space>
          </template>
          <p v-if="equipmentStore.equippedArtifacts[type]">
            {{ equipmentStore.equippedArtifacts[type].name }}
          </p>
          <p v-else>未装备</p>
          <template #footer>
            <n-space justify="space-between">
              <span>{{ equipmentTypes[type] }}</span>
              <n-button
                size="small"
                type="info"
                @click.stop="() => showEquippedEquipmentDetails(type)"
                v-if="equipmentStore.equippedArtifacts[type]"
              >
                详细
              </n-button>
              <n-button
                size="small"
                type="error"
                @click.stop="() => unequipItem(type)"
                v-if="equipmentStore.equippedArtifacts[type]"
              >
                卸下
              </n-button>
            </n-space>
          </template>
        </n-card>
      </n-grid-item>
    </n-grid>

    <!-- 装备列表弹窗 -->
    <n-modal
      v-model:show="showEquipmentModal"
      preset="dialog"
      :title="`${equipmentTypes[selectedEquipmentType]}列表`"
      style="width: 800px"
      @after-leave="onCloseEquipmentModal"
    >
      <n-space vertical>
        <n-space justify="space-between">
          <n-select v-model:value="selectedQuality" :options="qualityOptions" style="width: 150px" />
          <n-button type="warning" :disabled="equipmentList.length === 0" @click="batchSellEquipments">一键卖出</n-button>
        </n-space>
        <n-pagination
          v-model:page="currentEquipmentPage"
          :page-size="equipmentPageSize"
          :item-count="filteredEquipmentList.length"
          v-if="equipmentList.length > 8"
          @update:page-size="onEquipmentPageSizeChange"
          :page-slot="7"
        />
        <n-grid :cols="2" :x-gap="12" :y-gap="8" v-if="pagedEquipmentList.length">
          <n-grid-item v-for="equipment in pagedEquipmentList" :key="equipment.id" @click="showEquipmentDetails(equipment)">
            <n-card hoverable>
              <template #header>
                <n-space justify="space-between">
                  <span @click.stop="() => logClickEvent('装备名称', equipment)">{{ equipment.name }}</span>
                  <n-button size="small" type="warning" @click.stop="sellEquipment(equipment)">卖出</n-button>
                </n-space>
              </template>
              <n-space vertical>
                <n-tag 
                  :style="{ color: (equipment.quality && equipmentQualities[equipment.quality]?.color) || '#000000' }"
                  @click.stop="() => logClickEvent('装备品质标签', equipment)"
                >
                  {{ (equipment.quality && equipmentQualities[equipment.quality]?.name) || (equipment.quality && equipmentQualities[equipment.quality]?.name) || '未知品质' }}
                </n-tag>
                <n-text @click.stop="() => logClickEvent('境界要求文本', equipment)">境界要求：{{ getRealmPeriodName(equipment.requiredRealm) || '未知境界' }}</n-text>
                <!-- 显示装备状态 -->
                <n-tag v-if="equipment.equipped" type="success" @click.stop="() => logClickEvent('装备状态标签', equipment)">
                  已装备
                </n-tag>
              </n-space>
            </n-card>
          </n-grid-item>
        </n-grid>
        <n-empty description="没有任何装备" v-else></n-empty>
      </n-space>
    </n-modal>

    <!-- 装备详情弹窗 -->
    <n-modal v-model:show="showEquipmentDetailModal" preset="dialog" :title="selectedEquipment?.name || '装备详情'">
      <n-descriptions bordered>
        <n-descriptions-item label="品质">
          <span :style="{ color: (selectedEquipment?.quality && equipmentQualities[selectedEquipment.quality]?.color) || '#000000' }">
            {{ (selectedEquipment?.quality && equipmentQualities[selectedEquipment.quality]?.name) || '未知品质' }}
          </span>
        </n-descriptions-item>
        <n-descriptions-item label="类型">
          {{ equipmentTypes[selectedEquipment?.type] }}
        </n-descriptions-item>
        <n-descriptions-item label="强化等级">+{{ selectedEquipment?.enhanceLevel || 0 }}</n-descriptions-item>
        <template v-if="selectedEquipment?.stats">
          <n-descriptions-item v-for="(value, stat) in selectedEquipment.stats" :key="stat" :label="getStatName(stat)">
            {{ formatStatValue(stat, value) }}
          </n-descriptions-item>
        </template>
      </n-descriptions>
      <div
        class="stats-comparison"
        v-if="equipmentComparison && selectedEquipment && selectedEquipment.id != equipmentStore.equippedArtifacts[selectedEquipment.type]?.id"
      >
        <n-divider>属性对比</n-divider>
        <n-table :bordered="false" :single-line="false">
          <thead>
            <tr>
              <th>属性</th>
              <th>当前装备</th>
              <th>选中装备</th>
              <th>属性变化</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(comparison, stat) in equipmentComparison" :key="stat">
              <td>{{ getStatName(stat) }}</td>
              <td>{{ formatStatValue(stat, comparison.current) }}</td>
              <td>{{ formatStatValue(stat, comparison.selected) }}</td>
              <td>
                <n-gradient-text :type="comparison.isPositive ? 'success' : 'error'">
                  {{ comparison.isPositive ? '+' : '' }}{{ formatStatValue(stat, comparison.diff) }}
                </n-gradient-text>
              </td>
            </tr>
          </tbody>
        </n-table>
      </div>
      <template #action>
        <n-space justify="space-between">
          <n-space>
            <n-button
              type="primary"
              @click="showEnhanceConfirm = true"
              :disabled="(selectedEquipment?.enhanceLevel || 0) >= 100"
            >
              强化
            </n-button>
            <n-button type="info" :disabled="playerInfoStore.refinementStones === 0" @click="handleReforgeEquipment">
              洗练
            </n-button>
          </n-space>
          <n-space>
            <n-button
              @click="equipItem(selectedEquipment)"
              :disabled="playerInfoStore.level < selectedEquipment?.requiredRealm"
              v-if="selectedEquipment && selectedEquipment.id != equipmentStore.equippedArtifacts[selectedEquipment.type]?.id"
            >
              装备
            </n-button>
            <n-button
              @click="unequipItem(selectedEquipment?.type)"
              :disabled="playerInfoStore.level < selectedEquipment?.requiredRealm"
              v-else-if="selectedEquipment"
            >
              卸下
            </n-button>
            <n-button
              type="error"
              @click="sellEquipment(selectedEquipment)"
              v-if="selectedEquipment && selectedEquipment.id != equipmentStore.equippedArtifacts[selectedEquipment.type]?.id"
            >
              出售
            </n-button>
          </n-space>
        </n-space>
      </template>
    </n-modal>

    <!-- 强化确认弹窗 -->
    <n-modal v-model:show="showEnhanceConfirm" preset="dialog" title="装备强化">
      <n-space vertical>
        <p>是否消耗 {{ ((selectedEquipment?.enhanceLevel || 0) + 1) * 10 }} 强化石强化装备？</p>
        <p>当前强化石数量：{{ inventoryStore.reinforceStones }}</p>
      </n-space>
      <template #action>
        <n-space justify="end">
          <n-button @click="showEnhanceConfirm = false">取消</n-button>
          <n-button
            type="primary"
            @click="handleEnhanceEquipment"
            :disabled="inventoryStore.reinforceStones < ((selectedEquipment?.enhanceLevel || 0) + 1) * 10"
          >
            确认强化
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 洗练确认弹窗 -->
    <n-modal v-model:show="showReforgeConfirm" preset="dialog" title="洗练结果确认">
      <template v-if="reforgeResult">
        <div class="reforge-compare">
          <div class="old-stats">
            <h3>原始属性</h3>
            <div v-for="(value, key) in reforgeResult.oldStats" :key="key">
              {{ getStatName(key) }}: {{ formatStatValue(key, value) }}
            </div>
          </div>
          <div class="new-stats">
            <h3>新属性</h3>
            <div v-for="(value, key) in reforgeResult.newStats" :key="key">
              {{ getStatName(key) }}: {{ formatStatValue(key, value) }}
            </div>
          </div>
        </div>
      </template>
      <template #action>
        <n-button type="primary" @click="confirmReforgeResult(true)">确认新属性</n-button>
        <n-button @click="confirmReforgeResult(false)">保留原属性</n-button>
      </template>
    </n-modal>
  </div>
</template>

<script setup>
  import { ref, computed, onMounted } from 'vue'
  import { useMessage } from 'naive-ui'
  import { getStatName, formatStatValue } from '../plugins/stats'
  import { getRealmPeriodName } from '../plugins/realm'
  import { itemQualities } from '../plugins/itemQualities'
  import APIService from '../services/api.js'
  import { getAuthToken } from '../stores/db'

  // Props
  const props = defineProps({
    playerInfoStore: {
      type: Object,
      required: true
    },
    inventoryStore: {
      type: Object,
      required: true
    },
    equipmentStore: {
      type: Object,
      required: true
    }
  })

  const emit = defineEmits([
    'loadEquippedArtifacts',
    'refreshPetList'
  ])

  const message = useMessage()

  // 装备类型
  const equipmentTypes = {
    faqi: '法宝',
    guanjin: '冠巾',
    daopao: '道袍',
    yunlv: '云履',
    fabao: '本命法宝'
  }

  // 装备品质配置（使用统一配置）
  const equipmentQualities = itemQualities.equipment

  // 装备弹窗相关
  const showEquipmentModal = ref(false)
  const selectedEquipmentType = ref(null)
  
  // 本地装备列表（从后端直接获取）
  const localEquipmentList = ref([])
  
  // 装备列表（只使用本地列表）
  const equipmentList = computed(() => {
    return localEquipmentList.value
  })

  // 装备品质选项
  const qualityOptions = computed(() => {
    const equipmentsByQuality = {}
    equipmentList.value
      .forEach(item => {
        equipmentsByQuality[item.quality] = (equipmentsByQuality[item.quality] || 0) + 1
      })
    
    return [
      { label: '全部品质', value: 'all' },
      { label: equipmentQualities.mythic.name, value: 'mythic', disabled: !equipmentsByQuality['mythic'] },
      { label: equipmentQualities.legendary.name, value: 'legendary', disabled: !equipmentsByQuality['legendary'] },
      { label: equipmentQualities.epic.name, value: 'epic', disabled: !equipmentsByQuality['epic'] },
      { label: equipmentQualities.rare.name, value: 'rare', disabled: !equipmentsByQuality['rare'] },
      { label: equipmentQualities.uncommon.name, value: 'uncommon', disabled: !equipmentsByQuality['uncommon'] },
      { label: equipmentQualities.common.name, value: 'common', disabled: !equipmentsByQuality['common'] }
    ]
  })

  // 过滤后的装备列表
  const filteredEquipmentList = computed(() => {
    let list = equipmentList.value.filter(item => {
      if (selectedQuality.value !== 'all' && item.quality !== selectedQuality.value) return false
      return true
    })
    
    // 对列表进行排序，已装备的装备排在前面
    list.sort((a, b) => {
      if (a.equipped && !b.equipped) return -1
      if (!a.equipped && b.equipped) return 1
      return 0
    })
    
    return list
  })

  // 装备列表分页
  const currentEquipmentPage = ref(1)
  const equipmentPageSize = ref(8)
  
  // 分页后的装备列表
  const pagedEquipmentList = computed(() => {
    const start = (currentEquipmentPage.value - 1) * equipmentPageSize.value
    const end = start + equipmentPageSize.value
    return filteredEquipmentList.value.slice(start, end)
  })

  // 装备品质筛选
  const selectedQuality = ref('all')

  // 页大小改变处理
  const onEquipmentPageSizeChange = size => {
    equipmentPageSize.value = size
    currentEquipmentPage.value = 1
  }

  // 装备详情相关
  const showEquipmentDetailModal = ref(false)
  const selectedEquipment = ref(null)

  // 强化确认弹窗
  const showEnhanceConfirm = ref(false)

  // 洗练确认弹窗
  const showReforgeConfirm = ref(false)
  const reforgeResult = ref(null)

  // 装备属性对比计算
  const equipmentComparison = computed(() => {
    if (!selectedEquipment.value || !selectedEquipmentType.value) return null
    const currentEquipment = props.equipmentStore.equippedArtifacts[selectedEquipmentType.value]
    if (!currentEquipment) return null
    const comparison = {}
    const allStats = new Set([
      ...Object.keys(selectedEquipment.value.stats || {}), 
      ...Object.keys(currentEquipment.stats || {})
    ])
    allStats.forEach(stat => {
      const selectedValue = (selectedEquipment.value.stats && selectedEquipment.value.stats[stat]) || 0
      const currentValue = (currentEquipment.stats && currentEquipment.stats[stat]) || 0
      const diff = selectedValue - currentValue
      comparison[stat] = {
        current: currentValue,
        selected: selectedValue,
        diff: diff,
        isPositive: diff > 0
      }
    })
    return comparison
  })

  // 日志记录函数
  const logClickEvent = (elementType, equipment) => {
    console.log(`[Inventory] 用户点击了装备列表中的${elementType}: ${equipment.name} (${equipment.id})`)
  }

  // 修改装备列表显示方法
  const showEquipmentList = async (type) => {
    selectedEquipmentType.value = type
    
    // 每次点击"更多"按钮都从后端获取最新数据
    try {
      const token = getAuthToken()
      if (!token) {
        message.error('未找到认证令牌')
        console.error('[Inventory] 未找到认证令牌，无法获取装备列表')
        return
      }
      
      // 使用新的API服务方法获取装备数据，获取所有装备（已装备和未装备）
      const data = await APIService.getEquipmentList(token, { equip_type: type })
      
      // 保存装备列表到本地变量
      if (data.equipment) {
        localEquipmentList.value = data.equipment;
        
        console.log(`[Inventory] 成功获取装备列表，共${data.equipment.length}件装备`, {
          equipmentCount: data.equipment.length,
          equipmentList: localEquipmentList.value.map(e => ({
            id: e.id,
            name: e.name,
            type: e.type,
            quality: e.quality,
            enhanceLevel: e.enhanceLevel,
            equipped: e.equipped
          }))
        });
      } else {
        localEquipmentList.value = [];
        console.log('[Inventory] 获取装备列表失败，返回空列表');
      }
    } catch (error) {
      console.error('获取装备数据时发生错误:', error)
      message.error('获取装备数据时发生错误')
    }
    
    showEquipmentModal.value = true
  }

  // 查看已装备的装备详情
  const showEquippedEquipmentDetails = async (type) => {
    const equippedItem = props.equipmentStore.equippedArtifacts[type];
    if (equippedItem) {
      selectedEquipment.value = equippedItem;
      showEquipmentDetailModal.value = true;
    } else {
      message.info('该位置暂未装备任何装备');
    }
  }

  // 关闭装备列表弹窗时，清空本地装备列表
  const onCloseEquipmentModal = () => {
    localEquipmentList.value = []
  }

  // 显示装备详情
  const showEquipmentDetails = async (equipment) => {
    try {
      console.log('[Inventory] 查看装备详情，传入的装备参数:', equipment);
      const token = getAuthToken()
      if (!token) {
        message.error('未找到认证令牌')
        console.error('[Inventory] 未找到认证令牌，无法获取装备详情')
        return
      }
      
      const itemId = equipment?.id
      if (!itemId) {
        message.error('无效的装备ID')
        console.error('[Inventory] 无效的装备ID，无法获取装备详情', { 
          equipmentId: equipment?.id, 
          selectedEquipmentId: selectedEquipment.value?.id,
          equipment: equipment,
          selectedEquipment: selectedEquipment.value
        })
        return
      }
      
      const response = await APIService.getEquipmentDetails(token, itemId)
      selectedEquipment.value = response.equipment
      showEquipmentDetailModal.value = true
    } catch (error) {
      console.error('获取装备详情时发生错误:', error)
      message.error('获取装备详情时发生错误: ' + error.message)
    }
  }

  // 卸下装备
  const unequipItem = async slot => {
    const token = getAuthToken()
    const result = await props.equipmentStore.unequipArtifact(
      slot,
      props.inventoryStore,
      
      token
    );

    if (result.success) {
      showEquipmentDetailModal.value = false
      message.success('当前装备已卸下')
      // 卸下装备后刷新已装备列表
      emit('loadEquippedArtifacts')
      // 清空本地装备列表以触发重新加载
      localEquipmentList.value = []
      
      // 更新玩家属性
      if (result.user) {
        props.playerInfoStore.$patch({
          baseAttributes: result.user.baseAttributes,
          combatAttributes: result.user.combatAttributes,
          combatResistance: result.user.combatResistance,
          specialAttributes: result.user.specialAttributes
        });
      }
      
      // 更新选中的装备信息
      if (result.equipment) {
        if (selectedEquipment.value && selectedEquipment.value.id === result.equipment.id) {
          selectedEquipment.value = result.equipment
        }
      }
    } else {
      message.error(result.message || '卸下装备失败')
      // 即使卸下失败也要清除缓存以确保列表刷新
      localEquipmentList.value = []
    }
  }

  // 使用装备
  const equipItem = async equipment => {
    const token = getAuthToken()
    
    // 添加令牌有效性检查
    if (!token) {
      message.error('认证已过期，请重新登录')
      console.error('[Inventory] 装备穿戴失败：认证令牌缺失')
      return
    }
    
    // 在穿戴装备前打印装备属性
    console.log('[Equipment Stats Before Equip] Equipment attributes before equipping:', {
      equipmentId: equipment.id,
      equipmentName: equipment.name,
      equipmentType: equipment.type,
      equipmentQuality: equipment.quality,
      equipmentStats: equipment.stats,
      equipmentEnhanceLevel: equipment.enhanceLevel,
      equipmentRequiredRealm: equipment.requiredRealm
    });
    
    // 在穿戴装备前打印玩家属性
    console.log('[Player Stats Before Equip] Player attributes before equipping:', {
      baseAttributes: props.playerInfoStore.baseAttributes,
      combatAttributes: props.playerInfoStore.combatAttributes,
      combatResistance: props.playerInfoStore.combatResistance,
      specialAttributes: props.playerInfoStore.specialAttributes
    });
    
    const result = await props.equipmentStore.equipArtifact(
      equipment,
      equipment.equipType || equipment.EquipType,
      props.inventoryStore,
      
      props.playerInfoStore,
      token
    )
    
    if (result.success) {
      message.success(result.message)
      showEquipmentModal.value = false
      showEquipmentDetailModal.value = false
      // 装备成功后刷新已装备列表
      emit('loadEquippedArtifacts')
      // 清空本地装备列表以触发重新加载
      localEquipmentList.value = []
      
      // 更新玩家属性
      if (result.user) {
        props.playerInfoStore.$patch({
          baseAttributes: result.user.baseAttributes,
          combatAttributes: result.user.combatAttributes,
          combatResistance: result.user.combatResistance,
          specialAttributes: result.user.specialAttributes
        });
      }
      
      // 在穿戴装备后打印玩家属性
      console.log('[Player Stats After Equip] Player attributes after equipping:', {
        baseAttributes: props.playerInfoStore.baseAttributes,
        combatAttributes: props.playerInfoStore.combatAttributes,
        combatResistance: props.playerInfoStore.combatResistance,
        specialAttributes: props.playerInfoStore.specialAttributes
      });
      
      // 更新选中的装备信息
      if (result.equipment) {
        if (selectedEquipment.value && selectedEquipment.value.id === result.equipment.id) {
          selectedEquipment.value = result.equipment
        }
      }
    } else {
      message.error(result.message || '装备失败')
      // 即使装备失败也要清除缓存以确保列表刷新
      localEquipmentList.value = []
    }
  }

  // 批量卖出装备
  const batchSellEquipments = async () => {
    const token = getAuthToken()
    const result = await props.equipmentStore.batchSellEquipments(
      selectedQuality.value === 'all' ? null : selectedQuality.value,
      selectedEquipmentType.value,
      props.inventoryStore,
      token
    )
    
    if (result.success) {
      message.success(result.message)
      // 批量卖出成功后刷新本地装备列表缓存
      localEquipmentList.value = []
      // 重新从后端获取装备列表
      await showEquipmentList(selectedEquipmentType.value)
    } else {
      message.error(result.message || '批量卖出失败')
    }
  }

  // 卖出单件装备
  const sellEquipment = async equipment => {
    const token = getAuthToken()
    const result = await props.equipmentStore.sellEquipment(equipment, props.inventoryStore, token)
    if (result.success) {
      message.success(result.message)
      showEquipmentDetailModal.value = false
      // 刷新本地装备列表缓存
      localEquipmentList.value = []
      // 重新从后端获取装备列表
      await showEquipmentList(selectedEquipmentType.value)
    } else {
      message.error(result.message || '卖出失败')
    }
  }

  // 强化装备
  const handleEnhanceEquipment = async () => {
    if (!selectedEquipment.value) return
    
    const token = getAuthToken()
    try {
      // 直接从后端获取最新的玩家数据，包括强化石数量
      const userData = await APIService.getUser(token)
      props.inventoryStore.reinforceStones = userData.reinforceStones || 0
      
      const result = await APIService.enhanceEquipment(token, selectedEquipment.value.id, props.inventoryStore.reinforceStones)
      
      if (result.success) {
        props.inventoryStore.reinforceStones -= result.cost
        selectedEquipment.value.stats = { ...result.newStats }
        selectedEquipment.value.enhanceLevel = result.newLevel
        // 更新装备的境界要求
        if (result.newRequiredRealm !== undefined) {
          selectedEquipment.value.requiredRealm = result.newRequiredRealm
        }
        message.success('强化成功')
      } else {
        message.error(result.message || '强化失败')
      }
    } catch (error) {
      console.error('装备强化失败:', error)
      message.error('装备强化失败: ' + error.message)
    }
  }

  // 洗练装备
  const handleReforgeEquipment = async () => {
    if (!selectedEquipment.value) return
    
    const token = getAuthToken()
    try {
      const result = await APIService.reforgeEquipment(token, selectedEquipment.value.id, props.playerInfoStore.refinementStones)
      
      if (result.success) {
        props.playerInfoStore.refinementStones -= result.cost
        reforgeResult.value = result
        showReforgeConfirm.value = true
      } else {
        message.error(result.message || '洗练失败')
      }
    } catch (error) {
      console.error('装备洗练失败:', error)
      message.error('装备洗练失败: ' + error.message)
    }
  }

  // 确认洗练结果
  const confirmReforgeResult = async confirm => {
    if (!reforgeResult.value) return
    
    const token = getAuthToken()
    try {
      const result = await APIService.confirmReforge(
        token, 
        selectedEquipment.value.id, 
        confirm, 
        confirm ? reforgeResult.value.newStats : null
      )
      
      if (result.success) {
        if (confirm) {
          // 用户确认后，应用新属性
          selectedEquipment.value.stats = reforgeResult.value.newStats
          message.success('已确认新属性')
        } else {
          // 用户取消，保留原属性
          message.info('已保留原有属性')
        }
      } else {
        message.error(result.message || '确认洗练结果失败')
      }
    } catch (error) {
      console.error('确认洗练结果失败:', error)
      message.error('确认洗练结果失败: ' + error.message)
    }
    
    showReforgeConfirm.value = false
    reforgeResult.value = null
  }

  defineExpose({
    showEquipmentList
  })
</script>

<style scoped>
  .reforge-compare {
    display: flex;
    justify-content: space-between;
    gap: 20px;
    margin: 16px 0;
  }

  .old-stats,
  .new-stats {
    flex: 1;
    padding: 16px;
    border-radius: 8px;
    background-color: rgba(0, 0, 0, 0.05);
  }

  .old-stats h3,
  .new-stats h3 {
    margin-top: 0;
    margin-bottom: 12px;
    font-size: 16px;
    color: #666;
  }
</style>