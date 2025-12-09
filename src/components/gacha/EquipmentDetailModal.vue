<template>
  <n-modal v-model:show="show" preset="dialog" title="装备详情" style="width: 600px">
    <template v-if="equipment">
      <n-descriptions bordered>
        <n-descriptions-item label="名称">{{ equipment.name }}</n-descriptions-item>
        <n-descriptions-item label="品质">
          <n-tag :style="{ color: equipment.qualityInfo ? equipment.qualityInfo.color : '#000000' }">
            {{ equipment.qualityInfo ? equipment.qualityInfo.name : '未知品质' }}
          </n-tag>
        </n-descriptions-item>
        <n-descriptions-item label="类型">{{ equipmentTypes[equipment.equipType]?.name || '未知类型' }}</n-descriptions-item>
        <n-descriptions-item label="等级">{{ equipment.level || 1 }}</n-descriptions-item>
      </n-descriptions>
      <n-divider>基础属性</n-divider>
      <n-descriptions bordered :column="2">
        <n-descriptions-item label="攻击力">{{ equipment.stats?.attack || 0 }}</n-descriptions-item>
        <n-descriptions-item label="生命值">{{ equipment.stats?.health || 0 }}</n-descriptions-item>
        <n-descriptions-item label="防御力">{{ equipment.stats?.defense || 0 }}</n-descriptions-item>
        <n-descriptions-item label="速度">{{ equipment.stats?.speed || 0 }}</n-descriptions-item>
      </n-descriptions>
      <n-divider v-if="hasExtraAttributes(equipment)">额外属性</n-divider>
      <n-descriptions bordered :column="3" v-if="hasExtraAttributes(equipment)">
        <n-descriptions-item v-for="(value, key) in equipment.extraAttributes" :key="key" :label="getAttributeLabel(key)">
          {{ (value * 100).toFixed(1) }}%
        </n-descriptions-item>
      </n-descriptions>
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
    equipment: {
      type: Object,
      default: null
    },
    equipmentTypes: {
      type: Object,
      required: true
    }
  })

  const emit = defineEmits(['update:show'])

  const show = computed({
    get: () => props.show,
    set: (value) => emit('update:show', value)
  })

  // 判断是否有额外属性
  const hasExtraAttributes = (equipment) => {
    return equipment.extraAttributes && Object.keys(equipment.extraAttributes).length > 0
  }

  // 获取属性标签
  const getAttributeLabel = (attribute) => {
    const labels = {
      critRate: '暴击率',
      comboRate: '连击率',
      counterRate: '反击率',
      stunRate: '眩晕率',
      dodgeRate: '闪避率',
      vampireRate: '吸血率'
    }
    return labels[attribute] || attribute
  }
</script>