<template>
  <!-- 玩家信息弹窗 -->
  <n-modal v-model:show="show" preset="card" title="玩家信息" style="width: 90%; max-width: 900px;">
    <n-space vertical>
      <!-- 基础信息 -->
      <n-descriptions bordered :column="2" v-if="player">
        <n-descriptions-item label="道号">{{ player.name }}</n-descriptions-item>
        <n-descriptions-item label="境界">{{ getRealmName(player.level).name }}</n-descriptions-item>
        <n-descriptions-item label="修为">{{ player.cultivation }}/{{ player.maxCultivation }}</n-descriptions-item>
        <n-descriptions-item label="灵石">{{ player.spiritStones }}</n-descriptions-item>
      </n-descriptions>

      <!-- 基础属性 -->
      <div v-if="player?.baseAttributes" class="attributes-section">
        <h4>基础属性</h4>
        <n-descriptions bordered :column="4" size="small">
          <n-descriptions-item label="生命">{{ player.baseAttributes.health }}</n-descriptions-item>
          <n-descriptions-item label="攻击">{{ player.baseAttributes.attack }}</n-descriptions-item>
          <n-descriptions-item label="防御">{{ player.baseAttributes.defense }}</n-descriptions-item>
          <n-descriptions-item label="速度">{{ player.baseAttributes.speed }}</n-descriptions-item>
        </n-descriptions>
      </div>

      <!-- 战斗属性 -->
      <div v-if="player?.combatAttributes" class="attributes-section">
        <h4>战斗属性</h4>
        <n-descriptions bordered :column="3" size="small">
          <n-descriptions-item label="暴击率">{{ (player.combatAttributes.critRate * 100).toFixed(2) }}%</n-descriptions-item>
          <n-descriptions-item label="连击率">{{ (player.combatAttributes.comboRate * 100).toFixed(2) }}%</n-descriptions-item>
          <n-descriptions-item label="反击率">{{ (player.combatAttributes.counterRate * 100).toFixed(2) }}%</n-descriptions-item>
          <n-descriptions-item label="眩晕率">{{ (player.combatAttributes.stunRate * 100).toFixed(2) }}%</n-descriptions-item>
          <n-descriptions-item label="闪避率">{{ (player.combatAttributes.dodgeRate * 100).toFixed(2) }}%</n-descriptions-item>
          <n-descriptions-item label="吸血率">{{ (player.combatAttributes.vampireRate * 100).toFixed(2) }}%</n-descriptions-item>
        </n-descriptions>
      </div>

      <!-- 抗性属性 -->
      <div v-if="player?.combatResistance" class="attributes-section">
        <h4>抗性属性</h4>
        <n-descriptions bordered :column="3" size="small">
          <n-descriptions-item label="暴击抗性">{{ (player.combatResistance.critResist * 100).toFixed(2) }}%</n-descriptions-item>
          <n-descriptions-item label="连击抗性">{{ (player.combatResistance.comboResist * 100).toFixed(2) }}%</n-descriptions-item>
          <n-descriptions-item label="反击抗性">{{ (player.combatResistance.counterResist * 100).toFixed(2) }}%</n-descriptions-item>
          <n-descriptions-item label="眩晕抗性">{{ (player.combatResistance.stunResist * 100).toFixed(2) }}%</n-descriptions-item>
          <n-descriptions-item label="闪避抗性">{{ (player.combatResistance.dodgeResist * 100).toFixed(2) }}%</n-descriptions-item>
          <n-descriptions-item label="吸血抗性">{{ (player.combatResistance.vampireResist * 100).toFixed(2) }}%</n-descriptions-item>
        </n-descriptions>
      </div>
    </n-space>
  </n-modal>
</template>

<script setup>
import { computed } from 'vue'
import { NModal, NDescriptions, NDescriptionsItem, NSpace } from 'naive-ui'
import { getRealmName } from '../../plugins/realm'

// 定义 props
const props = defineProps({
  // 是否显示模态框
  show: {
    type: Boolean,
    default: false
  },
  // 玩家信息
  player: {
    type: Object,
    default: null
  }
})

// 定义 emit 事件
const emit = defineEmits(['update:show'])

// 处理 show 属性变化
const show = computed({
  get: () => props.show,
  set: (value) => emit('update:show', value)
})
</script>

<style scoped>
.attributes-section {
  margin-top: 16px;
}

.attributes-section h4 {
  margin: 0 0 12px 0;
  font-size: 14px;
  font-weight: 600;
  color: #333;
}
</style>
