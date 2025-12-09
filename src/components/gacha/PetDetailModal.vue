<template>
  <n-modal v-model:show="show" preset="dialog" title="灵宠详情" style="width: 600px">
    <template v-if="pet">
      <n-descriptions bordered>
        <n-descriptions-item label="名称">{{ pet.name }}</n-descriptions-item>
        <n-descriptions-item label="品质">
          <n-tag :style="{ color: pet.rarity && petRarities[pet.rarity] ? petRarities[pet.rarity].color : '#000000' }">
            {{ pet.rarity && petRarities[pet.rarity] ? petRarities[pet.rarity].name : '未知品质' }}
          </n-tag>
        </n-descriptions-item>
        <n-descriptions-item label="等级">{{ pet.level || 1 }}</n-descriptions-item>
        <n-descriptions-item label="星级">{{ pet.star || 0 }}</n-descriptions-item>
        <n-descriptions-item label="境界">{{ Math.floor((pet.star || 0) / 5) }}阶</n-descriptions-item>
      </n-descriptions>
      <n-divider>属性加成</n-divider>
      <n-descriptions bordered>
        <n-descriptions-item label="攻击加成">
          +{{ (getPetBonus(pet).attack * 100).toFixed(1) }}%
        </n-descriptions-item>
        <n-descriptions-item label="防御加成">
          +{{ (getPetBonus(pet).defense * 100).toFixed(1) }}%
        </n-descriptions-item>
        <n-descriptions-item label="生命加成">
          +{{ (getPetBonus(pet).health * 100).toFixed(1) }}%
        </n-descriptions-item>
      </n-descriptions>
      <n-divider>灵宠属性</n-divider>
      <n-collapse>
        <n-collapse-item title="展开" name="1">
          <n-divider>基础属性</n-divider>
          <n-descriptions bordered :column="2">
            <n-descriptions-item label="攻击力">{{ pet.combatAttributes?.attack || 0 }}</n-descriptions-item>
            <n-descriptions-item label="生命值">{{ pet.combatAttributes?.health || 0 }}</n-descriptions-item>
            <n-descriptions-item label="防御力">{{ pet.combatAttributes?.defense || 0 }}</n-descriptions-item>
            <n-descriptions-item label="速度">{{ pet.combatAttributes?.speed || 0 }}</n-descriptions-item>
          </n-descriptions>
          <n-divider>战斗属性</n-divider>
          <n-descriptions bordered :column="3">
            <n-descriptions-item label="暴击率">
              {{ ((pet.combatAttributes?.critRate || 0) * 100).toFixed(1) }}%
            </n-descriptions-item>
            <n-descriptions-item label="连击率">
              {{ ((pet.combatAttributes?.comboRate || 0) * 100).toFixed(1) }}%
            </n-descriptions-item>
            <n-descriptions-item label="反击率">
              {{ ((pet.combatAttributes?.counterRate || 0) * 100).toFixed(1) }}%
            </n-descriptions-item>
            <n-descriptions-item label="眩晕率">
              {{ ((pet.combatAttributes?.stunRate || 0) * 100).toFixed(1) }}%
            </n-descriptions-item>
            <n-descriptions-item label="闪避率">
              {{ ((pet.combatAttributes?.dodgeRate || 0) * 100).toFixed(1) }}%
            </n-descriptions-item>
            <n-descriptions-item label="吸血率">
              {{ ((pet.combatAttributes?.vampireRate || 0) * 100).toFixed(1) }}%
            </n-descriptions-item>
          </n-descriptions>
          <n-divider>战斗抗性</n-divider>
          <n-descriptions bordered :column="3">
            <n-descriptions-item label="抗暴击">
              {{ ((pet.combatAttributes?.critResist || 0) * 100).toFixed(1) }}%
            </n-descriptions-item>
            <n-descriptions-item label="抗连击">
              {{ ((pet.combatAttributes?.comboResist || 0) * 100).toFixed(1) }}%
            </n-descriptions-item>
            <n-descriptions-item label="抗反击">
              {{ ((pet.combatAttributes?.counterResist || 0) * 100).toFixed(1) }}%
            </n-descriptions-item>
            <n-descriptions-item label="抗眩晕">
              {{ ((pet.combatAttributes?.stunResist || 0) * 100).toFixed(1) }}%
            </n-descriptions-item>
            <n-descriptions-item label="抗闪避">
              {{ ((pet.combatAttributes?.dodgeResist || 0) * 100).toFixed(1) }}%
            </n-descriptions-item>
            <n-descriptions-item label="抗吸血">
              {{ ((pet.combatAttributes?.vampireResist || 0) * 100).toFixed(1) }}%
            </n-descriptions-item>
          </n-descriptions>
          <n-divider>特殊属性</n-divider>
          <n-descriptions bordered :column="3">
            <n-descriptions-item label="强化治疗">
              {{ ((pet.combatAttributes?.healBoost || 0) * 100).toFixed(1) }}%
            </n-descriptions-item>
            <n-descriptions-item label="强化爆伤">
              {{ ((pet.combatAttributes?.critDamageBoost || 0) * 100).toFixed(1) }}%
            </n-descriptions-item>
            <n-descriptions-item label="弱化爆伤">
              {{ ((pet.combatAttributes?.critDamageReduce || 0) * 100).toFixed(1) }}%
            </n-descriptions-item>
            <n-descriptions-item label="最终增伤">
              {{ ((pet.combatAttributes?.finalDamageBoost || 0) * 100).toFixed(1) }}%
            </n-descriptions-item>
            <n-descriptions-item label="最终减伤">
              {{ ((pet.combatAttributes?.finalDamageReduce || 0) * 100).toFixed(1) }}%
            </n-descriptions-item>
            <n-descriptions-item label="战斗属性提升">
              {{ ((pet.combatAttributes?.combatBoost || 0) * 100).toFixed(1) }}%
            </n-descriptions-item>
            <n-descriptions-item label="战斗抗性提升">
              {{ ((pet.combatAttributes?.resistanceBoost || 0) * 100).toFixed(1) }}%
            </n-descriptions-item>
          </n-descriptions>
        </n-collapse-item>
      </n-collapse>
    </template>
  </n-modal>
</template>

<script setup>
  import { computed, defineProps, defineEmits } from 'vue'

  const props = defineProps({
    show: {
      type: Boolean,
      default: false
    },
    pet: {
      type: Object,
      default: null
    },
    petRarities: {
      type: Object,
      required: true
    }
  })

  const emit = defineEmits(['update:show'])

  const show = computed({
    get: () => props.show,
    set: (value) => emit('update:show', value)
  })

  // 获取灵宠加成
  const getPetBonus = (pet) => {
    const rarity = props.petRarities[pet.rarity]
    if (!rarity) return { attack: 0, defense: 0, health: 0 }

    // 基础属性加成
    const baseAttack = 0.05 * rarity.statMod
    const baseDefense = 0.03 * rarity.statMod
    const baseHealth = 0.02 * rarity.statMod

    // 星级加成
    const starBonus = (pet.star || 0) * 0.01

    return {
      attack: baseAttack + starBonus,
      defense: baseDefense + starBonus,
      health: baseHealth + starBonus
    }
  }
</script>