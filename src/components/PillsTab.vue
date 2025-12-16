<template>
  <div>
    <n-grid :cols="2" :x-gap="12" :y-gap="8" v-if="groupedPills.length">
      <n-grid-item v-for="pill in groupedPills" :key="pill.id">
        <n-card hoverable>
          <template #header>
            <n-space justify="space-between">
              <span>{{ pill.name }}({{ pill.count }})</span>
              <n-button size="small" type="primary" @click="usePill(pill)">服用</n-button>
            </n-space>
          </template>
          <p>{{ pill.description }}</p>
        </n-card>
      </n-grid-item>
    </n-grid>
    <n-empty v-else />
  </div>
</template>

<script setup>
  import { computed } from 'vue'
  import { useMessage } from 'naive-ui'

  const props = defineProps({
    playerInfoStore: {
      type: Object,
      required: true
    }
  })

  const message = useMessage()

  // 丹药分组
  const groupedPills = computed(() => {
    const pills = {}
    props.playerInfoStore.items
      .filter(item => item.type === 'pill')
      .forEach(pill => {
        if (pills[pill.id]) {
          pills[pill.id].count++
        } else {
          pills[pill.id] = { ...pill, count: 1 }
        }
      })
    return Object.values(pills)
  })

  // 使用丹药
  const usePill = pill => {
    // 丹药使用功能已移到后端，前端仅负责打开库存中的丹药 UI
    message.info('丹药使用功能正在开发中')
  }
</script>

<style scoped>
</style>