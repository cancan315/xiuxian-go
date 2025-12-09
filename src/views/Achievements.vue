<script setup>
  // 修改为使用模块化store
  import { usePlayerInfoStore } from '../stores/playerInfo'
  import { useInventoryStore } from '../stores/inventory'
  import { useEquipmentStore } from '../stores/equipment'
  import { usePetsStore } from '../stores/pets'
  import { usePillsStore } from '../stores/pills'
  import { useSettingsStore } from '../stores/settings'
  import { useStatsStore } from '../stores/stats'
  import { ref, computed, onMounted } from 'vue'
  import { useMessage } from 'naive-ui'

  const playerInfoStore = usePlayerInfoStore()
  const inventoryStore = useInventoryStore()
  const equipmentStore = useEquipmentStore()
  const petsStore = usePetsStore()
  const pillsStore = usePillsStore()
  const settingsStore = useSettingsStore()
  const statsStore = useStatsStore()
  
  const message = useMessage()
  const activeTab = ref('achievements')

  // 成就列表
  const achievements = computed(() => [
    {
      id: 'first_breakthrough',
      name: '初窥门径',
      description: '完成第一次境界突破',
      condition: '突破次数 >= 1',
      achieved: statsStore.breakthroughCount >= 1,
      reward: '100灵石'
    },
    {
      id: 'explorer',
      name: '探险家',
      description: '完成10次探索',
      condition: '探索次数 >= 10',
      achieved: statsStore.explorationCount >= 10,
      reward: '50灵石'
    },
    {
      id: 'collector',
      name: '收藏家',
      description: '获得50件物品',
      condition: '获得物品数 >= 50',
      achieved: statsStore.itemsFound >= 50,
      reward: '100灵石'
    },
    {
      id: 'alchemist',
      name: '炼丹师',
      description: '炼制10颗丹药',
      condition: '炼制丹药数 >= 10',
      achieved: pillsStore.pillsCrafted >= 10,
      reward: '200灵石'
    },
    {
      id: 'equip_master',
      name: '装备大师',
      description: '装备一件仙品装备',
      condition: '拥有仙品装备',
      achieved: Object.values(equipmentStore.equippedArtifacts).some(a => a && a.quality === 'mythic'),
      reward: '500灵石'
    },
    {
      id: 'pet_tamer',
      name: '灵宠大师',
      description: '拥有一只神品灵宠',
      condition: '拥有神品灵宠',
      achieved: petsStore.pets.some(p => p.rarity === 'divine') || (playerInfoStore.activePet && playerInfoStore.activePet.rarity === 'divine'),
      reward: '1000灵石'
    },
    {
      id: 'realm_master',
      name: '境界大师',
      description: '达到练气十层',
      condition: '境界等级 >= 10',
      achieved: playerInfoStore.level >= 10,
      reward: '1000灵石'
    },
    {
      id: 'wealthy',
      name: '富豪',
      description: '拥有10000灵石',
      condition: '灵石数量 >= 10000',
      achieved: inventoryStore.spiritStones >= 10000,
      reward: '500灵石'
    }
  ])

  // 已完成成就
  const completedAchievements = computed(() => {
    return achievements.value.filter(a => a.achieved)
  })

  // 未完成成就
  const incompleteAchievements = computed(() => {
    return achievements.value.filter(a => !a.achieved)
  })

  // 统计数据
  const stats = computed(() => [
    { name: '境界等级', value: playerInfoStore.level },
    { name: '当前境界', value: playerInfoStore.realm },
    { name: '总修炼时间', value: `${Math.floor(statsStore.totalCultivationTime / 60)}小时${statsStore.totalCultivationTime % 60}分钟` },
    { name: '突破次数', value: statsStore.breakthroughCount },
    { name: '探索次数', value: statsStore.explorationCount },
    { name: '获得物品数', value: statsStore.itemsFound },
    { name: '触发事件数', value: statsStore.eventTriggered },
    { name: '炼制丹药数', value: pillsStore.pillsCrafted },
    { name: '服用丹药数', value: pillsStore.pillsConsumed },
    { name: '灵石数量', value: inventoryStore.spiritStones },
    { name: '强化石数量', value: inventoryStore.reinforceStones },
    { name: '洗练石数量', value: inventoryStore.refinementStones },
    { name: '灵宠精华', value: playerInfoStore.petEssence },
    { name: '背包物品数', value: inventoryStore.items.length },
    { name: '灵宠数量', value: petsStore.pets.length },
    { name: '灵草种类', value: inventoryStore.herbs.length },
    { name: '掌握丹方数', value: pillsStore.pillRecipes.length }
  ])

  // 解锁境界
  const unlockedRealms = computed(() => {
    return playerInfoStore.unlockedRealms.map(realm => ({
      name: realm
    }))
  })

  // 解锁地点
  const unlockedLocations = computed(() => {
    return playerInfoStore.unlockedLocations.map(location => ({
      name: location
    }))
  })
</script>