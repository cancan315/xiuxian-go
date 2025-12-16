<template>
  <div>
    <n-tabs type="segment">
      <n-tab-pane name="complete" tab="完整丹方">
        <n-grid :cols="2" :x-gap="12" :y-gap="8" v-if="groupedFormulas.complete.length">
          <n-grid-item v-for="formula in groupedFormulas.complete" :key="formula.id">
            <n-card hoverable>
              <template #header>
                <n-space justify="space-between">
                  <span>{{ formula.name }}</span>
                  <n-space>
                    <n-tag type="success" size="small">完整</n-tag>
                    <n-tag type="info" size="small">{{ pillGrades[formula.grade].name }}</n-tag>
                    <n-tag type="warning" size="small">{{ pillTypes[formula.type].name }}</n-tag>
                  </n-space>
                </n-space>
              </template>
              <p>{{ formula.description }}</p>
            </n-card>
          </n-grid-item>
        </n-grid>
        <n-empty v-else />
      </n-tab-pane>
      <n-tab-pane name="incomplete" tab="残缺丹方">
        <n-grid :cols="2" :x-gap="12" :y-gap="8" v-if="groupedFormulas.incomplete.length">
          <n-grid-item v-for="formula in groupedFormulas.incomplete" :key="formula.id">
            <n-card hoverable>
              <template #header>
                <n-space justify="space-between">
                  <span>{{ formula.name }}</span>
                  <n-space>
                    <n-tag type="warning" size="small">残缺</n-tag>
                    <n-tag type="info" size="small">{{ pillGrades[formula.grade].name }}</n-tag>
                    <n-tag type="warning" size="small">{{ pillTypes[formula.type].name }}</n-tag>
                  </n-space>
                </n-space>
              </template>
              <p>{{ formula.description }}</p>
              <n-progress
                type="line"
                :percentage="Number(((formula.fragments / formula.fragmentsNeeded) * 100).toFixed(2))"
                :show-indicator="true"
                indicator-placement="inside"
              >
                收集进度: {{ formula.fragments }}/{{ formula.fragmentsNeeded }}
              </n-progress>
            </n-card>
          </n-grid-item>
        </n-grid>
        <n-empty v-else />
      </n-tab-pane>
    </n-tabs>
  </div>
</template>

<script setup>
  import { computed } from 'vue'
  import { pillGrades, pillTypes } from '../plugins/pills'

  const props = defineProps({
    playerInfoStore: {
      type: Object,
      required: true
    }
  })

  // 丹方分组
  const groupedFormulas = computed(() => {
    const complete = []
    const incomplete = []
    
    props.playerInfoStore.items
      .filter(item => item.type === 'formula')
      .forEach(formula => {
        if (formula.fragments >= formula.fragmentsNeeded) {
          complete.push(formula)
        } else {
          incomplete.push(formula)
        }
      })
      
    return { complete, incomplete }
  })
</script>

<style scoped>
</style>