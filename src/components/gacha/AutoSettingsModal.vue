<template>
  <n-modal v-model:show="show" preset="dialog" title="自动处理设置" style="width: 800px">
    <n-card :bordered="false">
      <n-space vertical>
        <n-divider>装备自动出售</n-divider>
        <n-checkbox-group v-model:value="localAutoSellQualities" @update:value="handleAutoSellChange">
          <n-space wrap>
            <n-checkbox
              value="all"
              :disabled="
                !!localAutoSellQualities?.length && !localAutoSellQualities.includes('all')
              "
            >
              全部品阶
            </n-checkbox>
            <n-checkbox
              v-for="(quality, key) in equipmentQualities"
              :key="key"
              :value="key"
              :disabled="localAutoSellQualities?.includes('all')"
            >
              <span :style="{ color: quality.color }">{{ quality.name }}</span>
            </n-checkbox>
          </n-space>
        </n-checkbox-group>
        <n-divider>灵宠自动放生</n-divider>
        <n-checkbox-group
          v-model:value="localAutoReleaseRarities"
          @update:value="handleAutoReleaseChange"
        >
          <n-space wrap>
            <n-checkbox
              value="all"
              :disabled="
                !!localAutoReleaseRarities?.length && !localAutoReleaseRarities.includes('all')
              "
            >
              全部品质
            </n-checkbox>
            <n-checkbox
              v-for="(rarity, key) in petRarities"
              :key="key"
              :value="key"
              :disabled="localAutoReleaseRarities?.includes('all')"
            >
              <span :style="{ color: rarity.color }">{{ rarity.name }}</span>
            </n-checkbox>
          </n-space>
        </n-checkbox-group>
      </n-space>
    </n-card>
    <template #footer>
      <n-space justify="end">
        <n-button @click="closeModal">关闭</n-button>
      </n-space>
    </template>
  </n-modal>
</template>

<script setup>
  import { ref, computed, defineProps, defineEmits } from 'vue'

  const props = defineProps({
    show: {
      type: Boolean,
      default: false
    },
    autoSellQualities: {
      type: Array,
      default: () => []
    },
    autoReleaseRarities: {
      type: Array,
      default: () => []
    },
    equipmentQualities: {
      type: Object,
      required: true
    },
    petRarities: {
      type: Object,
      required: true
    }
  })

  const emit = defineEmits(['update:show', 'update:auto-sell-qualities', 'update:auto-release-rarities'])

  const show = computed({
    get: () => props.show,
    set: (value) => emit('update:show', value)
  })

  const localAutoSellQualities = computed({
    get: () => props.autoSellQualities,
    set: (value) => emit('update:auto-sell-qualities', value)
  })

  const localAutoReleaseRarities = computed({
    get: () => props.autoReleaseRarities,
    set: (value) => emit('update:auto-release-rarities', value)
  })

  const handleAutoSellChange = (values) => {
    // 如果选择了"全部"，则清除其他选项
    if (values.includes('all')) {
      localAutoSellQualities.value = ['all']
    }
    // 如果之前选择了"全部"而现在选择了其他选项，则移除"全部"
    else if (
      localAutoSellQualities.value.includes('all') &&
      values.length > 0
    ) {
      localAutoSellQualities.value = values.filter(v => v !== 'all')
    }
  }

  const handleAutoReleaseChange = (values) => {
    // 如果选择了"全部"，则清除其他选项
    if (values.includes('all')) {
      localAutoReleaseRarities.value = ['all']
    }
    // 如果之前选择了"全部"而现在选择了其他选项，则移除"全部"
    else if (
      localAutoReleaseRarities.value.includes('all') &&
      values.length > 0
    ) {
      localAutoReleaseRarities.value = values.filter(v => v !== 'all')
    }
  }

  const closeModal = () => {
    show.value = false
  }
</script>