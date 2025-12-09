<template>
  <n-layout>
    <n-layout-header bordered>
      <n-page-header>
        <template #title>抽奖系统</template>
      </n-page-header>
    </n-layout-header>
    <n-layout-content class="gacha-content">
      <n-card :bordered="false">
        <div class="gacha-container">
          <div class="gacha-type-selector">
            <n-radio-group v-model:value="gachaType" name="gachaType">
              <n-radio-button value="equipment">装备池</n-radio-button>
              <n-radio-button value="pet">灵宠池</n-radio-button>
            </n-radio-group>
          </div>
          <div class="spirit-stones">
            <n-statistic label="灵石" :value="inventoryStore.spiritStones" />
          </div>
          
          <GachaCard 
            :gacha-type="gachaType"
            :is-shaking="gachaStore.isShaking"
            :is-opening="gachaStore.isOpening"
          />
          
          <GachaButtons
            :spirit-stones="inventoryStore.spiritStones"
            :wishlist-enabled="inventoryStore.wishlistEnabled"
            :is-drawing="gachaStore.isDrawing"
            @gacha="performGacha"
            @show-probability="gachaStore.showProbabilityInfo = true"
            @show-wishlist="gachaStore.showWishlistSettings = true"
            @show-auto-settings="gachaStore.showAutoSettings = true"
          />
        </div>
      </n-card>
    </n-layout-content>
    
    <!-- 弹窗组件 -->
    <GachaResultModal
      v-model:show="gachaStore.showResultModal"
      v-model:current-page="gachaStore.currentPage"
      :gacha-type="gachaType"
      :gacha-number="gachaNumber"
      :spirit-stones="inventoryStore.spiritStones"
      :wishlist-enabled="inventoryStore.wishlistEnabled"
      :is-drawing="gachaStore.isDrawing"
      :current-page-results="gachaStore.currentPageResults"
      :total-pages="gachaStore.totalPages"
      :page-size="gachaStore.pageSize"
      :selected-wish-equip-quality="inventoryStore.selectedWishEquipQuality"
      :selected-wish-pet-rarity="inventoryStore.selectedWishPetRarity"
      :pet-rarities="petRarities"
      :equipment-quality-options="equipmentQualityOptions"
      :pet-rarity-options="petRarityOptions"
      :equipment-types-map="equipmentTypes"
      @perform-gacha="performGacha"
      @show-pet-detail="showPetDetail"
      @show-equipment-detail="showEquipmentDetail"
    />
    
    <ProbabilityInfoModal
      v-model:show="gachaStore.showProbabilityInfo"
      :wishlist-enabled="inventoryStore.wishlistEnabled"
      :selected-wish-equip-quality="inventoryStore.selectedWishEquipQuality"
      :selected-wish-pet-rarity="inventoryStore.selectedWishPetRarity"
      :equipment-qualities="equipmentQualities"
      :pet-rarities="petRarities"
      :adjusted-equip-probabilities="adjustedEquipProbabilities"
      :adjusted-pet-probabilities="adjustedPetProbabilities"
    />
    
    <WishlistSettingsModal
      v-model:show="gachaStore.showWishlistSettings"
      v-model:wishlist-enabled="inventoryStore.wishlistEnabled"
      v-model:selected-wish-equip-quality="inventoryStore.selectedWishEquipQuality"
      v-model:selected-wish-pet-rarity="inventoryStore.selectedWishPetRarity"
      :equipment-qualities="equipmentQualities"
      :pet-rarities="petRarities"
      :equipment-quality-options="equipmentQualityOptions"
      :pet-rarity-options="petRarityOptions"
    />
    
    <AutoSettingsModal
      v-model:show="gachaStore.showAutoSettings"
      v-model:auto-sell-qualities="inventoryStore.autoSellQualities"
      v-model:auto-release-rarities="inventoryStore.autoReleaseRarities"
      :equipment-qualities="equipmentQualities"
      :pet-rarities="petRarities"
    />
    
    <PetDetailModal
      v-model:show="gachaStore.showPetDetails"
      :pet="gachaStore.selectedPet"
      :pet-rarities="petRarities"
    />
    
    <EquipmentDetailModal
      v-model:show="gachaStore.showEquipmentDetails"
      :equipment="gachaStore.selectedEquipment"
      :equipment-types="equipmentTypes"
    />
  </n-layout>
</template>

<script setup>
  import { ref, computed, watch } from 'vue'
  import { useMessage } from 'naive-ui'
  // 修改为使用模块化store
  import { usePlayerInfoStore } from '../../stores/playerInfo'
  import { useInventoryStore } from '../../stores/inventory'
  import { useEquipmentStore } from '../../stores/equipment'
  import { usePetsStore } from '../../stores/pets'
  import { usePillsStore } from '../../stores/pills'
  import { useSettingsStore } from '../../stores/settings'
  import { useStatsStore } from '../../stores/stats'
  import { usePersistenceStore } from '../../stores/persistence'
  import { useGachaStore } from '../../stores/gacha'
  import GachaCard from './GachaCard.vue'
  import GachaButtons from './GachaButtons.vue'
  import GachaResultModal from './GachaResultModal.vue'
  import ProbabilityInfoModal from './ProbabilityInfoModal.vue'
  import WishlistSettingsModal from './WishlistSettingsModal.vue'
  import AutoSettingsModal from './AutoSettingsModal.vue'
  import PetDetailModal from './PetDetailModal.vue'
  import EquipmentDetailModal from './EquipmentDetailModal.vue'
  // 使用后端API的抽奖功能
  import { useGacha } from './useGacha'
  import { getAdjustedEquipProbabilities, getAdjustedPetProbabilities } from './useProbability'
  import { itemQualities } from '../../plugins/itemQualities'
  import { getAuthToken } from '../../stores/db'
  import APIService from '../../services/api'

  const playerInfoStore = usePlayerInfoStore()
  const inventoryStore = useInventoryStore()
  const equipmentStore = useEquipmentStore()
  const petsStore = usePetsStore()
  const pillsStore = usePillsStore()
  const settingsStore = useSettingsStore()
  const statsStore = useStatsStore()
  const persistenceStore = usePersistenceStore()
  const gachaStore = useGachaStore()
  
  const message = useMessage()
  
  // 使用后端API的抽奖功能
  const { performGacha: performGachaAPI, processAutoActions: processAutoActionsAPI } = useGacha()
  
  // 抽卡类型
  const gachaType = ref('equipment') // 'equipment' 或 'pet'
  
  // 抽卡次数
  const gachaNumber = ref(10)
  
  // 装备品质配置（使用统一配置）
  const equipmentQualities = itemQualities.equipment
  
  // 装备类型
  const equipmentTypes = {
  faqi: {
    name: '法宝',
    slot: 'faqi',
    prefixes: ['三清', '五雷', '七星', '八卦', '九宫', '招魂', '镇妖', '降魔']
  },
  guanjin: {
    name: '冠巾',
    slot: 'guanjin',
    prefixes: ['清心', '明性', '护神', '静思', '悟道', '通玄', '守一', '凝神']
  },
  daopao: {
    name: '道袍',
    slot: 'daopao',
    prefixes: ['青云', '鹤氅', '星斗', '霞光', '云纹', '清风', '明月', '松风']
  },
  yunlv: {
    name: '云履',
    slot: 'yunlv',
    prefixes: ['踏云', '凌波', '追风', '缩地', '登天', '步虚', '御风', '赶月']
  },
  fabao: {
    name: '本命法宝',
    slot: 'fabao',
    prefixes: ['翻天印', '捆仙绳', '打神鞭', '阴阳镜', '斩仙剑', '乾坤圈', '混天绫', '定海珠']
  }
}
  
  // 灵宠稀有度（使用统一配置）
  const petRarities = itemQualities.pet
  
  // 装备品质选项
  const equipmentQualityOptions = computed(() => {
    return Object.keys(equipmentQualities).map(key => ({
      label: equipmentQualities[key].name,
      value: key
    }))
  })
  
  // 灵宠稀有度选项
  const petRarityOptions = computed(() => {
    return Object.keys(petRarities).map(key => ({
      label: petRarities[key].name,
      value: key
    }))
  })
  
  // 获取调整后的装备概率
  const adjustedEquipProbabilities = computed(() => {
    return getAdjustedEquipProbabilities(inventoryStore.wishlistEnabled, inventoryStore.selectedWishEquipQuality)
  })
  
  // 获取调整后的灵宠概率
  const adjustedPetProbabilities = computed(() => {
    return getAdjustedPetProbabilities(inventoryStore.wishlistEnabled, inventoryStore.selectedWishPetRarity)
  })
  

  // 处理抽卡
  const performGacha = async (count = 1) => {
    console.log(`[GachaMain] 用户开始抽卡，次数: ${count}`);
    if (gachaStore.isDrawing) {
      console.warn('[GachaMain] 抽卡正在进行中，忽略重复请求');
      return;
    }

    gachaStore.setDrawingState(true)
    gachaNumber.value = count

    try {
      // 调用后端API执行抽卡
      const token = getAuthToken()
      if (!token) {
        console.error('[GachaMain] 未找到认证令牌，无法执行抽卡');
        message.error('认证失败，请重新登录');
        return;
      }
      
      console.log(`[GachaMain] 调用后端API执行抽卡，类型: ${gachaType.value}`);
      const response = await performGachaAPI(token, gachaType.value, count, inventoryStore.wishlistEnabled)
      
      if (!response.success) {
        console.error(`[GachaMain] 抽卡API返回失败，消息: ${response.message}`);
        throw new Error(response.message || '抽卡失败')
      }

      // 更新灵石数量（从后端返回的数据中获取）
      if (response.spiritStones !== undefined) {
        console.log(`[GachaMain] 更新灵石数量: ${response.spiritStones}`);
        inventoryStore.spiritStones = response.spiritStones
      }

      // 显示抽卡动画
      console.log('[GachaMain] 显示抽卡动画');
      gachaStore.setShakingState(true)
      await new Promise(resolve => setTimeout(resolve, 1000))
      gachaStore.setShakingState(false)
      gachaStore.setOpeningState(true)
      await new Promise(resolve => setTimeout(resolve, 500))

      // 设置抽卡结果
      console.log(`[GachaMain] 设置抽卡结果，物品数量: ${response.items?.length || 0}`);
      // 处理物品品质信息，确保前端可以正确显示
      const processedItems = response.items.map(item => {
        if (item.type === 'equipment') {
          // 为装备添加品质信息
          return {
            ...item,
            qualityInfo: equipmentQualities[item.quality] || equipmentQualities.common
          };
        } else if (item.type === 'pet') {
          // 为灵宠添加品质信息（如果需要）
          return item;
        }
        return item;
      });
      
      gachaStore.setGachaResults(processedItems)
      gachaStore.toggleResultModal(true)
      gachaStore.resetPagination()

      // 重新获取玩家完整数据以同步数据库中的最新数据
      console.log('[GachaMain] 重新获取玩家完整数据以同步数据');
      try {
        const playerDataResponse = await APIService.initializePlayer(token);
        if (playerDataResponse.items) {
          console.log(`[GachaMain] 成功获取玩家物品列表，物品数量: ${playerDataResponse.items.length}`);
          inventoryStore.items = playerDataResponse.items;
        }
        
        // 更新灵宠列表
        if (playerDataResponse.pets) {
          console.log(`[GachaMain] 成功获取玩家灵宠列表，灵宠数量: ${playerDataResponse.pets.length}`);
          petsStore.pets = playerDataResponse.pets;
        }
      } catch (inventoryError) {
        console.error('[GachaMain] 获取玩家完整数据失败:', inventoryError);
        // 如果获取最新数据失败，仍然使用抽卡返回的结果
        inventoryStore.items = [...inventoryStore.items, ...response.items];
      }

      // 处理自动操作（如果后端也支持的话）

      // 显示自动操作结果

      message.success(`抽卡完成，获得了${count}件物品！`)
      console.log(`[GachaMain] 抽卡完成，获得了${count}件物品！`);
    } catch (error) {
      console.error('[GachaMain] 抽卡过程中发生错误:', error)
      message.error('抽卡失败: ' + error.message)
    } finally {
      gachaStore.setOpeningState(false)
      gachaStore.setDrawingState(false)
      console.log('[GachaMain] 抽卡流程结束');
    }
  }
  
  // 显示宠物详情
  const showPetDetail = (pet) => {
    console.log(`[GachaMain] 显示宠物详情: ${pet.name} (${pet.id})`);
    gachaStore.selectedPet = pet
    gachaStore.showPetDetails = true
  }
  
  // 显示装备详情
  const showEquipmentDetail = (equipment) => {
    console.log(`[GachaMain] 显示装备详情: ${equipment.name} (${equipment.id})`);
    gachaStore.selectedEquipment = equipment
    gachaStore.showEquipmentDetails = true
  }
  
  // 处理自动出售变更
  const handleAutoSellChange = (values) => {
    // 如果选择了"全部"，则清除其他选项
    if (values.includes('all')) {
      inventoryStore.autoSellQualities = ['all']
    }
    // 如果之前选择了"全部"而现在选择了其他选项，则移除"全部"
    else if (
      inventoryStore.autoSellQualities.includes('all') &&
      values.length > 0
    ) {
      inventoryStore.autoSellQualities = values.filter(v => v !== 'all')
    }
  }

  // 处理自动放生变更
  const handleAutoReleaseChange = (values) => {
    // 如果选择了"全部"，则清除其他选项
    if (values.includes('all')) {
      inventoryStore.autoReleaseRarities = ['all']
    }
    // 如果之前选择了"全部"而现在选择了其他选项，则移除"全部"
    else if (
      inventoryStore.autoReleaseRarities.includes('all') &&
      values.length > 0
    ) {
      inventoryStore.autoReleaseRarities = values.filter(v => v !== 'all')
    }
  }
  
  // 监听自动设置的变化
  watch(() => inventoryStore.autoSellQualities, (newVal) => {
    // 可以在这里添加额外的逻辑
  }, { deep: true })
  
  watch(() => inventoryStore.autoReleaseRarities, (newVal) => {
    // 可以在这里添加额外的逻辑
  }, { deep: true })
</script>

<style scoped>
  .gacha-content {
    padding: 20px;
    max-width: 1200px;
    margin: 0 auto;
  }
  
  .gacha-container {
    max-width: 800px;
    margin: 0 auto;
    text-align: center;
  }
  
  .gacha-type-selector {
    margin-bottom: 20px;
  }
  
  .spirit-stones {
    margin-bottom: 20px;
  }
</style>