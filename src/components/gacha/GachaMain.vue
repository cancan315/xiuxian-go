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
            <n-statistic label="灵石" :value="playerInfoStore.spiritStones" />
          </div>
          
          <GachaCard 
            :gacha-type="gachaType"
            :is-shaking="playerInfoStore.isShaking"
            :is-opening="playerInfoStore.isOpening"
          />
          
          <GachaButtons
            :spirit-stones="playerInfoStore.spiritStones"
            :wishlist-enabled="playerInfoStore.wishlistEnabled"
            :is-drawing="playerInfoStore.isDrawing"
            @gacha="performGacha"
            @show-probability="playerInfoStore.showProbabilityInfo = true"
            @show-wishlist="playerInfoStore.showWishlistSettings = true"
            @show-auto-settings="playerInfoStore.showAutoSettings = true"
          />
        </div>
      </n-card>
    </n-layout-content>
    
    <!-- 弹窗组件 -->
    <GachaResultModal
      v-model:show="playerInfoStore.showResultModal"
      v-model:current-page="playerInfoStore.currentPage"
      :gacha-type="gachaType"
      :gacha-number="gachaNumber"
      :spirit-stones="playerInfoStore.spiritStones"
      :wishlist-enabled="playerInfoStore.wishlistEnabled"
      :is-drawing="playerInfoStore.isDrawing"
      :current-page-results="playerInfoStore.currentPageResults"
      :total-pages="playerInfoStore.totalPages"
      :page-size="playerInfoStore.pageSize"
      :selected-wish-equip-quality="playerInfoStore.selectedWishEquipQuality"
      :selected-wish-pet-rarity="playerInfoStore.selectedWishPetRarity"
      :pet-rarities="petRarities"
      :equipment-quality-options="equipmentQualityOptions"
      :pet-rarity-options="petRarityOptions"
      :equipment-types-map="equipmentTypes"
      @perform-gacha="performGacha"
      @show-pet-detail="showPetDetail"
      @show-equipment-detail="showEquipmentDetail"
    />
    
    <ProbabilityInfoModal
      v-model:show="playerInfoStore.showProbabilityInfo"
      :wishlist-enabled="playerInfoStore.wishlistEnabled"
      :selected-wish-equip-quality="playerInfoStore.selectedWishEquipQuality"
      :selected-wish-pet-rarity="playerInfoStore.selectedWishPetRarity"
      :equipment-qualities="equipmentQualities"
      :pet-rarities="petRarities"
      :adjusted-equip-probabilities="adjustedEquipProbabilities"
      :adjusted-pet-probabilities="adjustedPetProbabilities"
    />
    
    <WishlistSettingsModal
      v-model:show="playerInfoStore.showWishlistSettings"
      v-model:wishlist-enabled="playerInfoStore.wishlistEnabled"
      v-model:selected-wish-equip-quality="playerInfoStore.selectedWishEquipQuality"
      v-model:selected-wish-pet-rarity="playerInfoStore.selectedWishPetRarity"
      :equipment-qualities="equipmentQualities"
      :pet-rarities="petRarities"
      :equipment-quality-options="equipmentQualityOptions"
      :pet-rarity-options="petRarityOptions"
    />
    
    <AutoSettingsModal
      v-model:show="playerInfoStore.showAutoSettings"
      v-model:auto-sell-qualities="playerInfoStore.autoSellQualities"
      v-model:auto-release-rarities="playerInfoStore.autoReleaseRarities"
      :equipment-qualities="equipmentQualities"
      :pet-rarities="petRarities"
    />
    
    <PetDetailModal
      v-model:show="playerInfoStore.showPetDetails"
      :pet="playerInfoStore.selectedPet"
      :pet-rarities="petRarities"
    />
    
    <EquipmentDetailModal
      v-model:show="playerInfoStore.showEquipmentDetails"
      :equipment="playerInfoStore.selectedEquipment"
      :equipment-types="equipmentTypes"
    />
  </n-layout>
</template>

<script setup>
  import { ref, computed, watch } from 'vue'
  import { useMessage } from 'naive-ui'
  // 修改为使用模块化store
  import { usePlayerInfoStore } from '../../stores/playerInfo'
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
    return getAdjustedEquipProbabilities(playerInfoStore.wishlistEnabled, playerInfoStore.selectedWishEquipQuality)
  })
  
  // 获取调整后的灵宠概率
  const adjustedPetProbabilities = computed(() => {
    return getAdjustedPetProbabilities(playerInfoStore.wishlistEnabled, playerInfoStore.selectedWishPetRarity)
  })
  

  // 处理抽卡
  const performGacha = async (count = 1) => {
    console.log(`[GachaMain] 用户开始抽卡，次数: ${count}`);
    console.log(`[GachaMain] 抽卡前灵石数量: ${playerInfoStore.spiritStones}`);
    if (playerInfoStore.isDrawing) {
      console.warn('[GachaMain] 抽卡正在进行中，忽略重复请求');
      return;
    }

    playerInfoStore.setDrawingState(true)
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
      const response = await performGachaAPI(token, gachaType.value, count, playerInfoStore.wishlistEnabled)
      
      if (!response.success) {
        console.error(`[GachaMain] 抽卡API返回失败，消息: ${response.message}`);
        throw new Error(response.message || '抽卡失败')
      }

      // 记录抽卡前后的灵石数量
      console.log(`[GachaMain] 抽卡后检查response:`, response);
      if (response.spiritStones !== undefined) {
        console.log(`[GachaMain] 抽卡后灵石数量 (spiritStones): ${response.spiritStones}`);
        playerInfoStore.spiritStones = response.spiritStones;
      } else if (response.spirit_stones !== undefined) {
        console.log(`[GachaMain] 抽卡后灵石数量 (spirit_stones): ${response.spirit_stones}`);
        playerInfoStore.spiritStones = response.spirit_stones;
      } else {
        console.log(`[GachaMain] response.spiritStones 和 response.spirit_stones 都不存在`);
      }
      console.log(`[GachaMain] 灵石数量更新完成`);

      // 显示抽卡动画
      console.log('[GachaMain] 显示抽卡动画');
      playerInfoStore.setShakingState(true)
      await new Promise(resolve => setTimeout(resolve, 1000))
      playerInfoStore.setShakingState(false)
      playerInfoStore.setOpeningState(true)
      await new Promise(resolve => setTimeout(resolve, 500))

      // 设置抽卡结果
      console.log(`[GachaMain] 设置抽卡结果，物品数量: ${response.items?.length || 0}`);
      // 处理物品品质信息，确保前端可以正确显示
      const processedItems = response.items.map(item => {
        if (item.type === 'equipment') {
          // 为装备添加品质信息
          const qualityInfo = equipmentQualities[item.quality];
          if (!qualityInfo) {
            console.warn(`[GachaMain] 未找到装备品质信息: ${item.quality}`);
            // 仅使用从后端获取的数据，不使用前端默认的 common 品质
            return item;
          }
          return {
            ...item,
            qualityInfo
          };
        } else if (item.type === 'pet') {
          // 为灵宠添加稀有度信息（如果需要）
          const rarityInfo = petRarities[item.rarity];
          if (!rarityInfo) {
            console.warn(`[GachaMain] 未找到灵宠稀有度信息: ${item.rarity}`);
            // 仅使用从后端获取的数据，不使用前端默认的 common 品质
            return item;
          }
          return {
            ...item,
            rarityInfo
          };
        }
        return item;
      });
      
      playerInfoStore.setGachaResults(processedItems)
      playerInfoStore.toggleResultModal(true)
      playerInfoStore.resetPagination()
      
      // 打印 playerInfoStore 中的数据日志
      console.log('[GachaMain] playerInfoStore 数据:', {
        gachaResults: playerInfoStore.gachaResults,
        showResultModal: playerInfoStore.showResultModal,
        currentPage: playerInfoStore.currentPage,
        pageSize: playerInfoStore.pageSize,
        totalPages: playerInfoStore.totalPages,
        currentPageResults: playerInfoStore.currentPageResults
      })
      
      // 重新获取玩家完整数据以同步数据库中的最新数据
      console.log('[GachaMain] 重新获取玩家完整数据以同步数据');
      try {
        const playerDataResponse = await APIService.getPlayerData(token);
        if (playerDataResponse.items) {
          console.log(`[GachaMain] 成功获取玩家物品列表，物品数量: ${playerDataResponse.items.length}`);
          playerInfoStore.items = playerDataResponse.items;
        }
        
        // 更新灵宠列表
        if (playerDataResponse.pets) {
          console.log(`[GachaMain] 成功获取玩家灵宠列表，灵宠数量: ${playerDataResponse.pets.length}`);
          playerInfoStore.pets = playerDataResponse.pets;
        }
      } catch (inventoryError) {
        console.error('[GachaMain] 获取玩家完整数据失败:', inventoryError);
        // 如果获取最新数据失败，仍然使用抽卡返回的结果
        if (response.items) {
          playerInfoStore.items = [...playerInfoStore.items, ...response.items];
        }
      }

      // 处理自动操作（如果后端也支持的话）

      // 显示自动操作结果
      // 更新灵石数量（从后端返回的数据中获取）
      
      message.success(`抽卡完成，获得了${count}件物品！`)
      console.log(`[GachaMain] 抽卡完成，获得了${count}件物品！`);
    } catch (error) {
      console.error('[GachaMain] 抽卡过程中发生错误:', error)
      message.error('抽卡失败: ' + error.message)
    } finally {
      playerInfoStore.setOpeningState(false)
      playerInfoStore.setDrawingState(false)
      console.log('[GachaMain] 抽卡流程结束');
    }
    
  }
  
  // 显示宠物详情
  const showPetDetail = (pet) => {
    console.log(`[GachaMain] 显示宠物详情: ${pet.name} (${pet.id})`);
    playerInfoStore.selectedPet = pet
    playerInfoStore.showPetDetails = true
  }
  
  // 显示装备详情
  const showEquipmentDetail = (equipment) => {
    console.log(`[GachaMain] 显示装备详情: ${equipment.name} (${equipment.id})`);
    playerInfoStore.selectedEquipment = equipment
    playerInfoStore.showEquipmentDetails = true
  }
  
  // 处理自动出售变更
  const handleAutoSellChange = (values) => {
    // 如果选择了"全部"，则清除其他选项
    if (values.includes('all')) {
      playerInfoStore.autoSellQualities = ['all']
    }
    // 如果之前选择了"全部"而现在选择了其他选项，则移除"全部"
    else if (
      playerInfoStore.autoSellQualities.includes('all') &&
      values.length > 0
    ) {
      playerInfoStore.autoSellQualities = values.filter(v => v !== 'all')
    }
  }

  // 处理自动放生变更
  const handleAutoReleaseChange = (values) => {
    // 如果选择了"全部"，则清除其他选项
    if (values.includes('all')) {
      playerInfoStore.autoReleaseRarities = ['all']
    }
    // 如果之前选择了"全部"而现在选择了其他选项，则移除"全部"
    else if (
      playerInfoStore.autoReleaseRarities.includes('all') &&
      values.length > 0
    ) {
      playerInfoStore.autoReleaseRarities = values.filter(v => v !== 'all')
    }
  }
  
  // 监听自动设置的变化
  watch(() => playerInfoStore.autoSellQualities, (newVal) => {
    // 可以在这里添加额外的逻辑
  }, { deep: true })
  
  watch(() => playerInfoStore.autoReleaseRarities, (newVal) => {
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