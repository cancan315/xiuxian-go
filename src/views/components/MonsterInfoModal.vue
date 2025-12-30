<template>
  <!-- 妖兽信息弹窗 -->
  <n-modal v-model:show="show" preset="card" title="妖兽信息" style="width: 90%; max-width: 900px;">
    <n-space vertical>
      <!-- 基础信息 -->
      <n-descriptions bordered :column="2" v-if="monster">
        <n-descriptions-item label="名称">{{ monster.name }}</n-descriptions-item>
        <n-descriptions-item label="难度">
          <n-tag :type="getDifficultyTagType(monster.difficulty)">
            {{ getDifficultyName(monster.difficulty) }}
          </n-tag>
        </n-descriptions-item>
        <n-descriptions-item label="等级">{{ monster.level }}</n-descriptions-item>
        <n-descriptions-item label="描述">{{ monster.description }}</n-descriptions-item>
      </n-descriptions>

      <!-- 基础属性 -->
      <div v-if="monster?.baseAttributes" class="attributes-section">
        <h4>基础属性</h4>
        <n-descriptions bordered :column="4" size="small">
          <n-descriptions-item label="生命">{{ monster.baseAttributes.health }}</n-descriptions-item>
          <n-descriptions-item label="攻击">{{ monster.baseAttributes.attack }}</n-descriptions-item>
          <n-descriptions-item label="防御">{{ monster.baseAttributes.defense }}</n-descriptions-item>
          <n-descriptions-item label="速度">{{ monster.baseAttributes.speed }}</n-descriptions-item>
        </n-descriptions>
      </div>

      <!-- 战斗属性 -->
      <div v-if="monster?.combatAttributes" class="attributes-section">
        <h4>战斗属性</h4>
        <n-descriptions bordered :column="3" size="small">
          <n-descriptions-item label="暴击率">{{ (monster.combatAttributes.critRate * 100).toFixed(2) }}%</n-descriptions-item>
          <n-descriptions-item label="连击率">{{ (monster.combatAttributes.comboRate * 100).toFixed(2) }}%</n-descriptions-item>
          <n-descriptions-item label="反击率">{{ (monster.combatAttributes.counterRate * 100).toFixed(2) }}%</n-descriptions-item>
          <n-descriptions-item label="眩晕率">{{ (monster.combatAttributes.stunRate * 100).toFixed(2) }}%</n-descriptions-item>
          <n-descriptions-item label="闪避率">{{ (monster.combatAttributes.dodgeRate * 100).toFixed(2) }}%</n-descriptions-item>
          <n-descriptions-item label="吸血率">{{ (monster.combatAttributes.vampireRate * 100).toFixed(2) }}%</n-descriptions-item>
        </n-descriptions>
      </div>

      <!-- 奖励信息 -->
      <div v-if="monster?.rewards" class="attributes-section">
        <h4>掉落奖励</h4>
        <n-descriptions bordered :column="2" size="small">
          <n-descriptions-item label="修为">{{ monster.rewards.cultivation }}</n-descriptions-item>
          <n-descriptions-item label="灵石">{{ monster.rewards.spiritStones }}</n-descriptions-item>
          <n-descriptions-item label="物品">{{ monster.rewards.dropItems }}</n-descriptions-item>
        </n-descriptions>
      </div>
    </n-space>
  </n-modal>
</template>

<script setup>
import { computed } from 'vue'
import { NModal, NDescriptions, NDescriptionsItem, NSpace, NTag } from 'naive-ui'
import { getDifficultyTagType, getDifficultyName } from '../utils/duelHelper'

// 定义 props
const props = defineProps({
  // 是否显示模态框
  show: {
    type: Boolean,
    default: false
  },
  // 妖兽信息
  monster: {
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
