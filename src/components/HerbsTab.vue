<template>
  <div>
    <n-grid :cols="2" :x-gap="12" :y-gap="8" v-if="groupedHerbs.length">
      <n-grid-item v-for="herb in groupedHerbs" :key="herb.id">
        <n-card hoverable>
          <template #header>
            <n-space justify="space-between">
              <span>{{ herb.name }}({{ herb.count }})</span>
            </n-space>
          </template>
          <p>{{ herb.description }}</p>
        </n-card>
      </n-grid-item>
    </n-grid>
    <n-empty v-else />
  </div>
</template>

<script setup>
  import { computed } from 'vue'

  const props = defineProps({
    inventoryStore: {
      type: Object,
      required: true
    }
  })

  // 灵草分组
  const groupedHerbs = computed(() => {
    const herbs = {}
    props.inventoryStore.items
      .filter(item => item.type === 'herb')
      .forEach(herb => {
        if (herbs[herb.id]) {
          herbs[herb.id].count++
        } else {
          herbs[herb.id] = { ...herb, count: 1 }
        }
      })
    return Object.values(herbs)
  })
</script>

<style scoped>
</style>