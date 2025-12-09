<template>
  <n-layout>
    <n-layout-header bordered>
      <n-page-header>
        <template #title>背包</template>
      </n-page-header>
    </n-layout-header>
    <n-layout-content>
      <n-card :bordered="false">
        <n-tabs type="line">
          <n-tab-pane name="equipment" tab="装备">
            <EquipmentTab 
              :player-info-store="playerInfoStore"
              :inventory-store="inventoryStore"
              :equipment-store="equipmentStore"
              :persistence-store="persistenceStore"
              @load-equipped-artifacts="loadEquippedArtifacts"
              @refresh-pet-list="loadPetList"
            />
          </n-tab-pane>
          <n-tab-pane name="herbs" tab="灵草">
            <HerbsTab :inventory-store="inventoryStore" />
          </n-tab-pane>
          <n-tab-pane name="pills" tab="丹药">
            <PillsTab 
              :inventory-store="inventoryStore"
              :pills-store="pillsStore"
              :pets-store="petsStore"
              :player-info-store="playerInfoStore"
              :persistence-store="persistenceStore"
            />
          </n-tab-pane>
          <n-tab-pane name="formulas" tab="丹方">
            <FormulasTab :inventory-store="inventoryStore" />
          </n-tab-pane>
          <n-tab-pane name="pets" tab="灵宠">
            <PetsTab
              :player-info-store="playerInfoStore"
              :pets-store="petsStore"
              @refresh-pet-list="loadPetList"
            />
          </n-tab-pane>
        </n-tabs>
      </n-card>
    </n-layout-content>
  </n-layout>
</template>

<script setup>
  // 修改为使用模块化store
  import { usePlayerInfoStore } from '../stores/playerInfo'
  import { useInventoryStore } from '../stores/inventory'
  import { useEquipmentStore } from '../stores/equipment'
  import { usePetsStore } from '../stores/pets'
  import { usePillsStore } from '../stores/pills'
  import { useSettingsStore } from '../stores/settings'
  import { useStatsStore } from '../stores/stats'
  import { usePersistenceStore } from '../stores/persistence'
  import { ref, computed, onMounted } from 'vue'
  import { useMessage } from 'naive-ui'
  import { getStatName, formatStatValue } from '../plugins/stats'
  import { getRealmName, getRealmPeriodName } from '../plugins/realm'
  import { pillRecipes, pillGrades, pillTypes, calculatePillEffect } from '../plugins/pills'
  import { itemQualities } from '../plugins/itemQualities'
  import APIService from '../services/api.js'
  import EquipmentTab from '../components/EquipmentTab.vue'
  import PetsTab from '../components/PetsTab.vue'
  import FormulasTab from '../components/FormulasTab.vue'
  import HerbsTab from '../components/HerbsTab.vue'
  import PillsTab from '../components/PillsTab.vue'
  
  const playerInfoStore = usePlayerInfoStore()
  const inventoryStore = useInventoryStore()
  const equipmentStore = useEquipmentStore()
  const petsStore = usePetsStore()
  const pillsStore = usePillsStore()
  const settingsStore = useSettingsStore()
  const statsStore = useStatsStore()
  const persistenceStore = usePersistenceStore()
  
  const message = useMessage()

  // 获取已装备的装备列表
  const loadEquippedArtifacts = async () => {
    try {
      const token = getAuthToken()
      if (!token) {
        console.error('[Inventory] 未找到认证令牌，无法获取已装备列表')
        return
      }
      
      console.log('[Inventory] 开始获取已装备的装备列表')
      const data = await APIService.getPlayerInventory(token, { equipped: 'true' })
      
      // 更新已装备的装备状态
      data.items.forEach(item => {
        if (item.slot) {
          equipmentStore.equippedArtifacts[item.slot] = item
        }
      })
      
      // 清空没有装备的槽位
      Object.keys(equipmentStore.equippedArtifacts).forEach(slot => {
        const isEquipped = data.items.some(item => item.slot === slot);
        if (!isEquipped) {
          equipmentStore.equippedArtifacts[slot] = null;
        }
      });
      
      console.log(`[Inventory] 成功获取已装备的装备列表，数量: ${data.items.length}`)
    } catch (error) {
      console.error('[Inventory] 获取已装备装备列表时发生错误:', error)
    }
  }
  
  // 页面加载时获取已装备的装备
  onMounted(() => {
    console.log('[Inventory] 页面挂载，开始加载已装备的装备和灵宠列表');
    loadEquippedArtifacts();
    
    // 加载灵宠列表
    loadPetList();
  })
  
  // 加载灵宠列表
  const loadPetList = async () => {
    try {
      console.log('[Inventory] 开始加载灵宠列表');
      const token = getAuthToken();
      if (!token) {
        console.error('[Inventory] 未找到认证令牌，无法获取灵宠列表');
        return;
      }
      
      const response = await APIService.getPlayerData(token);
      console.log('[Inventory] 获取到玩家完整数据:', response);
      
      if (response.pets) {
        console.log('[Inventory] 更新灵宠列表，数量:', response.pets.length);
        petsStore.pets = response.pets;
        
        // 打印灵宠列表状态用于调试
        console.log('[Inventory] 当前灵宠列表状态:');
        petsStore.pets.forEach(pet => {
          console.log(`  ${pet.name}(${pet.id}): isActive=${pet.isActive}`);
        });
      } else {
        console.warn('[Inventory] 响应中未包含灵宠数据');
      }
    } catch (error) {
      console.error('[Inventory] 加载灵宠列表时发生错误:', error);
    }
  }
  
  // 导入必要的模块
  import { getAuthToken } from '../stores/db'
 
</script>

<style scoped>
  .n-card {
    cursor: pointer;
  }
</style>