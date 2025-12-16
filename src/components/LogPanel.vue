<template>
  <n-divider>{{ title }}</n-divider>
  <n-card class="log-panel" :style="$attrs.style">
    <n-space justify="end" style="margin-bottom: 8px;">
      <n-button size="small" @click="clearLogs" type="error" secondary>清空日志</n-button>
    </n-space>
    <n-scrollbar ref="scrollRef" trigger="none" style="max-height: 200px">
      <div class="log-container" v-if="logList.length">
        <div v-for="(log, index) in logList" :key="index" class="log-item">
          <n-tag :type="log.type" size="small" class="log-type">
            {{ log.time }}
          </n-tag>
          <span class="log-content">
            <n-gradient-text :type="log.type">
              {{ log.content }}
            </n-gradient-text>
          </span>
        </div>
      </div>
      <n-empty v-else description="暂无日志" />
    </n-scrollbar>
  </n-card>
</template>

<script setup>
  import { ref, onMounted, onUnmounted, watch } from 'vue'

  const props = defineProps({
    title: {
      type: String,
      default: '系统日志'
    },
    logs: {
      type: Array,
      default: () => []
    }
  })

  // 日志数组和滚动引用
  const logList = ref([])
  const scrollRef = ref(null)

  // 监听外部传入的 logs，同步到内部
  watch(
    () => props.logs,
    (newLogs) => {
      logList.value = newLogs
    },
    { immediate: true }
  )

  // 添加清空日志方法
  const clearLogs = () => {
    logList.value = []
  }

  // 优化滚动逻辑：只在 logs 数量变化时滚动
  watch(
    () => logList.value.length,
    () => {
      setTimeout(() => {
        if (scrollRef.value) {
          scrollRef.value.scrollTo({ top: 99999, behavior: 'smooth' })
        }
      })
    }
  )

  // 暒露方法
  defineExpose({
    clearLogs
  })
</script>

<style scoped>
  .log-panel {
    margin-top: 12px;
  }

  .log-container {
    padding: 8px;
  }

  .log-item {
    margin-bottom: 8px;
    display: flex;
    align-items: flex-start;
  }

  .log-type {
    margin-right: 8px;
    flex-shrink: 0;
  }

  .log-content {
    flex-grow: 1;
    word-break: break-all;
  }
</style>
