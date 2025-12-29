<template>
  <div>
    <n-space style="margin-bottom: 16px">
      <n-select
        v-model:value="selectedRarityToRelease"
        :options="options"
        placeholder="选择卖出品阶"
        style="width: 150px"
      />
      <n-button
        @click="showBatchReleaseConfirm = true"
        :disabled="!props.playerInfoStore.pets.length"
      >
        一键卖出
      </n-button>
    </n-space>
    <n-modal v-model:show="showBatchReleaseConfirm" preset="dialog" title="批量卖出确认" style="width: 600px">
      <p>
        确定要卖出{{
          selectedRarityToRelease === 'all' ? '所有' : petRarities[selectedRarityToRelease].name
        }}品阶的未出战灵宠吗？此操作不可撤销。
      </p>
      <n-space justify="end" style="margin-top: 16px">
        <n-button size="small" @click="showBatchReleaseConfirm = false">取消</n-button>
        <n-button size="small" type="error" @click="batchReleasePets">确认卖出</n-button>
      </n-space>
    </n-modal>
    <n-pagination
      v-if="filteredPets.length > 12"
      v-model:page="currentPage"
      :page-size="pageSize"
      :item-count="filteredPets.length"
      @update:page-size="onPageSizeChange"
      :page-slot="7"
    />
    <div v-if="displayPets.length === 0" style="text-align: center; padding: 20px;">
      <n-empty description="暂无灵宠" />
      <p style="color: #999; margin-top: 10px;">通过抽奖可以获得灵宠</p>
    </div>
    <n-grid v-else-if="displayPets.length > 0" :cols="2" :x-gap="12" :y-gap="8" style="margin-top: 16px">
      <n-grid-item v-for="pet in displayPets" :key="pet.id">
        <n-card hoverable>
          <template #header>
            <n-space justify="space-between">
              <span>{{ pet.name }}</span>
              <n-button size="small" type="primary" @click="useItem(pet)">
                {{ pet.isActive ? '召回' : '出战' }}
              </n-button>
            </n-space>
          </template>
          <p>{{ pet.description }}</p>
          <n-space vertical>
            <n-tag :style="{ color: pet.rarity && petRarities[pet.rarity] ? petRarities[pet.rarity].color : '#000000' }">
              {{ pet.rarity && petRarities[pet.rarity] ? petRarities[pet.rarity].name : '未知品质' }}
            </n-tag>
            <n-space justify="space-between">
              <n-text>等级: {{ pet.level || 1 }}</n-text>
              <n-text>星级: {{ pet.star || 0 }}</n-text>
              <n-button size="small" @click="showPetDetails(pet)">详情</n-button>
            </n-space>
          </n-space>
        </n-card>
      </n-grid-item>
    </n-grid>

    <!-- 灵宠详情弹窗 -->
    <n-modal v-model:show="showPetModal" preset="dialog" title="灵宠详情" style="width: 600px">
      <template v-if="selectedPet">
        <n-descriptions bordered>
          <n-descriptions-item label="名称">{{ selectedPet.name }}</n-descriptions-item>
          <n-descriptions-item label="品质">
            <n-tag :style="{ color: selectedPet.rarity && petRarities[selectedPet.rarity] ? petRarities[selectedPet.rarity].color : '#000000' }">
              {{ selectedPet.rarity && petRarities[selectedPet.rarity] ? petRarities[selectedPet.rarity].name : '未知品质' }}
            </n-tag>
          </n-descriptions-item>
          <n-descriptions-item label="等级">{{ selectedPet.level || 1 }}</n-descriptions-item>
          <n-descriptions-item label="星级">{{ selectedPet.star || 0 }}</n-descriptions-item>
          <n-descriptions-item label="境界">{{ Math.floor((selectedPet.star || 0) / 5) }}阶</n-descriptions-item>
        </n-descriptions>
        <n-divider>属性加成</n-divider>
        <n-descriptions bordered>
          <n-descriptions-item label="攻击加成">
            +{{ ((selectedPet.bonus?.attack || selectedPet.attackBonus || 0) * 100).toFixed(1) }}%
          </n-descriptions-item>
          <n-descriptions-item label="防御加成">
            +{{ ((selectedPet.bonus?.defense || selectedPet.defenseBonus || 0) * 100).toFixed(1) }}%
          </n-descriptions-item>
          <n-descriptions-item label="生命加成">
            +{{ ((selectedPet.bonus?.health || selectedPet.healthBonus || 0) * 100).toFixed(1) }}%
          </n-descriptions-item>
        </n-descriptions>
        <n-divider>灵宠属性</n-divider>
        <n-collapse>
          <n-collapse-item title="展开" name="1">
            <n-divider>基础属性</n-divider>
            <n-descriptions bordered :column="2">
              <n-descriptions-item label="攻击力">{{ selectedPet.combatAttributes?.attack || 0 }}</n-descriptions-item>
              <n-descriptions-item label="生命值">{{ selectedPet.combatAttributes?.health || 0 }}</n-descriptions-item>
              <n-descriptions-item label="防御力">{{ selectedPet.combatAttributes?.defense || 0 }}</n-descriptions-item>
              <n-descriptions-item label="速度">{{ selectedPet.combatAttributes?.speed || 0 }}</n-descriptions-item>
            </n-descriptions>
            <n-divider>战斗属性</n-divider>
            <n-descriptions bordered :column="3">
              <n-descriptions-item label="暴击率">
                {{ ((selectedPet.combatAttributes?.critRate || 0) * 100).toFixed(1) }}%
              </n-descriptions-item>
              <n-descriptions-item label="连击率">
                {{ ((selectedPet.combatAttributes?.comboRate || 0) * 100).toFixed(1) }}%
              </n-descriptions-item>
              <n-descriptions-item label="反击率">
                {{ ((selectedPet.combatAttributes?.counterRate || 0) * 100).toFixed(1) }}%
              </n-descriptions-item>
              <n-descriptions-item label="眩晕率">
                {{ ((selectedPet.combatAttributes?.stunRate || 0) * 100).toFixed(1) }}%
              </n-descriptions-item>
              <n-descriptions-item label="闪避率">
                {{ ((selectedPet.combatAttributes?.dodgeRate || 0) * 100).toFixed(1) }}%
              </n-descriptions-item>
              <n-descriptions-item label="吸血率">
                {{ ((selectedPet.combatAttributes?.vampireRate || 0) * 100).toFixed(1) }}%
              </n-descriptions-item>
            </n-descriptions>
            <n-divider>战斗抗性</n-divider>
            <n-descriptions bordered :column="3">
              <n-descriptions-item label="抗暴击">
                {{ ((selectedPet.combatAttributes?.critResist || 0) * 100).toFixed(1) }}%
              </n-descriptions-item>
              <n-descriptions-item label="抗连击">
                {{ ((selectedPet.combatAttributes?.comboResist || 0) * 100).toFixed(1) }}%
              </n-descriptions-item>
              <n-descriptions-item label="抗反击">
                {{ ((selectedPet.combatAttributes?.counterResist || 0) * 100).toFixed(1) }}%
              </n-descriptions-item>
              <n-descriptions-item label="抗眩晕">
                {{ ((selectedPet.combatAttributes?.stunResist || 0) * 100).toFixed(1) }}%
              </n-descriptions-item>
              <n-descriptions-item label="抗闪避">
                {{ ((selectedPet.combatAttributes?.dodgeResist || 0) * 100).toFixed(1) }}%
              </n-descriptions-item>
              <n-descriptions-item label="抗吸血">
                {{ ((selectedPet.combatAttributes?.vampireResist || 0) * 100).toFixed(1) }}%
              </n-descriptions-item>
            </n-descriptions>
            <n-divider>特殊属性</n-divider>
            <n-descriptions bordered :column="3">
              <n-descriptions-item label="强化治疗">
                {{ ((selectedPet.combatAttributes?.healBoost || 0) * 100).toFixed(1) }}%
              </n-descriptions-item>
              <n-descriptions-item label="强化爆伤">
                {{ ((selectedPet.combatAttributes?.critDamageBoost || 0) * 100).toFixed(1) }}%
              </n-descriptions-item>
              <n-descriptions-item label="弱化爆伤">
                {{ ((selectedPet.combatAttributes?.critDamageReduce || 0) * 100).toFixed(1) }}%
              </n-descriptions-item>
              <n-descriptions-item label="最终增伤">
                {{ ((selectedPet.combatAttributes?.finalDamageBoost || 0) * 100).toFixed(1) }}%
              </n-descriptions-item>
              <n-descriptions-item label="最终减伤">
                {{ ((selectedPet.combatAttributes?.finalDamageReduce || 0) * 100).toFixed(1) }}%
              </n-descriptions-item>
              <n-descriptions-item label="战斗属性提升">
                {{ ((selectedPet.combatAttributes?.combatBoost || 0) * 100).toFixed(1) }}%
              </n-descriptions-item>
              <n-descriptions-item label="战斗抗性提升">
                {{ ((selectedPet.combatAttributes?.resistanceBoost || 0) * 100).toFixed(1) }}%
              </n-descriptions-item>
            </n-descriptions>
          </n-collapse-item>
        </n-collapse>
        <n-divider>操作</n-divider>
        <n-space vertical>
          <n-space justify="space-between">
            <span>升级（消耗{{ getUpgradeCost(selectedPet) }} / {{ playerInfoStore.petEssence }}灵宠精华）</span>
            <n-button size="small" type="primary" @click="upgradePet(selectedPet)" :disabled="!canUpgrade(selectedPet)">
              升级
            </n-button>
          </n-space>
          <n-space justify="space-between">
            <span>升星（同品质灵宠，相同名字成功率100%，不同名字成功率30%）</span>
            <n-select
              v-model:value="selectedFoodPet"
              :options="getAvailableFoodPets(selectedPet)"
              placeholder="选择升星材料"
              style="width: 200px"
            />
            <n-button size="small" type="warning" @click="evolvePet(selectedPet)" :disabled="!selectedFoodPet">
              升星
            </n-button>
          </n-space>
          <n-space justify="space-between">
            <span>卖出灵宠（不会返还已消耗的道具）</span>
            <n-button size="small" type="error" @click="confirmReleasePet(selectedPet)">卖出灵宠</n-button>
            <n-modal v-model:show="showReleaseConfirm" preset="dialog" title="灵宠卖出" style="width: 600px">
              <template v-if="petToRelease">
                <p>确定要卖出 {{ petToRelease.name }} 吗？此操作不可撤销，且不会返还已消耗的道具。</p>
                <n-space justify="end" style="margin-top: 16px">
                  <n-button size="small" @click="cancelReleasePet">取消</n-button>
                  <n-button size="small" type="error" @click="releasePet">确认卖出</n-button>
                </n-space>
              </template>
            </n-modal>
          </n-space>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script setup>
  import { ref, computed } from 'vue'
  import { useMessage } from 'naive-ui'
  import { itemQualities } from '../plugins/itemQualities'
  import APIService from '../services/api.js'
  import { getAuthToken } from '../stores/db'

  // Props
  const props = defineProps({
    playerInfoStore: {
      type: Object,
      required: true
    }
  })

  const emit = defineEmits([
    'refreshPetList'
  ])

  const message = useMessage()

  // 分页相关
  const currentPage = ref(1)
  const pageSize = ref(12)

  // 选中的卖出品阶
  const selectedRarityToRelease = ref('all')

  // 灵宠品质配置（使用统一配置）
  const petRarities = itemQualities.pet

  // 灵宠详情相关
  const showPetModal = ref(false)
  const selectedPet = ref(null)
  const selectedFoodPet = ref(null)

  // 卖出确认弹窗
  const showReleaseConfirm = ref(false)
  const showBatchReleaseConfirm = ref(false)
  const petToRelease = ref(null)

  // 过滤后的灵宠列表
  const filteredPets = computed(() => {
    const pets = props.playerInfoStore.pets
    if (selectedRarityToRelease.value === 'all') {
      return pets
    }
    const filtered = pets.filter(pet => pet.rarity === selectedRarityToRelease.value);
    return filtered;
  })

  // 稀有度排序权重（数字越小，排序优先级越高）
  const rarityOrder = {
    mythic: 1,
    legendary: 2,
    epic: 3,
    rare: 4,
    uncommon: 5,
    common: 6
  }

  // 当前页显示的灵宠
  const displayPets = computed(() => {
    const start = (currentPage.value - 1) * pageSize.value
    const end = start + pageSize.value
    let displayed = filteredPets.value.slice(start, end);
    
    // 对灵宠列表进行排序：先按稀有度高低排序，再按出战状态排序
    displayed = displayed.sort((a, b) => {
      // 按稀有度排序（稀有度高的排在前面）
      const rarityA = rarityOrder[a.rarity] || 999
      const rarityB = rarityOrder[b.rarity] || 999
      if (rarityA !== rarityB) {
        return rarityA - rarityB
      }
      
      // 同稀有度下，已出战的灵宠排在前面
      if (a.isActive && !b.isActive) return -1;
      if (!a.isActive && b.isActive) return 1;
      return 0;
    });
    
    return displayed;
  })

  // 页大小改变处理
  const onPageSizeChange = size => {
    pageSize.value = size
    currentPage.value = 1
  }

  const options = [
    { label: '全部品阶', value: 'all' },
    // 使用统一的灵宠品质配置
    { label: petRarities.mythic.name, value: 'mythic' },
    { label: petRarities.legendary.name, value: 'legendary' },
    { label: petRarities.epic.name, value: 'epic' },
    { label: petRarities.rare.name, value: 'rare' },
    { label: petRarities.uncommon.name, value: 'uncommon' },
    { label: petRarities.common.name, value: 'common' }
  ]

  // 显示卖出确认弹窗
  const confirmReleasePet = pet => {
    petToRelease.value = pet
    showReleaseConfirm.value = true
  }

  // 取消卖出
  const cancelReleasePet = () => {
    petToRelease.value = null
    showReleaseConfirm.value = false
  }

  // 执行卖出
  const releasePet = async () => {
    if (petToRelease.value) {
      const token = getAuthToken()
      try {
        // 调用API卖出灵宠
        const response = await APIService.deletePets(token, [petToRelease.value.id])
        if (response.success) {
          // ✅ 优化：显示灵宠名称和灵石奖励
          const spiritStones = response.spiritStones || 0
          message.success(`已卖出 ${petToRelease.value.name}，获得 ${spiritStones} 灵石`)
        } else {
          message.error(response.message || '卖出失败')
        }
      } catch (error) {
        console.error('卖出灵宠失败:', error)
        message.error('卖出灵宠失败: ' + error.message)
      }
      
      // 关闭所有相关弹窗
      showReleaseConfirm.value = false
      showPetModal.value = false
      petToRelease.value = null
    }
  }

  // 显示灵宠详情
  const showPetDetails = pet => {
    selectedPet.value = pet
    selectedFoodPet.value = null
    showPetModal.value = true
  }

  // 获取升级所需精华数量
  const getUpgradeCost = pet => {
    return (pet.level || 1) * 10
  }

  // 检查是否可以升级
  const canUpgrade = pet => {
    const cost = getUpgradeCost(pet)
    return props.playerInfoStore.petEssence >= cost
  }

  // 获取可用作升星材料的灵宠列表
  const getAvailableFoodPets = pet => {
    if (!pet) return []
    return props.playerInfoStore.pets
      .filter(
        item =>
          item.id !== pet.id &&
          item.star === pet.star &&
          item.rarity === pet.rarity
      )
      .map(item => ({
        label: `${item.name} (${item.level || 1}级 ${item.star || 0}星)${item.name !== pet.name ? ' [成功率30%]' : ''}`,
        value: item.id
      }))
  }

  // 升级灵宠
  const upgradePet = async pet => {
    const token = getAuthToken()
    try {
      const response = await APIService.upgradePet(token, pet.id, getUpgradeCost(pet))
      if (response.success) {
        message.success('升级成功')
        // ✅ 立即更新本地灵宠数据
        if (response.pet) {
          const petIndex = props.playerInfoStore.pets.findIndex(p => p.id === pet.id)
          if (petIndex !== -1) {
            props.playerInfoStore.pets[petIndex] = response.pet
            // ✅ 同时更新弹窗中显示的灵宠详情
            if (selectedPet.value && selectedPet.value.id === pet.id) {
              selectedPet.value = response.pet
            }
          }
        }
        // ✅ 新增：如果灵宠是出战状态，后端会返回更新后的玩家属性，需要同步
        if (response.user) {
          props.playerInfoStore.$patch({
            baseAttributes: response.user.baseAttributes,
            combatAttributes: response.user.combatAttributes,
            combatResistance: response.user.combatResistance,
            specialAttributes: response.user.specialAttributes
          })
        }
      } else {
        message.error(response.message || '升级失败')
      }
    } catch (error) {
      console.error('升级灵宠失败:', error)
      message.error('升级灵宠失败: ' + error.message)
    }
  }

  // 升星灵宠
  const evolvePet = async pet => {
    if (!selectedFoodPet.value) {
      message.error('请选择用于升星的灵宠')
      return
    }
    // 通过id查找对应的灵宠对象
    const foodPet = props.playerInfoStore.pets.find(item => item.id === selectedFoodPet.value)
    if (!foodPet) {
      message.error('升星材料灵宠不存在')
      return
    }
      
    const token = getAuthToken()
    try {
      const response = await APIService.evolvePet(token, pet.id, foodPet.id)
      if (response.success) {
        message.success('升星成功')
        // ✅ 立即更新本地灵宠数据
        if (response.pet) {
          const petIndex = props.playerInfoStore.pets.findIndex(p => p.id === pet.id)
          if (petIndex !== -1) {
            props.playerInfoStore.pets[petIndex] = response.pet
            // ✅ 同时更新弹窗中显示的灵宠详情（关闭ui前）
            if (selectedPet.value && selectedPet.value.id === pet.id) {
              selectedPet.value = response.pet
            }
          }
          // 移除材料灵宠
          const foodPetIndex = props.playerInfoStore.pets.findIndex(p => p.id === foodPet.id)
          if (foodPetIndex !== -1) {
            props.playerInfoStore.pets.splice(foodPetIndex, 1)
          }
        }
        // ✅ 新增：如果灵宠是出战状态，后端会返回更新后的玩家属性，需要同步
        if (response.user) {
          props.playerInfoStore.$patch({
            baseAttributes: response.user.baseAttributes,
            combatAttributes: response.user.combatAttributes,
            combatResistance: response.user.combatResistance,
            specialAttributes: response.user.specialAttributes
          })
        }
        selectedFoodPet.value = null
        showPetModal.value = false
      } else {
        // 升星失败，昺示成功率信息
        const successRateInfo = response.successRate ? ` (成功率${response.successRate}%)` : ''
        message.error(response.message + successRateInfo)
      }
    } catch (error) {
      console.error('升星灵宠失败:', error)
      message.error('升星灵宠失败: ' + error.message)
    }
  }

  // 批量释放宠物
  const batchReleasePets = async () => {
    const rarity = selectedRarityToRelease.value === 'all' ? null : selectedRarityToRelease.value
    const token = getAuthToken()
    try {
      const response = await APIService.batchReleasePets(token, { rarity })
      if (response.success) {
        message.success('批量卖出成功')
        showBatchReleaseConfirm.value = false
        // 刷新灵宠列表
        emit('refreshPetList')
      } else {
        message.error(response.message || '批量卖出失败')
      }
    } catch (error) {
      console.error('批量卖出灵宠失败:', error)
      message.error('批量卖出灵宠失败: ' + error.message)
    }
  }

  // 使用物品
  const useItem = async item => {
    // 处理灵宠出战/召回
    const token = getAuthToken();

    // 检查灵宠是否处于出战状态
    if (item.isActive) {
      // 召回灵宠
      try {
        const response = await APIService.recallPet(token, item.id);
        if (response.success) {
          message.success('召回成功');
          // 更新玩家数据
          // 清除当前出战的灵宠
          props.playerInfoStore.activePet = response.pet || null;
          
          // 确保被召回的灵宠对象的isActive属性也被更新
          const petInStore = props.playerInfoStore.pets.find(p => p.id === item.id);
          if (petInStore) {
            petInStore.isActive = false;
          }
          
          // 重新加载玩家数据以获取更新后的属性
          try {
            const playerDataResponse = await APIService.getPlayerData(token);
            if (playerDataResponse.user) {
              Object.assign(props.playerInfoStore, playerDataResponse.user);
            }
          } catch (error) {
            console.error('获取更新后的玩家数据失败:', error);
          }
          
          // 重新加载灵宠列表以确保数据同步
          emit('refreshPetList');
        } else {
          message.error(response.message || '召回失败');
        }
      } catch (error) {
        console.error('召回灵宠失败:', error);
        message.error('召回灵宠失败: ' + error.message);
      }
    } else {
      // 出战灵宠前打印玩家当前属性
      console.log('出战灵宠前的玩家属性:', {
        name: props.playerInfoStore.name,
        level: props.playerInfoStore.level,
        realm: props.playerInfoStore.realm,
        baseAttributes: {
          attack: props.playerInfoStore.baseAttributes.attack,
          health: props.playerInfoStore.baseAttributes.health,
          defense: props.playerInfoStore.baseAttributes.defense,
          speed: props.playerInfoStore.baseAttributes.speed
        },
        combatAttributes: {
          critRate: props.playerInfoStore.combatAttributes.critRate,
          comboRate: props.playerInfoStore.combatAttributes.comboRate,
          counterRate: props.playerInfoStore.combatAttributes.counterRate,
          stunRate: props.playerInfoStore.combatAttributes.stunRate,
          dodgeRate: props.playerInfoStore.combatAttributes.dodgeRate,
          vampireRate: props.playerInfoStore.combatAttributes.vampireRate
        },
        combatResistance: {
          critResist: props.playerInfoStore.combatResistance.critResist,
          comboResist: props.playerInfoStore.combatResistance.comboResist,
          counterResist: props.playerInfoStore.combatResistance.counterResist,
          stunResist: props.playerInfoStore.combatResistance.stunResist,
          dodgeResist: props.playerInfoStore.combatResistance.dodgeResist,
          vampireResist: props.playerInfoStore.combatResistance.vampireResist
        },
        specialAttributes: {
          healBoost: props.playerInfoStore.specialAttributes.healBoost,
          critDamageBoost: props.playerInfoStore.specialAttributes.critDamageBoost,
          critDamageReduce: props.playerInfoStore.specialAttributes.critDamageReduce,
          finalDamageBoost: props.playerInfoStore.specialAttributes.finalDamageBoost,
          finalDamageReduce: props.playerInfoStore.specialAttributes.finalDamageReduce,
          combatBoost: props.playerInfoStore.specialAttributes.combatBoost,
          resistanceBoost: props.playerInfoStore.specialAttributes.resistanceBoost
        },
        activePet: props.playerInfoStore.activePet
      });

      // 出战灵宠前打印灵宠属性
      console.log('准备出战的灵宠属性:', {
        id: item.id,
        name: item.name,
        rarity: item.rarity,
        level: item.level,
        star: item.star,
        description: item.description,
        combatAttributes: item.combatAttributes,
        bonus: {
          attackBonus: item.attackBonus,
          defenseBonus: item.defenseBonus,
          healthBonus: item.healthBonus
        }
      });
      
      // 出战灵宠
      try {
        const response = await APIService.deployPet(token, item.id);
        if (response.success) {
          message.success('出战成功');
          // 更新玩家数据
          // 更新当前出战的灵宠
          props.playerInfoStore.activePet = response.pet || item;
          
          // 确保出战的灵宠对象的isActive属性也被更新
          const petInStore = props.playerInfoStore.pets.find(p => p.id === item.id);
          if (petInStore) {
            petInStore.isActive = true;
          }
          
          // 重新加载玩家数据以获取更新后的属性
          try {
            const playerDataResponse = await APIService.getPlayerData(token);
            if (playerDataResponse.user) {
              Object.assign(props.playerInfoStore, playerDataResponse.user);
            }
          } catch (error) {
            console.error('获取更新后的玩家数据失败:', error);
          }
          
          // 出战灵宠后打印玩家更新的属性
          console.log('出战灵宠后的玩家属性:', {
            name: props.playerInfoStore.name,
            level: props.playerInfoStore.level,
            realm: props.playerInfoStore.realm,
            baseAttributes: {
              attack: props.playerInfoStore.baseAttributes.attack,
              health: props.playerInfoStore.baseAttributes.health,
              defense: props.playerInfoStore.baseAttributes.defense,
              speed: props.playerInfoStore.baseAttributes.speed
            },
            combatAttributes: {
              critRate: props.playerInfoStore.combatAttributes.critRate,
              comboRate: props.playerInfoStore.combatAttributes.comboRate,
              counterRate: props.playerInfoStore.combatAttributes.counterRate,
              stunRate: props.playerInfoStore.combatAttributes.stunRate,
              dodgeRate: props.playerInfoStore.combatAttributes.dodgeRate,
              vampireRate: props.playerInfoStore.combatAttributes.vampireRate
            },
            combatResistance: {
              critResist: props.playerInfoStore.combatResistance.critResist,
              comboResist: props.playerInfoStore.combatResistance.comboResist,
              counterResist: props.playerInfoStore.combatResistance.counterResist,
              stunResist: props.playerInfoStore.combatResistance.stunResist,
              dodgeResist: props.playerInfoStore.combatResistance.dodgeResist,
              vampireResist: props.playerInfoStore.combatResistance.vampireResist
            },
            specialAttributes: {
              healBoost: props.playerInfoStore.specialAttributes.healBoost,
              critDamageBoost: props.playerInfoStore.specialAttributes.critDamageBoost,
              critDamageReduce: props.playerInfoStore.specialAttributes.critDamageReduce,
              finalDamageBoost: props.playerInfoStore.specialAttributes.finalDamageBoost,
              finalDamageReduce: props.playerInfoStore.specialAttributes.finalDamageReduce,
              combatBoost: props.playerInfoStore.specialAttributes.combatBoost,
              resistanceBoost: props.playerInfoStore.specialAttributes.resistanceBoost
            },
            activePet: props.playerInfoStore.activePet
          });
          
          // 重新加载灵宠列表以确保数据同步
          emit('refreshPetList');
        } else {
          message.error(response.message || '出战失败');
        }
      } catch (error) {
        console.error('出战灵宠失败:', error);
        message.error('出战灵宠失败: ' + error.message);
      }
    }
  }
</script>

<style scoped>
</style>